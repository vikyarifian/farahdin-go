package service

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/vikyarifian/farahdin-go/internal/domain"
	"github.com/vikyarifian/farahdin-go/internal/external"
)

// PrimbonInput is the Primbon form. Which fields are used depends on the topic.
type PrimbonInput struct {
	Name     string
	Dream    string
	Partner  string
	Birthday time.Time // shown for topics 5, 6, 7; also used (from the profile) by 8 and 9
	Date     time.Time // "Tanggal" / partner birthdate, shown for topics 4, 5, 8, 9
}

// PrimbonFields reports which inputs the form shows for a topic
// (ports the [1,3,5,7] / [2] / [5,6,7] / [3,5] / [4,5,8,9] checks).
type PrimbonFields struct {
	Name, Dream, Birthday, Partner, Date bool
}

// PrimbonFieldsFor returns the visible fields for topic.
func PrimbonFieldsFor(topic int) PrimbonFields {
	return PrimbonFields{
		Name:     domain.TopicIn(topic, 1, 3, 5, 7),
		Dream:    topic == 2,
		Birthday: domain.TopicIn(topic, 5, 6, 7),
		Partner:  domain.TopicIn(topic, 3, 5),
		Date:     domain.TopicIn(topic, 4, 5, 8, 9),
	}
}

// ValidatePrimbon checks the inputs a topic needs. The source app sent empty
// values upstream; the PWA asks for them instead (docs/migration/deviations.md).
func ValidatePrimbon(topic int, in PrimbonInput, lang string) error {
	f := PrimbonFieldsFor(topic)
	id := lang == "ID"
	switch {
	case f.Name && strings.TrimSpace(in.Name) == "":
		return &ErrValidation{pick(id, "Nama wajib diisi.", "Name is required.")}
	case f.Dream && strings.TrimSpace(in.Dream) == "":
		return &ErrValidation{pick(id, "Mimpi wajib diisi.", "Dream is required.")}
	case f.Partner && strings.TrimSpace(in.Partner) == "":
		return &ErrValidation{pick(id, "Nama pasangan wajib diisi.", "Partner name is required.")}
	}
	return nil
}

// Primbon ports app/pages/primbon.tsx generate().
func (r *Readings) Primbon(ctx context.Context, topic int, in PrimbonInput, lang string) (res Reading, err error) {
	ctx, st := withTranslateStatus(ctx)
	defer func() { noteTranslation(st, lang, &res) }()
	if err := ValidatePrimbon(topic, in, lang); err != nil {
		return Reading{}, err
	}
	switch topic {
	case 1:
		return r.primbonNameMeaning(ctx, in, lang)
	case 2:
		return r.primbonDream(ctx, in, lang)
	case 3:
		return r.primbonMatch(ctx, in, lang)
	case 4:
		return r.primbonImportantDate(ctx, in, lang)
	case 5:
		return r.primbonSoulmate(ctx, in, lang)
	case 6:
		return r.primbonWetonFortune(ctx, in, lang)
	case 7:
		return r.primbonNameCompatibility(ctx, in, lang)
	case 8:
		return r.primbonGoodDay(ctx, in, lang)
	case 9:
		return r.primbonProhibitedDay(ctx, in, lang)
	}
	return Reading{}, fmt.Errorf("primbon: unknown topic %d", topic)
}

func (r *Readings) primbonDoc(ctx context.Context, get bool, path string, form url.Values) (*external.Doc, error) {
	var body string
	var err error
	if get {
		body, err = r.Fetch.Get(ctx, r.Src.Primbon+path)
	} else {
		body, err = r.Fetch.PostForm(ctx, r.Src.PrimbonPost+path, form)
	}
	if err != nil {
		return nil, err
	}
	return external.ParseHTML(body)
}

// case 1: Arti Nama.
func (r *Readings) primbonNameMeaning(ctx context.Context, in PrimbonInput, lang string) (Reading, error) {
	doc, err := r.primbonDoc(ctx, true, "/arti_nama.php?nama1="+url.QueryEscape(in.Name)+"&proses=+Submit%21+", nil)
	if err != nil {
		return Reading{}, err
	}
	body := external.Str(doc.Find("#body").Text()).Split("Nama:", 0)
	cleaned := body.Replace("ARTI NAMA", "").Replace(strings.Repeat(" ", 32), "").
		Replace(adsPush, "").ReplaceRe(external.ReNewline, "").TrimStart()
	if lang == "ID" {
		lines, err := cleaned.SplitAll("\n")
		return Reading{Lines: lines}, err
	}
	text, err := cleaned.ReplaceReAll(external.ReNewline, "").Value()
	if err != nil {
		return Reading{}, err
	}
	lines, err := r.translate(ctx, text, "id", "en")
	return Reading{Lines: lines}, err
}

