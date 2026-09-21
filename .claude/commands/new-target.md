---
description: Scaffold a new `rnvim connect <kind>` target
argument-hint: <target-name>
---

Add a new rnvim connect target: $ARGUMENTS

Follow the `add-connect-target` skill (`.agents/skills/add-connect-target/SKILL.md`)
step by step: read `internal/transport/ssh.go` and `internal/transport/transport.go`
first, decide whether a new `transport.Transport` implementation is actually
needed (most targets should reuse `ContainerTransport`/`SSHTransport`), wire
the subcommand into `internal/cmd` (see `.agents/skills/cobra-cli`), add
tests per the skill's testing section, update `README.md`'s usage block and
`AGENTS.md`'s code layout if needed, and finish by running the checks from
`/check`.
