---
name: add-connect-target
description: Add a new `rnvim connect <kind>` target (a new way to reach a remote/container). Use when asked to support a new transport (e.g. WSL, a cloud dev-environment API, etc.).
---

# Adding a new connect target

rnvim has four targets today: `devcontainer`, `docker`, `podman`, `ssh`. Adding
a fifth one touches three layers. Read `internal/transport/ssh.go` (the
simplest end-to-end example) before starting.

## 1. Implement `transport.Transport` (if the new target needs one)

`devcontainer` doesn't implement its own `Transport` — it resolves a
container id via `devcontainer up` and then reuses `ContainerTransport`. Most
new targets should do the same: only write a new `Transport` implementation
if the target genuinely can't be reached through `docker`/`podman` exec or
`ssh`/`scp`.

If you do need a new one, implement all five methods
(`internal/transport/transport.go`'s interface): `RunInteractive`, `Run`,
`CopyTo`, `MkdirAll`, `Home`, `Uname`. Two non-obvious requirements, both
covered in `AGENTS.md`'s "Design invariants worth knowing":

- `Run` must return **stdout only** on success, with stderr folded into the
  error on failure (see `internal/transport/util.go`'s `runCapture` — reuse
  it if your transport shells out to a CLI at all).
- `CopyTo(local, remote)` must create `remote` as a copy of `local`'s
  *contents* when `remote` doesn't already exist (matching `docker cp` and
  `scp -r`'s behavior) — `internal/nvimsetup`'s push strategy depends on this.

## 2. Wire it into `internal/nvimsetup` or `internal/devcontainer`

If the target provisions nvim like docker/podman/ssh do (network install,
fall back to push), no new code is needed here — `nvimsetup.Ensure` already
works against any `transport.Transport`. If it needs a genuinely different
provisioning flow (like `devcontainer`'s feature-injection), follow
`internal/devcontainer/devcontainer.go` as the reference.

## 3. Add the subcommand in `internal/cmd`

Follow `.agents/skills/cobra-cli`. If the target is exec/copy-based like
docker/podman/ssh, reuse `provisionFlags` (`container.go`) for
`--strategy`/`--config`/`--nvim-bin`. Wire the new `new<Kind>Cmd()` into
`connect.go`'s `AddCommand(...)` call.

## 4. Add tests

- If you wrote a new `Transport`, add it to `internal/transport`'s shared
  suite: call `testTransportSuite(t, tr)` (`transport_test.go`) against a
  real instance, gated the same way `container_test.go`/`ssh_test.go` are
  (`internal/dockertest.RequireDocker`/`RequireDaemon`/`RequireBin`, so it
  skips instead of failing when unavailable).
- If you touched `internal/nvimsetup` or `internal/devcontainer` logic beyond
  reuse, add fast fake-`Transport` tests (see
  `internal/nvimsetup/fake_transport_test.go`) for branch logic, plus one
  real-target integration test if practical.

## 5. Finish

- Update the `Usage` block and the relevant section of `README.md`.
- Update `AGENTS.md`'s code-layout list if you added a package.
- Run `.agents/skills/pre-pr-check`.
