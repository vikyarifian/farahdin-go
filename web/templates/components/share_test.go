package components

import (
	"net/url"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestExcerpt(t *testing.T) {
	if got := Excerpt([]string{"  Short  text. ", "Two"}, 100); got != "Short text. Two" {
		t.Errorf("short = %q", got)
	}
	long := strings.Repeat("kata ", 100)
	got := Excerpt([]string{long}, 50)
	if !strings.HasSuffix(got, "…") || utf8.RuneCountInString(got) > 51 || strings.Contains(got, "kat…") {
		t.Errorf("long = %q (must cut at a word boundary)", got)
	}
	if Excerpt(nil, 10) != "" {
		t.Error("empty lines must give an empty excerpt")
	}
}

func TestShareLinks(t *testing.T) {
	s := Share{Title: "Farahdin · Tarot · Love", Text: "🔮 Tarot · Love\n\nThe Hermit & you", Short: "🔮 Tarot — The Hermit", URL: "https://farahdin.example/"}
	wa, _ := url.Parse(s.WhatsApp())
	if wa.Host != "wa.me" || wa.Query().Get("text") != s.Text+"\n"+s.URL {
		t.Errorf("whatsapp = %s", s.WhatsApp())
	}
	x, _ := url.Parse(s.X())
	if x.Query().Get("text") != s.Short || x.Query().Get("url") != s.URL {
		t.Errorf("x = %s", s.X())
	}
	fb, _ := url.Parse(s.Facebook())
	if fb.Query().Get("u") != s.URL {
		t.Errorf("facebook = %s", s.Facebook())
	}
	tg, _ := url.Parse(s.Telegram())
	if tg.Query().Get("text") != s.Text || tg.Query().Get("url") != s.URL {
		t.Errorf("telegram = %s", s.Telegram())
	}
	if (Share{}).Enabled() {
		t.Error("empty share must be disabled")
	}
}
