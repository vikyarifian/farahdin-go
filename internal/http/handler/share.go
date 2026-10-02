package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/vikyarifian/farahdin-go/internal/domain"
	"github.com/vikyarifian/farahdin-go/internal/i18n"
	"github.com/vikyarifian/farahdin-go/web/templates/components"
)

// Excerpt lengths: the full message for WhatsApp/Telegram/native share, a
// shorter one that keeps an X post under 280 characters with the link, and
// the text drawn on the share image (share-card.js fits it to the space left).
const (
	shareExcerpt      = 260
	shareExcerptShort = 140
	shareCardText     = 3000
)

// shareInput describes one result for sharing.
type shareInput struct {
	Title, Subtitle string
	Highlight       string   // e.g. love score or matrix purposes
	Lines           []string // cleaned reading lines
	Images          []string // card images for the share image
	Icons           []string // small images (zodiac signs)
	Love            int
	Chart           bool
	Decor           string // feature illustration (web.Asset path)
}

// share builds the social-media message and share image for a result.
func (a *App) share(r *http.Request, in shareInput) components.Share {
	body := in.Lines
	if in.Highlight != "" {
		lead := in.Highlight
		if in.Love > 0 {
			lead += " " + strings.Repeat("❤", in.Love) // the image draws heart icons instead
		}
		body = append([]string{lead}, in.Lines...)
	}
	excerpt := components.Excerpt(body, shareExcerpt)
	if excerpt == "" {
		return components.Share{}
	}
	heading := in.Title
	if in.Subtitle != "" {
		heading += " · " + in.Subtitle
	}
	cta := i18n.T(r.Context(), "Cek ramalanmu sendiri di Farahdin:", "Get your own reading on Farahdin:")
	appURL := a.Cfg.BaseURL + "/"
	host := appURL
	if u, err := url.Parse(a.Cfg.BaseURL); err == nil && u.Host != "" {
		host = u.Host
	}
	return components.Share{
		Title: "Farahdin · " + heading,
		Text:  "🔮 " + heading + "\n\n" + excerpt + "\n\n" + cta,
		Short: "🔮 " + heading + " — " + components.Excerpt(body, shareExcerptShort) + " " + cta,
		URL:   appURL,
		Card: components.ShareCard{
			Heading:   heading,
			Highlight: in.Highlight,
			Body:      cardBody(in.Lines),
			Images:    in.Images,
			Icons:     in.Icons,
			Love:      in.Love,
			Chart:     in.Chart,
			Decor:     in.Decor,
			CTA:       i18n.T(r.Context(), "Cek ramalanmu di Farahdin", "Get your reading on Farahdin"),
			Host:      host,
		},
	}
}

// cardBody keeps whole paragraphs (one per line) up to shareCardText runes;
// a paragraph that does not fit is cut after its last complete sentence,
// never in the middle of one. share-card.js fits the result to the image.
func cardBody(lines []string) string {
	var out []string
	n := 0
	for _, l := range lines {
		l = strings.Join(strings.Fields(l), " ")
		if l == "" {
			continue
		}
		if size := len([]rune(l)); n+size > shareCardText {
			if head := wholeSentences(l, shareCardText-n); head != "" {
				out = append(out, head)
			}
			break
		}
		out = append(out, l)
		n += len([]rune(l))
	}
	return strings.Join(out, "\n")
}

// sentenceEnd matches the end of a sentence: . ! ? (optionally closed by a
// quote or bracket) followed by a space or the end of the text.
var sentenceEnd = regexp.MustCompile(`[.!?]["'”’)\]]*(\s|$)`)

// wholeSentences returns the longest prefix of s made of complete sentences
// that is at most limit runes, or "" when not even one sentence fits.
func wholeSentences(s string, limit int) string {
	best := ""
	for _, loc := range sentenceEnd.FindAllStringIndex(s, -1) {
		head := strings.TrimSpace(s[:loc[1]])
		if len([]rune(head)) > limit {
			break
		}
		best = head
	}
	return best
}

// loveHighlight ports the Primbon "Jodoh" heart meter into words.
func loveHighlight(r *http.Request, love int) string {
	if love <= 0 {
		return ""
	}
	return fmt.Sprintf("%s %d/5", i18n.T(r.Context(), "Kecocokan", "Compatibility"), love)
}

// matrixHighlight summarises the purposes of a destiny matrix.
func matrixHighlight(r *http.Request, m domain.Matrix) string {
	t := func(id, en string) string { return i18n.T(r.Context(), id, en) }
	return fmt.Sprintf("%s %d · %s %d · %s %d · %s %d.",
		t("Tujuan pribadi", "Personal purpose"), m.Purposes["perspurpose"],
		t("Masyarakat", "Society"), m.Purposes["socialpurpose"],
		t("Hidup", "Lifetime"), m.Purposes["generalpurpose"],
		t("Planet", "Planetary"), m.Purposes["planetarypurpose"])
}
