package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/vikyarifian/farahdin-go/internal/domain"
	"github.com/vikyarifian/farahdin-go/internal/http/response"
	"github.com/vikyarifian/farahdin-go/internal/i18n"
	"github.com/vikyarifian/farahdin-go/internal/service"
	"github.com/vikyarifian/farahdin-go/web/templates/components"
	"github.com/vikyarifian/farahdin-go/web/templates/pages"
)

func (a *App) topics(titleID, titleEN string, topics []domain.Topic, base string, large bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response.Render(w, r, http.StatusOK, pages.Topics(i18n.T(r.Context(), titleID, titleEN), topics, base, large))
	}
}

func invalidDate(r *http.Request) string {
	return i18n.T(r.Context(), "Tanggal tidak valid.", "Date is invalid.")
}

// syncBirthday ports the Horoscope/Primbon effect that saved a changed birthday.
func (a *App) syncBirthday(r *http.Request, u *domain.User, birthday string) {
	b, ok := domain.ParseDate(birthday)
	if !ok {
		return
	}
	if err := a.Profiles.SyncBirthday(r.Context(), u, b); err != nil {
		slog.ErrorContext(r.Context(), "sync birthday", "err", err)
	}
}

func readingResult(title, subtitle, closeHref string, reading service.Reading, err error, r *http.Request) pages.ResultView {
	if err != nil {
		return pages.ResultView{Error: readingError(r, err)}
	}
	reading.Lines = service.CleanLines(reading.Lines)
	return pages.ResultView{Title: title, Subtitle: subtitle, CloseHref: closeHref, Reading: reading, Show: true}
}

// --- Primbon ---

func (a *App) primbonView(r *http.Request, t domain.Topic) pages.PrimbonView {
	u := user(r)
	birthday := domain.FormatDate(u.BirthdayOr(a.today()))
	return pages.PrimbonView{
		Topic:    t,
		Fields:   service.PrimbonFieldsFor(t.Key),
		Name:     u.Fullname,
		Birthday: birthday,
		Date:     birthday, // the source initialised both dates from the profile birthday
		Today:    domain.FormatDate(a.today()),
	}
}

func (a *App) primbonForm(w http.ResponseWriter, r *http.Request) {
	t, ok := topicParam(r, domain.PrimbonTopics)
	if !ok {
		a.notFound(w, r)
		return
	}
	response.Render(w, r, http.StatusOK, pages.Primbon(a.primbonView(r, t)))
}

func (a *App) primbonSubmit(w http.ResponseWriter, r *http.Request) {
	t, ok := topicParam(r, domain.PrimbonTopics)
	if !ok {
		a.notFound(w, r)
		return
	}
	u := user(r)
	v := a.primbonView(r, t)
	v.Name = strings.TrimSpace(r.FormValue("name"))
	v.Dream = strings.TrimSpace(r.FormValue("dream"))
	v.Partner = strings.TrimSpace(r.FormValue("partner"))
	if v.Fields.Birthday {
		v.Birthday = r.FormValue("birthday")
	}
	if v.Fields.Date {
		v.Date = r.FormValue("date")
	}

	birthday, okB := domain.ParseDate(v.Birthday)
	date, okD := domain.ParseDate(v.Date)
	if (v.Fields.Birthday && !okB) || (v.Fields.Date && !okD) {
		v.Result = pages.ResultView{Error: invalidDate(r)}
	} else {
		if v.Fields.Birthday {
			a.syncBirthday(r, u, v.Birthday)
		}
		reading, err := a.Readings.Primbon(r.Context(), t.Key, service.PrimbonInput{
			Name: v.Name, Dream: v.Dream, Partner: v.Partner, Birthday: birthday, Date: date,
		}, lang(r))
		v.Result = readingResult("Primbon", t.Label(lang(r)), "/primbon/"+strconv.Itoa(t.Key), reading, err, r)
	}
	fragmentOrPage(w, r, pages.Result(v.Result), pages.Primbon(v))
}

// --- Horoscope ---

