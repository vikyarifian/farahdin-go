---
description: Audit the repository against the migration rules (stack, security, parity, ledger)
---

# /audit

Audit the repository against the migration rules.

## Check

- forbidden dependencies (`go.mod` beyond ADR 0005; anything from npm shipped to the browser);
- framework violations (no React/Vue/Alpine/jQuery, no Go web framework);
- architecture drift from `docs/architecture.md` (handler → service → repository/external → domain);
- untested routes (compare `handler.Routes()` with `internal/http/handler/handler_test.go`);
- missing validation;
- authorization gaps (every non-public route wrapped by `RequireUser`);
- unsafe PWA caching (`internal/http/handler/pwa.go`);
- inline `<script>`/`style=""` that the CSP would block;
- duplicated business logic (rules in templates or `app.js`);
- hand-edited generated files (`*_templ.go`, `matrix_formulas.go`, `matrix_chart.templ`, `app.css`);
- undocumented behavior differences;
- migration ledger accuracy.

## Output

Group findings by severity and provide exact file locations.
