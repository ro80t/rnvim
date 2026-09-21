package transport

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"
	"strings"
)

// shQuote single-quotes s for embedding in a posix shell command line.
func shQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// parentDir returns the posix parent directory of a remote path.
func parentDir(p string) string {
	return path.Dir(p)
}

// buildRemoteCommand renders env exports followed by a shell-quoted argv,
// for transports (ssh) that must ship a single command string.
func buildRemoteCommand(env map[string]string, args []string) string {
	var sb strings.Builder
	for k, v := range env {
		fmt.Fprintf(&sb, "export %s=%s; ", k, shQuote(v))
	}
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = shQuote(a)
	}
	sb.WriteString(strings.Join(parts, " "))
	return sb.String()
}

// runCapture runs cmd with stdout and stderr captured separately, so a
// remote process's diagnostic/warning output (e.g. ssh's "Warning:
// Permanently added ... to the list of known hosts") never corrupts stdout
// that callers parse (Home, Uname, ...). On failure, stderr is folded into
// the returned error.
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
