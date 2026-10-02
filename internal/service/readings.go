// Package service holds the application use cases. Reading services port the
// `generate()` functions of the source app's feature screens one-to-one; each
// port keeps the original string pipeline so parity can be reviewed side by side.
package service

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/vikyarifian/farahdin-go/internal/external"
)

// Translator translates text sentence by sentence (utils/Translate.ts).
type Translator interface {
	Translate(ctx context.Context, text, from, to string) ([]string, error)
}

// Fetcher performs upstream HTTP calls.
type Fetcher interface {
	Get(ctx context.Context, rawURL string) (string, error)
	PostForm(ctx context.Context, rawURL string, form url.Values) (string, error)
}

// Readings bundles the dependencies shared by every reading feature.
type Readings struct {
	Fetch     Fetcher
	Translate Translator
	Src       external.Sources
	Now       func() time.Time
}

// Reading is the result rendered on a result sheet.
type Reading struct {
	Lines []string
	// ImageURL is an upstream image shown under the text (Primbon "Rejeki Weton").
	ImageURL string
	// Love is the 1..5 heart score for Primbon "Jodoh"; 0 means "do not show".
	Love int
	// Notice is a non-fatal message shown with the result (e.g. part of it failed).
	Notice string
}

// ErrValidation wraps user-input problems that should be shown on the form.
type ErrValidation struct{ Msg string }

func (e *ErrValidation) Error() string { return e.Msg }

// IsValidation reports whether err is a validation error and returns its message.
func IsValidation(err error) (string, bool) {
	var v *ErrValidation
	if errors.As(err, &v) {
		return v.Msg, true
	}
	return "", false
}

// adsPush is the AdSense snippet that cheerio's .text() leaks from <script> tags.
const adsPush = "(adsbygoogle = window.adsbygoogle || []).push({});"

// CleanLines ports the result renderer:
// (a.trimStart().trimEnd() !== '.' ? a.trimStart() : ”) and drops empty lines.
func CleanLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		t := external.Str(l).TrimStart().TrimEnd().String()
		if t == "." || t == "" || t == "<br>" {
			continue
		}
		out = append(out, external.Str(l).TrimStart().String())
	}
	return out
}

// translateStatus records, per reading, whether a translation fell back to
// the original text.
type translateStatus struct{ failed atomic.Bool }

type translateStatusKey struct{}

func withTranslateStatus(ctx context.Context) (context.Context, *translateStatus) {
	st := &translateStatus{}
	return context.WithValue(ctx, translateStatusKey{}, st), st
}

func translationFailed(ctx context.Context) bool {
	st, _ := ctx.Value(translateStatusKey{}).(*translateStatus)
	return st != nil && st.failed.Load()
}

// translate calls the translator; when it still fails after its retries the
// original text is returned (one line per paragraph) and the reading is
// marked, so the user gets the result in its source language with a notice
// instead of an error. The error result is always nil.
func (r *Readings) translate(ctx context.Context, text, from, to string) ([]string, error) {
	lines, err := r.Translate.Translate(ctx, text, from, to)
	if err == nil {
		return lines, nil
	}
	slog.WarnContext(ctx, "translation unavailable, showing original text", "from", from, "to", to, "err", err)
	if st, _ := ctx.Value(translateStatusKey{}).(*translateStatus); st != nil {
		st.failed.Store(true)
	}
	return strings.Split(text, newline), nil
}

// newline separates paragraphs in the ported string pipelines.
const newline = "\x0a"

// noteTranslation adds the fallback notice to a finished reading.
func noteTranslation(st *translateStatus, lang string, res *Reading) {
	if st.failed.Load() && res.Notice == "" {
		res.Notice = pick(lang == "ID",
			"Terjemahan sedang tidak tersedia, hasil ditampilkan dalam bahasa aslinya.",
			"Translation is unavailable right now; the result is shown in its original language.")
	}
}

// translateOrSplit ports the common tail:
// if (lang === 'ID') translate(text, 'en', 'id') else text.split('\n').
func (r *Readings) translateEnToID(ctx context.Context, lang, text string) ([]string, error) {
	if lang == "ID" {
		return r.translate(ctx, text, "en", "id")
	}
	return strings.Split(text, "\n"), nil
}

// translateIDToEN ports the Primbon tail used by topics 5–9:
// EN: translate(content.replace('\n','. .%.').split('.%.').join('. ').replace('...','.').replace('..','.'), 'id', 'en')
// ID: content.split('\n').
func (r *Readings) translateIDToEN(ctx context.Context, lang, content string) ([]string, error) {
	if lang == "EN" {
		text := external.Str(content).Replace("\n", ". .%.").ReplaceAll(".%.", ". ").
			Replace("...", ".").Replace("..", ".").String()
		return r.translate(ctx, text, "id", "en")
	}
	return strings.Split(content, "\n"), nil
}

// birthChartPromo is a horoscope.com upsell added after tarot and gem
// readings (seen 2026-10-02, not in the source app's cut list; deviation D-29).
// Everything from it onwards is promo, so the text is cut there.
const birthChartPromo = "Discover the key to your unique life path"

// horoscomPromos are the promo sentences stripped from horoscope.com pages.
func horoscomPromos(s external.S) external.S {
	return s.Split("The Year of the", 0).Split("Heal your heart!", 0).
		Split("Get a FREE Love & Relationship Reading", 0).Split("Gain the clarity you seek", 0).
		Split("Grow your astrology knowledge", 0).Split("Gain the clarity you seek", 0).
		Split("More Horoscopes for", 0).Split("Get Love Insights for FREE", 0).
		Split(birthChartPromo, 0)
}
