# ADR 0005 — SQLite storage and approved non-stdlib dependencies

## Status

Accepted (2026-10-01). **Storage part superseded by ADR 0008 (PostgreSQL, 2026-10-02)**; the dependency table is updated accordingly.

## Context

`CLAUDE.md` asks to preserve the existing database unless the migration plan changes it, and to
use the standard library unless an exception is documented.

The source database is **Convex** (hosted document DB, functions in `convex/`). Convex is accessed
through its JS client and authenticates calls with the Clerk JWT; ADR 0004 removes Clerk. The data
set is small: `users`, `categories`, `inboxes` (and only `users` is used by any screen).

## Decision

1. **Storage: SQLite through `database/sql`**, one file (`DATABASE_PATH`), schema in
   `migrations/0001_init.sql` mirroring `convex/schema.ts` (same tables, fields and indexes,
   plus `convex_id` for traceability, `google_sub`, `image_url`, and `sessions`).
   Data moves with `cmd/import-convex` from a `npx convex export` ZIP; the import is idempotent.
2. **Approved dependencies** (everything else is the standard library):

| Module | Why | Scope |
|---|---|---|
| `github.com/a-h/templ` | required by the target stack (templ runtime) | rendering |
| ~~`modernc.org/sqlite`~~ → `github.com/jackc/pgx/v5` | `database/sql` driver for PostgreSQL (ADR 0008); pure Go, no cgo | repository |
| `golang.org/x/net/html` | HTML5 parser for the scraped pages (the source used cheerio). Writing a spec-compliant parser by hand is not reasonable. | `internal/external` |
| `htmx.org` 2.0.11 (vendored file) | target stack | browser |
| `@tailwindcss/cli` 4.3.3 (npm, build time only) | target stack; nothing from npm ships | build |

Versions are pinned to releases that support Go 1.25 (the local toolchain).

## Consequences

- One process, one file to back up (`data/farahdin.db*`); WAL mode, `busy_timeout` 5s.
- Horizontal scaling needs a different store; if that is ever required, the repository
  interfaces (`service.UserStore`, `auth.SessionStore`) are the seam.
- Convex is retired after the import; nothing writes to it any more.
