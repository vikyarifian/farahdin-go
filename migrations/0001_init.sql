-- Ports convex/schema.ts to PostgreSQL. Convex document ids are kept in
-- convex_id so data imported from a Convex export stays traceable
-- (docs/migration/data-map.md).

CREATE TABLE users (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    convex_id   TEXT UNIQUE,
    username    TEXT NOT NULL,
    fullname    TEXT NOT NULL,
    email       TEXT NOT NULL,
    birthday    TEXT,            -- "YYYY-MM-DD" as in Convex; legacy values like "NaN-NaN-NaN" are kept as-is
    birthplace  TEXT,
    gender      TEXT,
    zodiac      TEXT,
    clerk_id    TEXT,            -- legacy identity (Clerk); kept for imported users
    google_sub  TEXT UNIQUE,     -- new identity (Google OIDC "sub"), ADR 0004
    image_url   TEXT NOT NULL DEFAULT '', -- profile photo, formerly read from Clerk
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX users_by_clerk_id ON users (clerk_id);
CREATE INDEX users_by_email ON users (email);

CREATE TABLE categories (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    convex_id  TEXT UNIQUE,
    code       TEXT NOT NULL,
    name_id    TEXT NOT NULL,
    name_en    TEXT NOT NULL
);
CREATE INDEX categories_by_code ON categories (code);

-- inboxes is not read or written by any screen of the source app (the Inbox
-- tab shows the creator page); the table is kept so no data is lost.
CREATE TABLE inboxes (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    convex_id    TEXT UNIQUE,
    user_id      BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    category_id  BIGINT NOT NULL REFERENCES categories (id),
    message_en   TEXT NOT NULL,
    message_id   TEXT NOT NULL,
    date         TEXT NOT NULL   -- free text from toLocaleDateString() in the source
);
CREATE INDEX inboxes_by_inbox_date ON inboxes (date);
CREATE INDEX inboxes_by_user_id ON inboxes (user_id);

-- Server-side sessions replace Clerk's client-held tokens (ADR 0004).
-- Only a SHA-256 hash of the cookie token is stored.
CREATE TABLE sessions (
    token_hash  TEXT PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL
);
CREATE INDEX sessions_by_user_id ON sessions (user_id);
CREATE INDEX sessions_by_expires_at ON sessions (expires_at);
