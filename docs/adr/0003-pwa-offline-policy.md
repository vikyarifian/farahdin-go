# ADR 0003 — PWA Offline Policy

## Status

Accepted

## Decision

Offline support is feature-specific. Static application assets may be cached aggressively, while authenticated or mutable business data defaults to network-first or no-cache behavior.

## Rationale

Caching sensitive or mutable application data without synchronization semantics can create stale-data and authorization problems.

## Consequences

Every migrated journey must explicitly classify its offline behavior.
