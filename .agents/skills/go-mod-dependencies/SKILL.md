---
name: go-mod-dependencies
description: Managing Go module dependencies in rnvim (go.mod/go.sum). Use when adding, updating, or removing a dependency, or investigating a build/version mismatch.
---

# Go module dependencies

rnvim's module is `github.com/ro80t/rnvim` (this path matters: it's what
makes `go install github.com/ro80t/rnvim@latest` work — don't change it
without updating the README's install instructions too). `go.mod` has exactly
one direct dependency: [`spf13/cobra`](https://github.com/spf13/cobra) (see
`.agents/skills/cobra-cli`). `mousetrap` and `pflag` are `// indirect`
(cobra's own dependencies).

Everything else — talking to docker/podman/ssh, copying files, parsing
devcontainer.json — is done with the standard library and by shelling out to
already-installed CLIs (`docker`, `podman`, `ssh`, `scp`, `devcontainer`).
This is deliberate: it keeps the binary small and avoids needing to trust/vet
a growing dependency tree for a tool that already needs to shell out to those
external programs anyway.

## Adding a dependency

1. Ask whether it's actually needed. Before reaching for a module, check
   whether the standard library or an already-required external CLI already
   covers it — e.g. ssh support shells out to the system `ssh`/`scp` instead
   of pulling in `golang.org/x/crypto/ssh`, specifically to inherit the
   user's `~/.ssh/config`/agent/known_hosts for free and avoid re-implementing
   the protocol.
2. `go get <module>@<version>` from the repo root.
3. `go mod tidy` to update `go.sum` and prune unused indirect entries.
4. Run the full check sequence (`.agents/skills/pre-pr-check`).

## Updating a dependency

- `go get <module>@latest` (or a specific version) then `go mod tidy`.
- Dependabot already opens weekly PRs for both `gomod` and `github-actions`
  ecosystems (`.github/dependabot.yml`), grouping Go dependency updates into
  one PR. Prefer reviewing/merging those over manually bumping versions
  unless something is urgent.
- Re-run `.agents/skills/pre-pr-check` after any bump.

## Removing a dependency

1. Remove the import and all usages.
2. `go mod tidy` — don't hand-edit `go.sum`.

## `go.mod`'s `go` directive

`go 1.26.5` pins the toolchain version. CI resolves its Go version from this
file (`go-version-file: go.mod` in both jobs of `.github/workflows/ci.yml`),
so bumping it also bumps what CI builds/tests/releases with — do that
deliberately.

## Checking what changed

`go mod tidy` and `go mod verify` are safe, read-mostly operations. `go list
-m all` shows the full resolved dependency graph; `go mod why <module>` shows
why something transitive is pulled in.
