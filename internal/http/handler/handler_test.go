package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/vikyarifian/farahdin-go/internal/auth"
	"github.com/vikyarifian/farahdin-go/internal/config"
	"github.com/vikyarifian/farahdin-go/internal/external"
	"github.com/vikyarifian/farahdin-go/internal/repository"
	"github.com/vikyarifian/farahdin-go/internal/repository/repotest"
	"github.com/vikyarifian/farahdin-go/internal/service"
)

type stubFetcher struct{ body string }

func (s stubFetcher) Get(context.Context, string) (string, error) { return s.body, nil }
func (s stubFetcher) PostForm(context.Context, string, url.Values) (string, error) {
	if s.body == "" {
		return "", errors.New("upstream down")
	}
	return s.body, nil
}

type stubTranslator struct{}

func (stubTranslator) Translate(_ context.Context, text, _, _ string) ([]string, error) {
	return []string{text}, nil
}

type testApp struct {
	*httptest.Server
	app *App
}

func newTestApp(t *testing.T, upstream string) *testApp {
	t.Helper()
	db := repotest.Open(t)
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := func() time.Time { return time.Date(2026, 10, 1, 9, 0, 0, 0, loc) }
	users := &repository.Users{DB: db}
	a := &App{
		Cfg:      config.Config{Env: "development", DevLogin: true, BaseURL: "http://example.test", Location: loc},
		DB:       db,
		Sessions: &auth.Sessions{Store: &repository.Sessions{DB: db}, Users: users, TTL: time.Hour, Now: time.Now},
		Google:   &auth.Google{},
		Profiles: &service.Profiles{Users: users, Now: now},
		Readings: &service.Readings{Fetch: stubFetcher{upstream}, Translate: stubTranslator{}, Src: external.DefaultSources(), Now: now},
		Now:      now,
	}
	srv := httptest.NewServer(a.Routes())
	t.Cleanup(srv.Close)
	return &testApp{Server: srv, app: a}
}

// client returns an HTTP client that keeps cookies and does not follow redirects.
func (ta *testApp) client(t *testing.T) *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{Jar: jar, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}

func (ta *testApp) do(t *testing.T, c *http.Client, method, path string, form url.Values, headers map[string]string) (*http.Response, string) {
	t.Helper()
	var body *strings.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	} else {
		body = strings.NewReader("")
	}
	req, _ := http.NewRequest(method, ta.URL+path, body)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp, string(b)
}

func (ta *testApp) signIn(t *testing.T, c *http.Client) {
	t.Helper()
	resp, _ := ta.do(t, c, "POST", "/auth/dev", url.Values{"email": {"viky@example.com"}}, nil)
	if resp.StatusCode != http.StatusSeeOther || resp.Header.Get("Location") != "/" {
		t.Fatalf("dev login: %d %s", resp.StatusCode, resp.Header.Get("Location"))
	}
}

func TestSignedOutRedirects(t *testing.T) {
	ta := newTestApp(t, "")
	c := ta.client(t)
	for _, p := range []string{"/", "/profile", "/primbon/1", "/matrix-destiny"} {
		resp, _ := ta.do(t, c, "GET", p, nil, nil)
		if resp.StatusCode != http.StatusSeeOther || resp.Header.Get("Location") != "/login" {
			t.Errorf("%s: %d -> %s", p, resp.StatusCode, resp.Header.Get("Location"))
		}
	}
	resp, _ := ta.do(t, c, "POST", "/primbon/1", url.Values{"name": {"x"}}, map[string]string{"HX-Request": "true"})
	if resp.StatusCode != http.StatusUnauthorized || resp.Header.Get("HX-Redirect") != "/login" {
		t.Errorf("htmx: %d %q", resp.StatusCode, resp.Header.Get("HX-Redirect"))
	}
}

func TestCrossOriginPostRejected(t *testing.T) {
	ta := newTestApp(t, "")
	c := ta.client(t)
	resp, _ := ta.do(t, c, "POST", "/auth/dev", url.Values{"email": {"v@example.com"}},
		map[string]string{"Sec-Fetch-Site": "cross-site", "Origin": "https://evil.example"})
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d, want 403", resp.StatusCode)
	}
}

func TestSessionLifecycle(t *testing.T) {
	ta := newTestApp(t, "")
	c := ta.client(t)
	ta.signIn(t, c)

	resp, body := ta.do(t, c, "GET", "/", nil, nil)
	if resp.StatusCode != 200 || !strings.Contains(body, "Good morning") || !strings.Contains(body, ", viky") {
		t.Fatalf("home: %d", resp.StatusCode)
	}
	cookie := resp.Request.Header.Get("Cookie")
	if !strings.Contains(cookie, auth.SessionCookie) {
		t.Errorf("session cookie not sent: %q", cookie)
	}
	resp, _ = ta.do(t, c, "POST", "/logout", url.Values{}, nil)
	if resp.StatusCode != http.StatusSeeOther || resp.Header.Get("Location") != "/login" {
		t.Errorf("logout: %d %s", resp.StatusCode, resp.Header.Get("Location"))
	}
	resp, _ = ta.do(t, c, "GET", "/profile", nil, nil)
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("after logout /profile = %d", resp.StatusCode)
	}
}