func (a *App) horoscopeView(r *http.Request, t domain.Topic) pages.HoroscopeView {
	birthday := domain.FormatDate(user(r).BirthdayOr(a.today()))
	return pages.HoroscopeView{
		Topic:       t,
		Birthday:    birthday,
		PartnerDate: birthday,
		Sign:        domain.Zodiac(birthday),
		Today:       domain.FormatDate(a.today()),
	}
}

func (a *App) horoscopeForm(w http.ResponseWriter, r *http.Request) {
	t, ok := topicParam(r, domain.HoroscopeTopics)
	if !ok {
		a.notFound(w, r)
		return
	}
	response.Render(w, r, http.StatusOK, pages.Horoscope(a.horoscopeView(r, t)))
}

func (a *App) horoscopeSubmit(w http.ResponseWriter, r *http.Request) {
	t, ok := topicParam(r, domain.HoroscopeTopics)
	if !ok {
		a.notFound(w, r)
		return
	}
	v := a.horoscopeView(r, t)
	v.Birthday = r.FormValue("birthday")
	if t.Key == 6 {
		v.PartnerDate = r.FormValue("partner_date")
	}
	birthday, okB := domain.ParseDate(v.Birthday)
	partner, okP := domain.ParseDate(v.PartnerDate)
	if !okB || !okP {
		v.Result = pages.ResultView{Error: invalidDate(r)}
	} else {
		a.syncBirthday(r, user(r), v.Birthday)
		in := service.HoroscopeInput{Birthday: birthday, PartnerDate: partner}
		v.Sign, v.PartnerSign = service.HoroscopeSigns(in)
		subtitle := t.Label(lang(r)) + " - " + v.Sign
		if t.Key == 6 {
			subtitle += " & " + v.PartnerSign
		}
		reading, err := a.Readings.Horoscope(r.Context(), t.Key, in, lang(r))
		v.Result = readingResult(i18n.T(r.Context(), "Horoskop", "Horoscope"), subtitle, "/horoscope/"+strconv.Itoa(t.Key), reading, err, r)
	}
	fragmentOrPage(w, r, pages.Result(v.Result), pages.Horoscope(v))
}

// horoscopeSign refreshes the zodiac image and label when the birthday changes.
func (a *App) horoscopeSign(w http.ResponseWriter, r *http.Request) {
	sign := ""
	if b, ok := formDate(r, "birthday"); ok {
		sign = domain.Zodiac(domain.FormatDate(b))
	}
	response.Render(w, r, http.StatusOK, pages.HoroscopeSignSwap(sign))
}

// --- Clairvoyance ---

func (a *App) clairvoyanceForm(w http.ResponseWriter, r *http.Request) {
	t, ok := topicParam(r, domain.ClairvoyanceTopics)
	if !ok {
		a.notFound(w, r)
		return
	}
	response.Render(w, r, http.StatusOK, pages.Clairvoyance(pages.ClairvoyanceView{Topic: t, Name: user(r).Fullname}))
}

func (a *App) clairvoyanceSubmit(w http.ResponseWriter, r *http.Request) {
	t, ok := topicParam(r, domain.ClairvoyanceTopics)
	if !ok {
		a.notFound(w, r)
		return
	}
	v := pages.ClairvoyanceView{Topic: t, Name: strings.TrimSpace(r.FormValue("name"))}
	reading, err := a.Readings.Clairvoyance(r.Context(), t.Key, v.Name, lang(r))
	v.Result = readingResult(i18n.T(r.Context(), "Kewaskitaan", "Clairvoyance"), t.Label(lang(r)), "/clairvoyance/"+strconv.Itoa(t.Key), reading, err, r)
	fragmentOrPage(w, r, pages.Result(v.Result), pages.Clairvoyance(v))
}

// --- Tarot ---

func (a *App) tarotView(t domain.Topic) pages.TarotView {
	return pages.TarotView{Topic: t, Cards: service.Shuffle(t.PickCard, t.LengthCard)}
}

func (a *App) tarotSpread(w http.ResponseWriter, r *http.Request) {
	t, ok := topicParam(r, domain.TarotTopics)
	if !ok {
		a.notFound(w, r)
		return
	}
	response.Render(w, r, http.StatusOK, pages.Tarot(a.tarotView(t)))
}

