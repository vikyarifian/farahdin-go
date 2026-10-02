package handler

import (
	"fmt"
	"net/http"
	"net/url"
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
	shareCardText     = 1400
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

// cardBody keeps whole paragraphs (one per line) up to shareCardText runes.
func cardBody(lines []string) string {
	var out []string
	n := 0
	for _, l := range lines {
		l = strings.Join(strings.Fields(l), " ")
		if l == "" {
			continue
		}
		if n+len([]rune(l)) > shareCardText {
			out = append(out, components.Excerpt([]string{l}, max(shareCardText-n, 40)))
			break
		}
		out = append(out, l)
		n += len([]rune(l))
	}
	return strings.Join(out, "\n")
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