func TestLanguageCookie(t *testing.T) {
	ta := newTestApp(t, "")
	c := ta.client(t)
	ta.signIn(t, c)
	resp, _ := ta.do(t, c, "POST", "/settings/language", url.Values{"lang": {"ID"}, "next": {"//evil.example"}}, nil)
	if resp.Header.Get("Location") != "/" {
		t.Errorf("open redirect not blocked: %q", resp.Header.Get("Location"))
	}
	_, body := ta.do(t, c, "GET", "/", nil, nil)
	if !strings.Contains(body, "Selamat pagi") || !strings.Contains(body, `lang="id"`) {
		t.Error("home must render in Indonesian after switching")
	}
	resp, _ = ta.do(t, c, "POST", "/settings/language", url.Values{"lang": {"EN"}}, map[string]string{"HX-Request": "true"})
	if resp.Header.Get("HX-Refresh") != "true" {
		t.Error("HTMX language switch must refresh the page")
	}
}

func TestEditProfile(t *testing.T) {
	ta := newTestApp(t, "")
	c := ta.client(t)
	ta.signIn(t, c)

	resp, body := ta.do(t, c, "POST", "/profile", url.Values{"name": {"Viky A"}, "birthday": {"2030-01-01"}, "gender": {"Male"}},
		map[string]string{"HX-Request": "true"})
	if resp.StatusCode != http.StatusUnprocessableEntity || !strings.Contains(body, "Birthdate is invalid.") {
		t.Errorf("future birthday: %d", resp.StatusCode)
	}
	if strings.Contains(body, "<html") {
		t.Error("HTMX response must be a fragment")
	}

	resp, body = ta.do(t, c, "POST", "/profile", url.Values{"name": {"Viky A"}, "birthday": {"1990-08-17"}, "gender": {"Female"}, "birthplace": {"Jakarta"}},
		map[string]string{"HX-Request": "true"})
	if resp.StatusCode != 200 || !strings.Contains(body, "Success") || !strings.Contains(body, `value="Leo"`) {
		t.Errorf("save: %d\n%s", resp.StatusCode, body)
	}
	_, body = ta.do(t, c, "GET", "/profile", nil, nil)
	if !strings.Contains(body, "Viky A") || !strings.Contains(body, ", 36") || !strings.Contains(body, "Female") {
		t.Error("profile must show name, age and gender")
	}
	_, body = ta.do(t, c, "GET", "/profile/zodiac?birthday=1990-03-25", nil, map[string]string{"HX-Request": "true"})
	if !strings.Contains(body, `value="Aries"`) {
		t.Errorf("zodiac fragment: %s", body)
	}
}

func TestReadingFragmentAndPage(t *testing.T) {
	upstream := `<html><body><div class="grid">Daily Love Tarot Reading The Lovers
A bright day.True Love Tarot Reading</div></body></html>`
	ta := newTestApp(t, upstream)
	c := ta.client(t)
	ta.signIn(t, c)

	resp, body := ta.do(t, c, "POST", "/tarot/1/read", url.Values{"card": {"6"}, "sheet": {"1"}}, map[string]string{"HX-Request": "true"})
	if resp.StatusCode != 200 || strings.Contains(body, "<html") || !strings.Contains(body, "A bright day.") || !strings.Contains(body, "tarot/6.png") {
		t.Errorf("htmx read: %d\n%s", resp.StatusCode, body)
	}
	resp, body = ta.do(t, c, "POST", "/tarot/1/read", url.Values{"card": {"6"}}, nil)
	if resp.StatusCode != 200 || !strings.Contains(body, "<html") || !strings.Contains(body, "data-result-sheet") {
		t.Errorf("no-JS read must render the full page with the sheet: %d", resp.StatusCode)
	}
	resp, _ = ta.do(t, c, "POST", "/tarot/1/read", url.Values{"card": {"99"}}, map[string]string{"HX-Request": "true"})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("out-of-deck card = %d", resp.StatusCode)
	}
	_, body = ta.do(t, c, "POST", "/tarot/2/read", url.Values{"you": {"3"}}, map[string]string{"HX-Request": "true"})
	if !strings.Contains(body, "Pick your card and your partner") {
		t.Errorf("true love needs two cards:\n%s", body)
	}
}

