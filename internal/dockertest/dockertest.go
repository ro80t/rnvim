// Package dockertest provides small helpers for tests that spin up real
// docker/podman containers instead of mocking the transport layer.
package dockertest

import (
	"os/exec"
	"strings"
	"testing"
)

// RequireBin skips the test if name isn't found in PATH.
func RequireBin(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s not found in PATH, skipping", name)
	}
}

// RequireDaemon skips the test in -short mode, if bin isn't installed, or
// if its daemon isn't reachable.
func RequireDaemon(t *testing.T, bin string) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping container-based test in -short mode")
	}
	RequireBin(t, bin)
	if err := exec.Command(bin, "info").Run(); err != nil {
		t.Skipf("%s daemon not reachable, skipping", bin)
	}
}

// RequireDocker is RequireDaemon(t, "docker").
func RequireDocker(t *testing.T) { RequireDaemon(t, "docker") }

// RunContainer starts a detached container (`<bin> run -d --rm ...`) and
// registers a cleanup that removes it. runArgs are extra `run` flags
// inserted before the image name (e.g. "-e", "FOO=bar").
func RunContainer(t *testing.T, bin string, runArgs []string, image string, cmdArgs ...string) string {
	t.Helper()
	args := append([]string{"run", "-d", "--rm"}, runArgs...)
	args = append(args, image)
	args = append(args, cmdArgs...)
	out, err := exec.Command(bin, args...).Output()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", bin, strings.Join(args, " "), err, out)
	}
	id := strings.TrimSpace(string(out))
	t.Cleanup(func() {
		_ = exec.Command(bin, "rm", "-f", id).Run()
	})
	return id
}

// HostPort returns the host-side "host:port" that containerPort (e.g.
// "2222/tcp") was published to.
func HostPort(t *testing.T, bin, id, containerPort string) string {
	t.Helper()
	out, err := exec.Command(bin, "port", id, containerPort).Output()
	if err != nil {
		t.Fatalf("%s port %s %s: %v", bin, id, containerPort, err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatalf("%s port %s %s: no mapping returned", bin, id, containerPort)
	}
	return strings.TrimSpace(lines[0])
}
