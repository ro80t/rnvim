---
name: pre-pr-check
description: Run rnvim's CI checks locally (gofmt, go vet, go build, go test) before considering a change finished or opening a PR. Use proactively after making Go code changes in this repository.
---

# Pre-PR check

Before treating a change to this repository as done, run the same checks CI
runs (`.github/workflows/ci.yml`), in this order:

```sh
gofmt -l .                    # 1. formatting — must print nothing
go vet ./...                   # 2. static analysis — must be clean
go build ./...                 # 3. build — must succeed
go test ./... -race -cover     # 4. tests — must pass
```

If `gofmt -l .` prints file names, run `gofmt -w .` to fix them and re-check
rather than hand-editing whitespace/formatting.

If any step fails, fix the root cause rather than working around it (e.g.
don't silence a `go vet` finding without understanding it, don't skip a
failing test).

## About step 4 on this repo

`go test ./...` here includes real docker/podman container and docker-hosted
sshd integration tests (`internal/transport`, `internal/nvimsetup`,
`internal/devcontainer` — see `.agents/skills/go-development`). They
self-skip when their dependency is missing (no docker daemon, no
`ssh-keygen`, no `devcontainer` CLI), so a clean pass locally without those
tools installed is expected and fine. If you *do* have docker running, prefer
running the full suite at least once before a PR, since that's what CI
actually exercises (CI installs the `devcontainer` CLI specifically so those
tests run there too). Use `go test ./... -short` to skip all of them
deliberately (e.g. for a quick inner-loop check while iterating).

New behavior needs test coverage — prefer exercising a real target
(`internal/dockertest`) over mocking `transport.Transport`; see
`.agents/skills/go-development` for when a fake is the right call instead.