func formCard(r *http.Request, name string, t domain.Topic) (int, bool) {
	n, err := strconv.Atoi(r.FormValue(name))
	return n, err == nil && service.ValidTarotCard(t, n)
}

// tarotPick ports pickCard() for single-card topics: open the sheet with the chosen card.
func (a *App) tarotPick(w http.ResponseWriter, r *http.Request) {
	t, ok := topicParam(r, domain.TarotTopics)
	if !ok || t.Key == 2 {
		a.notFound(w, r)
		return
	}
	card, ok := formCard(r, "card", t)
	v := a.tarotView(t)
	if !ok {
		msg := i18n.T(r.Context(), "Pilih satu kartu dulu.", "Pick a card first.")
		v.Error = msg
		fragmentOrPage(w, r, components.Alert(msg), pages.Tarot(v))
		return
	}
	sheet := pages.TarotSheet{Topic: t, Card: card}
	v.Sheet = &sheet
	fragmentOrPage(w, r, pages.TarotResult(sheet), pages.Tarot(v))
}

// tarotRead ports generate() (topics 1, 3, 4) and the True Love branch of pickCard().
func (a *App) tarotRead(w http.ResponseWriter, r *http.Request) {
	t, ok := topicParam(r, domain.TarotTopics)
	if !ok {
		a.notFound(w, r)
		return
	}
	v := a.tarotView(t)
	sheet := pages.TarotSheet{Topic: t}

	if t.Key == 2 {
		you, okY := formCard(r, "you", t)
		partner, okP := formCard(r, "partner", t)
		if !okY || !okP {
			sheet.Error = i18n.T(r.Context(), "Pilih kartumu dan kartu pasangan.", "Pick your card and your partner's card.")
		} else {
			sheet.You, sheet.Partner = you, partner
			reading, err := a.Readings.TarotTrueLove(r.Context(), you, partner, lang(r))
			if err != nil {
				sheet.Error = readingError(r, err)
			} else {
				sheet.Read, sheet.Reading = true, service.Reading{Lines: service.CleanLines(reading.Lines)}
			}
		}
		v.Sheet = &sheet
		fragmentOrPage(w, r, pages.TarotResult(sheet), pages.Tarot(v))
		return
	}

	card, ok := formCard(r, "card", t)
	if !ok {
		a.notFound(w, r)
		return
	}
	sheet.Card = card
	reading, err := a.Readings.Tarot(r.Context(), t.Key, card, lang(r))
	if err != nil {
		sheet.Error = readingError(r, err)
	} else {
		sheet.Read, sheet.Reading = true, service.Reading{Lines: service.CleanLines(reading.Lines)}
	}
	v.Sheet = &sheet
	fragmentOrPage(w, r, pages.TarotReading(sheet), pages.Tarot(v))
}

// --- Matrix Destiny ---

func (a *App) matrixView(r *http.Request) pages.MatrixView {
	u := user(r)
	t, _ := domain.FindTopic(domain.MatrixTopics, 1)
	return pages.MatrixView{
		Topic:    t,
		Name:     u.Fullname,
		Birthday: domain.FormatDate(u.BirthdayOr(a.today())),
		Today:    domain.FormatDate(a.today()),
	}
}

func (a *App) matrixForm(w http.ResponseWriter, r *http.Request) {
	response.Render(w, r, http.StatusOK, pages.Matrix(a.matrixView(r)))
}

func (a *App) matrixSubmit(w http.ResponseWriter, r *http.Request) {
	v := a.matrixView(r)
	v.Name = strings.TrimSpace(r.FormValue("name"))
	v.Birthday = r.FormValue("birthday")
	birthday, _ := domain.ParseDate(v.Birthday) // zero on failure; validation reports it
	res, err := a.Readings.MatrixDestiny(r.Context(), v.Name, birthday, lang(r))
	if err != nil {
		v.Error = readingError(r, err)
	} else {
		v.Result = &res
	}
	fragmentOrPage(w, r, pages.MatrixResult(v), pages.Matrix(v))
}
