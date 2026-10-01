---
description: Review current changes for behaviour parity with the React Native app, security, accessibility and responsiveness
---

# /parity-review

Perform a focused review of the current changes. (Named `parity-review` so it does not shadow the
built-in `/review`.)

## Steps

1. inspect the diff (or the files named in $ARGUMENTS);
2. inspect the relevant source behaviour in `../farahdin-react-native`;
3. inspect target architecture (`docs/architecture.md`, ADRs);
4. inspect tests and run them;
5. check security (`docs/security-checklist.md`);
6. check accessibility (labels, focus, keyboard, contrast);
7. check responsive behavior (390px and desktop);
8. report required changes in the reviewer format (`APPROVE` / `CHANGES_REQUIRED` / `BLOCKED`).

Never approve based only on compilation.
