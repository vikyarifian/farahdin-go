# Risks

| ID | Risk | Impact | Likelihood | Mitigation |
|---|---|---|---|---|
| R-01 | Scraped sites change their markup | Readings fail for everyone (friendly error) | High (two changes already, D-08) | Literal ports + `external` unit tests; run `scratchpad`-style live smoke test (`docs/testing-strategy.md`) after deploys and on a schedule; fix in one service file |
| R-02 | Upstream sites rate-limit or block the server IP; their terms may forbid automated access | All readings fail | Medium | 20s timeouts; honest User-Agent; consider caching identical public readings (e.g. daily horoscope per sign) — needs a product decision; review the sites' terms |
| R-03 | Unofficial Google Translate endpoint (`client=gtx`) throttles or disappears. **Happened 2026-10-02:** from one server IP, bursts get HTTP 429 (the RN app spread calls over users' phones), so every reading that needed translation failed after a language switch | Translated readings fail | High under load | `external.Translator`: 24h LRU cache, max 4 parallel calls, and 11 endpoints tried in order (primary gtx + 10 backups: gtx and dict-chrome-ex on 5 Google hosts each, then MyMemory, chunked to 450 chars). Google limits per client kind across hosts, so these are 3 independent pools; a 429 cools the whole kind for 60s, other errors cool one endpoint for 20s; if all fail the reading is shown in its original language with a notice (D-32). Long term: the official Cloud Translation API (needs key and budget) |
| R-04 | Font licences (Adobe Garamond Small Caps, Titling; Javassoul) may not allow web embedding | Legal | Unknown (Q-06) | Confirm licences; fallbacks are already declared (EB Garamond is OFL) |
| R-05 | ~~SQLite is single-node~~ | – | – | Resolved by ADR 0008 (PostgreSQL) |
| R-06 | Accounts are linked by email | Wrong link if an email is reused | Low | Only Google-verified emails are accepted |
| R-07 | Convex export ZIP contains personal data | Data leak | Low | `*.zip` git-ignored; delete after import |
| R-08 | Google OAuth flow not yet exercised against real Google | Sign-in could fail on first deploy | Medium | Tested against a fake provider (`internal/auth/auth_test.go`); do a real sign-in on staging before switching users (ledger MIG-001) |
| R-09 | `POSTGRES_URL` has no `sslmode`; if TLS ever fails the driver falls back to plain text over the internet | Credentials and data exposed | Low (TLS works today) | Add `?sslmode=require` |
| R-10 | The app depends on a remote database | Outage when it is unreachable (`/healthz` returns 503) | Medium | Backups and monitoring on the database server |
