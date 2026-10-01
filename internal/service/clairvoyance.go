package service

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/vikyarifian/farahdin-go/internal/external"
)

// Clairvoyance ports app/pages/clairvoyance.tsx generate(): three random
// cards and the first name are posted to a horoscope.com tarot page and the
// topic's section is cut out of the reading.
func (r *Readings) Clairvoyance(ctx context.Context, topic int, name, lang string) (Reading, error) {
	if strings.TrimSpace(name) == "" {
		return Reading{}, &ErrValidation{pick(lang == "ID", "Nama wajib diisi.", "Name is required.")}
	}
	cards := Shuffle(3, 22)
	form := url.Values{
		"CardNumber_1_numericalint": {strconv.Itoa(cards[0])},
		"CardNumber_2_numericalint": {strconv.Itoa(cards[1])},
		"CardNumber_3_numericalint": {strconv.Itoa(cards[2])},
		"FirstName":                 {strings.Split(name, " ")[0]},
	}
	path, start, end := "/us/tarot/tarot-gems.aspx", "Gems Oracle", "Today's Tip:"
	if topic == 1 {
		// The source hard-coded "2025 Tarot Reading" (deviation D-07).
		path, start, end = "/us/tarot/tarot-daily.aspx", "Your reading for Today:", strconv.Itoa(r.Now().Year())+" Tarot Reading"
	}
	html, err := r.Fetch.PostForm(ctx, r.Src.Horoscope+path, form)
	if err != nil {
		return Reading{}, err
	}
	doc, err := external.ParseHTML(html)
	if err != nil {
		return Reading{}, err
	}
	content := external.Str(doc.Find(".grid").Text()).Split(start, 1).Split(end, 0).TrimStart().TrimEnd()
	switch topic {
	case 2:
		content = content.Split("Career: ", 0)
	case 3:
		content = content.Replace("Career: ", "Career:x Career: ").Split("Career:x ", 1).Split("Wellness: ", 0)
	case 4:
		content = content.Replace("Wellness: ", "Wellness:x Wellness: ").Split("Wellness:x ", 1)
	}
	text, err := content.Value()
	if err != nil {
		return Reading{}, err
	}
	for _, cut := range []string{"Reignite passion", "Change are coming", "Get a Free", birthChartPromo} {
		if strings.Contains(text, cut) {
			text = strings.Split(text, cut)[0]
		}
	}
	lines, err := r.translateEnToID(ctx, lang, text)
	return Reading{Lines: lines}, err
}
