# ADR 0002 — Vertical Slice Migration

## Status

Accepted

## Decision

Migrate by complete user journey rather than migrating all screens, then all APIs, then all components.

## Rationale

A vertical slice exposes integration problems early:

```text
source behavior
→ backend contract
→ Go service
→ database
→ templ
→ HTMX
→ browser
→ PWA
→ tests
```

## Consequences

Each completed migration unit is independently testable and deployable where project constraints allow.
