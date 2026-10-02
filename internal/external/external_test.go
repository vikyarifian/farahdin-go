package external

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const page = `<!DOCTYPE html><html><body>
<div id="body" class="width">
<h1>ARTI NAMA</h1>
<script>(adsbygoogle = window.adsbygoogle || []).push({});</script>
Nama Viky, memiliki arti: Pemimpin.<br>Kalimat dua.<br>
<img border="0" src="ramalan_kecocokan_cinta4.png"><br><br>Saran &amp; nasihat<a href="x">&lt; Hitung Kembali</a>
<font><i>Tidak ditemukan</i></font>
<span><img src="rejeki/15.png"></span>
</div>
<div class="main-horoscope"><p>Oct 1, 2026 - Today is good.</p><p>More Horoscopes for Leo</p></div>
<div class="grid"><span>A</span></div><section class="grid x">B</section>
</body></html>`

func TestFindAndText(t *testing.T) {
	doc, err := ParseHTML(page)
	if err != nil {
		t.Fatal(err)
	}
	body := doc.Find("#body").Text()
	if !strings.Contains(body, "(adsbygoogle = window.adsbygoogle || []).push({});") {
		t.Error("script text must be included, like cheerio .text()")
	}
	if !strings.Contains(body, "Saran & nasihat") {
		t.Error("entities must be decoded in .text()")
	}
	if got := doc.Find("#body > font > i").Text(); got != "Tidak ditemukan" {
		t.Errorf("child selector: %q", got)
	}
	if got := doc.Find(".main-horoscope > p").Text(); got != "Oct 1, 2026 - Today is good.More Horoscopes for Leo" {
		t.Errorf("multiple matches concatenate: %q", got)
	}
	if got := doc.Find(".grid").Text(); got != "AB" {
		t.Errorf("class selector: %q", got)
	}
	if src, ok := doc.Find("#body > span > img").Attr("src"); !ok || src != "rejeki/15.png" {
		t.Errorf("attr: %q %v", src, ok)
	}
	if _, ok := doc.Find("#body > img.nope").Attr("src"); ok {
		t.Error("attr on empty selection must report false")
	}
}

func TestInnerHTMLMatchesParse5(t *testing.T) {
	doc, _ := ParseHTML(page)
	h, ok := doc.Find("#body").HTML()
	if !ok {
		t.Fatal("no #body")
	}
	for _, want := range []string{
		`<br>Kalimat dua.<br>`,                                // void elements without a slash
		`<img border="0" src="ramalan_kecocokan_cinta4.png">`, // attribute order kept
		`png"><br><br>Saran &amp; nasihat`,                    // text re-escaped
		`&lt; Hitung Kembali`,                                 // < escaped in text
		`push({});</script>`,                                  // script text raw
	} {
		if !strings.Contains(h, want) {
			t.Errorf("inner HTML missing %q in\n%s", want, h)
		}
	}
}

func TestStrPipeline(t *testing.T) {
	got, err := Str("a-b-c").Split("-", 1).Value()
	if err != nil || got != "b" {
		t.Errorf("Split: %q %v", got, err)
	}
	if _, err := Str("abc").Split("-", 1).Replace("a", "b").Value(); err != ErrUnexpectedContent {
		t.Errorf("missing index must fail the chain, got %v", err)
	}
	if got := Str("x.x.x").Replace(".", "!").String(); got != "x!x.x" {
		t.Errorf("Replace replaces the first match only: %q", got)
	}
	if got := Str("\n\n  a\n \n\nb").ReplaceReAll(ReBlankLines, "").String(); got != "  a\nb" {
		t.Errorf("ReBlankLines: %q", got)
	}
	if got := Str("x < Hitung Kembali\nmore").ReplaceRe(ReHitungKembali, "").String(); got != "x " {
		t.Errorf("ReHitungKembali: %q", got)
	}
	if got := Str("  a ").TrimStart().TrimEnd().String(); got != "a" {
		t.Errorf("trim: %q", got)
	}
}

func TestClientRejectsErrorsAndCapsBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/fail" {
			http.Error(w, "nope", http.StatusBadGateway)
			return
		}
		if r.Method == http.MethodPost && r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("content type %q", r.Header.Get("Content-Type"))
		}
		w.Write([]byte(strings.Repeat("x", 100)))
	}))
	defer srv.Close()
	c := NewClient(5 * time.Second)
	c.MaxBody = 10
	if _, err := c.Get(context.Background(), srv.URL+"/fail"); err == nil {
		t.Error("expected error for 502")
	}
	body, err := c.PostForm(context.Background(), srv.URL+"/ok", map[string][]string{"a": {"b"}})
	if err != nil || len(body) != 10 {
		t.Errorf("body len %d err %v", len(body), err)
	}
}

// translateServer answers for every endpoint kind by path; status overrides
// the HTTP status per kind (0 = OK).
func translateServer(t *testing.T, status map[string]int, calls map[string]int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		kind := map[string]string{"/translate_a/single": KindGTX, "/translate_a/t": KindDict, "/get": KindMyMemory}[r.URL.Path]
		calls[kind]++
		if code := status[kind]; code != 0 {
			http.Error(w, "nope", code)
			return
		}
		switch kind {
		case KindGTX:
			if r.URL.Query().Get("q") != "Halo & #dunia" && !strings.HasPrefix(r.URL.Query().Get("q"), "Hello") && r.URL.Query().Get("q") != "x" {
				t.Errorf("gtx q = %q", r.URL.Query().Get("q"))
			}
			w.Write([]byte(`{"sentences":[{"trans":"Hello. "},{"trans":"World"},{"src_translit":"x"}]}`))
		case KindDict:
			w.Write([]byte(`["Halo dunia.` + `\` + `nBaris dua."]`))
		case KindMyMemory:
			w.Write([]byte(`{"responseData":{"translatedText":"Halo (` + r.URL.Query().Get("langpair") + `)"},"responseStatus":200}`))
		}
	}))
}

func endpoints(base string, kinds ...string) []*Endpoint {
	var eps []*Endpoint
	for _, k := range kinds {
		eps = append(eps, &Endpoint{Kind: k, Base: base})
	}
	return eps
}

func TestTranslateGTX(t *testing.T) {
	calls := map[string]int{}
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls["gtx"]++
		gotQuery = r.URL.Query().Get("q")
		w.Write([]byte(`{"sentences":[{"trans":"Hello. "},{"trans":"World"},{"src_translit":"x"}]}`))
	}))
	defer srv.Close()
	tr := &Translator{Client: NewClient(5 * time.Second), Endpoints: endpoints(srv.URL, KindGTX)}
	got, err := tr.Translate(context.Background(), "Halo & #dunia", "id", "en")
	if err != nil || strings.Join(got, "|") != "Hello. |World" {
		t.Fatalf("got %q, %v", got, err)
	}
	if gotQuery != "Halo dan #dunia" {
		t.Errorf("query = %q (& must become dan, # must survive encoding)", gotQuery)
	}
}

func TestTranslateFallsBackAndCoolsDown(t *testing.T) {
	calls := map[string]int{}
	srv := translateServer(t, map[string]int{KindGTX: http.StatusTooManyRequests}, calls)
	defer srv.Close()
	tr := &Translator{Client: NewClient(5 * time.Second), Endpoints: endpoints(srv.URL, KindGTX, KindDict), Backoff: time.Millisecond}

	nl := string(rune(10))
	got, err := tr.Translate(context.Background(), "Hello world."+nl+"Line two.", "en", "id")
	if err != nil || strings.Join(got, "|") != "Halo dunia.|Baris dua." {
		t.Fatalf("got %q, %v", got, err)
	}
	if calls[KindGTX] != 1 || calls[KindDict] != 1 {
		t.Errorf("calls = %v", calls)
	}
	if _, err := tr.Translate(context.Background(), "Hello world."+nl+"Line two.", "en", "id"); err != nil || calls[KindDict] != 1 {
		t.Errorf("second call must come from the cache: %v", calls)
	}
	if _, err := tr.Translate(context.Background(), "Hello again", "en", "id"); err != nil || calls[KindGTX] != 1 || calls[KindDict] != 2 {
		t.Errorf("gtx must be skipped while cooling down: %v", calls)
	}
}

