# Behavior Map

Observable behaviour of the source app and the PWA target, per journey. Intentional differences
are listed in `deviations.md` (D-xx); open questions are at the end (Q-xx).

---

### Journey: Sign in / sign out (MIG-001)

**Actor:** visitor with a Google account
**Preconditions:** signed out
**Entry point:** app start (`app/index.tsx` → `/(auth)/login`)

**Happy path:**

1. Login shows "Farahdin" (Javassoul) and "Continue with Google" / "Lanjut pakai Google".
2. Google SSO completes; Clerk session becomes active; app replaces route with `/(tabs)`.
3. Clerk webhook `user.created` → `users.createUser` (new accounts only).
4. Sign out (Profile) → `signOut()` → open `/` → guard sends to login.

**Validation:** none (provider-side).

**Errors:**

| Condition | Source behaviour | PWA behaviour |
|---|---|---|
| OAuth cancelled/failed | `console.log`, stays on login | redirect to `/login?error=signin`, message shown |
| Webhook fails | user row missing; screens show "Guest" | n/a — the user is created during the callback |

**PWA target:** `GET /login`, `GET /auth/google/start`, `GET /auth/google/callback`, `POST /logout` (ADR 0004).
**Permissions:** every other screen requires a session (`middleware.RequireUser`).
**Offline behavior:** `ONLINE_REQUIRED`.

---

### Journey: Home and language (MIG-002)

**Entry point:** tab 1.

**Happy path:**

1. Header: greeting by device hour (5–12 morning, 12–17 afternoon, 17–21 evening, else night) + first name (`fullname.split(' ')[0]`) or "Guest".
2. Flag button opens a dropdown with the other language; choosing it stores `lang` and re-renders.
3. Hero image with "Farahdin"; five feature cards open their screens.

**PWA target:** `GET /`; language: `POST /settings/language` (cookie `lang`, default `EN`); greeting refined by `app.js` from the device clock (D-10).
**Offline behavior:** `ONLINE_REQUIRED` (page is personal; offline fallback page shown).

---

### Journey: Profile, Edit Profile, Settings (MIG-003, MIG-004)

**Profile:** photo (Clerk image → Google picture; fallback icon, D-02), "`fullname || 'Guest'`, `age`" and "`gender, ` `zodiac`", links Edit Profile, Settings, Sign out.

**Edit Profile happy path:**

1. Email (read-only), Name, Birthdate (picker, min 1900-01-01), Zodiac (read-only, recomputed when the birthday changes), Location, Gender (Male/Female, default Male).
2. Save → `users.updateUser({username, email, fullname, gender, birthday 'YYYY-MM-DD', birthplace, zodiac})` → `alert('Success')`.

**Validation (source):** none. **PWA:** birthday required, 1900-01-01 … today; gender Male/Female; name ≤ 100, location ≤ 200 chars (D-05). Username and email are not accepted from the client (D-05).

**Errors:**

| Condition | Source | PWA |
|---|---|---|
| mutation throws | `alert('Error: ' + error)` | 500 page (logged) |
| invalid input | n/a | 422, inline field errors |

**Settings:** radio Indonesia/English; selection is applied immediately (PWA: `hx-trigger="change"`, page refresh).

**PWA target:** `GET /profile`, `GET /profile/edit`, `POST /profile`, `GET /profile/zodiac`, `GET /settings`.

---

### Journey: Creator page (MIG-005)

Tab 3 ("inbox") shows "Farahdin — Created by @vikyarifian" and contacts (website, Twitter, Instagram). Static. → `GET /creator` (old `/inbox` redirects, D-25).

---

### Journey: Primbon (MIG-006)

**Entry:** Home → Primbon → topic list (9) → topic form → Generate → result sheet ("Primbon", topic label) → Done.

| # | Topic | Fields shown | Upstream | Language handling |
|---|---|---|---|---|
| 1 | Arti Nama | Name | GET `arti_nama.php?nama1=` | ID raw lines; EN translate id→en |
| 2 | Tafsir Mimpi | Dream | translate dream auto→id, GET `tafsir_mimpi.php?mimpi=` | not-found message fixed text per language |
| 3 | Jodoh | Name, Partner | GET `kecocokan_nama_pasangan.php` | + love meter 1–5 hearts from `ramalan_kecocokan_cinta{N}.png` |
| 4 | Tanggal Jadi | Date | GET `tanggal_jadian_pernikahan.php?tgl&bln&thn` | EN translate |
| 5 | Ramalan Jodoh | Name, Birthdate, Partner, Partner Birthdate | POST `ramalan_jodoh.php` | EN translate |
| 6 | Rejeki Weton | Birthdate | POST `rejeki_hoki_weton.php` | + upstream chart image |
| 7 | Kecocokan Nama | Name, Birthdate | POST `kecocokan_nama.php` | EN translate |
| 8 | Hari Baik | Date (**request uses the birthday**, Q-03) | POST `petung_hari_baik.php` | EN translate |
| 9 | Hari Larangan | Date (**request uses the birthday**, Q-03) | POST `hari_sangar_taliwangke.php` | EN translate |

