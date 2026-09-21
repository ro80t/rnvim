package transport

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

var envKeyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func shQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// path.Dir, not filepath.Dir: the target is always posix regardless of the
// local OS (e.g. a Windows host still needs "/" parent-dir logic here).
func parentDir(p string) string {
	return path.Dir(p)
}

// ssh has no argv-exec like docker/podman: everything ships as one
// command string, so env exports and args are shell-quoted and joined here.
func buildRemoteCommand(env map[string]string, args []string) string {
	var sb strings.Builder
	for k, v := range env {
		// export's left-hand side must be a bare identifier, so it can't be
		// shell-quoted like a value; reject anything else instead of risking
		// an env key that breaks out of the export statement.
		if !envKeyPattern.MatchString(k) {
			panic(fmt.Sprintf("transport: invalid env var name %q", k))
		}
		fmt.Fprintf(&sb, "export %s=%s; ", k, shQuote(v))
	}
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = shQuote(a)
	}
	sb.WriteString(strings.Join(parts, " "))
	return sb.String()
}

// Captures stdout/stderr separately, not combined: ssh's "Warning:
// Permanently added ... to the list of known hosts" on stderr once
// corrupted Home/Uname's stdout parsing. Stderr is folded into the error
// on failure.
func runCapture(cmd *exec.Cmd) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return stdout.String(), fmt.Errorf("%w: %s", err, msg)
		}
		return stdout.String(), err
	}
	return stdout.String(), nil
}