// case 2: Tafsir Mimpi.
func (r *Readings) primbonDream(ctx context.Context, in PrimbonInput, lang string) (Reading, error) {
	trans, err := r.translate(ctx, in.Dream, "auto", "id")
	if err != nil {
		return Reading{}, err
	}
	params := strings.Join(trans, " ")
	doc, err := r.primbonDoc(ctx, true, "/tafsir_mimpi.php?mimpi="+url.QueryEscape(params)+"&submit=+Submit+", nil)
	if err != nil {
		return Reading{}, err
	}
	if strings.Contains(doc.Find("#body > font > i").Text(), "Tidak ditemukan") {
		if lang == "ID" {
			return Reading{Lines: []string{"Tidak ditemukan tafsir mimpi " + params + ". Cari dengan kata kunci yang lain.."}}, nil
		}
		return Reading{Lines: []string{"No interpretation of the dream of " + params + ". Search with other keywords.."}}, nil
	}
	base := external.Str(doc.Find("#body").Text()).Replace(adsPush, "").
		Split("Solusi - Menanggulangi akibat dari tafsir mimpi yang buruk", 0).
		Split("Hasil pencarian untuk kata kunci: "+params, 1)
	if lang == "ID" {
		lines, err := base.ReplaceAll(".", ".\n").TrimStart().SplitAll("\n")
		return Reading{Lines: lines}, err
	}
	text, err := base.ReplaceAll(".", ". ").TrimStart().Value()
	if err != nil {
		return Reading{}, err
	}
	lines, err := r.translate(ctx, text, "id", "en")
	return Reading{Lines: lines}, err
}

// case 3: Jodoh (name compatibility of a couple), with the love meter image.
func (r *Readings) primbonMatch(ctx context.Context, in PrimbonInput, lang string) (Reading, error) {
	path := "/kecocokan_nama_pasangan.php?nama1=" + url.QueryEscape(in.Name) + "&nama2=" + url.QueryEscape(in.Partner) + "&proses=+Submit%21+"
	doc, err := r.primbonDoc(ctx, true, path, nil)
	if err != nil {
		return Reading{}, err
	}
	src, _ := doc.Find("#body > img").Attr("src")
	image := "https://www.primbon.com/" + src
	body := external.Str(doc.Find("#body").Text()).Split(in.Partner, 1)
	html, _ := doc.Find("#body").HTML()

	id := lang == "ID"
	lines := []string{
		pick(id, "Nama Anda: ", "Your Name: ") + in.Name,
		pick(id, "Pasangan: ", "Partner: ") + in.Partner,
		"",
	}
	negative := external.Str(html).Replace(adsPush, "").Split(`<img border="0"`, 0).
		Split("Sisi Negatif Anda:", 1).ReplaceAll("<br>", "").ReplaceAll("</b>", "")
	// bodyHtml.split('&lt; Hitung Kembali')[0].split('png"><br><br>')[1]?.split('<a href="')[0]... || ''
	// The source cut at `<a href="`; primbon.com now writes `<a class="button" href=…>`,
	// so the cut also stops at any `<a ` (deviation D-08).
	advice := external.Str(html).Replace(adsPush, "").Split("&lt; Hitung Kembali", 0).
		Split(`png"><br><br>`, 1).Split(`<a href="`, 0).Split("<a ", 0).ReplaceAll("<br>", "").ReplaceAll("</b>", "").String()

	if id {
		positive, err := body.Replace(adsPush, "").Split("Sisi Negatif Anda: ", 0).Value()
		if err != nil {
			return Reading{}, err
		}
		neg, err := negative.Value()
		if err != nil {
			return Reading{}, err
		}
		lines = append(lines, positive+". ", "Sisi Negatif Anda: "+neg+". ", advice)
	} else {
		positive, err := body.Split("Sisi Negatif Anda: ", 0).Value()
		if err != nil {
			return Reading{}, err
		}
		neg, err := negative.Value()
		if err != nil {
			return Reading{}, err
		}
		trans, err := r.translate(ctx, positive+". "+"Sisi Negatif Anda: "+neg+". "+advice, "id", "en")
		if err != nil {
			return Reading{}, err
		}
		lines = append(lines, trans...)
	}
	return Reading{Lines: lines, Love: loveScore(image)}, nil
}

// loveScore ports parseInt(image.split('https://www.primbon.com/ramalan_kecocokan_cinta')[1].split('.png')[0]).
// The meter is hidden for the generic kecocokan_jodoh.jpg image.
func loveScore(image string) int {
	if image == "" || image == "https://www.primbon.com/image/kecocokan_jodoh.jpg" {
		return 0
	}
	s, err := external.Str(image).Split("https://www.primbon.com/ramalan_kecocokan_cinta", 1).Split(".png", 0).Value()
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0
	}
	return min(n, 5)
}

