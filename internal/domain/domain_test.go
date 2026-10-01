package domain

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

func TestZodiac(t *testing.T) {
	tests := []struct {
		date, want string
	}{
		{"1990-03-21", "Aries"},
		{"1990-04-19", "Aries"},
		{"1990-04-20", "Taurus"},
		{"1990-06-21", "Cancer"},
		{"1990-07-23", "Leo"},
		{"1990-12-21", "Sagittarius"},
		{"1990-12-22", "Capricorn"},
		{"1990-01-19", "Capricorn"},
		{"1990-01-20", "Aquarius"},
		{"1990-02-19", "Pisces"},
		{"1990-03-20", "Pisces"},
		{"NaN-NaN-NaN", ""},
		{"", ""},
	}
	for _, tt := range tests {
		if got := Zodiac(tt.date); got != tt.want {
			t.Errorf("Zodiac(%q) = %q, want %q", tt.date, got, tt.want)
		}
	}
}

func TestZodiacNumberAndDate(t *testing.T) {
	if got := ZodiacNumber("Aries"); got != 1 {
		t.Errorf("ZodiacNumber(Aries) = %d, want 1", got)
	}
	if got := ZodiacNumber("pisces"); got != 12 {
		t.Errorf("ZodiacNumber(pisces) = %d, want 12", got)
	}
	if got := ZodiacNumber(""); got != 0 {
		t.Errorf("ZodiacNumber(\"\") = %d, want 0", got)
	}
	if got := ZodiacDate("Leo"); got != "Jul 23 - Aug 22" {
		t.Errorf("ZodiacDate(Leo) = %q", got)
	}
}

func TestCalcAge(t *testing.T) {
	today := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		birth string
		want  int
	}{
		{"1990-10-01", 36},
		{"1990-10-02", 35},
		{"1990-09-30", 36},
		{"2000-02-29", 26},
	}
	for _, tt := range tests {
		b, _ := ParseDate(tt.birth)
		if got := CalcAge(b, today); got != tt.want {
			t.Errorf("CalcAge(%s) = %d, want %d", tt.birth, got, tt.want)
		}
	}
}

func TestReduceNumber(t *testing.T) {
	tests := map[int]int{1: 1, 22: 22, 23: 5, 29: 11, 31: 4, 88: 16}
	for in, want := range tests {
		if got := ReduceNumber(in); got != want {
			t.Errorf("ReduceNumber(%d) = %d, want %d", in, got, want)
		}
	}
	if got := CalculateYear(1990); got != 19 {
		t.Errorf("CalculateYear(1990) = %d, want 19", got)
	}
	if got := CalculateYear(1999); got != 10 { // 28 -> 10
		t.Errorf("CalculateYear(1999) = %d, want 10", got)
	}
}

// TestMatrixMatchesSource compares the Go port with values produced by
// evaluating the React Native expressions in Node (testdata/matrix_reference.json).
func TestMatrixMatchesSource(t *testing.T) {
	raw, err := os.ReadFile("testdata/matrix_reference.json")
	if err != nil {
		t.Fatal(err)
	}
	var refs []struct {
		Date     string         `json:"date"`
		Points   map[string]int `json:"points"`
		Years    map[string]int `json:"years"`
		Purposes map[string]int `json:"purposes"`
		Chakras  map[string]int `json:"chakras"`
	}
	if err := json.Unmarshal(raw, &refs); err != nil {
		t.Fatal(err)
	}
	for _, ref := range refs {
		d, ok := ParseDate(ref.Date)
		if !ok {
			t.Fatalf("bad date %q", ref.Date)
		}
		m := ComputeMatrix(d)
		compare(t, ref.Date+" points", m.Points, ref.Points)
		compare(t, ref.Date+" years", m.Years, ref.Years)
		compare(t, ref.Date+" purposes", m.Purposes, ref.Purposes)

		got := map[string]int{}
		keys := map[string]string{"Sahastrara": "sah", "Ajna": "aj", "Vishuddha": "vish", "Anahata": "anah",
			"Manipura": "man", "Svadhishtana": "svad", "Muladhara": "mul"}
		for _, c := range m.Chakras {
			k := keys[c.Name]
			got[k+"physics"], got[k+"energy"], got[k+"emotions"] = c.Physics, c.Energy, c.Emotions
		}
		compare(t, ref.Date+" chakras", got, ref.Chakras)
	}
}

func compare(t *testing.T, label string, got, want map[string]int) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s: %d values, want %d", label, len(got), len(want))
	}
	for k, w := range want {
		if g, ok := got[k]; !ok || g != w {
			t.Errorf("%s[%s] = %d, want %d", label, k, g, w)
		}
	}
}

func TestValidateMatrixInput(t *testing.T) {
	now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	ok, _ := ParseDate("1990-01-01")
	if msg := ValidateMatrixInput(ok, "Fufu-Fafa", now, "EN"); msg != "" {
		t.Errorf("valid input rejected: %q", msg)
	}
	if msg := ValidateMatrixInput(ok, "", now, "EN"); !strings.Contains(msg, "one of the fields is empty") {
		t.Errorf("empty name: %q", msg)
	}
	future := now.AddDate(0, 0, 1)
	if msg := ValidateMatrixInput(future, "Ana", now, "ID"); !strings.Contains(msg, "masa depan") {
		t.Errorf("future date: %q", msg)
	}
	old, _ := ParseDate("1900-01-01")
	if msg := ValidateMatrixInput(old, "Ana", now, "EN"); !strings.Contains(msg, "so far in the past") {
		t.Errorf("old date: %q", msg)
	}
	if msg := ValidateMatrixInput(ok, "Ana123", now, "EN"); !strings.Contains(msg, "Name format is incorrect") {
		t.Errorf("bad name: %q", msg)
	}
}

func TestTitleCase(t *testing.T) {
	if got := TitleCase("fufu-fafa mage wita"); got != "Fufu-Fafa Mage Wita" {
		t.Errorf("TitleCase = %q", got)
	}
}

func TestIdentity(t *testing.T) {
	id := Identity{Email: "viky.arifian@example.com", FirstName: "Viky", LastName: ""}
	if id.Fullname() != "Viky" || id.Username() != "viky.arifian" {
		t.Errorf("got %q / %q", id.Fullname(), id.Username())
	}
}
