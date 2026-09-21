---
name: cobra-cli
description: Cobra CLI structure in rnvim (internal/cmd). Use when adding flags, subcommands, or touching --help/--version output.
---

# Cobra CLI in rnvim

`main.go` is a two-line entry point: it calls `cmd.NewRoot(version).Execute()`
from `internal/cmd`. All [Cobra](https://github.com/spf13/cobra) wiring lives
in `internal/cmd`, one file per concern:

- `cmd.go` — `NewRoot(version string) *cobra.Command` builds the root `rnvim`
  command (`SilenceUsage: true`, `Version: version`) and attaches `connect`.
- `connect.go` — the `connect` parent command; only wires up its four leaf
  subcommands, no flags of its own.
- `container.go` — `newContainerCmd(bin string)` builds both `docker` and
  `podman` (same CLI surface, parameterized by binary name), plus
  `provisionFlags`, the `--strategy`/`--config`/`--nvim-bin` flag set shared
  by docker/podman/ssh.
- `ssh.go` — `newSSHCmd()`; reuses `provisionFlags` and adds `-p/--port`,
  `-i/--identity`.
- `devcontainer.go` — `newDevcontainerCmd()`; its own `--feature`/`--config`
  flags (no `provisionFlags`, since devcontainer provisioning goes through the
  `devcontainer` CLI instead of `internal/nvimsetup`).
- `config.go` — `defaultLocalConfigDir()`, the shared default for `--config`
  across all four subcommands.

`version` stays a package-level `var` in `main.go` (default `"dev"`),
overridable at build time via `-ldflags "-X main.version=..."` — see the
`assets` job in `.github/workflows/ci.yml`, which doesn't currently set it
(worth doing before it matters for a real release).

## Adding a flag or subcommand

- A flag shared by docker/podman/ssh belongs on `provisionFlags`
  (`container.go`), not duplicated per subcommand.
- A new connect target (a new leaf under `connect`) follows
  `.agents/skills/add-connect-target` — don't wire it directly here without
  reading that first, it also covers the `internal/transport` and
  `internal/nvimsetup`/`internal/devcontainer` side.
- Every subcommand's `RunE` should stay a thin adapter: build a
  `transport.Transport`, call `nvimsetup.Ensure` (or
  `devcontainer.Connect`), then `RunInteractive`. Put actual logic in
  `internal/*`, not in `internal/cmd`.

## Checking it

```sh
go build -o rnvim.exe .
./rnvim.exe --help
./rnvim.exe connect --help
./rnvim.exe connect ssh --help
```

Cobra's own flag/help handling doesn't need app-level tests. `internal/cmd`
currently has no `_test.go` — if you add non-trivial logic to a `RunE` beyond
wiring (rare; it should stay thin), test it there.