func TestTranslateReachesMyMemoryLast(t *testing.T) {
	calls := map[string]int{}
	srv := translateServer(t, map[string]int{KindGTX: 429, KindDict: 503}, calls)
	defer srv.Close()
	tr := &Translator{Client: NewClient(5 * time.Second), Endpoints: endpoints(srv.URL, KindGTX, KindDict, KindMyMemory), Backoff: time.Millisecond}
	got, err := tr.Translate(context.Background(), "snake", "auto", "id")
	if err != nil || len(got) != 1 || got[0] != "Halo (autodetect|id)" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestTranslateGivesUpWhenEverythingFails(t *testing.T) {
	calls := map[string]int{}
	srv := translateServer(t, map[string]int{KindGTX: 429, KindDict: 429, KindMyMemory: 500}, calls)
	defer srv.Close()
	tr := &Translator{Client: NewClient(5 * time.Second), Endpoints: endpoints(srv.URL, KindGTX, KindDict, KindMyMemory), Backoff: time.Millisecond, Retries: 2}
	if _, err := tr.Translate(context.Background(), "x", "en", "id"); err == nil {
		t.Fatal("expected an error")
	}
	// Each endpoint is tried once; then all are cooling down, so the retries skip them.
	if calls[KindGTX] != 1 || calls[KindDict] != 1 || calls[KindMyMemory] != 1 {
		t.Errorf("calls = %v", calls)
	}
}

func TestDefaultEndpoints(t *testing.T) {
	eps := DefaultTranslateEndpoints()
	if len(eps) != 11 || eps[0].Kind != KindGTX || eps[0].Base != "https://translate.googleapis.com" || eps[10].Kind != KindMyMemory {
		t.Errorf("want the source app's endpoint first and 10 backups ending with MyMemory, got %d", len(eps))
	}
}

func TestChunkText(t *testing.T) {
	nl := string(rune(10))
	long := strings.Repeat("Sentence number one is here. ", 30) // ~870 bytes on one line
	chunks := chunkText("Short line."+nl+long+nl+"Tail.", 450)
	for _, c := range chunks {
		if len(c) > 450 {
			t.Errorf("chunk of %d bytes", len(c))
		}
	}
	joined := strings.Join(chunks, nl)
	if !strings.HasPrefix(joined, "Short line."+nl) || !strings.HasSuffix(joined, nl+"Tail.") {
		t.Errorf("structure lost: %q", joined[:40])
	}
	if strings.Count(strings.ReplaceAll(joined, nl, " "), "Sentence number one is here.") != 30 {
		t.Error("text lost while chunking")
	}
}

func TestRateLimitCoolsTheWholeKind(t *testing.T) {
	calls := map[string]int{}
	srv := translateServer(t, map[string]int{KindGTX: 429}, calls)
	defer srv.Close()
	// Three gtx "hosts" then a dict endpoint: one 429 must skip the other gtx ones.
	tr := &Translator{Client: NewClient(5 * time.Second), Endpoints: endpoints(srv.URL, KindGTX, KindGTX, KindGTX, KindDict), Backoff: time.Millisecond}
	if _, err := tr.Translate(context.Background(), "Hello", "en", "id"); err != nil {
		t.Fatal(err)
	}
	if calls[KindGTX] != 1 || calls[KindDict] != 1 {
		t.Errorf("calls = %v, want one gtx attempt then dict", calls)
	}
}
