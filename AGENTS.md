# AGENTS.md

Instructions for AI coding agents working in this repository.

## What this is

rnvim is a Go CLI for quickly connecting a local `nvim` to a devcontainer, a
docker/podman container, or an ssh host, without hand-typing the
exec/install/copy steps every time.

## Build, test, run

```sh
go build ./...               # build
go vet ./...                  # static analysis
go test ./... -race -cover    # tests (fast + docker/ssh integration tests)
go test ./... -short          # fast tests only, skips anything needing docker/ssh
gofmt -l .                     # formatting check (fix with gofmt -w .)
go run . connect --help        # run the CLI
```

These are the same checks CI runs (`.github/workflows/ci.yml`). Run them before
considering a change done — see `.agents/skills/pre-pr-check` for the full
sequence and what each step catches.

## Code layout

- `main.go` — entry point; builds and executes the root command from
  `internal/cmd`.
- `internal/cmd` — the cobra command tree (`rnvim connect
  devcontainer|docker|podman|ssh`), one file per subcommand (`docker`/`podman`
  share `container.go`). See `.agents/skills/cobra-cli`.
- `internal/transport` — talks to a target: `ContainerTransport` (shared by
  docker/podman via their CLI) and `SSHTransport` (shells out to the system
  `ssh`/`scp`). Both implement the `Transport` interface (`transport.go`).
- `internal/nvimsetup` — ensures nvim is available on the target: tries a
  network install via the target's package manager first, falls back to
  pushing a local nvim binary + runtime + config
  (`--strategy auto|install|push`).
- `internal/devcontainer` — injects a neovim feature into `devcontainer.json`,
  drives `devcontainer up` (the `@devcontainers/cli` binary), and attaches via
  `internal/transport`.
- `internal/dockertest` — test-only helpers for spinning up real
  docker/podman containers and a real sshd in tests, instead of mocking the
  transport layer.

## Comment style

Comments in this codebase are intentionally minimal: default to none, only add
one when the *why* isn't derivable from the code itself (a hidden constraint,
a workaround, a non-obvious invariant). Don't restate what a well-named
function or type already says. See `.agents/skills/hush` for the full rules
this follows.

## Design invariants worth knowing

- **`transport.Transport.Run` returns stdout only; stderr is folded into the
  error on failure** (`internal/transport/util.go`'s `runCapture`). This was a
  real bug fix, not a stylistic choice: ssh prints `Warning: Permanently added
  ... to the list of known hosts.` to stderr on a host's first connection, and
  when stdout/stderr were merged that warning corrupted `Uname()`'s parsing.
  Any new `Transport` implementation must keep stdout and stderr separate for
  the same reason.
- **`CopyTo(local, remote)` semantics**: when `remote` doesn't already exist,
  both `docker cp` and `scp -r` create it as a copy of `local`'s *contents*
  (not `remote/<basename of local>`). `internal/nvimsetup`'s `push()` relies on
  this: it copies a local `.../share/nvim` directory straight onto
  `remoteBase+"/share/nvim"` and expects `remoteBase/share/nvim/runtime/...`
  to exist afterward.
- **push strategy requires matching OS/arch** (`checkArchMatch` in
  `internal/nvimsetup/nvimsetup.go`) because it copies the local `nvim` binary
  as-is — no cross-compilation or emulation. This is why a Windows host can't
  push directly into a Linux container/host.
- **`--strategy auto` tries install, then push, and only push is required to
  succeed** — `install` failures are expected (no network, no supported
  package manager, no sudo) and silently trigger the fallback; only `push`
  failing (or `install`/`push` failing when forced explicitly) is a hard
  error. Don't make the install path noisier or more fatal than it already is.
- **`devcontainer` CLI is invoked via `node` directly, bypassing its own
  shim** (`devcontainerCommand` in `internal/devcontainer/devcontainer.go`).
  npm's generated `devcontainer.cmd`/`.ps1`/sh shims all forward to the same
  `<npm-global-dir>/node_modules/@devcontainers/cli/devcontainer.js`; on
  Windows, the `.cmd` shim was confirmed (via a minimal repro) to fail
  outright — `'powershell.exe' is not recognized...` — specifically when
  rnvim itself is run from git-bash/MSYS2, even though the exact same shim
  works fine from cmd.exe/PowerShell. Resolving `devcontainer` via
  `exec.LookPath`, finding its sibling `node_modules/@devcontainers/cli/
  devcontainer.js`, and running that through `node` directly sidesteps the
  shim (and its batch-interpretation quirks) entirely, so behavior is now
  identical regardless of which shell launched rnvim. Falls back to invoking
  `devcontainer` directly if that layout isn't found (e.g. installed some
  other way than npm).
- **No RPC-based remote UI.** Connecting attaches a PTY straight to a remote
  TUI `nvim` process (`exec -it` for docker/podman, `ssh -t` for ssh). There's
  deliberately no `nvim --server`/`--remote-ui` client — that would require
  writing a full terminal UI client in Go, out of proportion to this project.
  `RunInteractive` therefore can't be exercised in an automated test (no real
  TTY in `go test`); it's a thin wrapper over the transport CLI and is
  reviewed by reading, not covered by an automated check.

## Testing

Prefer a real docker/podman container or a real docker-hosted sshd over
mocking the `Transport` interface — see `.agents/skills/go-development` and
`internal/dockertest`. Fall back to a fake `Transport` (see
`internal/nvimsetup/fake_transport_test.go`) only for logic that's expensive
or awkward to exercise for real (e.g. architecture-mismatch branches).

## Adding a new connect target

See `.agents/skills/add-connect-target` and/or run `/new-target <name>`.

See also `.github/CONTRIBUTING.md` for the human-facing version of this.
