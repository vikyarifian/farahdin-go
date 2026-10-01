package domain

import "time"

// DateLayout is the storage format the source app used for birthdays.
const DateLayout = "2006-01-02"

// MinBirthDate mirrors the date pickers' minimumDate={new Date(1900, 0, 1)}.
var MinBirthDate = time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)

// ParseDate parses "YYYY-MM-DD"; legacy values like "NaN-NaN-NaN" fail.
func ParseDate(s string) (time.Time, bool) {
	t, err := time.Parse(DateLayout, s)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// FormatDate renders a date as "YYYY-MM-DD".
func FormatDate(t time.Time) string {
	return t.Format(DateLayout)
}

// DateOnly drops the clock part, keeping the calendar day in t's location.
func DateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// CalcAge ports utils/Utils.ts calcAge().
func CalcAge(birth, today time.Time) int {
	age := today.Year() - birth.Year()
	m := int(today.Month()) - int(birth.Month())
	if m < 0 || (m == 0 && today.Day() < birth.Day()) {
		age--
	}
	return age
}

// DisplayDate ports Date.toDateString(), e.g. "Mon Jan 02 2006".
func DisplayDate(t time.Time) string {
	return t.Format("Mon Jan 02 2006")
}
