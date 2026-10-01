package domain

// Topic is one selectable reading inside a feature. Keys, labels and icons
// are copied from the `topics` arrays in the source app's feature screens.
type Topic struct {
	Key  int
	ID   string // Indonesian label
	EN   string // English label
	Icon string // Ionicons name, rendered from /static/icons/sprite.svg
	// Tarot only: how many cards are dealt (PickCard) out of a deck of LengthCard.
	PickCard   int
	LengthCard int
}

// Label returns the topic label for the language ("ID" or "EN").
func (t Topic) Label(lang string) string {
	if lang == "ID" {
		return t.ID
	}
	return t.EN
}

// PrimbonTopics ports app/pages/primbon.tsx.
var PrimbonTopics = []Topic{
	{Key: 1, ID: "Arti Nama", EN: "Name Meaning", Icon: "finger-print"},
	{Key: 2, ID: "Tafsir Mimpi", EN: "Dream Interpretation", Icon: "bed-outline"},
	{Key: 3, ID: "Jodoh", EN: "Match", Icon: "heart-outline"},
	{Key: 4, ID: "Tanggal Jadi", EN: "Important Date", Icon: "calendar-number-outline"},
	{Key: 5, ID: "Ramalan Jodoh", EN: "Soulmate Prediction", Icon: "male-female-outline"},
	{Key: 6, ID: "Rejeki Weton", EN: "Weton Fortune", Icon: "dice-outline"},
	{Key: 7, ID: "Kecocokan Nama", EN: "Name Compatibility", Icon: "body-outline"},
	{Key: 8, ID: "Hari Baik", EN: "Good Day", Icon: "sunny-outline"},
	{Key: 9, ID: "Hari Larangan", EN: "Prohibited Day", Icon: "thunderstorm-outline"},
}

// HoroscopeTopics ports app/pages/horoscope.tsx.
var HoroscopeTopics = []Topic{
	{Key: 1, ID: "Kemarin", EN: "Yesterday", Icon: "arrow-back"},
	{Key: 2, ID: "Hari ini", EN: "Today", Icon: "today-outline"},
	{Key: 3, ID: "Besok", EN: "Tomorrow", Icon: "arrow-forward"},
	{Key: 4, ID: "Minggu ini", EN: "Weekly", Icon: "calendar-outline"},
	{Key: 5, ID: "Bulan ini", EN: "Monthly", Icon: "calendar-number-outline"},
	{Key: 6, ID: "Kecocokan", EN: "Match", Icon: "heart-outline"},
}

// TarotTopics ports app/pages/tarot.tsx.
var TarotTopics = []Topic{
	{Key: 1, ID: "Cinta", EN: "Love", Icon: "heart-outline", PickCard: 22, LengthCard: 22},
	{Key: 2, ID: "Cinta Sejati", EN: "True Love", Icon: "male-female-outline", PickCard: 2, LengthCard: 22},
	{Key: 3, ID: "Malaikat", EN: "Angel", Icon: "eye-outline", PickCard: 22, LengthCard: 22},
	{Key: 4, ID: "Kehidupan Lampau", EN: "Past Lives", Icon: "accessibility-outline", PickCard: 22, LengthCard: 30},
}

// ClairvoyanceTopics ports app/pages/clairvoyance.tsx.
var ClairvoyanceTopics = []Topic{
	{Key: 1, ID: "Umum", EN: "General", Icon: "sparkles-outline"},
	{Key: 2, ID: "Cinta", EN: "Love", Icon: "heart-outline"},
	{Key: 3, ID: "Karir", EN: "Career", Icon: "briefcase-outline"},
	{Key: 4, ID: "Kesehatan", EN: "Health", Icon: "medkit-outline"},
}

// MatrixTopics ports app/pages/matrix-destiny.tsx. Only Personal is
// reachable in the source UI (topic starts at 1 and the picker is hidden).
var MatrixTopics = []Topic{
	{Key: 1, ID: "Pribadi", EN: "Personal", Icon: "sparkles-outline"},
}

// FindTopic returns the topic with key k.
func FindTopic(topics []Topic, k int) (Topic, bool) {
	for _, t := range topics {
		if t.Key == k {
			return t, true
		}
	}
	return Topic{}, false
}

// TopicIn reports whether k is one of keys (ports [..].includes(topic)).
func TopicIn(k int, keys ...int) bool {
	for _, x := range keys {
		if x == k {
			return true
		}
	}
	return false
}
