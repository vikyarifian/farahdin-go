# ADR 0001 — Target Stack

## Status

Accepted

## Decision

The PWA rewrite uses:

- Go standard library;
- templ;
- HTMX;
- vanilla JavaScript;
- Tailwind CSS;
- browser PWA APIs.

## Rationale

The target favors server-rendered HTML, simple operational characteristics, minimal client-side state, and a small dependency surface.

## Consequences

Positive:

- simpler frontend runtime;
- server-side business logic;
- progressive enhancement;
- reduced JavaScript complexity.

Trade-offs:

- some highly interactive features require explicit vanilla JS;
- teams must understand HTMX response boundaries;
- complex offline synchronization requires deliberate architecture.
