---
name: conventional-commits
description: Write commit messages in Conventional Commits format for rnvim. Use whenever creating a git commit in this repository.
---

# Conventional Commits

Format: `<type>[optional scope]: <description>`

Pick the type that matches the change; don't invent new ones:

- `feat` — new behavior (a new connect target, a new flag, a new strategy)
- `fix` — bug fix
- `refactor` — code change with no behavior change
- `docs` — README/CONTRIBUTING/AGENTS.md/comments only
- `test` — test-only changes
- `chore` — everything else (deps, CI, tooling)

Optional scope in parentheses when it narrows the change to one area, e.g.
`feat(ssh): add ProxyJump support`, `fix(nvimsetup): handle missing $HOME`.
Skip it when the type alone is clear.

Description: imperative mood, lowercase, no trailing period — "add x", not
"added x" or "adds x".

Breaking change: `!` after the type/scope (`feat!: ...`) or a `BREAKING
CHANGE:` footer, whichever fits the amount of explanation needed.

## Body

Only add a body when the *why* isn't obvious from the diff — same bar as code
comments in this repo (see `AGENTS.md` and `.agents/skills/hush`). A one-line
`feat: add wsl connect target` needs nothing more. Reserve the body for
reasoning a reviewer couldn't get from the diff alone (e.g. why a bug fix
changed a public function's error-return contract).

## What not to do

- Don't title-case or capitalize the description (`feat: Add X` → `feat: add
  x`).
- Don't bundle unrelated changes under one type; split into separate commits.
- Don't use `feat` for a pure internal refactor, even one touching many
  files — that's `refactor`.
