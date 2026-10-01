# CLAUDE.md — Farahdin PWA (Go)

## Mission

Rewrite of the React Native app `../farahdin-react-native` (Expo + Clerk + Convex) as a production-ready
Progressive Web App, preserving business behaviour, user flows, validation rules, permissions and
data semantics. Orchestration rules come from `../pwa-agent-orchestration`.

## Status (2026-10-01)

Phases 0–3 are complete: every journey is implemented (`docs/migration/ledger.md`). Phase 4
(hardening) is in progress: rows in REVIEW need a side-by-side parity run and reviewer approval;
rows in TESTING list what is not verified yet (real Google sign-in, Lighthouse/installability,
real Convex import).

## Target Stack

- Backend: Go standard library (`net/http`, `database/sql`, `embed`, `log/slog`, …), Go 1.25
- Server-rendered UI: `templ`
- Progressive enhancement: `HTMX` 2 (vendored)
- Client behaviour: vanilla JavaScript only (`web/static/js/app.js`)
- Styling: Tailwind CSS v4 (CLI at build time)
- Database: PostgreSQL via `database/sql` + pgx (ADR 0008; Convex data imported with `cmd/import-convex`)
- Auth: Google OIDC + server sessions (ADR 0004)
- PWA: manifest, generated service worker, offline fallback (`docs/pwa-strategy.md`)

## Commands

```bash
templ generate                         # after editing *.templ
npm run css                            # after changing classes in templates or input.css
GOTOOLCHAIN=local go vet ./...
TEST_POSTGRES_URL=… GOTOOLCHAIN=local go test ./...   # DB tests skip without it; never use production
go test -tags live -run TestLive -v ./internal/service/   # real upstream sites
set -a; . ./.env; set +a; go run ./cmd/app                 # http://localhost:8080, needs POSTGRES_URL
go run ./cmd/import-convex convex-export.zip               # uses POSTGRES_URL
```

`GOTOOLCHAIN=local` keeps Go from downloading a newer toolchain; dependencies are pinned to Go 1.25-compatible releases.

## Map

- `docs/architecture.md`: structure, layering, HTMX conventions
- `docs/migration/`: inventory, behavior map (Q-xx), API map, data map, ledger (MIG-xxx), deviations (D-xx), risks (R-xx)
- `docs/adr/`: decisions 0001–0008
- `.claude/agents/`: orchestrator, analyst, architect, backend, frontend, migration, pwa, qa, security, reviewer, devops
- `.claude/commands/`: `/migrate`, `/audit`, `/parity-review`, `/live-check`

## Hard Constraints

1. No React, React Native, Vue, Angular, Svelte, Alpine.js, jQuery or other frontend framework.
2. No Go web frameworks (Gin, Echo, Fiber, Chi, Buffalo, …).
3. Standard library for routing, middleware, cookies, sessions, JSON, logging. The only approved modules are in ADR 0005 (amended by ADR 0008); anything else needs a new ADR.
4. `templ` for HTML; HTMX for server interaction; vanilla JS only where HTML/HTMX cannot do it.
5. Tailwind for styling. The CSP forbids inline `<script>` and `style=""`.
6. Preserve the React Native app's observable behaviour unless a requirement changes it; record every intentional difference in `docs/migration/deviations.md`.
7. Never rewrite a feature from assumptions when the source can be inspected.
8. Do not silently change API contracts, validation, authorization, business rules or data semantics.
9. HTML must stay meaningful without JavaScript (every HTMX form also works as a plain form post).
10. Accessibility, responsive behaviour, security and performance are first-class.

## Project rules

- Upstream readings are literal ports of the TypeScript `generate()` functions using `external.Str(...)`; keep that shape so they can be compared line by line (ADR 0007).
- Generated files are never edited by hand: `*_templ.go`, `web/static/css/app.css`, `internal/domain/matrix_formulas.go`, `web/templates/pages/matrix_chart.templ`.
- Text in templates: `i18n.T(ctx, "Indonesian", "English")`.
- Static asset URLs: `web.Asset("path")`.
- Handlers stay thin: parse → service → `fragmentOrPage`.

## Source-of-Truth Hierarchy

1. Explicit product/business requirements.
2. Existing backend/API/database behaviour (Convex functions).
3. Existing React Native behaviour.
4. Existing documentation.
5. New implementation assumptions.

Document unresolved conflicts (behavior-map Q-xx) instead of silently choosing.

## Definition of Done

A migrated feature is complete only when: behaviour matches the acceptance criteria; authorization
and validation are server-side; HTML works without unnecessary client state; HTMX interactions have
loading/error/empty states; JS is minimal; responsive and keyboard behaviour are verified; errors are
observable and friendly; critical paths are tested; no forbidden framework was introduced; differences are documented.

## Agent Rules

Agents must inspect existing code first, make small reviewable changes, avoid unrelated refactors,
update the relevant docs, report assumptions and blockers, and never claim tests passed without
running them. Agents must not invent API fields or database columns, replace working business rules,
add dependencies for convenience, or change authentication/security behaviour without an ADR.

Every agent report has: `Context`, `Findings`, `Changes`, `Tests`, `Risks`, `Open Questions`.

## Git Discipline

Focused commits: `feat: migrate <journey>`, `fix: preserve <behavior>`, `test: cover <journey>`,
`docs: document <decision>`, `chore: …`. Never combine unrelated migrations in one commit.