**Defaults:** name = profile fullname; birthday and date = profile birthday (or today).
**Side effect:** a changed birthday is saved to the profile with its zodiac (PWA: on submit, D-03).
**Validation:** source none; PWA requires the visible name/dream/partner (D-05).
**Loading:** button disabled + spinner (source: ActivityIndicator in the button).
**Errors:** source silently stops loading; PWA shows "Could not get a reading right now…" (D-04).
**Offline behavior:** `ONLINE_REQUIRED`.

---

### Journey: Horoscope (MIG-007)

Topics: Yesterday, Today, Tomorrow, Weekly, Monthly (horoscope.com `?sign=1..12`, Aries=1) and Match (californiapsychics.com `{sign}-{partner}-compatibility.html`).
Header shows the zodiac image, topic, sign and date range; changing the birthday updates them live (PWA: `GET /horoscope/sign`, out-of-band swap).
Result subtitle: "`topic` - `sign`" (+ " & `partnerSign`" for Match).
Daily/monthly take the text after the first " - "; weekly keeps the date range prefix (source parity).
ID: translate en→id. Birthday side effect as in Primbon. `ONLINE_REQUIRED`.

---

### Journey: Tarot (MIG-008)

Topics: Love (22 of 22), True Love (2 of 22), Angel (22 of 22), Past Lives (22 of 30).

1. Topic → deck dealt face down (`tarot.png`); selecting a card raises it 30px.
2. **Pick** → result sheet with the card face (Love) or the angel/past-life illustration (Angel, Past Lives; cards > 22 show the back) and **Read Card**.
3. **Read Card** → POST card to horoscope.com (`tarot-daily-love`, `tarot-angel`, `tarot-past-lives`) → text; Angel/Past Lives now also show the card face.
4. True Love: two placeholders ("your card", "partner's card"); tapping reveals each; **Read** posts both to `tarot-true-love.aspx`; sheet shows both cards labelled "Your Card"/"Partner's Card" (English in both languages, source parity; typo fixed D-18).

Validation: source ignores Pick without a selection; PWA shows "Pick a card first." `ONLINE_REQUIRED`.

---

### Journey: Clairvoyance (MIG-009)

Topics General/Love/Career/Health; field Name (default fullname). Three random cards (1–22) + first name are posted to `tarot-daily.aspx` (General) or `tarot-gems.aspx` (others); Love/Career/Health cut their section ("Career: ", "Wellness: "). ID: translate en→id. `ONLINE_REQUIRED`.

---

### Journey: Matrix Destiny (MIG-010)

Only **Personal** is reachable (topic starts at 1, topic list hidden; Q-04).

1. Name (default fullname), Birthdate (default profile birthday).
2. Validation (`valide`): invalid date; empty name; future date; > 120 years ago; name must match `^[а-яё\- ]*[a-z\- ]*$` (i). Messages concatenated, per language.
3. Chart computed locally: a = reduce(day), b = month, c = reduce(digit sum of year); `reduceNumber` reduces once when > 22. 32 points, 56 age points, 8 purposes, 7 chakras × 3 + reduced totals.
4. Reading: GET matrixdestinychart.com (nonce from `var ajax_var = …;`), POST `admin-ajax.php` (`action=matrix_calc, postid=63, chartid=0`), sections infg1–6, infg8. ID: translate.
5. Result sheet: name (title case) + date of birth, SVG chart, chakra table, personal/society/general/planetary purposes, reading.

Errors: source `alert()`s validation and leaves the spinner running; reading failures are swallowed. PWA: inline validation; chart still shown with a notice when the reading fails (D-04).
**Offline behavior:** `ONLINE_REQUIRED` (computation is server-side).

---

## Behavioral Equivalence

A PWA implementation is equivalent when the user-visible business outcome and important interaction semantics match the source behavior.

UI appearance does not need pixel-level equality unless explicitly required.

## Intentional Differences

Every intentional difference must be recorded in `docs/migration/deviations.md`.

## Open Questions

| ID | Question | Current choice |
|---|---|---|
| Q-01 | `users.getUser` returns `createUser` instead of the user. Does the deployed function differ? | Return the user (intended behaviour) — D-17 |
| Q-02 | Tafsir Mimpi translates the dream with `sl=auto`; short Indonesian words can be mis-detected (e.g. "ular" → Uzbek "they" → "mereka"). Force `sl=id`? | Source parity (`auto`) |
| Q-03 | Hari Baik / Hari Larangan show a "Date" field but send the birthday. Which date is intended? | Source parity (birthday) |
| Q-04 | Matrix "Compatibility" topic exists in code but is unreachable. Needed? | Not migrated |
| Q-05 | `categories`/`inboxes` are unused by the UI. Keep, drop, or build an inbox? | Kept in schema and importer |
| Q-06 | Are the Adobe Garamond font licences valid for web serving? | Served as-is (risk R-04) |
| Q-07 | Login mentions "Terms and Privacy Policy" without links. Where are they? | Text only, as in source |
