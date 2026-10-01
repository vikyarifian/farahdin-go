# ADR 0008 — PostgreSQL replaces SQLite

## Status

Accepted (2026-10-02). Supersedes the storage part of ADR 0005.

## Context

ADR 0005 chose SQLite as a single-file store. The project owner provided a PostgreSQL 16 server
(`POSTGRES_URL`) and asked to move the database there, which also removes SQLite's single-node limit
(former risk R-05).

## Decision

- Storage is **PostgreSQL** via `database/sql` with the **pgx** driver (`github.com/jackc/pgx/v5/stdlib`,
  v5.11.0, Go 1.25 compatible). This replaces `modernc.org/sqlite` in the approved dependency list.
- Connection: `POSTGRES_URL` (required). Pool: 10 open / 5 idle, 30 min lifetime.
- Schema: `migrations/0001_init.sql` rewritten for PostgreSQL (identity `BIGINT` keys, `TIMESTAMPTZ`).
  Column names and semantics are unchanged; `birthday` stays `TEXT` so legacy Convex values survive.
  No deployment had used the SQLite schema, so 0001 was rewritten instead of adding a 0002.
- Migrations run at startup, one transaction each, serialized with `pg_advisory_xact_lock` so several
  instances can start together.
- Tests use `TEST_POSTGRES_URL`; every test gets its own schema (`repotest.Open`) that is dropped
  afterwards. Without it, database tests are skipped. The production user cannot create schemas,
  so tests never run there.

## Consequences

- Backups and availability are the database server's concern; the app is stateless.
- Local development needs a reachable PostgreSQL.
- Data moved on 2026-10-02: the one real Google account from the old local SQLite file was copied
  (profile intact); development accounts and sessions were not.
- The server supports TLS and the current connection uses it (checked 2026-10-02 via `pg_stat_ssl`),
  but the URL has no `sslmode`, so the default `prefer` would silently fall back to plain text.
  Add `?sslmode=require` to `POSTGRES_URL` to make TLS mandatory (risk R-09).
