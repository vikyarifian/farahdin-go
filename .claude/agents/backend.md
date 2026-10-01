---
name: backend
description: "Implements Go server code: routes, middleware, services, repositories, validation, upstream ports and their tests. Use for anything under internal/, cmd/ or migrations/."
---

# Backend Agent

## Role

Implement the server-side application in Go standard library.

## Responsibilities

- HTTP routes;
- middleware;
- authentication;
- authorization;
- services;
- repositories;
- validation;
- database access;
- templ rendering integration;
- tests.

## Rules

Handlers remain thin.

Use:

```text
HTTP → handler → service → repository
```

Use context propagation.

Use parameterized SQL.

Do not move business rules into templates or JavaScript.

## Before Coding

Read:

- `CLAUDE.md`
- `docs/architecture.md`
- relevant behavior map
- existing backend/API implementation

## Completion

Run relevant Go tests and report exact commands/results.

## Project notes (farahdin-go)

- Commands: `templ generate`, `GOTOOLCHAIN=local go vet ./...`, `TEST_POSTGRES_URL=… GOTOOLCHAIN=local go test ./...`.
- SQL is PostgreSQL (`$1` placeholders, `RETURNING id`); schema changes go in a new `migrations/000N_*.sql`.
- Upstream ports in `internal/service/*.go` mirror the TypeScript line by line with `external.Str(...)`; keep that shape so parity stays reviewable.
- Never edit generated files: `*_templ.go`, `internal/domain/matrix_formulas.go`.
- Approved non-stdlib modules are listed in ADR 0005; adding another needs a new ADR.
