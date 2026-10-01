# Backend Conventions

## Go

Use idiomatic Go and the standard library.

Prefer:

- small packages;
- explicit dependencies;
- context propagation;
- wrapped errors;
- structured logging;
- table-driven tests.

Avoid:

- global mutable state;
- unnecessary abstractions;
- framework-style dependency injection;
- magic reflection;
- hidden side effects.

## HTTP

Handlers should be thin.

```text
handler
  → service
    → repository
```

## Context

Use request context for cancellation, deadlines, authentication identity, and tracing metadata when appropriate.

Do not store arbitrary application state in context.

## Validation

Validate at trust boundaries.

Never rely only on browser validation.

## Authentication

Authentication design must be documented before implementation.

Session identifiers should be opaque.

Cookies should use appropriate `Secure`, `HttpOnly`, and `SameSite` attributes.

## Authorization

Check authorization server-side for every protected operation.

Do not infer authorization from hidden UI controls.

## Database

Use `database/sql` unless the existing application requires another approved approach.

Use parameterized queries.

Transactions must have clear boundaries.

## Errors

Separate:

- domain errors;
- validation errors;
- authorization errors;
- infrastructure errors.

Do not expose infrastructure details to users.

## Tests

Minimum:

- service unit tests;
- repository integration tests where practical;
- HTTP handler tests;
- critical journey tests.