// case 4: Tanggal Jadi.
func (r *Readings) primbonImportantDate(ctx context.Context, in PrimbonInput, lang string) (Reading, error) {
	d := in.Date
	path := fmt.Sprintf("/tanggal_jadian_pernikahan.php?tgl=%d&bln=%d&thn=%d&proses=+Submit%%21+", d.Day(), int(d.Month()), d.Year())
	doc, err := r.primbonDoc(ctx, true, path, nil)
	if err != nil {
		return Reading{}, err
	}
	text := external.Str(doc.Find("#body").Text()).Replace("Karakteristik:", "\nKarakteristik:").
		Replace("Hubungan", "\nHubungan").ReplaceRe(external.ReHitungKembali, "")
	// The source took .split(ads)[1]. primbon.com no longer puts the AdSense
	// snippet before the content (checked 2026-10-01), which made the source
	// throw; fall back to the whole text (deviation D-08).
	if after := text.Split(adsPush, 1); after.Err() == nil {
		text = after
	}
	content, err := text.TrimStart().ReplaceReAll(external.ReBlankLines, ".. ").SplitAll(". ")
	if err != nil {
		return Reading{}, err
	}
	if lang == "EN" {
		lines, err := r.translate(ctx, strings.Join(content, "\n"), "id", "en")
		return Reading{Lines: lines}, err
	}
	return Reading{Lines: content}, nil
}

// case 5: Ramalan Jodoh.
func (r *Readings) primbonSoulmate(ctx context.Context, in PrimbonInput, lang string) (Reading, error) {
	b, p := in.Birthday, in.Date
	form := url.Values{
		"nama1":  {in.Name},
		"tgl1":   {strconv.Itoa(b.Day())},
		"bln1":   {strconv.Itoa(int(b.Month()))},
		"thn1":   {strconv.Itoa(b.Year())},
		"nama2":  {in.Partner},
		"tgl2":   {strconv.Itoa(p.Day())},
		"bln2":   {strconv.Itoa(int(p.Month()))},
		"thn2":   {strconv.Itoa(p.Year())},
		"submit": {" RAMALAN JODOH &#62;&#62; "},
	}
	doc, err := r.primbonDoc(ctx, false, "/ramalan_jodoh.php", form)
	if err != nil {
		return Reading{}, err
	}
	s := external.Str(doc.Find("#body").Text()).ReplaceReAll(external.ReBlankLines, "").
		Replace(in.Name, "\n"+in.Name).ReplaceReAll(reTglLahir, "\nTanggal Lahir:").
		Replace(in.Partner, "\n"+in.Partner).Replace("Dibawah", "\n\nDibawah")
	for _, re := range reNumbered {
		s = s.ReplaceReAll(re.re, re.repl)
	}
	content, err := s.ReplaceRe(reStar, "\n\n*").ReplaceRe(external.ReHitungKembali, "").
		Replace("Konsultasi Hari Baik Akad Nikah >>>", "").Replace(adsPush, "").
		Replace("RAMALAN JODOH", "").TrimStart().ReplaceAll(".[", "\n [").
		Replace("...", ".").Replace("..", ".").Value()
	if err != nil {
		return Reading{}, err
	}
	lines, err := r.translateIDToEN(ctx, lang, content)
	return Reading{Lines: lines}, err
}

// case 6: Rejeki Weton, with the upstream fortune chart image.
func (r *Readings) primbonWetonFortune(ctx context.Context, in PrimbonInput, lang string) (Reading, error) {
	b := in.Birthday
	form := url.Values{
		"tgl":    {strconv.Itoa(b.Day())},
		"bln":    {strconv.Itoa(int(b.Month()))},
		"thn":    {strconv.Itoa(b.Year())},
		"submit": {" Submit! "},
	}
	doc, err := r.primbonDoc(ctx, false, "/rejeki_hoki_weton.php", form)
	if err != nil {
		return Reading{}, err
	}
	src, _ := doc.Find("#body > span > img").Attr("src")
	content := external.Str(doc.Find("#body").Text()).ReplaceReAll(external.ReBlankLines, "").
		Replace("Hari Lahir:", "\nHari Lahir:").Replace(adsPush, "").
		Replace("Seseorang", "\nSeseorang").Replace("Fluktuasi", "\n\nFluktuasi").
		Replace("Hover\n", "").ReplaceRe(external.ReHitungKembali, "").String()
	lines, err := r.translateIDToEN(ctx, lang, content)
	return Reading{Lines: lines, ImageURL: "https://www.primbon.com/" + src}, err
}

