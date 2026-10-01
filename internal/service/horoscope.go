package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/vikyarifian/farahdin-go/internal/domain"
	"github.com/vikyarifian/farahdin-go/internal/external"
)

// HoroscopeInput is the Horoscope form.
type HoroscopeInput struct {
	Birthday    time.Time
	PartnerDate time.Time // only for topic 6 (Match)
}

// HoroscopeSigns derives the user's and partner's signs from the dates.
func HoroscopeSigns(in HoroscopeInput) (sign, partner string) {
	return domain.Zodiac(domain.FormatDate(in.Birthday)), domain.Zodiac(domain.FormatDate(in.PartnerDate))
}

var horoscopePages = map[int]string{
	1: "/us/horoscopes/general/horoscope-general-daily-yesterday.aspx",
	2: "/us/horoscopes/general/horoscope-general-daily-today.aspx",
	3: "/us/horoscopes/general/horoscope-general-daily-tomorrow.aspx",
	4: "/us/horoscopes/general/horoscope-general-weekly.aspx",
	5: "/us/horoscopes/general/horoscope-general-monthly.aspx",
}

// Horoscope ports app/pages/horoscope.tsx generate().
func (r *Readings) Horoscope(ctx context.Context, topic int, in HoroscopeInput, lang string) (Reading, error) {
	sign, partner := HoroscopeSigns(in)
	if topic == 6 {
		return r.horoscopeMatch(ctx, sign, partner, lang)
	}
	page, ok := horoscopePages[topic]
	if !ok {
		return Reading{}, fmt.Errorf("horoscope: unknown topic %d", topic)
	}
	html, err := r.Fetch.Get(ctx, r.Src.Horoscope+page+"?sign="+strconv.Itoa(domain.ZodiacNumber(sign)))
	if err != nil {
		return Reading{}, err
	}
	doc, err := external.ParseHTML(html)
	if err != nil {
		return Reading{}, err
	}
	// The source used the birth year here, so the promo was never removed;
	// the current year is what horoscope.com prints (deviation D-07).
	promo := "Reveal what " + strconv.Itoa(r.Now().Year()) + " has in the stars for you with your Yearly Horoscope!"
	content := horoscomPromos(external.Str(doc.Find(".main-horoscope > p").Text()).Replace(promo, ""))
	if topic != 4 {
		// Daily and monthly pages start with "<date> - "; weekly does not.
		content = content.Split(" - ", 1)
	}
	text, err := content.Value()
	if err != nil {
		return Reading{}, err
	}
	lines, err := r.translateEnToID(ctx, lang, text)
	return Reading{Lines: lines}, err
}

// horoscopeMatch ports case 6 (californiapsychics.com compatibility article).
func (r *Readings) horoscopeMatch(ctx context.Context, sign, partner, lang string) (Reading, error) {
	if sign == "" || partner == "" {
		return Reading{}, &ErrValidation{pick(lang == "ID", "Tanggal lahir tidak valid.", "Birthdate is invalid.")}
	}
	page := fmt.Sprintf("/blog/zodiac-sign-compatibility-blog/%s-%s-compatibility.html", strings.ToLower(sign), strings.ToLower(partner))
	html, err := r.Fetch.Get(ctx, r.Src.CaliforniaPsychic+page)
	if err != nil {
		return Reading{}, err
	}
	doc, err := external.ParseHTML(strings.ReplaceAll(strings.ReplaceAll(html, ".</p>", ".\n"), "</p>", "\n"))
	if err != nil {
		return Reading{}, err
	}
	text, err := external.Str(doc.Text()).ReplaceAll("</a>.", "").ReplaceAll("</a>", "").ReplaceAll("</div>", "").
		Split("Matters of the", 0).Split("LJ Innes", 2).Replace("</a></p></div></div></div>", "").TrimStart().
		Split("Have you been searching for your soulmate?", 0).Split("About California Psychics", 0).Value()
	if err != nil {
		return Reading{}, err
	}
	lines, err := r.translateEnToID(ctx, lang, text)
	return Reading{Lines: lines}, err
}
