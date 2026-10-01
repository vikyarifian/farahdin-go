package domain

import (
	"strings"
	"time"
)

// User mirrors the Convex `users` table plus the identity fields that the
// Google sign-in replacement needs (see docs/adr/0004-auth-google-oidc.md).
type User struct {
	ID         int64
	Username   string
	Fullname   string
	Email      string
	Birthday   string // "YYYY-MM-DD"; may hold legacy values such as "NaN-NaN-NaN"
	Birthplace string
	Gender     string
	Zodiac     string
	ImageURL   string
	GoogleSub  string
	ClerkID    string
}

// FirstName ports fullname.split(' ')[0].
func (u *User) FirstName() string {
	if u == nil {
		return ""
	}
	return strings.Split(u.Fullname, " ")[0]
}

// BirthdayOr returns the parsed birthday or fallback when it is missing or
// invalid, mirroring the source app's `new Date()` fallback.
func (u *User) BirthdayOr(fallback time.Time) time.Time {
	if u == nil {
		return fallback
	}
	if t, ok := ParseDate(u.Birthday); ok {
		return t
	}
	return fallback
}

// Identity is what an identity provider tells us about a signed-in person.
type Identity struct {
	Subject   string
	Email     string
	FirstName string
	LastName  string
	ImageURL  string
}

// Fullname ports `${first_name || ""} ${last_name || ""}`.trim().
func (i Identity) Fullname() string {
	return strings.TrimSpace(i.FirstName + " " + i.LastName)
}

// Username ports email.split("@")[0].
func (i Identity) Username() string {
	return strings.Split(i.Email, "@")[0]
}

// ProfileUpdate is the editable part of a profile (users.updateUser args
// minus the fields the server now owns: username, email, zodiac).
type ProfileUpdate struct {
	Fullname   string
	Birthday   time.Time
	Birthplace string
	Gender     string
}
