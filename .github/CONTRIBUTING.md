# Contributing to rnvim

Thanks for your interest in contributing to rnvim! This document covers how to set up your
development environment and the guidelines for submitting changes.

## Development setup

Go 1.26.5 or later is required (see `go.mod`).

```sh
git clone https://github.com/<your-fork>/rnvim.git
cd rnvim
go build ./...
```

## Build, test, and run

```sh
go build ./...               # build
go vet ./...                  # static analysis
go test ./... -race -cover    # tests
gofmt -l .                     # formatting check (fix with gofmt -w .)
go run . connect --help        # run the CLI
```

These are the same checks run in CI (`.github/workflows/ci.yml`), so make sure they pass locally
before opening a PR.

### Docker/ssh-backed tests

`internal/transport`, `internal/nvimsetup`, and `internal/devcontainer` have integration tests
that spin up real containers instead of mocking the transport layer:

- `internal/transport`: runs the same suite (`Run`, `CopyTo`, `MkdirAll`, `Home`, `Uname`) against
  a real docker/podman container and against a real sshd hosted in
  `ghcr.io/linuxserver/openssh-server` (an ephemeral ssh keypair is generated per run).
- `internal/nvimsetup`: `TestEnsure_InstallStrategy_Docker` installs nvim into a live alpine
  container via `--strategy install` and checks the config was copied.
- `internal/devcontainer`: `TestUpAndGetContainerID_Docker` runs a real `devcontainer up`
  (requires the `devcontainer` CLI: `npm i -g @devcontainers/cli`).

Each of these skips itself (not fails) when its dependency isn't available: no `docker`, no
`podman` daemon, no `ssh-keygen`/`ssh`/`scp`, no `devcontainer` CLI, or `go test -short`. Use
`go test -short ./...` to run only the fast, non-container tests.

`internal/dockertest` holds the shared helpers (`RequireDocker`, `RunContainer`, `HostPort`, ...)
these tests are built on.

## Before submitting a pull request

- `gofmt -l .` reports no differences
- `go vet ./...` passes with no warnings
- `go test ./... -race -cover` passes
- New behavior is covered by tests
- Commit messages clearly describe the change

## Code layout

- `main.go` — entry point; builds and executes the root command from `internal/cmd`.
- `internal/cmd` — the cobra command tree (`rnvim connect devcontainer|docker|podman|ssh`), one
  file per subcommand (`docker`/`podman` share `container.go`).
- `internal/transport` — talks to a target: `ContainerTransport` (shared by docker/podman via
  their CLI) and `SSHTransport` (shells out to the system `ssh`/`scp`).
- `internal/nvimsetup` — ensures nvim is available on the target, implementing the
  install-then-push-fallback strategy (`--strategy auto|install|push`).
- `internal/devcontainer` — injects a neovim feature into `devcontainer.json`, drives
  `devcontainer up`, and attaches via `internal/transport`.
- `internal/dockertest` — test-only helpers for spinning up real docker/podman containers.

## Reporting issues

Please include reproduction steps, expected behavior, actual behavior, and your OS/Go version.
