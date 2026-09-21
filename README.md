# rnvim

A Go CLI for quickly connecting nvim to a devcontainer, docker, podman, or ssh target.

## Install

```sh
go install github.com/ro80t/rnvim@latest
```

## Usage

```
rnvim connect devcontainer [workspace]   [--feature <ref>] [--config <dir>]
rnvim connect docker <container>         [--strategy auto|install|push] [--config <dir>] [--nvim-bin <path>]
rnvim connect podman <container>         [--strategy auto|install|push] [--config <dir>] [--nvim-bin <path>]
rnvim connect ssh <[user@]host>          [--strategy auto|install|push] [--config <dir>] [--nvim-bin <path>] [-p port] [-i identity]
```

- `--config` defaults to the local `$XDG_CONFIG_HOME/nvim` (`%LOCALAPPDATA%\nvim` on Windows).
- `devcontainer` adds a neovim feature (default `ghcr.io/devcontainers-extra/features/neovim:1`)
  to `.devcontainer/devcontainer.json`'s `features` if it's not already there, then runs
  `devcontainer up`. The file is only rewritten when the feature is missing (comments/formatting
  are lost on rewrite). Requires the `devcontainer` CLI (`npm i -g @devcontainers/cli`).

## nvim provisioning strategy for docker/podman/ssh (`--strategy`)

- `auto` (default): try installing `nvim` over the network first, using the target's package
  manager (apt/dnf/apk/pacman/brew). If that fails, automatically fall back to pushing the local
  `nvim` binary + runtime + config to `/tmp/rnvim` on the target.
- `install`: network install only. Exits with an error if it fails (no fallback).
- `push`: always push the local `nvim` binary + config.

Since `push` copies the local `nvim` binary as-is, **the local and remote OS/CPU architecture
must match** (e.g. you can't push directly from a Windows host into a Linux target — run rnvim
from WSL instead, or point `--nvim-bin` at a binary built for the target's architecture).

## How connecting works

docker/podman use `exec -it` and ssh uses `ssh -t` to attach a PTY straight to the remote nvim
TUI. There's no RPC-based remote UI client (`nvim --server`/`--remote-ui`) — building a real
terminal UI client was out of scope for the effort involved.

## Requirements

- Locally: the `docker`/`podman` CLI, or an OpenSSH client (`ssh`/`scp`). The `devcontainer` CLI
  if you use the `devcontainer` subcommand.
- For the `push` strategy, a local nvim install with the standard layout
  (`<prefix>/bin/nvim` + `<prefix>/share/nvim/runtime`).

## Contributing

See [.github/CONTRIBUTING.md](.github/CONTRIBUTING.md) for development setup and guidelines.
