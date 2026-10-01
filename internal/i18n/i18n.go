// Package i18n carries the ID/EN language choice that the source app kept
// in AsyncStorage under the key "lang" (default "EN").
package i18n

import (
	"context"
	"net/http"
	"time"
)

// Supported languages.
const (
	ID = "ID"
	EN = "EN"
)

// Cookie is the cookie that replaces AsyncStorage "lang".
const Cookie = "lang"

type ctxKey struct{}

// Normalize maps any value to a supported language, defaulting to EN.
func Normalize(v string) string {
	if v == ID {
		return ID
	}
	return EN
}

// FromRequest reads the language cookie.
func FromRequest(r *http.Request) string {
	if c, err := r.Cookie(Cookie); err == nil {
		return Normalize(c.Value)
	}
	return EN
}

// SetCookie stores the language for a year.
func SetCookie(w http.ResponseWriter, lang string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     Cookie,
		Value:    Normalize(lang),
		Path:     "/",
		MaxAge:   int((365 * 24 * time.Hour).Seconds()),
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// WithLang stores the language in ctx.
func WithLang(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, ctxKey{}, Normalize(lang))
}

// Lang returns the language stored in ctx (EN when absent).
func Lang(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKey{}).(string); ok {
		return v
	}
	return EN
}

// IsID reports whether ctx is in Indonesian.
func IsID(ctx context.Context) bool { return Lang(ctx) == ID }

// T picks the Indonesian or English text for ctx, like the source app's
// inline `lang === 'ID' ? '…' : '…'` expressions.
func T(ctx context.Context, idText, enText string) string {
	if IsID(ctx) {
		return idText
	}
	return enText
}
