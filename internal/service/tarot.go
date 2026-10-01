package service

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/url"
	"strconv"

	"github.com/vikyarifian/farahdin-go/internal/domain"
	"github.com/vikyarifian/farahdin-go/internal/external"
)

// Shuffle ports shuffle(number, length): `number` distinct cards from 1..length.
// The source sorted with Math.random()-0.5; a Fisher–Yates permutation gives the
// same kind of result without the bias.
func Shuffle(number, length int) []int {
	cards := rand.Perm(length)
	if number > length {
		number = length
	}
	out := make([]int, number)
	for i := range out {
		out[i] = cards[i] + 1
	}
	return out
}

// ValidTarotCard reports whether card can be dealt for the topic.
func ValidTarotCard(t domain.Topic, card int) bool {
	return card >= 1 && card <= t.LengthCard
}

var tarotPages = map[int]struct{ path, start, end string }{
	1: {"/us/tarot/tarot-daily-love.aspx", "Daily Love Tarot Reading", "True Love Tarot Reading"},
	3: {"/us/tarot/tarot-angel.aspx", "Angel Tarot Reading", "Guardian Angel Tarot"},
	4: {"/us/tarot/tarot-past-lives.aspx", "Past Lives Tarot", "Crystal Ball Tarot"},
}

// Tarot ports tarot.tsx generate() for the single-card topics (1, 3, 4).
func (r *Readings) Tarot(ctx context.Context, topic, card int, lang string) (Reading, error) {
	page, ok := tarotPages[topic]
	if !ok {
		page = tarotPages[1] // the source's switch default
	}
	html, err := r.Fetch.PostForm(ctx, r.Src.Horoscope+page.path, url.Values{
		"CardNumber_1_numericalint": {strconv.Itoa(card)},
	})
	if err != nil {
		return Reading{}, err
	}
	doc, err := external.ParseHTML(html)
	if err != nil {
		return Reading{}, err
	}
	text, err := horoscomPromos(external.Str(doc.Find(".grid").Text()).Split(page.start, 1)).
		Split(page.end, 0).TrimStart().Value()
	if err != nil {
		return Reading{}, err
	}
	lines, err := r.translateEnToID(ctx, lang, text)
	return Reading{Lines: lines}, err
}

// TarotTrueLove ports tarot.tsx pickCard() for topic 2 (two cards).
func (r *Readings) TarotTrueLove(ctx context.Context, you, partner int, lang string) (Reading, error) {
	html, err := r.Fetch.PostForm(ctx, r.Src.Horoscope+"/us/tarot/tarot-true-love.aspx", url.Values{
		"CardNumber_1_numericalint": {strconv.Itoa(you)},
		"CardNumber_2_numericalint": {strconv.Itoa(partner)},
	})
	if err != nil {
		return Reading{}, err
	}
	doc, err := external.ParseHTML(html)
	if err != nil {
		return Reading{}, err
	}
	text, err := horoscomPromos(external.Str(doc.Find(".grid").Text()).Split("Your Reading", 1)).
		Split("Is it true love", 0).Split("Heal your Heart", 0).Split("Reignite passion", 0).
		Split("Get day-by-day", 0).Split(strconv.Itoa(r.Now().Year())+" is a", 0).TrimStart().Value()
	if err != nil {
		return Reading{}, err
	}
	lines, err := r.translateEnToID(ctx, lang, text)
	return Reading{Lines: lines}, err
}

// TarotCardImage ports tarotCard(no): faces 1..22, the two True Love
// placeholders (-1, -2), otherwise the card back.
func TarotCardImage(no int) string {
	switch {
	case no >= 1 && no <= 22:
		return fmt.Sprintf("/static/images/tarot/%d.png", no)
	case no == -1:
		return "/static/images/tarot/your-card.png"
	case no == -2:
		return "/static/images/tarot/partner-card.png"
	}
	return "/static/images/tarot/tarot.png"
}

// TarotExtraImage ports the Angel (3) and Past Lives (4) illustration shown
// next to the card; "" for other topics.
func TarotExtraImage(topic, no int) string {
	dir := map[int]string{3: "angels", 4: "past-life"}[topic]
	if dir == "" {
		return ""
	}
	if no >= 1 && no <= 22 {
		return fmt.Sprintf("/static/images/tarot/%s/%d.png", dir, no)
	}
	return "/static/images/tarot/tarot.png"
}
