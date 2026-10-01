# React Native Inventory

> Source: `../farahdin-react-native` (inspected 2026-10-01). Facts only; unknowns are listed at the end.

## Application

- Name: farahdin (Expo slug `farahdin`, iOS bundle `com.vikyarifi.farahdin`)
- Version: 1.0.0
- Entry point: `expo-router/entry` → `app/_layout.tsx`
- Build system: Expo SDK 53, EAS (`eas.json`), Metro
- React Native version: 0.79.2 (React 19, new architecture enabled)
- Navigation library: expo-router 5 (file-based), bottom tabs
- State management: local `useState` only; server state through Convex `useQuery`/`useMutation`
- Networking: `fetch` to third-party sites + Convex client
- Storage: AsyncStorage (`lang`), Expo SecureStore (Clerk token cache)
- Authentication: Clerk (`@clerk/clerk-expo`), Google SSO only
- Push notifications: none
- Native modules: date time picker, SVG, secure store, vector icons (no camera/location/etc.)

## Screens

| ID | Screen | Route | Entry Points | API Calls | Local State | Native Dependency | Status |
|---|---|---|---|---|---|---|---|
| S-01 | Login | `/(auth)/login` | app start when signed out | Clerk `startSSOFlow(oauth_google)` | lang | SVG | migrated → `/login` |
| S-02 | Home | `/(tabs)/index` | tab 1, after sign-in | `users.getUser` | lang, 5 modal flags, dropdown | SVG | migrated → `/` |
| S-03 | Profile | `/(tabs)/profile` | tab 2 | `users.getUser`, Clerk `useUser` (photo), `signOut` | lang, 2 modal flags | expo-image | migrated → `/profile` |
| S-04 | Inbox (Creator page) | `/(tabs)/inbox` | tab 3 | none | lang | – | migrated → `/creator` |
| S-05 | Edit Profile (modal) | `pages/edit-profile` | Profile → Edit Profile | `users.updateUser` | form, picker | DateTimePicker | migrated → `/profile/edit` |
| S-06 | Settings (modal) | `pages/settings` | Profile → Settings | none (AsyncStorage) | lang | – | migrated → `/settings` |
| S-07 | Primbon (modal) | `pages/primbon` | Home card | primbon.com ×9, Google Translate, `users.updateUser` | topic, form, result, 2 pickers | DateTimePicker | migrated → `/primbon[/n]` |
| S-08 | Horoscope (modal) | `pages/horoscope` | Home card | horoscope.com ×5, californiapsychics.com, Translate, `users.updateUser` | topic, form, result, 2 pickers | DateTimePicker | migrated → `/horoscope[/n]` |
| S-09 | Tarot (modal) | `pages/tarot` | Home card | horoscope.com tarot ×4, Translate | topic, shuffled deck, selection, result | – | migrated → `/tarot[/n]` |
| S-10 | Clairvoyance (modal) | `pages/clairvoyance` | Home card | horoscope.com tarot-daily / tarot-gems, Translate | topic, name, result | – | migrated → `/clairvoyance[/n]` |
| S-11 | Matrix Destiny (modal) | `pages/matrix-destiny` | Home card | matrixdestinychart.com (nonce + admin-ajax), Translate | name, birthday, ~120 computed values | DateTimePicker, SVG | migrated → `/matrix-destiny` |
| C-01 | Result sheet | `components/Result.tsx` | every feature | – | – | Modal | → `components.ResultSheet` |
| C-02 | Feature modal | `components/Modal.tsx` | every feature | – | – | Modal | → full page (`layouts.Feature`) |

## Navigation

- Root: `app/_layout.tsx` (fonts, Clerk+Convex providers, safe area) → `InitialLayout` (Stack, no headers).
- Guard: `InitialLayout` redirects signed-out users to `/(auth)/login` and signed-in users away from it to `/(tabs)`.
- `app/index.tsx` always redirects to `/(auth)/login` (the guard then forwards signed-in users).
- Tabs (`app/(tabs)/_layout.tsx`): `index` (Entypo home), `profile` (FontAwesome user), `inbox` (MaterialIcons alternate-email); labels hidden.
- Feature screens and Edit Profile/Settings are **modals** (`components/Modal.tsx`), not routes; back arrow in a feature goes from form → topic list → close.
- Deep links: scheme `myapp` declared; no deep-link handling in code.
- Back behaviour: Android back closes the modal (`onRequestClose`).

## Device Features

| Feature | RN Implementation | Required in PWA? | Browser Equivalent | Decision |
|---|---|---|---|---|
| Camera | – | no | – | – |
| Location | – (birthplace is free text) | no | – | – |
| Files | – | no | – | – |
| Push | – | no | – | – |
| Biometrics | – | no | – | – |
| Share | – | no | – | – |
| Date picker | `@react-native-community/datetimepicker` (min 1900-01-01) | yes | `<input type="date" min="1900-01-01">` | native input |
| Secure token storage | expo-secure-store (Clerk) | yes | HttpOnly session cookie | ADR 0004 |
| Local preference | AsyncStorage `lang` | yes | `lang` cookie | cookie |

## State

- Global: none besides providers. Language is re-read from AsyncStorage by every screen.
- Server state: current user (`users.getUser`).
- Persisted: `lang` (AsyncStorage), Clerk session (SecureStore).
- Screen state: topic, form fields, picker visibility, loading, result lines.
- Derived: zodiac from birthday (`utils/Zodiac.ts`), age (`utils/Utils.ts`), matrix values (`matrix-destiny.tsx`).

## API

See `api-map.md` for every request (method, URL, parameters, parsing, errors).

Summary:

- Convex: `users.getUser` (query), `users.updateUser` (mutation), `users.createUser` (mutation, webhook only),
  `inboxes.getInboxes` / `inboxes.createInbox` (not used by any screen), HTTP `POST /clerk-webhook`.
- Third-party: 9 primbon.com, 5 horoscope.com horoscope pages, 4 horoscope.com tarot pages,
  1 californiapsychics.com page, 2 matrixdestinychart.com calls, Google Translate.
- No retries, no pagination, no caching anywhere. Errors are `console.log`ged and the spinner stops.

## Unknowns

- U-01 `users.getUser` returns `createUser` (the mutation object) instead of `currentUser`
  (`convex/users.ts:6`). As written, screens would never receive the user; the deployed Convex
  function may differ from the repository. See behavior-map Q-01.
- U-02 How React Native sent `FormData` with a manually set `application/x-www-form-urlencoded`
  header (RN may override it with multipart). The PWA sends urlencoded; all endpoints accepted it on 2026-10-01.
- U-03 Whether the Matrix "Compatibility" topic was ever reachable (it is not in this code).
