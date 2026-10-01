// Package auth implements sign-in with Google (OpenID Connect, authorization
// code + PKCE) and opaque server-side sessions. See ADR 0004.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/vikyarifian/farahdin-go/internal/domain"
	"github.com/vikyarifian/farahdin-go/internal/repository"
)

// SessionCookie is the name of the session cookie.
const SessionCookie = "farahdin_session"

// SessionStore persists sessions.
type SessionStore interface {
	Create(ctx context.Context, tokenHash string, userID int64, now, expires time.Time) error
	UserID(ctx context.Context, tokenHash string, now time.Time) (int64, error)
	Delete(ctx context.Context, tokenHash string) error
}

// UserLoader loads a user by id.
type UserLoader interface {
	ByID(ctx context.Context, id int64) (*domain.User, error)
}

// Sessions issues and resolves session cookies.
type Sessions struct {
	Store  SessionStore
	Users  UserLoader
	TTL    time.Duration
	Secure bool // set the Secure cookie attribute (true outside local HTTP development)
	Now    func() time.Time
}

// ErrNoSession means the request carries no valid session.
var ErrNoSession = errors.New("no session")

// Start creates a session for userID and sets the cookie.
func (s *Sessions) Start(ctx context.Context, w http.ResponseWriter, userID int64) error {
	token, err := randomToken(32)
	if err != nil {
		return err
	}
	now := s.Now()
	expires := now.Add(s.TTL)
	if err := s.Store.Create(ctx, hashToken(token), userID, now, expires); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		MaxAge:   int(s.TTL.Seconds()),
		HttpOnly: true,
		Secure:   s.Secure,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// User resolves the signed-in user of r.
func (s *Sessions) User(r *http.Request) (*domain.User, error) {
	c, err := r.Cookie(SessionCookie)
	if err != nil || c.Value == "" {
		return nil, ErrNoSession
	}
	id, err := s.Store.UserID(r.Context(), hashToken(c.Value), s.Now())
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNoSession
	}
	if err != nil {
		return nil, err
	}
	u, err := s.Users.ByID(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNoSession
	}
	return u, err
}

// End deletes the session of r and clears the cookie.
func (s *Sessions) End(w http.ResponseWriter, r *http.Request) error {
	if c, err := r.Cookie(SessionCookie); err == nil && c.Value != "" {
		if err := s.Store.Delete(r.Context(), hashToken(c.Value)); err != nil {
			return err
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.Secure,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
