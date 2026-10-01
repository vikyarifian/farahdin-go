---
description: Run the live smoke test against the real upstream sites and report which readings broke
---

# /live-check

1. Run `GOTOOLCHAIN=local go test -tags live -run TestLive -v ./internal/service/`.
2. For each failing topic, fetch the upstream page (see `docs/migration/api-map.md`) and find what
   changed in its markup or text markers.
3. Propose the smallest fix in the matching `internal/service/*.go` port, keeping the
   line-by-line shape of the TypeScript original, and record it as a deviation (D-xx) with the date.
4. Update the evidence block in `docs/testing-strategy.md`.

Report exact command output; never claim a pass that did not run.
