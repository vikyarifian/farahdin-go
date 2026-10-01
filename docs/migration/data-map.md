# Data Map

Source: `convex/schema.ts`. Target: `migrations/0001_init.sql` (PostgreSQL 16, ADR 0008).

## users

| Convex field | Type | PostgreSQL column | Notes |
|---|---|---|---|
| `_id` | Id<users> | `convex_id` TEXT UNIQUE | kept for traceability; new users have NULL |
| – | – | `id` BIGINT identity PK | new surrogate key |
| `username` | string | `username` | email local part |
| `fullname` | string | `fullname` | |
| `email` | string | `email` (index `users_by_email`) | not unique in Convex either; lookups take the first |
| `birthday` | string? | `birthday` | `YYYY-MM-DD`; legacy `NaN-NaN-NaN` values are tolerated (treated as unknown) |
| `birthplace` | string? | `birthplace` | shown as "Location" |
| `gender` | string? | `gender` | `Male` / `Female` / '' |
| `zodiac` | string? | `zodiac` | always derived from birthday on write |
| `clerkId` | string (index `by_clerk_id`) | `clerk_id` (index) | legacy identity, kept |
| – | – | `google_sub` TEXT UNIQUE | new identity (ADR 0004) |
| – (Clerk `imageUrl`) | – | `image_url` | Google `picture` |
| `_creationTime` | number | `created_at` | not imported (export time is used); `updated_at` added |

## categories

| Convex | PostgreSQL |
|---|---|
| `_id` | `convex_id` |
| `code` (index `by_code`) | `code` (index) |
| `nameID` | `name_id` |
| `nameEN` | `name_en` |

## inboxes

| Convex | PostgreSQL |
|---|---|
| `_id` | `convex_id` |
| `userId` → users | `user_id` → `users.id` (resolved through `convex_id`), ON DELETE CASCADE |
| `categoryId` → categories | `category_id` → `categories.id` |
| `messageEN` | `message_en` |
| `messageID` | `message_id` |
| `date` (index `by_inbox_date`) | `date` (index) — free text from `toLocaleDateString()` |

## sessions (new)

`token_hash` (SHA-256 of the cookie token), `user_id`, `created_at`, `expires_at` (`TIMESTAMPTZ`). Purged hourly.

## Client-side data

| Source | Target |
|---|---|
| AsyncStorage `lang` (JSON string `"ID"`/`"EN"`) | cookie `lang` (`ID`/`EN`, default `EN`, 1 year) |
| SecureStore Clerk tokens | cookie `farahdin_session` (HttpOnly) |

## Migration procedure

1. In `farahdin-react-native`: `npx convex export --path convex-export.zip` (contains personal data — store securely, delete after).
2. `go run ./cmd/import-convex convex-export.zip` with `POSTGRES_URL` set (idempotent; one transaction).
3. Users sign in with Google; the first sign-in links the imported row by email.

## SQLite → PostgreSQL (2026-10-02)

The first PWA builds stored data in a local SQLite file (`data/farahdin.db`). When moving to
PostgreSQL (ADR 0008) the one real Google account was copied with its profile; two development
accounts and all sessions were left behind (users sign in again). The SQLite file is no longer read.