// case 7: Kecocokan Nama.
func (r *Readings) primbonNameCompatibility(ctx context.Context, in PrimbonInput, lang string) (Reading, error) {
	b := in.Birthday
	form := url.Values{
		"nama":  {in.Name},
		"tgl":   {strconv.Itoa(b.Day())},
		"bln":   {strconv.Itoa(int(b.Month()))},
		"thn":   {strconv.Itoa(b.Year())},
		"kirim": {" Submit! "},
	}
	doc, err := r.primbonDoc(ctx, false, "/kecocokan_nama.php", form)
	if err != nil {
		return Reading{}, err
	}
	content := external.Str(doc.Find("#body").Text()).ReplaceReAll(external.ReBlankLines, "").
		Replace("Tgl. Lahir:", "\nTanggal Lahir:").ReplaceRe(external.ReHitungKembali, "").
		Replace(adsPush, "").String()
	lines, err := r.translateIDToEN(ctx, lang, content)
	return Reading{Lines: lines}, err
}

// case 8: Hari Baik. Like the source app this uses the birthday, not the
// visible "Tanggal" field (open question Q-03 in docs/migration/behavior-map.md).
func (r *Readings) primbonGoodDay(ctx context.Context, in PrimbonInput, lang string) (Reading, error) {
	b := in.Birthday
	form := url.Values{
		"tgl":    {strconv.Itoa(b.Day())},
		"bln":    {strconv.Itoa(int(b.Month()))},
		"thn":    {strconv.Itoa(b.Year())},
		"submit": {" Submit! "},
	}
	doc, err := r.primbonDoc(ctx, false, "/petung_hari_baik.php", form)
	if err != nil {
		return Reading{}, err
	}
	year := strconv.Itoa(b.Year())
	content := external.Str(doc.Find("#body").Text()).ReplaceReAll(external.ReBlankLines, "").
		Replace("Kamarokam", "Kamarokam\n").Replace(year, year+"\n").Replace("Tgl.", "Tanggal").
		ReplaceRe(external.ReHitungKembali, "").Replace(adsPush, "").String()
	lines, err := r.translateIDToEN(ctx, lang, content)
	return Reading{Lines: lines}, err
}

// case 9: Hari Larangan. Uses the birthday, like case 8.
func (r *Readings) primbonProhibitedDay(ctx context.Context, in PrimbonInput, lang string) (Reading, error) {
	b := in.Birthday
	form := url.Values{
		"tgl":   {strconv.Itoa(b.Day())},
		"bln":   {strconv.Itoa(int(b.Month()))},
		"thn":   {strconv.Itoa(b.Year())},
		"kirim": {" Submit! "},
	}
	doc, err := r.primbonDoc(ctx, false, "/hari_sangar_taliwangke.php", form)
	if err != nil {
		return Reading{}, err
	}
	year := strconv.Itoa(b.Year())
	body := external.Str(doc.Find("#body").Text()).ReplaceReAll(external.ReBlankLines, "").
		Replace("Watak", "\nWatak").Replace("Kamarokam", "Kamarokam\n").Replace(year, year+"\n").
		ReplaceRe(external.ReHitungKembali, "").Replace(adsPush, "").
		Replace("Termasuk hari", "\nTermasuk hari").Replace("\n =", " =")
	head, err := body.Split("Untuk mengetahui watak hari, masukkan:", 0).Value()
	if err != nil {
		return Reading{}, err
	}
	ref, err := body.Replace("Referensi", "Referensi refrensii").Split(" refrensii", 1).
		Split("Catatan:", 0).Replace(": Kitab Primbon Jawa", "*Referensi: Kitab Primbon Jawa").Value()
	if err != nil {
		return Reading{}, err
	}
	lines, err := r.translateIDToEN(ctx, lang, head+"\n"+ref)
	return Reading{Lines: lines}, err
}

// Regular expressions used by the Ramalan Jodoh port.
var (
	reTglLahir = regexp.MustCompile(`Tgl\. Lahir:`)
	reStar     = regexp.MustCompile(`\*`)
	// reNumbered ports .replace(/1\. /g, "\n1. ") ... .replace(/10\. /g, "\n10. "), in source order.
	reNumbered = func() []numberedRule {
		var out []numberedRule
		for i := 1; i <= 10; i++ {
			n := strconv.Itoa(i)
			out = append(out, numberedRule{regexp.MustCompile(n + `\. `), "\n" + n + ". "})
		}
		return out
	}()
)

type numberedRule struct {
	re   *regexp.Regexp
	repl string
}

func pick(id bool, idText, enText string) string {
	if id {
		return idText
	}
	return enText
}
