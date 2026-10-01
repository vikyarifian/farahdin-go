package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/vikyarifian/farahdin-go/internal/domain"
)

// Google endpoints (https://accounts.google.com/.well-known/openid-configuration).
const (
	googleAuthURL     = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL    = "https://oauth2.googleapis.com/token"
	googleUserInfoURL = "https://openidconnect.googleapis.com/v1/userinfo"
)

// flowCookie holds the state and PKCE verifier between redirect and callback.
const flowCookie = "farahdin_oauth"

// Google runs the OpenID Connect authorization-code flow with PKCE.
type Google struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Secure       bool
	HTTP         *http.Client
	// Endpoint overrides for tests.
	AuthURL, TokenURL, UserInfoURL string
}

// Enabled reports whether Google sign-in is configured.
func (g *Google) Enabled() bool {
	return g != nil && g.ClientID != "" && g.ClientSecret != "" && g.RedirectURL != ""
}

func (g *Google) endpoint(override, def string) string {
	if override != "" {
		return override
	}
	return def
}

// Redirect starts the flow: it stores state + verifier in a short-lived
// cookie and redirects the browser to Google.
func (g *Google) Redirect(w http.ResponseWriter, r *http.Request) error {
	state, err := randomToken(24)
	if err != nil {
		return err
	}
	verifier, err := randomToken(48)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     flowCookie,
		Value:    state + "." + verifier,
		Path:     "/auth/google",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   g.Secure,
		SameSite: http.SameSiteLaxMode,
	})
	challenge := sha256.Sum256([]byte(verifier))
	q := url.Values{
		"client_id":             {g.ClientID},
		"redirect_uri":          {g.RedirectURL},
		"response_type":         {"code"},
		"scope":                 {"openid email profile"},
		"state":                 {state},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(challenge[:])},
		"code_challenge_method": {"S256"},
		"prompt":                {"select_account"},
	}
	http.Redirect(w, r, g.endpoint(g.AuthURL, googleAuthURL)+"?"+q.Encode(), http.StatusFound)
	return nil
}

// ErrFlow means the callback did not belong to a flow started by this browser.
var ErrFlow = errors.New("invalid or expired sign-in attempt")

// Callback completes the flow and returns the verified identity.
func (g *Google) Callback(w http.ResponseWriter, r *http.Request) (domain.Identity, error) {
	c, err := r.Cookie(flowCookie)
	http.SetCookie(w, &http.Cookie{Name: flowCookie, Path: "/auth/google", MaxAge: -1, HttpOnly: true, Secure: g.Secure, SameSite: http.SameSiteLaxMode})
	if err != nil {
		return domain.Identity{}, ErrFlow
	}
	state, verifier, ok := strings.Cut(c.Value, ".")
	got := r.URL.Query().Get("state")
	if !ok || got == "" || subtle.ConstantTimeCompare([]byte(state), []byte(got)) != 1 {
		return domain.Identity{}, ErrFlow
	}
	if e := r.URL.Query().Get("error"); e != "" {
		return domain.Identity{}, fmt.Errorf("google: %s", e)
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		return domain.Identity{}, ErrFlow
	}
	access, err := g.exchange(r.Context(), code, verifier)
	if err != nil {
		return domain.Identity{}, err
	}
	return g.userInfo(r.Context(), access)
}

func (g *Google) exchange(ctx context.Context, code, verifier string) (string, error) {
	form := url.Values{
		"code":          {code},
		"client_id":     {g.ClientID},
		"client_secret": {g.ClientSecret},
		"redirect_uri":  {g.RedirectURL},
		"grant_type":    {"authorization_code"},
		"code_verifier": {verifier},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.endpoint(g.TokenURL, googleTokenURL), strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var tok struct {
		AccessToken string `json:"access_token"`
	}
	if err := g.doJSON(req, &tok); err != nil {
		return "", fmt.Errorf("google token: %w", err)
	}
	if tok.AccessToken == "" {
		return "", errors.New("google token: empty access token")
	}
	return tok.AccessToken, nil
}

func (g *Google) userInfo(ctx context.Context, accessToken string) (domain.Identity, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.endpoint(g.UserInfoURL, googleUserInfoURL), nil)
	if err != nil {
		return domain.Identity{}, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	var info struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
	}
	if err := g.doJSON(req, &info); err != nil {
		return domain.Identity{}, fmt.Errorf("google userinfo: %w", err)
	}
	// Accounts are linked by email (see SignIn), so the email must be verified.
	if info.Sub == "" || info.Email == "" || !info.EmailVerified {
		return domain.Identity{}, errors.New("google userinfo: missing or unverified email")
	}
	return domain.Identity{
		Subject:   info.Sub,
		Email:     info.Email,
		FirstName: info.GivenName,
		LastName:  info.FamilyName,
		ImageURL:  info.Picture,
	}, nil
}

func (g *Google) doJSON(req *http.Request, v any) error {
	client := g.HTTP
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return json.Unmarshal(body, v)
}
