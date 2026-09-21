---
name: go-development
description: Go language conventions used in rnvim (error handling, testing, comments). Use when writing or reviewing Go code anywhere in this repository.
---

# Go development conventions

rnvim is a small, dependency-light Go CLI (module `github.com/ro80t/rnvim`,
Go 1.26.5, one direct dependency: `spf13/cobra`). These are the conventions
actually followed in this codebase — match them rather than defaulting to
generic Go style.

## Comments

Default to none. Only add a comment when the *why* isn't derivable from the
code — a hidden constraint, a real bug being worked around, an invariant the
type system can't express. See `.agents/skills/hush` for the full rules and
`AGENTS.md`'s "Design invariants worth knowing" for a catalog of the
non-obvious behavior in this codebase (the stdout/stderr split in
`transport.Run`, `CopyTo`'s "create as copy of contents" semantics, etc.) that
should stay documented at the point they'd surprise someone, not scattered as
comments everywhere.

## Error handling

- Wrap errors with `fmt.Errorf("%s: %w", context, err)` or a sentence that
  says what failed and how to fix it — see `internal/nvimsetup`'s messages
  (they name the exact flag or file to check, not just "operation failed").
- `transport.Transport.Run` returns stdout on success; on failure the
  returned error already has stderr folded in (`internal/transport/util.go`'s
  `runCapture`). Callers should not re-wrap a `Run` error with the same
  output a second time.
- A missing local file/dir is often not an error — see `push()` and
  `install()` in `internal/nvimsetup/nvimsetup.go` treating a missing
  `LocalConfig` as "skip copying config", not a failure.

## Testing

- Prefer a real target over a mock. `internal/transport`, `internal/nvimsetup`,
  and `internal/devcontainer` have tests that spin up a real docker/podman
  container or a real sshd (via `internal/dockertest`) rather than mocking
  `transport.Transport` — see `internal/transport/container_test.go` and
  `ssh_test.go`, which run the exact same `testTransportSuite` against both.
  Only fall back to a fake (`internal/nvimsetup/fake_transport_test.go`) for
  branches that are awkward or slow to hit for real (e.g. an architecture
  mismatch).
- Every docker/ssh-backed test must skip (not fail) when its dependency is
  missing: use `dockertest.RequireDocker`/`RequireDaemon`/`RequireBin`, which
  also honor `go test -short`. Never make CI depend on something that isn't
  guaranteed to exist (a real network, a specific package manager, `podman`).
- Use `t.TempDir()` for filesystem isolation and `t.Cleanup(...)` for teardown
  (removing a container, deleting a key) rather than manual defer/cleanup
  bookkeeping.
- New behavior needs test coverage — see `.agents/skills/pre-pr-check`.

## Security

- Never build a remote shell command by concatenating untrusted strings
  without quoting; use `shQuote` (`internal/transport/util.go`) for anything
  embedded in a script passed to `Run`/`RunInteractive`.
- `internal/devcontainer`'s `stripJSONC` is a deliberately narrow comment
  stripper (no trailing-comma tolerance) — don't extend it into a general
  JSON5/JSONC parser; if a devcontainer.json needs more, pull in a real JSONC
  library instead of growing this by hand.

## Style

- `gofmt` is enforced by CI — run `gofmt -w .` before finishing, not by
  hand-formatting.
- `go vet` must be clean.
- Small structs implementing a shared interface (`transport.Transport`'s two
  implementations) over interfaces-for-their-own-sake; don't add an
  abstraction layer for a single implementation.

See also `.agents/skills/pre-pr-check` (verification loop) and
`.agents/skills/go-mod-dependencies` (dependency changes).
