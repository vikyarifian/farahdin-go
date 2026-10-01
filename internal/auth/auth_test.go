package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// fakeGoogle implements the token and userinfo endpoints and checks PKCE.
func fakeGoogle(t *testing.T, challenge *string, verified bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			r.ParseForm()
			sum := sha256.Sum256([]byte(r.PostForm.Get("code_verifier")))
			if base64.RawURLEncoding.EncodeToString(sum[:]) != *challenge {
				t.Error("PKCE verifier does not match the challenge")
				http.Error(w, "bad verifier", http.StatusBadRequest)
				return
			}
			if r.PostForm.Get("code") != "the-code" || r.PostForm.Get("client_secret") != "secret" {
				http.Error(w, "bad code", http.StatusBadRequest)
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"access_token": "at"})
		case "/userinfo":
			if r.Header.Get("Authorization") != "Bearer at" {
				http.Error(w, "no token", http.StatusUnauthorized)
				return
			}
			json.NewEncoder(w).Encode(map[string]any{"sub": "123", "email": "v@example.com", "email_verified": verified,
				"given_name": "Viky", "family_name": "Arifian", "picture": "https://lh3.googleusercontent.com/x"})
		}
	}))
}

func startFlow(t *testing.T, g *Google) (state string, cookie *http.Cookie, challenge string) {
	t.Helper()
	rec := httptest.NewRecorder()
	if err := g.Redirect(rec, httptest.NewRequest("GET", "/auth/google/start", nil)); err != nil {
		t.Fatal(err)
	}
	loc, _ := url.Parse(rec.Header().Get("Location"))
	q := loc.Query()
	if q.Get("code_challenge_method") != "S256" || q.Get("scope") != "openid email profile" {
		t.Errorf("auth params = %v", q)
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == flowCookie {
			cookie = c
		}
	}
	if cookie == nil || !cookie.HttpOnly {
		t.Fatal("flow cookie must be set and HttpOnly")
	}
	return q.Get("state"), cookie, q.Get("code_challenge")
}

func TestGoogleFlow(t *testing.T) {
	var challenge string
	srv := fakeGoogle(t, &challenge, true)
	defer srv.Close()
	g := &Google{ClientID: "id", ClientSecret: "secret", RedirectURL: "http://app/auth/google/callback",
		TokenURL: srv.URL + "/token", UserInfoURL: srv.URL + "/userinfo"}

	state, cookie, ch := startFlow(t, g)
	challenge = ch

	req := httptest.NewRequest("GET", "/auth/google/callback?code=the-code&state="+state, nil)
	req.AddCookie(cookie)
	id, err := g.Callback(httptest.NewRecorder(), req)
	if err != nil {
		t.Fatal(err)
	}
	if id.Subject != "123" || id.Email != "v@example.com" || id.Fullname() != "Viky Arifian" {
		t.Errorf("identity = %+v", id)
	}
}

func TestGoogleCallbackRejectsWrongState(t *testing.T) {
	g := &Google{ClientID: "id", ClientSecret: "secret", RedirectURL: "x"}
	_, cookie, _ := startFlow(t, g)
	req := httptest.NewRequest("GET", "/auth/google/callback?code=c&state=forged", nil)
	req.AddCookie(cookie)
	if _, err := g.Callback(httptest.NewRecorder(), req); err != ErrFlow {
		t.Errorf("err = %v, want ErrFlow", err)
	}
	// No cookie at all (callback opened in another browser).
	req = httptest.NewRequest("GET", "/auth/google/callback?code=c&state=x", nil)
	if _, err := g.Callback(httptest.NewRecorder(), req); err != ErrFlow {
		t.Errorf("err = %v, want ErrFlow", err)
	}
}

func TestGoogleRejectsUnverifiedEmail(t *testing.T) {
	var challenge string
	srv := fakeGoogle(t, &challenge, false)
	defer srv.Close()
	g := &Google{ClientID: "id", ClientSecret: "secret", RedirectURL: "x", TokenURL: srv.URL + "/token", UserInfoURL: srv.URL + "/userinfo"}
	state, cookie, ch := startFlow(t, g)
	challenge = ch
	req := httptest.NewRequest("GET", "/auth/google/callback?code=the-code&state="+state, nil)
	req.AddCookie(cookie)
	if _, err := g.Callback(httptest.NewRecorder(), req); err == nil || !strings.Contains(err.Error(), "unverified") {
		t.Errorf("err = %v", err)
	}
}
