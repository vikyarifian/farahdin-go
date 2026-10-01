# API Map

Every request the source app makes, and where the PWA makes it now. All third-party calls moved
from the phone to the server (ADR 0007). Go ports live in `internal/service/*.go`.

## Convex (source backend, retired)

| Function | Kind | Args | Behaviour | PWA equivalent |
|---|---|---|---|---|
| `users.getUser` | query | – | current user by Clerk subject (see inventory U-01) | `middleware.Authenticate` → `CurrentUser(ctx)` |
| `users.updateUser` | mutation | username, email, fullname, gender, birthday, birthplace, zodiac | patches the current user; no validation | `service.Profiles.UpdateProfile` (username/email ignored, zodiac derived) and `SyncBirthday` |
| `users.createUser` | mutation | username, fullname, email, clerkId | by clerkId → no-op; by email → patch username/fullname/clerkId; else insert with birthday=today | `service.Profiles.SignIn` |
| `inboxes.getInboxes` | query | – | current user's inboxes, newest first | not used by any screen; not exposed |
| `inboxes.createInbox` | mutation | userId, categoryId, messageEn, messageId, date | inserts for the current user (ignores `userId`, uses `toLocaleDateString()`) | not used by any screen; not exposed |
| `POST /clerk-webhook` | HTTP | Svix-signed Clerk event | `user.created` → createUser | removed (ADR 0004) |

## Google Translate (unofficial)

`GET https://translate.googleapis.com/translate_a/single?client=gtx&sl={from}&tl={to}&dj=1&dt=t&ie=UTF-8&q={text}`
→ `{"sentences":[{"trans": "..."}]}`; one result line per sentence. `&` in text → "dan" (from id) / "and".
PWA: `external.Translator`; `q` is now URL-encoded (source interpolated raw text).

## primbon.com

Base for GET: `https://www.primbon.com`; base for POST: `https://primbon.com` (as in source). Content is read with
`$('#body').text()` (includes the AdSense `push({})` script text, which is then removed).

| Topic | Request | Parsing summary |
|---|---|---|
| 1 Arti Nama | `GET /arti_nama.php?nama1={name}&proses=+Submit%21+` | text before "Nama:", drop "ARTI NAMA", 32 spaces, ads, first newline |
| 2 Tafsir Mimpi | `GET /tafsir_mimpi.php?mimpi={translated}&submit=+Submit+` | not found if `#body > font > i` contains "Tidak ditemukan"; else text between "Hasil pencarian untuk kata kunci: {q}" and "Solusi - …" |
| 3 Jodoh | `GET /kecocokan_nama_pasangan.php?nama1&nama2&proses=+Submit%21+` | `#body > img[src]` → love meter; positive/negative sides from text and `#body` inner HTML |
| 4 Tanggal Jadi | `GET /tanggal_jadian_pernikahan.php?tgl&bln&thn&proses=+Submit%21+` | text after ads snippet (fallback: whole text, D-08), split on ". " |
| 5 Ramalan Jodoh | `POST /ramalan_jodoh.php` nama1,tgl1,bln1,thn1,nama2,tgl2,bln2,thn2,submit=" RAMALAN JODOH &#62;&#62; " | numbered list reformatting |
| 6 Rejeki Weton | `POST /rejeki_hoki_weton.php` tgl,bln,thn,submit=" Submit! " | text + `#body > span > img[src]` chart |
| 7 Kecocokan Nama | `POST /kecocokan_nama.php` nama,tgl,bln,thn,kirim=" Submit! " | text |
| 8 Hari Baik | `POST /petung_hari_baik.php` tgl,bln,thn (birthday),submit=" Submit! " | text |
| 9 Hari Larangan | `POST /hari_sangar_taliwangke.php` tgl,bln,thn (birthday),kirim=" Submit! " | head + "Referensi" section |

## horoscope.com

| Use | Request | Parsing |
|---|---|---|
| Horoscope 1–5 | `GET /us/horoscopes/general/horoscope-general-{daily-yesterday,daily-today,daily-tomorrow,weekly,monthly}.aspx?sign={1..12}` | `.main-horoscope > p` text, promo cuts, text after first " - " (not weekly) |
| Tarot Love / Angel / Past Lives | `POST /us/tarot/{tarot-daily-love,tarot-angel,tarot-past-lives}.aspx` `CardNumber_1_numericalint` | `.grid` text between section markers |
| Tarot True Love | `POST /us/tarot/tarot-true-love.aspx` `CardNumber_1_numericalint`, `CardNumber_2_numericalint` | `.grid` text after "Your Reading", promo cuts |
| Clairvoyance | `POST /us/tarot/{tarot-daily,tarot-gems}.aspx` `CardNumber_{1,2,3}_numericalint`, `FirstName` | `.grid` text between markers, topic section |

## californiapsychics.com

`GET /blog/zodiac-sign-compatibility-blog/{sign}-{partner}-compatibility.html` (lower-case signs);
`</p>` → newline, whole-document text, third "LJ Innes" segment before "Matters of the".

## matrixdestinychart.com

1. `GET /` → nonce from `var ajax_var = {...};`
2. `POST /wp-admin/admin-ajax.php` `yourname`, `yourbirthday=DD/MM/YYYY`, `action=matrix_calc`, `postid=63`, `chartid=0`, `nonce_code` → JSON with `infg1..infg8` (`-1` when rejected).

## PWA HTTP routes

| Method | Path | Auth | Response |
|---|---|---|---|
| GET | `/` | session → home; else 303 `/login` | page |
| GET | `/login` | public | page |
| GET | `/auth/google/start`, `/auth/google/callback` | public | redirects |
| POST | `/auth/dev` | public, dev only | redirect |
| POST | `/logout` | public | 303 / HX-Redirect `/login` |
| POST | `/settings/language` | public | 303 `next` / `HX-Refresh` |
| GET | `/profile`, `/profile/edit`, `/settings`, `/creator` | user | page (`/inbox` → 301 `/creator`) |
| POST | `/profile` | user | form fragment (HTMX) or page; 422 on validation |
| GET | `/profile/zodiac?birthday=` | user | fragment |
| GET | `/{primbon,horoscope,tarot,clairvoyance}` | user | topic list |
| GET | `/{primbon,horoscope,clairvoyance}/{topic}`, `/tarot/{topic}`, `/matrix-destiny` | user | form page |
| POST | `/{primbon,horoscope,clairvoyance}/{topic}`, `/matrix-destiny` | user | result fragment (HTMX) or page |
| GET | `/horoscope/sign?birthday=` | user | out-of-band fragment |
| POST | `/tarot/{topic}/pick`, `/tarot/{topic}/read` | user | sheet fragment or page |
| GET | `/offline`, `/manifest.webmanifest`, `/sw.js`, `/healthz`, `/static/*` | public | PWA / ops |

Signed-out HTMX requests get `401` + `HX-Redirect: /login`.
