# ADR 0004 — Sign-in with Google OIDC and server-side sessions

## Status

Accepted (2026-10-01)

## Context

The source app signs users in with **Clerk** (`@clerk/clerk-expo`, `useSSO` with `oauth_google`).
Clerk's `user.created` webhook (`convex/http.ts`) calls `users.createUser`, and Convex resolves
the current user from the Clerk JWT `subject` (`utils/User.ts`).

The target stack is Go standard library + templ + HTMX. Clerk has no Go-stdlib-friendly
server SDK, and its hosted flow relies on client-side JavaScript that this project avoids.
The only sign-in method the source offers is Google.

## Decision

- Replace Clerk with a direct **Google OpenID Connect** authorization-code flow with **PKCE (S256)**,
  written against `net/http` (`internal/auth/google.go`).
  - `state` + PKCE verifier are kept in a 10-minute `HttpOnly`, `SameSite=Lax` cookie scoped to `/auth/google`.
  - The identity comes from Google's `userinfo` endpoint over TLS with the access token; the email must be verified.
- `createUser` semantics are kept in `service.Profiles.SignIn`:
  1. match by provider subject (`users.google_sub`, replaces `by_clerk_id`) → return unchanged;
  2. else match by email (`by_email`, `.first()`) → overwrite `username`, `fullname`, provider id;
  3. else insert with `birthday = today`, `gender = zodiac = birthplace = ''`.
- Sessions are **opaque server-side sessions** (`sessions` table). The cookie `farahdin_session`
  holds 32 random bytes; only its SHA-256 is stored. `HttpOnly`, `SameSite=Lax`, `Secure` when
  `BASE_URL` is https. Default TTL 30 days (`SESSION_TTL`). Logout deletes the row.
- `users.image_url` stores Google's `picture` (the source read `user.imageUrl` from Clerk).
- A development-only sign-in (`POST /auth/dev`, `DEV_LOGIN=true`) exists for local testing and
  is refused by configuration when `APP_ENV=production`.

## Consequences

- Existing users migrate without action: the Convex export is imported (`cmd/import-convex`),
  and their first Google sign-in links them by email, which is what the Clerk webhook did.
- Clerk-only features (other providers, MFA, user management UI) are not available. None were used.
- The `/clerk-webhook` endpoint is retired.
- Requires a Google Cloud OAuth client; redirect URI `${BASE_URL}/auth/google/callback`.
