package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/vikyarifian/farahdin-go/internal/domain"
	"github.com/vikyarifian/farahdin-go/internal/external"
)

// MatrixResult is everything the Matrix Destiny result sheet shows.
type MatrixResult struct {
	Name     string // title-cased
	Birthday time.Time
	Matrix   domain.Matrix
	Reading  Reading
}

// MatrixDestiny ports matrix-destiny.tsx generate(): validate, compute the
// chart locally, then fetch the text reading. A failed reading does not
// block the chart (the source swallowed that error too).
func (r *Readings) MatrixDestiny(ctx context.Context, name string, birthday time.Time, lang string) (MatrixResult, error) {
	if msg := domain.ValidateMatrixInput(birthday, name, r.Now(), lang); msg != "" {
		return MatrixResult{}, &ErrValidation{msg}
	}
	res := MatrixResult{
		Name:     domain.TitleCase(name),
		Birthday: birthday,
		Matrix:   domain.ComputeMatrix(birthday),
	}
	ctx, st := withTranslateStatus(ctx)
	reading, err := r.matrixReading(ctx, name, birthday, lang)
	noteTranslation(st, lang, &reading)
	if err != nil {
		slog.WarnContext(ctx, "matrix reading unavailable", "err", err)
		res.Reading.Notice = pick(lang == "ID",
			"Pembacaan matriks sedang tidak tersedia.", "The matrix reading is unavailable right now.")
		return res, nil
	}
	res.Reading = reading
	return res, nil
}

var matrixSections = []struct{ title, key string }{
	{"1. Soul comfort", "infg1"},
	{"2. First impression", "infg2"},
	{"3. Karmic tasks & previous life", "infg3"},
	{"4. Relationship", "infg4"},
	{"5. Money", "infg5"},
	{"6. Talents", "infg6"},
	{"7. Purpose", "infg8"},
}

func (r *Readings) matrixReading(ctx context.Context, name string, birthday time.Time, lang string) (Reading, error) {
	page, err := r.Fetch.Get(ctx, r.Src.MatrixDestiny+"/")
	if err != nil {
		return Reading{}, err
	}
	raw, err := external.Str(page).Split("var ajax_var = ", 1).Split(";", 0).Value()
	if err != nil {
		return Reading{}, err
	}
	var ajax struct {
		Nonce string `json:"nonce"`
	}
	if err := json.Unmarshal([]byte(raw), &ajax); err != nil {
		return Reading{}, fmt.Errorf("matrix nonce: %w", err)
	}

	body, err := r.Fetch.PostForm(ctx, r.Src.MatrixDestiny+"/wp-admin/admin-ajax.php", url.Values{
		"yourname":     {name},
		"yourbirthday": {birthday.Format("02/01/2006")},
		"action":       {"matrix_calc"},
		"postid":       {"63"},
		"chartid":      {"0"},
		"nonce_code":   {ajax.Nonce},
	})
	if err != nil {
		return Reading{}, err
	}
	if strings.TrimSpace(body) == "-1" {
		return Reading{}, fmt.Errorf("matrix reading: rejected nonce")
	}
	var fields map[string]any
	if err := json.Unmarshal([]byte(body), &fields); err != nil {
		return Reading{}, fmt.Errorf("matrix reading: decode: %w", err)
	}

	var content strings.Builder
	for i, s := range matrixSections {
		v, ok := fields[s.key].(string)
		if !ok {
			return Reading{}, fmt.Errorf("matrix reading: missing %s", s.key)
		}
		if i > 0 {
			content.WriteString("<br>")
		}
		content.WriteString(s.title + "<br>" + strings.ReplaceAll(strings.ReplaceAll(v, "<br />", ""), "<br><br>", "<br>"))
	}
	text := content.String()
	if lang != "ID" {
		return Reading{Lines: strings.Split(text, "<br>")}, nil
	}
	trans, err := r.translate(ctx, strings.ReplaceAll(text, "<br>", ". <br>"), "en", "id")
	if err != nil {
		return Reading{}, err
	}
	if translationFailed(ctx) {
		return Reading{Lines: strings.Split(text, "<br>")}, nil
	}
	joined := strings.ReplaceAll(strings.ReplaceAll(strings.Join(trans, "<br>"), ". <br>", ". "), "<br><br>", "<br>")
	return Reading{Lines: strings.Split(joined, "<br>")}, nil
}
