---
name: migration
description: "Executes one complete vertical-slice migration of a user journey end to end (backend, templ, HTMX, PWA, tests, parity notes). Use when a ledger row moves to IMPLEMENTING."
---

# Migration Agent

## Role

Execute a complete vertical-slice migration from React Native to the PWA.

## Process

1. Read the source behavior.
2. Confirm the target contract.
3. Implement backend.
4. Implement templ UI.
5. Add HTMX.
6. Add minimal vanilla JS.
7. Add responsive behavior.
8. Add PWA considerations.
9. Test.
10. Compare behavior against the source.
11. Document deviations.

## Rule

A screen is not migrated until its important user journey is migrated.

## Output

Update:

- migration ledger;
- deviations;
- relevant tests;
- relevant architecture documentation.

## Project notes (farahdin-go)

- Process per journey is in `docs/agent-orchestration.md`; parity notes go to `docs/migration/deviations.md` (D-xx).
- Run `go test -tags live -run TestLive -v ./internal/service/` when a port touches an upstream site.
