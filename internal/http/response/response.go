// Package response holds rendering helpers shared by HTTP handlers.
package response

import (
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
)

// IsHTMX reports whether r was issued by HTMX (and is not a history restore).
func IsHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-History-Restore-Request") != "true"
}

// Render writes a templ component as HTML with the given status.
func Render(w http.ResponseWriter, r *http.Request, status int, c templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Method == http.MethodGet && IsHTMX(r) {
		w.Header().Add("Vary", "HX-Request")
	}
	w.WriteHeader(status)
	if err := c.Render(r.Context(), w); err != nil {
		slog.ErrorContext(r.Context(), "render failed", "err", err, "path", r.URL.Path)
	}
}

// Redirect sends the browser to url: HX-Redirect for HTMX requests,
// 303 See Other otherwise (so a POST becomes a GET).
func Redirect(w http.ResponseWriter, r *http.Request, url string) {
	if IsHTMX(r) {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}
