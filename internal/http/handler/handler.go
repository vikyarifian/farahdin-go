// Package handler maps HTTP routes to use cases and templates. Handlers stay
// thin: parse, call a service, render.
package handler

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"

	"github.com/vikyarifian/farahdin-go/internal/auth"
	"github.com/vikyarifian/farahdin-go/internal/config"
	"github.com/vikyarifian/farahdin-go/internal/domain"
	"github.com/vikyarifian/farahdin-go/internal/http/middleware"
	"github.com/vikyarifian/farahdin-go/internal/http/response"
	"github.com/vikyarifian/farahdin-go/internal/i18n"
	"github.com/vikyarifian/farahdin-go/internal/service"
	"github.com/vikyarifian/farahdin-go/web/templates/pages"
)

// App holds the dependencies of every handler.
type App struct {
	Cfg      config.Config
	DB       *sql.DB
	Sessions *auth.Sessions
	Google   *auth.Google
	Profiles *service.Profiles
	Readings *service.Readings
	Now      func() time.Time // in the configured time zone
}

// Routes builds the router with the full middleware chain.
func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	protected := func(h http.HandlerFunc) http.Handler { return middleware.RequireUser(h) }

	// Public.
	mux.HandleFunc("GET /{$}", a.root)
	mux.HandleFunc("GET /login", a.login)
	mux.HandleFunc("GET /auth/google/start", a.googleStart)
	mux.HandleFunc("GET /auth/google/callback", a.googleCallback)
	mux.HandleFunc("POST /auth/dev", a.devLogin)
	mux.HandleFunc("POST /logout", a.logout)
	mux.HandleFunc("POST /settings/language", a.setLanguage)
	mux.HandleFunc("GET /offline", a.offline)
	mux.HandleFunc("GET /healthz", a.healthz)
	mux.HandleFunc("GET /manifest.webmanifest", a.manifest)
	mux.HandleFunc("GET /sw.js", a.serviceWorker)
	mux.Handle("GET /static/", staticHandler())

	// Signed-in.
	mux.Handle("GET /profile", protected(a.profile))
	mux.Handle("GET /profile/edit", protected(a.editProfile))
	mux.Handle("POST /profile", protected(a.updateProfile))
	mux.Handle("GET /profile/zodiac", protected(a.profileZodiac))
	mux.Handle("GET /settings", protected(a.settings))
	mux.Handle("GET /creator", protected(a.creator))
	// The source tab was named "inbox"; keep old links working.
	mux.Handle("GET /inbox", http.RedirectHandler("/creator", http.StatusMovedPermanently))

	mux.Handle("GET /primbon", protected(a.topics("Primbon", "Primbon", domain.PrimbonTopics, "/primbon", false)))
	mux.Handle("GET /primbon/{topic}", protected(a.primbonForm))
	mux.Handle("POST /primbon/{topic}", protected(a.primbonSubmit))

	mux.Handle("GET /horoscope", protected(a.topics("Horoskop", "Horoscope", domain.HoroscopeTopics, "/horoscope", false)))
	mux.Handle("GET /horoscope/sign", protected(a.horoscopeSign))
	mux.Handle("GET /horoscope/{topic}", protected(a.horoscopeForm))
	mux.Handle("POST /horoscope/{topic}", protected(a.horoscopeSubmit))

	mux.Handle("GET /tarot", protected(a.topics("Tarot", "Tarot", domain.TarotTopics, "/tarot", true)))
	mux.Handle("GET /tarot/{topic}", protected(a.tarotSpread))
	mux.Handle("POST /tarot/{topic}/pick", protected(a.tarotPick))
	mux.Handle("POST /tarot/{topic}/read", protected(a.tarotRead))

	mux.Handle("GET /clairvoyance", protected(a.topics("Kewaskitaan", "Clairvoyance", domain.ClairvoyanceTopics, "/clairvoyance", true)))
	mux.Handle("GET /clairvoyance/{topic}", protected(a.clairvoyanceForm))
	mux.Handle("POST /clairvoyance/{topic}", protected(a.clairvoyanceSubmit))

	mux.Handle("GET /matrix-destiny", protected(a.matrixForm))
	mux.Handle("POST /matrix-destiny", protected(a.matrixSubmit))

	mux.HandleFunc("/", a.notFound)

	return middleware.Chain(mux,
		middleware.Recover,
		middleware.Logger,
		middleware.SecurityHeaders(a.Cfg.SecureCookies()),
		middleware.MaxBody(1<<20),
		middleware.CrossOrigin,
		middleware.Lang,
		middleware.Authenticate(a.Sessions),
	)
}

// --- shared helpers ---

func user(r *http.Request) *domain.User { return middleware.CurrentUser(r.Context()) }

func lang(r *http.Request) string { return i18n.Lang(r.Context()) }

func (a *App) today() time.Time { return domain.DateOnly(a.Now()) }

// topicParam resolves the {topic} path value against a feature's topics.
func topicParam(r *http.Request, topics []domain.Topic) (domain.Topic, bool) {
	k, err := strconv.Atoi(r.PathValue("topic"))
	if err != nil {
		return domain.Topic{}, false
	}
	return domain.FindTopic(topics, k)
}

// formDate parses a "YYYY-MM-DD" form value.
func formDate(r *http.Request, name string) (time.Time, bool) {
	return domain.ParseDate(strings.TrimSpace(r.FormValue(name)))
}

// fragmentOrPage renders the fragment for HTMX requests and the full page otherwise.
func fragmentOrPage(w http.ResponseWriter, r *http.Request, fragment, page templ.Component) {
	if response.IsHTMX(r) {
		response.Render(w, r, http.StatusOK, fragment)
		return
	}
	response.Render(w, r, http.StatusOK, page)
}

// readingError turns a service error into a user-facing message and logs
// upstream failures (the source app swallowed them silently).
func readingError(r *http.Request, err error) string {
	if msg, ok := service.IsValidation(err); ok {
		return msg
	}
	slog.ErrorContext(r.Context(), "reading failed", "path", r.URL.Path, "err", err)
	return i18n.T(r.Context(),
		"Bacaan tidak bisa diambil sekarang. Coba lagi sebentar lagi.",
		"Could not get a reading right now. Please try again in a moment.")
}

func (a *App) notFound(w http.ResponseWriter, r *http.Request) {
	response.Render(w, r, http.StatusNotFound, pages.ErrorPage(http.StatusNotFound,
		i18n.T(r.Context(), "Halaman tidak ditemukan", "Page not found")))
}

func (a *App) serverError(w http.ResponseWriter, r *http.Request, err error) {
	slog.ErrorContext(r.Context(), "internal error", "path", r.URL.Path, "err", err)
	response.Render(w, r, http.StatusInternalServerError, pages.ErrorPage(http.StatusInternalServerError,
		i18n.T(r.Context(), "Terjadi kesalahan. Coba lagi.", "Something went wrong. Please try again.")))
}

// safeNext accepts only local paths for post-action redirects.
func safeNext(next string) string {
	if !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") || strings.HasPrefix(next, "/\\") {
		return "/"
	}
	return next
}
