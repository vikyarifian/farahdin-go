package domain

import "strings"

// ZodiacSigns is the order horoscope.com uses for its numeric sign parameter.
var ZodiacSigns = []string{
	"aries", "taurus", "gemini", "cancer", "leo", "virgo",
	"libra", "scorpio", "sagittarius", "capricorn", "aquarius", "pisces",
}

// Zodiac ports utils/Zodiac.ts zodiac(): it compares the "MM-DD" part of a
// "YYYY-MM-DD" string lexically, exactly like the source app.
func Zodiac(date string) string {
	parts := strings.Split(date, "-")
	if len(parts) < 3 {
		return ""
	}
	md := strings.Join(parts[1:], "-")
	between := func(lo, hi string) bool { return md >= lo && md <= hi }
	switch {
	case between("03-21", "04-19"):
		return "Aries"
	case between("04-20", "05-20"):
		return "Taurus"
	case between("05-21", "06-20"):
		return "Gemini"
	case between("06-21", "07-22"):
		return "Cancer"
	case between("07-23", "08-22"):
		return "Leo"
	case between("08-23", "09-22"):
		return "Virgo"
	case between("09-23", "10-22"):
		return "Libra"
	case between("10-23", "11-21"):
		return "Scorpio"
	case between("11-22", "12-21"):
		return "Sagittarius"
	case between("12-22", "12-31") || between("01-01", "01-19"):
		return "Capricorn"
	case between("01-20", "02-18"):
		return "Aquarius"
	case between("02-19", "03-20"):
		return "Pisces"
	}
	return ""
}

// ZodiacDate ports zodiacDate(): the date range label for a sign.
func ZodiacDate(sign string) string {
	switch strings.ToLower(sign) {
	case "aries":
		return "Mar 21 - Apr 19"
	case "taurus":
		return "Apr 20 - May 20"
	case "gemini":
		return "May 21 - Jun 20"
	case "cancer":
		return "Jun 21 - Jul 22"
	case "leo":
		return "Jul 23 - Aug 22"
	case "virgo":
		return "Aug 23 - Sep 22"
	case "libra":
		return "Sep 23 - Oct 22"
	case "scorpio":
		return "Oct 23 - Nov 21"
	case "sagittarius":
		return "Nov 22 - Dec 21"
	case "capricorn":
		return "Dec 22 - Jan 19"
	case "aquarius":
		return "Jan 20 - Feb 18"
	case "pisces":
		return "Feb 19 - Mar 20"
	}
	return ""
}

// ZodiacNumber ports zodiacNumber(): 1-based index of the sign, 0 if unknown.
func ZodiacNumber(sign string) int {
	s := strings.ToLower(sign)
	for i, z := range ZodiacSigns {
		if z == s {
			return i + 1
		}
	}
	return 0
}
