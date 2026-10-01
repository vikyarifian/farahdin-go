# ADR 0007 — Readings are fetched and parsed on the server

## Status

Accepted (2026-10-01)

## Context

All five features call third-party websites **from the phone** and scrape the HTML with cheerio
(primbon.com, horoscope.com, californiapsychics.com, matrixdestinychart.com) and translate with
the unofficial `translate.googleapis.com/translate_a/single?client=gtx` endpoint.
A browser cannot do this (CORS), and the business logic must stay out of client JavaScript.

## Decision

- `internal/external` performs the requests (20s timeout, 5 MB cap, honest User-Agent) and offers a
  cheerio-compatible subset: selectors (`tag`, `#id`, `.class`, `>` and descendant), `.text()`
  (including `<script>` text, as cheerio does) and `.html()` serialized like parse5.
- `internal/service` ports each `generate()` **literally**, using a small JS-string pipeline
  (`external.Str(...).Split(sep, i).Replace(...)`) so the Go code can be compared line by line with
  the TypeScript. `split(x)[i]` on a missing index fails the chain, mirroring the source's TypeError.
- Upstream failures are logged and shown as a friendly error instead of being swallowed (deviation D-04).

## Consequences

- Scraping is fragile by nature; upstream markup changes break readings (see risks R-01). Two
  drift fixes were already needed (deviation D-08). `cmd`-level live smoke tests are documented in
  `docs/testing-strategy.md`.
- The server's IP, not the user's phone, now makes the requests: rate limiting or blocking by the
  upstream sites affects everyone (risk R-02).
