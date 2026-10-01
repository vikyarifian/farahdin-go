# Target PWA Architecture

## Principles

1. Server-render first.
2. Enhance with HTMX.
3. Use vanilla JS only where required.
4. Keep domain logic independent from HTTP and templates.
5. Keep authorization on the server.
6. Keep browser state small.
7. Prefer simple Go standard-library primitives.

## Structure (as built)

```text
farahdin-go/
├── cmd/
│   ├── app/                  server: config → DB → services → router → graceful shutdown
│   └── import-convex/        Convex export ZIP → PostgreSQL (data-map.md)
├── internal/
│   ├── config/               env vars (.env.example), production guards
│   ├── domain/               pure rules: zodiac, age, dates, topics, matrix (+ generated formulas)
│   ├── external/             upstream HTTP client, cheerio-like HTML queries, JS-string pipeline, Translate
│   ├── service/              use cases: one file per feature (literal ports of generate()), profiles
│   ├── repository/           PostgreSQL (pgx): users, sessions, migrations runner, Convex importer; repotest/ for tests
│   ├── auth/                 Google OIDC + PKCE, opaque sessions
│   ├── i18n/                 ID/EN language (cookie + context)
│   └── http/
│       ├── handler/          routes, thin handlers, PWA endpoints (manifest, sw.js, offline, healthz)
│       ├── middleware/       recover, log, security headers, body limit, CSRF, lang, auth
│       └── response/         render / HTMX helpers
├── migrations/               *.sql, embedded
├── web/
│   ├── web.go                embeds static/, content-hashed asset URLs
│   ├── templates/{layouts,components,pages}/   templ (+ generated *_templ.go)
│   └── static/{css,js,icons,images,fonts}/
└── docs/
```

Dependency direction: `handler → service → (repository | external) → domain`. `domain` imports nothing
from the app. Templates import `domain`/`service` types for view models only.

## HTTP

- `net/http.ServeMux` with method + wildcard patterns (Go 1.22+), registered in `handler.Routes()`.
- Middleware order: Recover → Logger → SecurityHeaders → MaxBody(1 MB) → CrossOrigin (CSRF) → Lang → Authenticate; `RequireUser` wraps signed-in routes.
- Handler pattern: parse form → call service → `fragmentOrPage(fragment, page)`: HTMX requests get the fragment, plain form posts get the whole page with the result in place (progressive enhancement).

## Authentication and authorization

ADR 0004. One role (signed-in user); every feature route requires a session; a user can only read
and change their own profile (the user id comes from the session, never from the request).

## Domain

`domain` has no HTTP, SQL or template imports. Matrix formulas are generated from the React Native
source and verified against values computed by the original expressions (`testdata/matrix_reference.json`).

## Upstream readings

ADR 0007. `service.Readings` depends on two interfaces (`Fetcher`, `Translator`) so tests use fakes.

## Templ

- `layouts`: `Base` (document shell, banners), `App` (bottom tab bar), `Feature` (header + back link).
- `components`: Icon (SVG sprite), Header, TopicGrid, FeatureHero, Field, TextInput, DateInput, SubmitButton, Alert, ResultSheet, Lines, flags.
- `pages`: one file per area; `matrix_chart.templ` is generated from the RN SVG.
- Text uses `i18n.T(ctx, id, en)`, mirroring the source's inline `lang === 'ID' ? … : …`.

## HTMX conventions

| Interaction | Request | Target / swap | Loading | Error |
|---|---|---|---|---|
| Reading forms | `hx-post` same URL | `#result` innerHTML | `hx-indicator` spinner + `hx-disabled-elt` | Alert fragment in `#result` |
| Tarot Read Card | `hx-post /tarot/{n}/read` | `#tarot-reading` outerHTML | same | Alert above the button |
| Edit Profile | `hx-post /profile` | form outerHTML | same | 422 + field errors |
| Zodiac preview | `hx-get` on birthday `change` | `#zodiac` outerHTML / out-of-band `#sign-icon`, `#sign-label` | – | empty sign |
| Language (Settings) | `hx-post` on `change` | `HX-Refresh` | – | – |

`htmx-config`: every status is swapped (the server renders friendly fragments), `allowEval: false`,
`includeIndicatorStyles: false` (CSP), `selfRequestsOnly: true`. Session loss → `401` + `HX-Redirect`.

## Vanilla JS (`web/static/js/app.js`, ~150 lines)

Service-worker registration + update banner, online/offline banner, greeting from the device clock,
`history.back()` for in-app back links, result-sheet focus/close/Escape, HTMX network-error message,
closing the language dropdown. No business rules.

## Tailwind

Tailwind v4 CLI (`web/static/css/input.css` → `app.css`), theme tokens from the source `COLORS`
(`primary`, `background`, `backdrop`, `surface`, `muted`, …) and fonts (`font-javassoul`,
`font-garamond`, `font-garamond-sc`). The CSP forbids inline styles, so dynamic colours use classes
(e.g. `.chakra-1…7`).

## PWA

See `pwa-strategy.md`.

## Security

See `security-checklist.md` (status of each item) and ADRs 0004, 0006.

## Error handling and observability

- `log/slog` JSON to stdout: one line per request (method, path, status, duration, htmx), upstream
  failures (`reading failed` with the error), panics with stack. No cookies, query strings or form values are logged.
- Users see generic messages; no stack traces, SQL or upstream errors.
- `GET /healthz` pings the database.

## Database

PostgreSQL via `database/sql` + pgx (ADR 0008). Migrations are embedded and applied at startup, one transaction each, under an advisory lock.
