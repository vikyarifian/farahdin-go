// Package pages holds the full-page and fragment templates.
package pages

import (
	"github.com/vikyarifian/farahdin-go/internal/domain"
	"github.com/vikyarifian/farahdin-go/internal/service"
)

// LoginView is the sign-in screen.
type LoginView struct {
	GoogleEnabled bool
	DevLogin      bool
	Error         string
}

// HomeView is the home tab.
type HomeView struct {
	FirstName string // "Guest" when unknown
	Hour      int    // server-side hour in the app time zone; app.js refines it
}

// ProfileView is the profile tab.
type ProfileView struct {
	User *domain.User
	Age  string // "" when the birthday is unknown
}

// EditProfileView is the Edit Profile form.
type EditProfileView struct {
	Email, Name, Birthday, Zodiac, Birthplace, Gender string
	Today                                             string
	Errors                                            map[string]string
	Saved                                             bool
}

// ResultView is the outcome of a reading form submission.
type ResultView struct {
	Title, Subtitle, CloseHref string
	Reading                    service.Reading
	Error                      string
	Show                       bool
}

// PrimbonView is a Primbon topic form.
type PrimbonView struct {
	Topic                                domain.Topic
	Fields                               service.PrimbonFields
	Name, Dream, Partner, Birthday, Date string
	Today                                string
	Result                               ResultView
}

// HoroscopeView is a Horoscope topic form.
type HoroscopeView struct {
	Topic                 domain.Topic
	Birthday, PartnerDate string
	Sign, PartnerSign     string
	Today                 string
	Result                ResultView
}

// ClairvoyanceView is a Clairvoyance topic form.
type ClairvoyanceView struct {
	Topic  domain.Topic
	Name   string
	Result ResultView
}

// TarotView is a Tarot card spread.
type TarotView struct {
	Topic domain.Topic
	Cards []int
	Error string
	Sheet *TarotSheet
}

// TarotSheet is the result sheet after picking (and possibly reading) cards.
type TarotSheet struct {
	Topic   domain.Topic
	Card    int // single-card topics
	You     int // True Love
	Partner int // True Love
	Read    bool
	Reading service.Reading
	Error   string
}

// MatrixView is the Matrix Destiny form.
type MatrixView struct {
	Topic    domain.Topic
	Name     string
	Birthday string
	Today    string
	Error    string
	Result   *service.MatrixResult
}
