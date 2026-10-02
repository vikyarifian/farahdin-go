// Package external talks to the third-party sites the source app scraped
// directly from the phone (primbon.com, horoscope.com, californiapsychics.com,
// matrixdestinychart.com) and to Google Translate (plus backups, see translate.go).
// In the PWA these calls run on the server: browsers would block them (CORS)
// and the server can apply timeouts and size limits.
package external

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Sources holds the upstream base URLs so tests can point them at httptest servers.
type Sources struct {
	Primbon           string // https://www.primbon.com (GET pages)
	PrimbonPost       string // https://primbon.com (form posts, as in the source app)
	Horoscope         string // https://www.horoscope.com
	CaliforniaPsychic string // https://www.californiapsychics.com
	MatrixDestiny     string // https://matrixdestinychart.com
}

// DefaultSources are the production upstreams used by the source app.
func DefaultSources() Sources {
	return Sources{
		Primbon:           "https://www.primbon.com",
		PrimbonPost:       "https://primbon.com",
		Horoscope:         "https://www.horoscope.com",
		CaliforniaPsychic: "https://www.californiapsychics.com",
		MatrixDestiny:     "https://matrixdestinychart.com",
	}
}

// Client performs bounded upstream requests.
type Client struct {
	HTTP      *http.Client
	UserAgent string
	MaxBody   int64
}

// NewClient returns a client with a request timeout and a body size cap.
func NewClient(timeout time.Duration) *Client {
	return &Client{
		HTTP:      &http.Client{Timeout: timeout},
		UserAgent: "Mozilla/5.0 (compatible; Farahdin/1.0; +https://profile.vikyarifian.web.id)",
		MaxBody:   5 << 20,
	}
}

// UpstreamError reports a non-2xx upstream response.
type UpstreamError struct {
	URL    string
	Status int
}

func (e *UpstreamError) Error() string {
	return fmt.Sprintf("upstream %s returned %d", e.URL, e.Status)
}

// Get fetches rawURL and returns the body decoded as UTF-8 text, like fetch().text().
func (c *Client) Get(ctx context.Context, rawURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	return c.do(req)
}

// PostForm posts an application/x-www-form-urlencoded body.
func (c *Client) PostForm(ctx context.Context, rawURL string, form url.Values) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.do(req)
}

func (c *Client) do(req *http.Request) (string, error) {
	req.Header.Set("User-Agent", c.UserAgent)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("request %s: %w", req.URL.Redacted(), err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, c.MaxBody))
	if err != nil {
		return "", fmt.Errorf("read %s: %w", req.URL.Redacted(), err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", &UpstreamError{URL: req.URL.Redacted(), Status: resp.StatusCode}
	}
	return string(body), nil
}
