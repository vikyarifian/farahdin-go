// Package config reads the runtime configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the application configuration. See .env.example.
type Config struct {
	Env                string // "development" or "production"
	Addr               string
	BaseURL            string
	PostgresURL        string
	GoogleClientID     string
	GoogleClientSecret string
	SessionTTL         time.Duration
	Location           *time.Location
	DevLogin           bool
	UpstreamTimeout    time.Duration
}

// Load reads and validates the configuration.
func Load() (Config, error) {
	c := Config{
		Env:                env("APP_ENV", "development"),
		Addr:               env("ADDR", ":8080"),
		BaseURL:            strings.TrimRight(env("BASE_URL", "http://localhost:8080"), "/"),
		PostgresURL:        os.Getenv("POSTGRES_URL"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	}
	var err error
	if c.SessionTTL, err = time.ParseDuration(env("SESSION_TTL", "720h")); err != nil {
		return c, fmt.Errorf("SESSION_TTL: %w", err)
	}
	if c.UpstreamTimeout, err = time.ParseDuration(env("UPSTREAM_TIMEOUT", "20s")); err != nil {
		return c, fmt.Errorf("UPSTREAM_TIMEOUT: %w", err)
	}
	if c.Location, err = time.LoadLocation(env("APP_TIMEZONE", "Asia/Jakarta")); err != nil {
		return c, fmt.Errorf("APP_TIMEZONE: %w", err)
	}
	if c.DevLogin, err = strconv.ParseBool(env("DEV_LOGIN", "false")); err != nil {
		return c, fmt.Errorf("DEV_LOGIN: %w", err)
	}
	if c.PostgresURL == "" {
		return c, fmt.Errorf("POSTGRES_URL is required (postgres://user:pass@host:5432/db)")
	}
	if c.Env != "development" && c.Env != "production" {
		return c, fmt.Errorf("APP_ENV must be development or production, got %q", c.Env)
	}
	if c.IsProduction() {
		if c.DevLogin {
			return c, fmt.Errorf("DEV_LOGIN must not be enabled in production")
		}
		if !c.SecureCookies() {
			return c, fmt.Errorf("BASE_URL must use https in production")
		}
		if c.GoogleClientID == "" || c.GoogleClientSecret == "" {
			return c, fmt.Errorf("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET are required in production")
		}
	}
	return c, nil
}

// IsProduction reports whether APP_ENV=production.
func (c Config) IsProduction() bool { return c.Env == "production" }

// SecureCookies reports whether cookies get the Secure attribute.
func (c Config) SecureCookies() bool { return strings.HasPrefix(c.BaseURL, "https://") }

// DevLoginEnabled reports whether the development-only sign-in is available.
func (c Config) DevLoginEnabled() bool { return c.DevLogin && !c.IsProduction() }

// GoogleRedirectURL is the OAuth callback registered in Google Cloud Console.
func (c Config) GoogleRedirectURL() string { return c.BaseURL + "/auth/google/callback" }

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}
