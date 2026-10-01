// Command app runs the Farahdin PWA server.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vikyarifian/farahdin-go/internal/auth"
	"github.com/vikyarifian/farahdin-go/internal/config"
	"github.com/vikyarifian/farahdin-go/internal/external"
	"github.com/vikyarifian/farahdin-go/internal/http/handler"
	"github.com/vikyarifian/farahdin-go/internal/repository"
	"github.com/vikyarifian/farahdin-go/internal/service"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := repository.Open(cfg.PostgresURL)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := repository.Migrate(ctx, db); err != nil {
		return err
	}

	now := func() time.Time { return time.Now().In(cfg.Location) }
	users := &repository.Users{DB: db}
	sessionStore := &repository.Sessions{DB: db}
	client := external.NewClient(cfg.UpstreamTimeout)
	src := external.DefaultSources()

	app := &handler.App{
		Cfg: cfg,
		DB:  db,
		Sessions: &auth.Sessions{
			Store:  sessionStore,
			Users:  users,
			TTL:    cfg.SessionTTL,
			Secure: cfg.SecureCookies(),
			Now:    time.Now,
		},
		Google: &auth.Google{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL(),
			Secure:       cfg.SecureCookies(),
		},
		Profiles: &service.Profiles{Users: users, Now: now},
		Readings: &service.Readings{
			Fetch:     client,
			Translate: &external.Translator{Client: client, Base: src.Translate},
			Src:       src,
			Now:       now,
		},
		Now: now,
	}

	go purgeSessions(ctx, sessionStore)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           app.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      90 * time.Second, // readings call slow upstream sites
		IdleTimeout:       120 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", cfg.Addr, "env", cfg.Env, "base_url", cfg.BaseURL,
			"google_signin", app.Google.Enabled(), "dev_login", cfg.DevLoginEnabled())
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		slog.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return err
		}
	}
	return nil
}

// purgeSessions deletes expired sessions once an hour.
func purgeSessions(ctx context.Context, s *repository.Sessions) {
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	for {
		if n, err := s.DeleteExpired(ctx, time.Now()); err != nil {
			slog.Error("purge sessions", "err", err)
		} else if n > 0 {
			slog.Info("purged sessions", "count", n)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}
