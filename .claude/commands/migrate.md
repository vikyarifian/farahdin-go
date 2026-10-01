---
description: Orchestrate one React Native → PWA migration unit (a ledger row) end to end
argument-hint: "[MIG-id]"
---

# /migrate

Use this command to orchestrate one React Native → PWA migration unit. Target: $ARGUMENTS
(if empty, pick the first ledger row that is not DONE).

## Procedure

1. Read `CLAUDE.md`.
2. Read `docs/agent-orchestration.md`.
3. Read `docs/migration/ledger.md`.
4. Select the journey.
5. Ask the `analyst` agent to verify source behavior in `../farahdin-react-native`.
6. Ask the `architect` agent to verify target boundaries if needed (new ADR if cross-cutting).
7. Implement backend and frontend as one vertical slice (`backend`, `frontend` or `migration` agent).
8. Add PWA behavior if applicable (`pwa` agent).
9. Run QA (`qa` agent): `go test ./...`, live tests if an upstream port changed, browser check.
10. Run security review for auth/data-sensitive features (`security` agent).
11. Run reviewer (`reviewer` agent).
12. Update ledger, deviations and testing evidence.

## Completion

Do not mark the migration unit done until QA and review have evidence.
