package domain

import (
	"regexp"
	"strings"
	"time"
)

// ReduceNumber ports reduceNumber from the source app: numbers above 22 are
// reduced once by summing their two digits. It intentionally does not loop.
func ReduceNumber(n int) int {
	if n > 22 {
		return n%10 + n/10
	}
	return n
}

// CalculateYear ports calculateYear: digit sum of the year, then reduced once.
func CalculateYear(year int) int {
	y := 0
	for year > 0 {
		y += year % 10
		year /= 10
	}
	return ReduceNumber(y)
}

// Chakra is one row of the chakra ("chart heart") table.
type Chakra struct {
	Number   int
	Name     string
	Color    string
	DarkText bool
	Physics  int
	Energy   int
	Emotions int
}

// Matrix holds every value rendered by the destiny matrix result.
type Matrix struct {
	Points   map[string]int
	Years    map[string]int
	Purposes map[string]int
	Chakras  []Chakra
	// Totals are the reduced column sums shown in the "Result" row.
	TotalPhysics  int
	TotalEnergy   int
	TotalEmotions int
}

// ComputeMatrix derives the full matrix from a birth date, exactly as the
// source app did from the day, month and year of the selected date.
func ComputeMatrix(birthday time.Time) Matrix {
	a := ReduceNumber(birthday.Day())
	b := int(birthday.Month())
	c := CalculateYear(birthday.Year())

	ch := matrixChakras(a, b, c)
	rows := []struct {
		n        int
		name     string
		key      string
		color    string
		darkText bool
	}{
		{7, "Sahastrara", "sah", "#b653f7", false},
		{6, "Ajna", "aj", "#3d54f5", false},
		{5, "Vishuddha", "vish", "#74e0f8", true},
		{4, "Anahata", "anah", "#b6fd57", true},
		{3, "Manipura", "man", "#fbe49d", true},
		{2, "Svadhishtana", "svad", "#ed7233", false},
		{1, "Muladhara", "mul", "#ea4631", false},
	}

	m := Matrix{
		Points:   matrixPoints(a, b, c),
		Years:    matrixYears(a, b, c),
		Purposes: matrixPurposes(a, b, c),
	}
	var sumP, sumEn, sumEm int
	for _, r := range rows {
		row := Chakra{
			Number:   r.n,
			Name:     r.name,
			Color:    r.color,
			DarkText: r.darkText,
			Physics:  ch[r.key+"physics"],
			Energy:   ch[r.key+"energy"],
			Emotions: ch[r.key+"emotions"],
		}
		sumP += row.Physics
		sumEn += row.Energy
		sumEm += row.Emotions
		m.Chakras = append(m.Chakras, row)
	}
	m.TotalPhysics = ReduceNumber(sumP)
	m.TotalEnergy = ReduceNumber(sumEn)
	m.TotalEmotions = ReduceNumber(sumEm)
	return m
}

// matrixNamePattern ports the source regex ^[а-яё\- ]*[a-z\- ]*$ (case-insensitive).
var matrixNamePattern = regexp.MustCompile(`(?i)^[а-яё\- ]*[a-z\- ]*$`)

// ValidateMatrixInput ports valide(): it returns the concatenated error
// message in the requested language, or "" when the input is acceptable.
func ValidateMatrixInput(birthday time.Time, name string, now time.Time, lang string) string {
	id := lang == "ID"
	t := func(idText, enText string) string {
		if id {
			return idText
		}
		return enText
	}
	var msg strings.Builder
	if birthday.IsZero() {
		msg.WriteString(t("Tanggal tidak valid", "Date is invalid"))
	}
	if name == "" || birthday.IsZero() {
		msg.WriteString(t("Tanggal tidak valid atau salah satu bidang kosong.", "Date is not valid or one of the fields is empty."))
	}
	if birthday.After(now) {
		msg.WriteString(t("Tanggal tidak bisa di masa depan.", "Date can't be in the future."))
	}
	if !birthday.IsZero() && now.Year()-birthday.Year() > 120 {
		msg.WriteString(t("Tanggal tidak bisa begitu jauh di masa lalu.", "Date can't be so far in the past."))
	}
	if !matrixNamePattern.MatchString(name) {
		msg.WriteString(t("Format nama salah, karakter yang diizinkan adalah huruf, tanda hubung dan spasi. Contoh: ",
			"Name format is incorrect, allowed characters are letters, dash and space. Example: "))
		msg.WriteString("Mulyono, Fufu-Fafa, Mage Wita.")
	}
	return msg.String()
}

var titleCasePattern = regexp.MustCompile(`^[a-zа-яё]|[\- ][a-zа-яё]`)

// TitleCase ports titleCase(): upper-cases the first letter and every letter
// following a dash or space.
func TitleCase(s string) string {
	return titleCasePattern.ReplaceAllStringFunc(s, strings.ToUpper)
}
