package devcontainer

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"rnvim/internal/dockertest"
)

// TestUpAndGetContainerID_Docker drives a real `devcontainer up` (requires
// the @devcontainers/cli `devcontainer` binary) and checks that we can
// parse a live container id out of its output. The final interactive nvim
// attach (Connect's last step) needs a real TTY and isn't covered here.
func TestUpAndGetContainerID_Docker(t *testing.T) {
	dockertest.RequireDocker(t)
	dockertest.RequireBin(t, "devcontainer")

	dir := t.TempDir()
	mustMkdir(t, filepath.Join(dir, ".devcontainer"))
	mustWrite(t, filepath.Join(dir, ".devcontainer", "devcontainer.json"), `{"image": "alpine:3.20"}`)

	id, err := upAndGetContainerID(dir)
	if err != nil {
		t.Fatalf("upAndGetContainerID: %v", err)
	}
	if id == "" {
		t.Fatal("got an empty container id")
	}
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-f", id).Run() })

	out, err := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", id).Output()
	if err != nil || strings.TrimSpace(string(out)) != "true" {
		t.Fatalf("container %s is not running: out=%q err=%v", id, out, err)
	}
}
