package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/vikyarifian/farahdin-go/internal/domain"
	"github.com/vikyarifian/farahdin-go/internal/http/response"
	"github.com/vikyarifian/farahdin-go/internal/i18n"
	"github.com/vikyarifian/farahdin-go/internal/repository"
	"github.com/vikyarifian/farahdin-go/web/templates/pages"
)

// root ports app/index.tsx + InitialLayout: home when signed in, else login.
func (a *App) root(w http.ResponseWriter, r *http.Request) {
	u := user(r)
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	first := u.FirstName()
	if first == "" {
		first = "Guest"
	}
	w.Header().Set("Cache-Control", "no-store")
	response.Render(w, r, http.StatusOK, pages.Home(pages.HomeView{FirstName: first, Hour: a.Now().Hour()}))
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	if user(r) != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	v := pages.LoginView{GoogleEnabled: a.Google.Enabled(), DevLogin: a.Cfg.DevLoginEnabled()}
	if r.URL.Query().Get("error") != "" {
		v.Error = i18n.T(r.Context(), "Gagal masuk. Silakan coba lagi.", "Sign-in failed. Please try again.")
	}
	response.Render(w, r, http.StatusOK, pages.Login(v))
}

func (a *App) googleStart(w http.ResponseWriter, r *http.Request) {
	if !a.Google.Enabled() {
		a.notFound(w, r)
		return
	}
	if err := a.Google.Redirect(w, r); err != nil {
		a.serverError(w, r, err)
	}
}

func (a *App) googleCallback(w http.ResponseWriter, r *http.Request) {
	if !a.Google.Enabled() {
		a.notFound(w, r)
		return
	}
	id, err := a.Google.Callback(w, r)
	if err != nil {
		slog.WarnContext(r.Context(), "google sign-in failed", "err", err)
		http.Redirect(w, r, "/login?error=signin", http.StatusSeeOther)
		return
	}
	a.signIn(w, r, id)
}

// devLogin is a development-only sign-in (DEV_LOGIN=true, never in production).
func (a *App) devLogin(w http.ResponseWriter, r *http.Request) {
	if !a.Cfg.DevLoginEnabled() {
		a.notFound(w, r)
		return
	}
	email := strings.TrimSpace(r.FormValue("email"))
	if email == "" || !strings.Contains(email, "@") {
		http.Redirect(w, r, "/login?error=dev", http.StatusSeeOther)
		return
	}
	a.signIn(w, r, domain.Identity{Subject: "dev:" + email, Email: email, FirstName: strings.Split(email, "@")[0]})
}

func (a *App) signIn(w http.ResponseWriter, r *http.Request, id domain.Identity) {
	u, err := a.Profiles.SignIn(r.Context(), id)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := a.Sessions.Start(r.Context(), w, u.ID); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// logout ports handleSignOut: end the session and return to the login screen.
func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	if err := a.Sessions.End(w, r); err != nil {
		slog.ErrorContext(r.Context(), "end session", "err", err)
	}
	w.Header().Set("Clear-Site-Data", `"cache"`)
	response.Redirect(w, r, "/login")
}

// setLanguage ports setItem('lang', option) from the home dropdown and Settings.
func (a *App) setLanguage(w http.ResponseWriter, r *http.Request) {
	i18n.SetCookie(w, r.FormValue("lang"), a.Cfg.SecureCookies())
	if response.IsHTMX(r) {
		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, safeNext(r.FormValue("next")), http.StatusSeeOther)
}

func (a *App) profile(w http.ResponseWriter, r *http.Request) {
	u := user(r)
	v := pages.ProfileView{User: u}
	if b, ok := domain.ParseDate(u.Birthday); ok {
		v.Age = strconv.Itoa(domain.CalcAge(b, a.today()))
	}
	response.Render(w, r, http.StatusOK, pages.Profile(v))
}

func (a *App) editProfileView(u *domain.User) pages.EditProfileView {
	birthday := u.BirthdayOr(a.today())
	gender := u.Gender
	if gender == "" {
		gender = "Male" // edit-profile.tsx default
	}
	return pages.EditProfileView{
		Email:      u.Email,
		Name:       u.Fullname,
		Birthday:   domain.FormatDate(birthday),
		Zodiac:     domain.Zodiac(domain.FormatDate(birthday)),
		Birthplace: u.Birthplace,
		Gender:     gender,
		Today:      domain.FormatDate(a.today()),
	}
}

func (a *App) editProfile(w http.ResponseWriter, r *http.Request) {
	response.Render(w, r, http.StatusOK, pages.EditProfile(a.editProfileView(user(r))))
}

func (a *App) updateProfile(w http.ResponseWriter, r *http.Request) {
	u := user(r)
	birthday, _ := formDate(r, "birthday")
	in := domain.ProfileUpdate{
		Fullname:   r.FormValue("name"),
		Birthday:   birthday,
		Birthplace: r.FormValue("birthplace"),
		Gender:     r.FormValue("gender"),
	}
	errs, err := a.Profiles.UpdateProfile(r.Context(), u, in, lang(r))
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if len(errs) > 0 {
		v := a.editProfileView(u)
		v.Name, v.Birthplace, v.Gender = in.Fullname, in.Birthplace, in.Gender
		v.Birthday = r.FormValue("birthday")
		v.Zodiac = domain.Zodiac(v.Birthday)
		v.Errors = errs
		status := http.StatusUnprocessableEntity
		if response.IsHTMX(r) {
			response.Render(w, r, status, pages.EditProfileForm(v))
		} else {
			response.Render(w, r, status, pages.EditProfile(v))
		}
		return
	}
	fresh, err := a.Profiles.Users.ByID(r.Context(), u.ID)
	if errors.Is(err, repository.ErrNotFound) {
		response.Redirect(w, r, "/login")
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	v := a.editProfileView(fresh)
	v.Saved = true
	fragmentOrPage(w, r, pages.EditProfileForm(v), pages.EditProfile(v))
}

// profileZodiac refreshes the read-only zodiac field when the birthday changes.
func (a *App) profileZodiac(w http.ResponseWriter, r *http.Request) {
	b, _ := formDate(r, "birthday")
	sign := ""
	if !b.IsZero() {
		sign = domain.Zodiac(domain.FormatDate(b))
	}
	response.Render(w, r, http.StatusOK, pages.ZodiacField(sign))
}

func (a *App) settings(w http.ResponseWriter, r *http.Request) {
	response.Render(w, r, http.StatusOK, pages.Settings())
}

func (a *App) creator(w http.ResponseWriter, r *http.Request) {
	response.Render(w, r, http.StatusOK, pages.Creator())
}
