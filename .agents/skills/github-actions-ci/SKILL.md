---
name: github-actions-ci
description: rnvim's GitHub Actions CI/CD workflow and Dependabot config. Use when modifying .github/workflows, adding a CI job or step, or diagnosing a CI failure.
---

# GitHub Actions CI/CD

## Current workflow (`.github/workflows/ci.yml`)

Two jobs, both `runs-on: ubuntu-latest`:

### `test` ("Build, Vet & Test") — push/PR to `main`, and on `release`

1. `actions/checkout@v7`
2. `actions/setup-go@v7` with `go-version-file: go.mod` (tracks `go.mod`'s
   `go` directive — see `.agents/skills/go-mod-dependencies`) and `cache: true`
3. `actions/setup-node@v5` + `npm install -g @devcontainers/cli` — needed so
   `internal/devcontainer`'s docker-backed test (which drives a real
   `devcontainer up`) actually runs in CI instead of skipping
4. `gofmt -l .` — fails the build if it prints any file names
5. `go vet ./...`
6. `go build ./...`
7. `go test ./... -race -cover` — ubuntu-latest ships docker and an OpenSSH
   client, so the container/ssh integration tests in `internal/transport`,
   `internal/nvimsetup`, and `internal/devcontainer` actually run here (they
   self-skip only when a dependency is genuinely missing — see
   `internal/dockertest` and `.agents/skills/go-development`)

`permissions: contents: read` at the workflow level — the minimum to check
out the repo.

### `assets` ("Upload Assets") — only `if: github.event_name == 'release'`, `needs: test`

Cross-builds `rnvim-<goos>-<goarch>[.exe]` for linux/darwin/windows ×
amd64/arm64 (`CGO_ENABLED=0`), compresses (`tar.gz` or `.zip` for Windows),
and uploads each as a release asset via `softprops/action-gh-release@v3`
(needs `permissions: contents: write`). This is rnvim's CD: there's no
separate `cd.yml` — cutting a GitHub Release *is* the release process.

This is exactly the sequence `.agents/skills/pre-pr-check` runs locally for
the `test` job — the two must stay in sync. If you change one, change the
other.

## Dependabot (`.github/dependabot.yml`)

Two weekly update checks: `gomod` at `/` (grouped into one PR via
`groups.go-dependencies.patterns: ["*"]`, capped at 10 open PRs) and
`github-actions` at `/` (keeps `actions/checkout`, `actions/setup-go`,
`actions/setup-node`, `softprops/action-gh-release` current).

## Adding a CI step

- Put a new check where it fails fastest: formatting/linting before build,
  build before test.
- If it needs another external tool (like the `devcontainer` CLI), install it
  explicitly as its own step with a clear name — don't assume `ubuntu-latest`
  has it preinstalled without checking the [runner image
  manifest](https://github.com/actions/runner-images).

## Diagnosing a CI failure

1. Reproduce locally first with the exact same commands
   (`.agents/skills/pre-pr-check`) — everything CI does here is reproducible
   without pushing, including the docker/ssh integration tests (this repo's
   dev machines have Docker Desktop; CI has native Docker).
2. `gofmt -l .` failures: run `gofmt -w .`, don't hand-fix whitespace.
3. `go vet` failures: fix the underlying issue; don't suppress.
4. A docker/ssh-backed test failing in CI but not locally (or vice versa)
   usually means an image pull/network flake or a missing `devcontainer` CLI
   step rather than the test logic itself — check the step's raw output
   before assuming the assertion is wrong. (Windows/git-bash used to be a
   third source of this via the `devcontainer` shim; see `AGENTS.md`'s note
   on `devcontainerCommand` — that's fixed now, not an open caveat.)
5. `gh run list` / `gh run view --log-failed` inspects a run that already
   happened on GitHub without leaving the terminal.
