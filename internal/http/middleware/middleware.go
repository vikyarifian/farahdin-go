// Package middleware holds the HTTP middleware chain.
package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/vikyarifian/farahdin-go/internal/auth"
	"github.com/vikyarifian/farahdin-go/internal/domain"
	"github.com/vikyarifian/farahdin-go/internal/http/response"
	"github.com/vikyarifian/farahdin-go/internal/i18n"
)

// Chain applies middleware so that the first one listed runs first.
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// Logger logs one structured line per request (no query strings, no cookies).
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		if strings.HasPrefix(r.URL.Path, "/static/") && rec.status < 400 {
			return
		}
		slog.InfoContext(r.Context(), "request", "method", r.Method, "path", r.URL.Path,
			"status", rec.status, "duration_ms", time.Since(start).Milliseconds(), "htmx", response.IsHTMX(r))
	})
}

// Recover turns panics into a 500 without leaking details.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				if err, ok := v.(error); ok && errors.Is(err, http.ErrAbortHandler) {
					panic(v)
				}
				slog.ErrorContext(r.Context(), "panic", "value", v, "stack", string(debug.Stack()))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// csp allows only same-origin code. Images may come from primbon.com
// (Rejeki Weton chart) and Google (profile photos).
const csp = "default-src 'self'; script-src 'self'; style-src 'self'; " +
	"img-src 'self' data: https://www.primbon.com https://primbon.com https://*.googleusercontent.com; " +
	"font-src 'self'; connect-src 'self'; manifest-src 'self'; worker-src 'self'; " +
	"frame-ancestors 'none'; base-uri 'self'; form-action 'self'; object-src 'none'"

// SecurityHeaders sets the baseline response headers.
func SecurityHeaders(hsts bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("Content-Security-Policy", csp)
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
			if hsts {
				h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	}
}

// CrossOrigin rejects cross-site state-changing requests (CSRF) using the
// standard library's Sec-Fetch-Site / Origin checks (ADR 0006).
func CrossOrigin(next http.Handler) http.Handler {
	p := http.NewCrossOriginProtection()
	p.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.WarnContext(r.Context(), "cross-origin request rejected", "method", r.Method, "path", r.URL.Path)
		http.Error(w, "Forbidden", http.StatusForbidden)
	}))
	return p.Handler(next)
}

// MaxBody bounds request bodies.
func MaxBody(n int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, n)
			next.ServeHTTP(w, r)
		})
	}
}

// Lang puts the language cookie into the request context.
func Lang(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(i18n.WithLang(r.Context(), i18n.FromRequest(r))))
	})
}

type userKey struct{}

// CurrentUser returns the signed-in user stored by Authenticate.
func CurrentUser(ctx context.Context) *domain.User {
	u, _ := ctx.Value(userKey{}).(*domain.User)
	return u
}

// WithUser stores u in ctx (also used by tests).
func WithUser(ctx context.Context, u *domain.User) context.Context {
	return context.WithValue(ctx, userKey{}, u)
}

// Authenticate loads the session user, if any, into the context.
func Authenticate(s *auth.Sessions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, err := s.User(r)
			if err != nil && !errors.Is(err, auth.ErrNoSession) {
				slog.ErrorContext(r.Context(), "load session", "err", err)
			}
			if u != nil {
				r = r.WithContext(WithUser(r.Context(), u))
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireUser ports components/InitialLayout.tsx: signed-out users are sent
// to the login screen. HTMX requests get HX-Redirect instead of a 302.
func RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if CurrentUser(r.Context()) == nil {
			if response.IsHTMX(r) {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
