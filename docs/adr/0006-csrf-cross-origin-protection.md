# ADR 0006 — CSRF protection with net/http CrossOriginProtection

## Status

Accepted (2026-10-01)

## Decision

Every state-changing request (POST) passes through Go 1.25's `http.NewCrossOriginProtection()`
(`middleware.CrossOrigin`). It rejects requests whose `Sec-Fetch-Site` is not `same-origin`/`none`,
or, for older browsers, whose `Origin` does not match `Host`. In addition the session cookie is
`SameSite=Lax`, and all state changes are POST.

No per-form CSRF tokens are used.

## Rationale

- Standard library, no token plumbing through templ/HTMX, works for HTMX and plain form posts alike.
- All supported browsers send `Sec-Fetch-Site` or `Origin` on cross-site POSTs.

## Consequences

- Requests without either header (non-browser clients) are allowed; they cannot carry a victim's
  cookies in a CSRF scenario.
- Covered by `TestCrossOriginPostRejected`.