func TestUpstreamFailureShowsFriendlyError(t *testing.T) {
	ta := newTestApp(t, "")
	c := ta.client(t)
	ta.signIn(t, c)
	resp, body := ta.do(t, c, "POST", "/clairvoyance/1", url.Values{"name": {"Viky"}}, map[string]string{"HX-Request": "true"})
	if resp.StatusCode != 200 || !strings.Contains(body, `role="alert"`) || !strings.Contains(body, "Could not get a reading") {
		t.Errorf("status %d\n%s", resp.StatusCode, body)
	}
	if strings.Contains(body, "upstream down") {
		t.Error("internal error details must not leak")
	}
}

func TestMatrixValidationAndChart(t *testing.T) {
	ta := newTestApp(t, "")
	c := ta.client(t)
	ta.signIn(t, c)
	_, body := ta.do(t, c, "POST", "/matrix-destiny", url.Values{"name": {"Viky99"}, "birthday": {"1990-08-17"}}, map[string]string{"HX-Request": "true"})
	if !strings.Contains(body, "Name format is incorrect") {
		t.Errorf("validation:\n%s", body)
	}
	_, body = ta.do(t, c, "POST", "/matrix-destiny", url.Values{"name": {"viky arifian"}, "birthday": {"1990-08-17"}}, map[string]string{"HX-Request": "true"})
	for _, want := range []string{`id="matrix"`, "Viky Arifian", "17.08.1990", "Sahastrara", "reading is unavailable"} {
		if !strings.Contains(body, want) {
			t.Errorf("matrix result missing %q", want)
		}
	}
}

func TestPWAEndpoints(t *testing.T) {
	ta := newTestApp(t, "")
	c := ta.client(t)
	resp, body := ta.do(t, c, "GET", "/manifest.webmanifest", nil, nil)
	if resp.StatusCode != 200 || !strings.Contains(body, `"display":"standalone"`) || !strings.Contains(body, "maskable") {
		t.Errorf("manifest: %d %s", resp.StatusCode, body)
	}
	resp, body = ta.do(t, c, "GET", "/sw.js", nil, nil)
	if resp.StatusCode != 200 || !strings.Contains(body, "farahdin-static-") || !strings.Contains(body, `"/offline"`) {
		t.Errorf("sw.js: %d", resp.StatusCode)
	}
	resp, _ = ta.do(t, c, "GET", "/offline", nil, nil)
	if resp.StatusCode != 200 {
		t.Errorf("offline must be public: %d", resp.StatusCode)
	}
	resp, _ = ta.do(t, c, "GET", "/static/css/app.css?v=1", nil, nil)
	if resp.StatusCode != 200 || !strings.Contains(resp.Header.Get("Cache-Control"), "immutable") {
		t.Errorf("static: %d %q", resp.StatusCode, resp.Header.Get("Cache-Control"))
	}
	resp, _ = ta.do(t, c, "GET", "/healthz", nil, nil)
	if resp.StatusCode != 200 {
		t.Errorf("healthz: %d", resp.StatusCode)
	}
}

func TestSecurityHeaders(t *testing.T) {
	ta := newTestApp(t, "")
	resp, _ := ta.do(t, ta.client(t), "GET", "/login", nil, nil)
	for _, h := range []string{"Content-Security-Policy", "X-Content-Type-Options", "Referrer-Policy", "X-Frame-Options"} {
		if resp.Header.Get(h) == "" {
			t.Errorf("missing %s", h)
		}
	}
}

func TestCreatorRoute(t *testing.T) {
	ta := newTestApp(t, "")
	c := ta.client(t)
	ta.signIn(t, c)
	resp, body := ta.do(t, c, "GET", "/creator", nil, nil)
	if resp.StatusCode != 200 || !strings.Contains(body, "@vikyarifian") || !strings.Contains(body, `href="/creator"`) {
		t.Errorf("/creator: %d", resp.StatusCode)
	}
	resp, _ = ta.do(t, c, "GET", "/inbox", nil, nil)
	if resp.StatusCode != http.StatusMovedPermanently || resp.Header.Get("Location") != "/creator" {
		t.Errorf("/inbox: %d -> %s", resp.StatusCode, resp.Header.Get("Location"))
	}
}

func TestTarotSheetReloadsSpreadOnClose(t *testing.T) {
	ta := newTestApp(t, "")
	c := ta.client(t)
	ta.signIn(t, c)
	_, body := ta.do(t, c, "POST", "/tarot/1/pick", url.Values{"card": {"6"}}, map[string]string{"HX-Request": "true"})
	if !strings.Contains(body, `data-reload-on-close="/tarot/1"`) {
		t.Errorf("tarot sheet must reload the spread on close:\n%s", body)
	}
	_, body = ta.do(t, c, "GET", "/", nil, nil)
	home, creator, profile := strings.Index(body, `href="/"`), strings.Index(body, `href="/creator"`), strings.Index(body, `href="/profile"`)
	if !(home < creator && creator < profile) {
		t.Errorf("tab order must be Home, Creator, Profile (%d, %d, %d)", home, creator, profile)
	}
}
