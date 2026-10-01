# PWA Strategy

## Goals

The PWA should provide installability, app-like navigation, fast repeat visits, reliable loading,
honest offline behaviour and safe updates.

## Manifest

`GET /manifest.webmanifest` (`internal/http/handler/pwa.go`), from the source `app.json`:
name/short name "Farahdin", `start_url` and `scope` `/`, `display: standalone`, `orientation: portrait`,
`background_color` and `theme_color` `#231d32`, icons 192/512 (`any`) and 512 maskable
(generated from `assets/images/icon.png`; maskable padded on the brand background).

## Cache Categories

| Resource | Strategy | Reason |
|---|---|---|
| `/static/*?v=<hash>` (CSS, JS, htmx, sprite, fonts, images) | cache-first; immutable HTTP caching (1 year) | URL changes when content changes (`web.Asset`) |
| Precache: `/offline`, app.css, app.js, htmx, sprite, icon-192, 3 WOFF2 fonts (preloaded in `<head>`, `Cache-Control` 30 days) | installed with the service worker | offline fallback renders styled |
| HTML pages | network-only, fallback `/offline` on navigation failure | personal and fresh; never cached |
| HTMX fragments / POSTs | not handled by the service worker | authenticated, mutable |
| `/sw.js`, manifest | `Cache-Control: no-cache` | fast update detection |

## Authentication and Cache

No authenticated response is ever written to the Cache Storage. Signed-in HTML is sent with
`Cache-Control: no-store`. Logout additionally sends `Clear-Site-Data: "cache"`.

## Service Worker

Served from the server at `/sw.js` (root scope) and generated with the current asset version:

- `install`: precache the list above into `farahdin-static-<version>`.
- `activate`: delete other `farahdin-*` caches, `clients.claim()`.
- `fetch`: same-origin GET only; static cache-first; navigations network with offline fallback.
- No business logic.

## Update

- Version = hash of all embedded static files (`web.Version()`); a new deploy with changed assets
  produces a new `sw.js`.
- The new worker waits; `app.js` shows "A new version is available — Reload"; clicking posts
  `SKIP_WAITING`, the page reloads on `controllerchange`.
- Rollback: deploying the previous build produces the previous version string; old caches are cleaned on activate.

## Offline

| Journey | Class | Behaviour offline |
|---|---|---|
| Sign in / out | `ONLINE_REQUIRED` | offline page |
| Home, Profile, Settings, Creator | `ONLINE_REQUIRED` | offline page (pages are personal, not cached) |
| Edit Profile | `ONLINE_REQUIRED` | HTMX shows "No connection…" |
| Primbon, Horoscope, Tarot, Clairvoyance | `ONLINE_REQUIRED` | readings come from upstream sites |
| Matrix Destiny | `ONLINE_REQUIRED` | chart is computed on the server; reading is upstream |

The network banner (`#network-status`) appears while the browser is offline. No offline writes are implemented.

## Not yet verified

Lighthouse PWA audit and the install prompt on Android/iOS (ledger MIG-011).
