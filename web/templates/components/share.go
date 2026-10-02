package components

import (
	"net/url"
	"strings"
	"unicode/utf8"
)

// Share is what a result sheet offers to share: a short text summary of the
// reading plus a link to the app. Results are not stored, so there is no
// per-result URL; the link invites the reader to get their own reading.
type Share struct {
	Title string // e.g. "Tarot · Love"
	Text  string // full message: title, excerpt, call to action
	Short string // tweet-sized message (X counts the URL separately)
	URL   string // public app URL

	// Card is drawn by web/static/js/share-card.js into a 1080×1350 image.
	Card ShareCard
}

// ShareCard is the content of the shareable image. Every image URL must be
// same-origin so the canvas stays exportable.
type ShareCard struct {
	Heading   string   // "Tarot · Love"
	Highlight string   // optional gold line, e.g. the love score
	Body      string   // reading text, one paragraph per line
	Images    []string // large images side by side (tarot cards)
	Icons     []string // small images side by side (zodiac signs)
	Love      int      // 1..5 hearts (Primbon Jodoh); 0 = none
	Chart     bool     // draw the page's #matrix SVG (Matrix Destiny)
	Decor     string   // feature illustration when there is no other media
	CTA       string   // footer call to action
	Host      string   // footer address, e.g. "farahdin.app"
}

// ImageList joins Images or Icons for a data attribute.
func ImageList(urls []string) string { return strings.Join(urls, " ") }

// Enabled reports whether there is anything to share.
func (s Share) Enabled() bool { return s.Text != "" && s.URL != "" }

// WhatsApp returns the wa.me share link.
func (s Share) WhatsApp() string {
	return "https://wa.me/?text=" + url.QueryEscape(s.Text+"\n"+s.URL)
}

// Telegram returns the Telegram share link.
func (s Share) Telegram() string {
	return "https://t.me/share/url?url=" + url.QueryEscape(s.URL) + "&text=" + url.QueryEscape(s.Text)
}

// X returns the X (Twitter) post intent link.
func (s Share) X() string {
	return "https://twitter.com/intent/tweet?text=" + url.QueryEscape(s.Short) + "&url=" + url.QueryEscape(s.URL)
}

// Facebook returns the Facebook sharer link (Facebook only accepts a URL).
func (s Share) Facebook() string {
	return "https://www.facebook.com/sharer/sharer.php?u=" + url.QueryEscape(s.URL)
}

// Excerpt joins reading lines into one paragraph and cuts it to at most
// max runes at a word boundary, adding "…" when it was shortened.
func Excerpt(lines []string, max int) string {
	text := strings.Join(strings.Fields(strings.Join(lines, " ")), " ")
	if utf8.RuneCountInString(text) <= max {
		return text
	}
	r := []rune(text)[:max]
	cut := string(r)
	if i := strings.LastIndexAny(cut, " "); i > max/2 {
		cut = cut[:i]
	}
	return strings.TrimRight(cut, " ,.;:-") + "…"
}

// CopyText is the message copied to the clipboard: text, then the link.
func (s Share) CopyText() string { return s.Text + "\n" + s.URL }
