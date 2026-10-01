# Farahdin (PWA)

Primbon, horoscope, tarot, clairvoyance and matrix destiny readings, in Indonesian and English.
A Progressive Web App rewrite of the Expo app in `../farahdin-react-native`, built with Go, templ,
HTMX, vanilla JS and Tailwind.

## Requirements

- Go 1.25+
- `templ` CLI v0.3.1020 (`go install github.com/a-h/templ/cmd/templ@v0.3.1020`)
- Node.js 20+ (build time only, for the Tailwind CLI)
- PostgreSQL 14+ (`POSTGRES_URL`)

## Run locally

```bash
npm ci                 # Tailwind CLI
templ generate
npm run css
cp .env.example .env    # set POSTGRES_URL (and Google credentials)
set -a; . ./.env; set +a
go run ./cmd/app
```

Open http://localhost:8080. Tables are created on first start. With `DEV_LOGIN=true` the login page offers an email-only sign-in for
development. For Google sign-in, create an OAuth client (type "Web application") with the redirect
URI `http://localhost:8080/auth/google/callback` and set `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET`
(see `.env.example`).

With `make`: `make setup`, then `make dev`.

## Test

```bash
TEST_POSTGRES_URL=postgres://… go test ./...                # unit + HTTP tests (DB tests skip without it)
go test -tags live -run TestLive -v ./internal/service/    # against the real upstream sites
```

## Configuration

All settings are environment variables; see `.env.example`. In production (`APP_ENV=production`)
the server refuses to start without an https `BASE_URL` and Google credentials, and refuses `DEV_LOGIN`.

## Deploy

```bash
docker build -t farahdin .
docker run -p 8080:8080 --env-file .env farahdin
```

Health check: `GET /healthz` (checks the database). Data lives in PostgreSQL; back it up there.

## Moving data from Convex

```bash
cd ../farahdin-react-native && npx convex export --path convex-export.zip
cd ../farahdin-go && go run ./cmd/import-convex ../farahdin-react-native/convex-export.zip   # uses POSTGRES_URL
```

Users keep their profile; on first Google sign-in they are linked by email.

## Documentation

- `CLAUDE.md`: rules for working on this repository (people and agents)
- `docs/architecture.md`, `docs/pwa-strategy.md`, `docs/security-checklist.md`, `docs/testing-strategy.md`
- `docs/migration/`: what the React Native app does, how each journey was migrated, deviations, risks
- `docs/adr/`: architecture decisions
