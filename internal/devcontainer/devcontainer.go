// Package devcontainer drives the official `devcontainer` CLI
// (@devcontainers/cli) to bring up a dev container with a neovim
// "additional feature" declared in devcontainer.json, then attaches nvim
// via the same container transport used for plain docker targets.
package devcontainer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ro80t/rnvim/internal/transport"
)

const DefaultNeovimFeature = "ghcr.io/devcontainers-extra/features/neovim:1"

// devcontainerCommand resolves `devcontainer` and, if it's an npm-generated
// shim (.cmd/.ps1/sh, all forwarding to the same node_modules layout),
// invokes its underlying devcontainer.js via node directly instead of the
// shim. On Windows the .cmd shim has been observed to fail outright when
// rnvim itself is run from git-bash/MSYS2 (its own internal relaunch breaks
// across that process boundary) despite working fine from cmd.exe/
// PowerShell; calling node directly skips the shim and behaves the same
// everywhere. Anything installed a different way (a standalone binary, a
// dev checkout) just runs as before.
func devcontainerCommand(args ...string) (*exec.Cmd, error) {
	binPath, err := exec.LookPath("devcontainer")
	if err != nil {
		return nil, err
	}
	jsPath := filepath.Join(filepath.Dir(binPath), "node_modules", "@devcontainers", "cli", "devcontainer.js")
	if nodePath, err := exec.LookPath("node"); err == nil {
		if _, err := os.Stat(jsPath); err == nil {
			return exec.Command(nodePath, append([]string{jsPath}, args...)...), nil
		}
	}
	return exec.Command(binPath, args...), nil
}

func Connect(workspace, feature, localConfig string) error {
	wsAbs, err := filepath.Abs(workspace)
	if err != nil {
		return err
	}

	if err := ensureFeature(wsAbs, feature); err != nil {
		return fmt.Errorf("update devcontainer.json: %w", err)
	}

	containerID, err := upAndGetContainerID(wsAbs)
	if err != nil {
		return fmt.Errorf("devcontainer up: %w (is @devcontainers/cli installed? try: npm i -g @devcontainers/cli)", err)
	}

	ct := transport.NewContainerTransport("docker", containerID)

	if localConfig != "" {
		if _, statErr := os.Stat(localConfig); statErr == nil {
			home, err := ct.Home()
			if err != nil || home == "" {
				home = "/root"
			}
			if err := ct.CopyTo(localConfig, home+"/.config/nvim"); err != nil {
				fmt.Fprintln(os.Stderr, "rnvim: warning: failed to copy local config into devcontainer:", err)
			}
		}
	}

	return ct.RunInteractive(nil, "nvim")
}

// upAndGetContainerID runs `devcontainer up` and parses the container id
// out of its JSON result line.
func upAndGetContainerID(wsAbs string) (string, error) {
	cmd, err := devcontainerCommand("up", "--workspace-folder", wsAbs, "--log-format", "json")
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &out)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return "", err
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, line := range slices.Backward(lines) {
		var result struct {
			ContainerID string `json:"containerId"`
		}
		if err := json.Unmarshal([]byte(line), &result); err == nil && result.ContainerID != "" {
			return result.ContainerID, nil
		}
	}
	return "", fmt.Errorf("could not determine container id from devcontainer up output")
}

// ensureFeature adds feature to devcontainer.json's "features" map if it's
// not already there. It rewrites the file (losing comments/formatting) only
// when a change is actually needed.
func ensureFeature(wsAbs, feature string) error {
	path := filepath.Join(wsAbs, ".devcontainer", "devcontainer.json")
	if _, err := os.Stat(path); err != nil {
		alt := filepath.Join(wsAbs, ".devcontainer.json")
		if _, err2 := os.Stat(alt); err2 != nil {
			return fmt.Errorf("no devcontainer.json found under %s", wsAbs)
		}
		path = alt
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var doc map[string]json.RawMessage
	if err := json.Unmarshal(stripJSONC(raw), &doc); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}

	features := map[string]json.RawMessage{}
	if fr, ok := doc["features"]; ok {
		_ = json.Unmarshal(fr, &features)
	}
	if _, exists := features[feature]; exists {
		return nil
	}
	features[feature] = json.RawMessage("{}")
	fb, err := json.Marshal(features)
	if err != nil {
		return err
	}
	doc["features"] = fb

	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0644)
}

// stripJSONC removes // and /* */ comments from JSONC so it can be parsed
// by encoding/json. It does not tolerate trailing commas.
func stripJSONC(b []byte) []byte {
	out := make([]byte, 0, len(b))
	inStr := false
	escaped := false
	for i := 0; i < len(b); i++ {
		c := b[i]
		if inStr {
			out = append(out, c)
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inStr = false
			}
			continue
		}
		if c == '"' {
			inStr = true
			out = append(out, c)
			continue
		}
		if c == '/' && i+1 < len(b) && b[i+1] == '/' {
			for i < len(b) && b[i] != '\n' {
				i++
			}
			out = append(out, '\n')
			continue
		}
		if c == '/' && i+1 < len(b) && b[i+1] == '*' {
			i += 2
			for i+1 < len(b) && !(b[i] == '*' && b[i+1] == '/') {
				i++
			}
			i++
			continue
		}
		out = append(out, c)
	}
	return out
}
