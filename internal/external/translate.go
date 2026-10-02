package external

import (
	"container/list"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Translator ports utils/Translate.ts.
//
// The source app called Google's unofficial endpoint from each phone, so every
// user had their own quota. From one server IP that endpoint rate-limits
// bursts (HTTP 429, seen 2026-10-02), so this translator caches results,
// limits concurrency and walks a list of endpoints: the source app's one
// first, then backups (other Google hosts/clients and MyMemory). An endpoint
// that fails cools down for a while and is skipped meanwhile.
type Translator struct {
	Client    *Client
	Endpoints []*Endpoint // tried in order; see DefaultTranslateEndpoints

	// Extra passes over the endpoints after the first (default 1), with a
	// backoff between passes (default 1s).
	Retries int
	Backoff time.Duration

	once  sync.Once
	slots chan struct{}
	cache *lru
}

// Endpoint kinds.
const (
	KindGTX      = "gtx"            // /translate_a/single?client=gtx, answers sentences
	KindDict     = "dict-chrome-ex" // /translate_a/t?client=dict-chrome-ex, answers one string
	KindMyMemory = "mymemory"       // api.mymemory.translated.net, 500 chars per request
)

// Endpoint is one translation backend.
type Endpoint struct {
	Kind string
	Base string // scheme + host, e.g. https://translate.googleapis.com

	coolTill atomic.Int64 // unix nanos
}

// DefaultTranslateEndpoints is the primary (the source app's endpoint) plus
// ten backups, all verified to answer on 2026-10-02. Google applies its rate
// limit per client kind across hosts, so these are three independent pools:
// gtx (5 hosts), dict-chrome-ex (5 hosts) and MyMemory.
func DefaultTranslateEndpoints() []*Endpoint {
	return []*Endpoint{
		{Kind: KindGTX, Base: "https://translate.googleapis.com"}, // primary
		{Kind: KindDict, Base: "https://translate.googleapis.com"},
		{Kind: KindGTX, Base: "https://translate.google.com"},
		{Kind: KindDict, Base: "https://translate.google.com"},
		{Kind: KindGTX, Base: "https://clients5.google.com"},
		{Kind: KindDict, Base: "https://clients5.google.com"},
		{Kind: KindGTX, Base: "https://translate.google.co.id"},
		{Kind: KindDict, Base: "https://translate.google.co.id"},
		{Kind: KindGTX, Base: "https://translate.google.co.uk"},
		{Kind: KindDict, Base: "https://translate.google.co.uk"},
		{Kind: KindMyMemory, Base: "https://api.mymemory.translated.net"}, // outside Google: separate quota
	}
}

// NewTranslator returns a translator over the default endpoints.
func NewTranslator(c *Client) *Translator {
	return &Translator{Client: c, Endpoints: DefaultTranslateEndpoints()}
}

const (
	rateLimitCooldown = time.Minute
	errorCooldown     = 20 * time.Second
	translateParallel = 4
	translateCacheMax = 5000
	translateCacheTTL = 24 * time.Hour
	myMemoryMaxChunk  = 450 // the API rejects queries over 500 characters
)

func (t *Translator) init() {
	t.once.Do(func() {
		t.slots = make(chan struct{}, translateParallel)
		t.cache = newLRU(translateCacheMax, translateCacheTTL)
		if t.Backoff == 0 {
			t.Backoff = time.Second
		}
		if t.Retries == 0 {
			t.Retries = 1
		}
		if t.Retries < 0 { // "no retries": still one pass over the endpoints
			t.Retries = 0
		}
		if len(t.Endpoints) == 0 {
			t.Endpoints = DefaultTranslateEndpoints()
		}
	})
}

// Translate returns the translation as lines: one per sentence from the gtx
// endpoint (like the source app's data.sentences[].trans), one per paragraph
// from the others. Like the source, "&" is replaced with "dan"/"and" first.
func (t *Translator) Translate(ctx context.Context, text, from, to string) ([]string, error) {
	t.init()
	key := cacheKey(text, from, to)
	if lines, ok := t.cache.get(key); ok {
		return lines, nil
	}

	select {
	case t.slots <- struct{}{}:
		defer func() { <-t.slots }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	text = strings.ReplaceAll(text, "&", ampWord(from))
	wait := t.Backoff
	err := errors.New("translate: no endpoint available")
	for pass := 0; pass <= t.Retries; pass++ {
		if pass > 0 {
			select {
			case <-time.After(wait):
				wait *= 2
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		for _, ep := range t.Endpoints {
			if time.Now().UnixNano() < ep.coolTill.Load() {
				continue
			}
			lines, e := t.call(ctx, ep, text, from, to)
			if e == nil {
				t.cache.put(key, lines)
				return lines, nil
			}
			err = e
			if errors.Is(e, context.Canceled) || errors.Is(e, context.DeadlineExceeded) {
				return nil, e
			}
			if rateLimited(e) {
				// Google rate-limits per client across all its hosts (seen
				// 2026-10-02: every gtx host answered 429 at once), so the whole
				// kind cools down and the next kind is tried right away.
				until := time.Now().Add(rateLimitCooldown).UnixNano()
				for _, other := range t.Endpoints {
					if other.Kind == ep.Kind {
						other.coolTill.Store(until)
					}
				}
			} else {
				ep.coolTill.Store(time.Now().Add(errorCooldown).UnixNano())
			}
		}
	}
	return nil, err
}

func (t *Translator) call(ctx context.Context, ep *Endpoint, text, from, to string) ([]string, error) {
	switch ep.Kind {
	case KindGTX:
		return t.gtx(ctx, ep.Base, text, from, to)
	case KindDict:
		return t.dict(ctx, ep.Base, text, from, to)
	case KindMyMemory:
		return t.myMemory(ctx, ep.Base, text, from, to)
	}
	return nil, fmt.Errorf("translate: unknown endpoint kind %q", ep.Kind)
}

// gtx is the source app's endpoint: {"sentences":[{"trans":"…"}]}.
func (t *Translator) gtx(ctx context.Context, base, text, from, to string) ([]string, error) {
	q := url.Values{"client": {"gtx"}, "sl": {from}, "tl": {to}, "dj": {"1"}, "dt": {"t"}, "ie": {"UTF-8"}, "q": {text}}
	body, err := t.Client.Get(ctx, base+"/translate_a/single?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("translate gtx %s: %w", base, err)
	}
	var data struct {
		Sentences []struct {
			Trans *string `json:"trans"`
		} `json:"sentences"`
	}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return nil, fmt.Errorf("translate gtx %s: decode: %w", base, err)
	}
	out := make([]string, 0, len(data.Sentences))
	for _, s := range data.Sentences {
		if s.Trans != nil {
			out = append(out, *s.Trans)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("translate gtx %s: empty response", base)
	}
	return out, nil
}

// dict answers ["text"], or [["text","lang"]] when the source language is
// detected. Line breaks are kept.
func (t *Translator) dict(ctx context.Context, base, text, from, to string) ([]string, error) {
	q := url.Values{"client": {"dict-chrome-ex"}, "sl": {from}, "tl": {to}, "q": {text}}
	body, err := t.Client.Get(ctx, base+"/translate_a/t?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("translate dict %s: %w", base, err)
	}
	var data []json.RawMessage
	if err := json.Unmarshal([]byte(body), &data); err != nil || len(data) == 0 {
		return nil, fmt.Errorf("translate dict %s: unexpected response", base)
	}
	var out string
	if json.Unmarshal(data[0], &out) != nil {
		var pair []string
		if json.Unmarshal(data[0], &pair) != nil || len(pair) == 0 {
			return nil, fmt.Errorf("translate dict %s: unexpected response", base)
		}
		out = pair[0]
	}
	return splitLines(out), nil
}

// myMemory translates chunk by chunk (≤450 characters, cut at line or
// sentence boundaries) and joins the chunks back.
func (t *Translator) myMemory(ctx context.Context, base, text, from, to string) ([]string, error) {
	src := from
	if src == "auto" {
		src = "autodetect"
	}
	var parts []string
	for _, chunk := range chunkText(text, myMemoryMaxChunk) {
		q := url.Values{"langpair": {src + "|" + to}, "q": {chunk}}
		body, err := t.Client.Get(ctx, base+"/get?"+q.Encode())
		if err != nil {
			return nil, fmt.Errorf("translate mymemory: %w", err)
		}
		var data struct {
			ResponseData struct {
				TranslatedText string `json:"translatedText"`
			} `json:"responseData"`
			ResponseStatus json.Number `json:"responseStatus"`
			QuotaFinished  bool        `json:"quotaFinished"`
		}
		if err := json.Unmarshal([]byte(body), &data); err != nil {
			return nil, fmt.Errorf("translate mymemory: decode: %w", err)
		}
		if data.QuotaFinished {
			return nil, &UpstreamError{URL: base, Status: 429}
		}
		if data.ResponseStatus.String() != "200" {
			return nil, fmt.Errorf("translate mymemory: status %s", data.ResponseStatus)
		}
		parts = append(parts, data.ResponseData.TranslatedText)
	}
	return splitLines(strings.Join(parts, nl)), nil
}

const nl = "\x0a"

func splitLines(s string) []string { return strings.Split(s, nl) }

// chunkText splits text into pieces of at most max bytes, preferring line
// breaks, then sentence ends, then spaces. Joining the pieces with a line
// break restores the line structure (sentence splits become line breaks).
func chunkText(text string, max int) []string {
	var chunks []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			chunks = append(chunks, cur.String())
			cur.Reset()
		}
	}
	add := func(piece string) {
		if cur.Len() > 0 && cur.Len()+1+len(piece) > max {
			flush()
		}
		if cur.Len() > 0 {
			cur.WriteString(nl)
		}
		cur.WriteString(piece)
	}
	for _, line := range strings.Split(text, nl) {
		if len(line) <= max {
			add(line)
			continue
		}
		for _, piece := range splitLong(line, max) {
			add(piece)
		}
	}
	flush()
	return chunks
}

// splitLong cuts one long line at sentence ends, or at spaces if a sentence is too long.
func splitLong(line string, max int) []string {
	var out []string
	var cur string
	for _, sentence := range strings.SplitAfter(line, ". ") {
		if len(cur)+len(sentence) <= max {
			cur += sentence
			continue
		}
		if cur != "" {
			out = append(out, strings.TrimSpace(cur))
		}
		cur = sentence
		for len(cur) > max {
			cut := strings.LastIndex(cur[:max], " ")
			if cut <= 0 {
				cut = max
			}
			out = append(out, strings.TrimSpace(cur[:cut]))
			cur = cur[cut:]
		}
	}
	if strings.TrimSpace(cur) != "" {
		out = append(out, strings.TrimSpace(cur))
	}
	return out
}

func ampWord(from string) string {
	if from == "id" {
		return "dan"
	}
	return "and"
}

// rateLimited reports an HTTP 429 (or an exhausted quota).
func rateLimited(err error) bool {
	var up *UpstreamError
	return errors.As(err, &up) && up.Status == 429
}

func cacheKey(text, from, to string) string {
	sum := sha256.Sum256([]byte(from + "\x00" + to + "\x00" + text))
	return string(sum[:])
}

// lru is a small size- and age-bounded cache.
type lru struct {
	mu    sync.Mutex
	max   int
	ttl   time.Duration
	order *list.List // front = most recent
	items map[string]*list.Element
}

type lruEntry struct {
	key     string
	lines   []string
	expires time.Time
}

func newLRU(max int, ttl time.Duration) *lru {
	return &lru{max: max, ttl: ttl, order: list.New(), items: map[string]*list.Element{}}
}

func (c *lru) get(key string) ([]string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.items[key]
	if !ok {
		return nil, false
	}
	e := el.Value.(*lruEntry)
	if time.Now().After(e.expires) {
		c.order.Remove(el)
		delete(c.items, key)
		return nil, false
	}
	c.order.MoveToFront(el)
	return append([]string(nil), e.lines...), true
}

func (c *lru) put(key string, lines []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		el.Value = &lruEntry{key, lines, time.Now().Add(c.ttl)}
		c.order.MoveToFront(el)
		return
	}
	c.items[key] = c.order.PushFront(&lruEntry{key, lines, time.Now().Add(c.ttl)})
	for c.order.Len() > c.max {
		last := c.order.Back()
		c.order.Remove(last)
		delete(c.items, last.Value.(*lruEntry).key)
	}
}
