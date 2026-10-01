# Migration Ledger

Updated 2026-10-01. "Tests" lists automated coverage; "Parity" is the comparison with the React
Native app. Evidence is in `docs/testing-strategy.md`.

| ID | Journey | RN Screens | APIs | PWA Routes | Backend | UI | Tests | Parity | Owner | Status |
|---|---|---|---|---|---|---|---|---|---|---|
| MIG-001 | Sign in / sign out | login, InitialLayout, profile | Clerk SSO, `createUser` webhook | `/login`, `/auth/google/*`, `/auth/dev`, `/logout` | done | done | auth flow vs fake Google, session lifecycle, CSRF | real Google sign-in not yet run (needs OAuth client) | backend | TESTING |
| MIG-002 | Home + language | (tabs)/index | `getUser` | `/`, `/settings/language` | done | done | handler tests | reviewed in browser screenshots | frontend | REVIEW |
| MIG-003 | Profile + Edit Profile | (tabs)/profile, pages/edit-profile | `getUser`, `updateUser` | `/profile`, `/profile/edit`, `POST /profile`, `/profile/zodiac` | done | done | service + handler tests | D-05 | backend | REVIEW |
| MIG-004 | Settings | pages/settings | – | `/settings` | done | done | handler test | – | frontend | REVIEW |
| MIG-005 | Creator page | (tabs)/inbox | – | `/creator` (`/inbox` redirects) | – | done | `TestCreatorRoute` | static content identical | frontend | REVIEW |
| MIG-006 | Primbon (9 topics) | pages/primbon | primbon.com ×9, Translate | `/primbon[/{1..9}]` | done | done | service tests; live run 2026-10-01: all 9 topics returned readings | D-03, D-05, D-08, Q-02, Q-03 | migration | REVIEW |
| MIG-007 | Horoscope (6 topics) | pages/horoscope | horoscope.com ×5, californiapsychics | `/horoscope[/{1..6}]`, `/horoscope/sign` | done | done | service tests; live run 2026-10-01: all 6 topics | D-06, D-07 | migration | REVIEW |
| MIG-008 | Tarot (4 topics) | pages/tarot | horoscope.com tarot ×4 | `/tarot[/{1..4}]`, `/pick`, `/read` | done | done | service + handler tests; live run 2026-10-01: all 4 topics | D-14, D-18 | migration | REVIEW |
| MIG-009 | Clairvoyance (4 topics) | pages/clairvoyance | horoscope.com tarot-daily/gems | `/clairvoyance[/{1..4}]` | done | done | live run 2026-10-01: all 4 topics | D-07 | migration | REVIEW |
| MIG-010 | Matrix Destiny | pages/matrix-destiny | matrixdestinychart.com ×2 | `/matrix-destiny` | done | done | formulas checked against the RN expressions for 10 dates; live reading run (EN + ID) | D-04, D-15, Q-04 | migration | REVIEW |
| MIG-011 | PWA platform | – | – | `/manifest.webmanifest`, `/sw.js`, `/offline` | done | done | handler tests | Lighthouse/installability not yet run in a browser | pwa | TESTING |
| MIG-012 | Data migration | – | Convex export | `cmd/import-convex` (PostgreSQL) | done | – | importer test with a synthetic export (PostgreSQL) | real Convex export not yet imported; local SQLite → PostgreSQL done 2026-10-02 | devops | TESTING |

## Status Values

- TODO
- DISCOVERY
- CONTRACT_READY
- IMPLEMENTING
- TESTING
- REVIEW
- DONE
- BLOCKED

## Completion Rule

A row becomes `DONE` only after the complete journey passes functional regression testing.
For REVIEW rows the remaining step is a side-by-side run against the React Native app and the
reviewer agent's `APPROVE`.
