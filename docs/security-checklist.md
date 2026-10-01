# Security Checklist

Status as of 2026-10-01. ✅ done and tested · ☑ done, not covered by an automated test · ⬜ open.

## Authentication

- ✅ Session model documented (ADR 0004)
- ☑ Password handling — no passwords; Google only
- ✅ Secure cookie attributes: `HttpOnly`, `SameSite=Lax`, `Secure` when `BASE_URL` is https (production requires https)
- ✅ Session expiration defined (`SESSION_TTL`, default 30 days; expired rows purged hourly)
- ✅ Logout invalidates the session server-side (`TestSessionLifecycle`)
- ✅ OAuth `state` and PKCE verified; unverified Google emails rejected (`internal/auth/auth_test.go`)
- ☑ Authentication errors do not leak details (generic "Sign-in failed")
- ✅ Development login refused in production (config guard)

## Authorization

- ✅ Protected routes require authentication (`TestSignedOutRedirects`)
- ✅ Object-level authorization: user id always from the session; no route takes a user id
- ✅ Email/username cannot be changed through the profile form (D-05)
- ☑ UI hiding is not treated as authorization

## CSRF

- ✅ State-changing requests protected by `http.CrossOriginProtection` (ADR 0006, `TestCrossOriginPostRejected`)
- ✅ SameSite policy reviewed (Lax)

## Input

- ✅ Request size limit 1 MB (`middleware.MaxBody`); upstream bodies capped at 5 MB
- ✅ Validation on the server for every form
- ✅ Parameterized SQL only
- ☑ File uploads — none
- ☑ Path traversal — static files from an embedded FS
- ✅ Safe redirect handling: `next` accepts local paths only (`TestLanguageCookie`)
- ☑ Upstream URLs are fixed; user input is only placed in query/form values (URL-encoded)

## Output

- ☑ templ escaping preserved; upstream text is rendered as text (no `templ.Raw`)
- ☑ No raw user HTML
- ☑ No secrets in HTML/JS
- ✅ No internal errors in responses (`TestUpstreamFailureShowsFriendlyError`)

## Headers

- ✅ Content-Security-Policy: `default-src 'self'`, no inline script/style, images from self + primbon.com + googleusercontent.com, `frame-ancestors 'none'`
- ✅ X-Content-Type-Options, Referrer-Policy, X-Frame-Options (`TestSecurityHeaders`)
- ☑ Strict-Transport-Security when `BASE_URL` is https
- ☑ Permissions-Policy (camera, microphone, geolocation off)

## PWA

- ☑ Service worker scope `/`, same-origin GET only
- ✅ Cache does not contain private data (only `/static/*` and `/offline`)
- ☑ Logout sends `Clear-Site-Data: "cache"`
- ☑ Offline data cannot bypass authorization (no offline data)

## Open

- ⬜ Rate limiting of reading endpoints (each request triggers upstream calls) — consider per-session limits before public launch
- ⬜ Real Google sign-in on staging (risk R-08)
