//go:build live

// Live smoke test against the real upstream sites (risk R-01). Not part of
// the default test run:
//
//	go test -tags live -run TestLive -v ./internal/service/
package service

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/vikyarifian/farahdin-go/internal/external"
)

func liveReadings() *Readings {
	client := external.NewClient(30 * time.Second)
	src := external.DefaultSources()
	return &Readings{Fetch: client, Translate: external.NewTranslator(client), Src: src, Now: time.Now}
}

func requireLines(t *testing.T, r Reading, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	lines := CleanLines(r.Lines)
	if len(lines) == 0 || len(strings.Join(lines, " ")) < 40 {
		t.Fatalf("reading too short: %q", lines)
	}
	t.Logf("%d lines: %.120s", len(lines), strings.Join(lines, " / "))
}

func TestLivePrimbon(t *testing.T) {
	r := liveReadings()
	b := time.Date(1990, 8, 17, 0, 0, 0, 0, time.UTC)
	d := time.Date(1992, 3, 3, 0, 0, 0, 0, time.UTC)
	in := PrimbonInput{Name: "Viky", Dream: "snake", Partner: "Farah", Birthday: b, Date: d}
	for topic := 1; topic <= 9; topic++ {
		t.Run(fmt.Sprintf("topic%d", topic), func(t *testing.T) {
			res, err := r.Primbon(context.Background(), topic, in, "ID")
			if topic == 2 && err == nil && len(res.Lines) == 1 {
				t.Logf("dream: %s", res.Lines[0]) // "not found" is a valid answer
				return
			}
			requireLines(t, res, err)
		})
	}
}

func TestLiveHoroscope(t *testing.T) {
	r := liveReadings()
	in := HoroscopeInput{Birthday: time.Date(1990, 8, 17, 0, 0, 0, 0, time.UTC), PartnerDate: time.Date(1992, 3, 3, 0, 0, 0, 0, time.UTC)}
	for topic := 1; topic <= 6; topic++ {
		res, err := r.Horoscope(context.Background(), topic, in, "EN")
		requireLines(t, res, err)
	}
}

func TestLiveTarotAndClairvoyance(t *testing.T) {
	r := liveReadings()
	for _, topic := range []int{1, 3, 4} {
		res, err := r.Tarot(context.Background(), topic, 7, "EN")
		requireLines(t, res, err)
	}
	res, err := r.TarotTrueLove(context.Background(), 3, 9, "EN")
	requireLines(t, res, err)
	for topic := 1; topic <= 4; topic++ {
		res, err := r.Clairvoyance(context.Background(), topic, "Viky", "EN")
		requireLines(t, res, err)
	}
}

func TestLiveMatrixAndTranslate(t *testing.T) {
	r := liveReadings()
	res, err := r.MatrixDestiny(context.Background(), "Viky Arifian", time.Date(1990, 8, 17, 0, 0, 0, 0, time.UTC), "ID")
	if err != nil {
		t.Fatal(err)
	}
	if res.Reading.Notice != "" {
		t.Fatalf("matrix reading failed: %s", res.Reading.Notice)
	}
	requireLines(t, res.Reading, nil)
}
