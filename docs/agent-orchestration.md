# Agent Orchestration

> Project: `farahdin-go`. Source app: `../farahdin-react-native`. Current status: `docs/migration/ledger.md`.

## Goal

Coordinate Claude Code agents so the React Native application can be rewritten incrementally into a Go + templ + HTMX + vanilla JS + Tailwind PWA.

## Agent Topology

```text
                        ┌──────────────────┐
                        │ Orchestrator     │
                        │ / Lead           │
                        └────────┬─────────┘
                                 │
          ┌──────────────────────┼──────────────────────┐
          │                      │                      │
          ▼                      ▼                      ▼
   Recon / Analyst        Architecture             Migration
          │                      │                      │
          └──────────────┬───────┴──────────────┬───────┘
                         ▼                      ▼
                    Backend                 Frontend
                         │                      │
                         └──────────┬───────────┘
                                    ▼
                                  QA
                                    │
                           ┌────────┴────────┐
                           ▼                 ▼
                       Security          Reviewer
```

## Agent Responsibilities

| Agent | Responsibility | Primary Output |
|---|---|---|
| orchestrator | plan, delegate, integrate | migration plan/status |
| analyst | reverse-engineer React Native behavior | inventory + behavior map |
| architect | define target architecture | architecture + ADRs |
| backend | Go HTTP/domain/data implementation | Go code + tests |
| frontend | templ/HTMX/JS/Tailwind implementation | UI + browser behavior |
| migration | execute feature migration | migrated vertical slices |
| pwa | manifest/service worker/offline/installability | PWA infrastructure |
| qa | functional/regression/accessibility testing | test reports |
| security | auth/input/CSRF/session/security review | security findings |
| reviewer | architecture/code review | review report |
| devops | build/deploy/runtime automation | CI/CD + deployment docs |

## Parallelization Rules

Safe to run in parallel:

- analyst + architect discovery;
- backend + frontend after contracts are stable;
- QA test design while implementation proceeds;
- security review after authentication boundaries exist.

Do not parallelize when two agents modify the same files or package without an explicit ownership boundary.

## Ownership

Use package/file ownership to avoid conflicts.

```text
internal/domain/, internal/service/, internal/repository/,
internal/auth/, internal/http/, internal/external/,
internal/config/, migrations/, cmd/                 backend
web/templates/, web/static/js/app.js, web/static/css/  frontend
internal/http/handler/pwa.go, web/static/icons/,
web/static/js/app.js (service-worker section)       pwa
*_test.go, internal/service/live_test.go            qa
Makefile, Dockerfile, package.json, .env.example    devops
docs/                                               orchestrator/architect
```

Generated files are never edited by hand: `web/templates/**/*_templ.go` (run `templ generate`),
`web/static/css/app.css` (run `npm run css`), `internal/domain/matrix_formulas.go`,
`web/templates/pages/matrix_chart.templ` (generated from the React Native source).

## Migration Unit

A migration unit is one complete user journey.

Example:

```text
Login
  ↓
Dashboard
  ↓
Open item
  ↓
Edit item
  ↓
Save
  ↓
Confirmation
```

Do not declare a feature migrated because only its screen exists.

## Status Format

Every agent reports:

```markdown
## Status
- State: DONE | IN_PROGRESS | BLOCKED
- Scope:
- Files changed:
- Tests:
- Risks:
- Open questions:
- Next recommended action:
```

## Blocking Rules

If an agent encounters an unknown API contract, database behavior, or business rule:

1. stop the risky implementation;
2. record the ambiguity;
3. inspect existing source/tests;
4. ask the orchestrator to resolve it;
5. do not invent behavior.

## Migration Ledger

Maintain `docs/migration/ledger.md`.

Each feature gets:

- source screen(s);
- source API(s);
- target route(s);
- status;
- behavior parity;
- test status;
- known differences;
- owner.

## ADR Rule

Architectural decisions that affect multiple agents require an ADR under `docs/adr/`.

Examples:

- session architecture;
- database access;
- HTMX conventions;
- PWA offline policy;
- file upload strategy;
- authorization model.
