package transport

import (
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"rnvim/internal/dockertest"
)

// TestSSHTransport runs the shared transport suite against a real sshd,
// hosted in a docker container, reachable with a freshly generated key.
func TestSSHTransport(t *testing.T) {
	dockertest.RequireDocker(t)
	dockertest.RequireBin(t, "ssh-keygen")
	dockertest.RequireBin(t, "ssh")
	dockertest.RequireBin(t, "scp")

	dir := t.TempDir()
	keyPath := filepath.Join(dir, "id_ed25519")
	if out, err := exec.Command("ssh-keygen", "-t", "ed25519", "-N", "", "-f", keyPath).CombinedOutput(); err != nil {
		t.Fatalf("ssh-keygen: %v\n%s", err, out)
	}
	pub, err := os.ReadFile(keyPath + ".pub")
	if err != nil {
		t.Fatal(err)
	}

	id := dockertest.RunContainer(t, "docker", []string{
		"-e", "PUBLIC_KEY=" + string(pub),
		"-e", "USER_NAME=testuser",
		"-e", "PASSWORD_ACCESS=false",
		"-p", "127.0.0.1::2222",
	}, "ghcr.io/linuxserver/openssh-server:latest")

	hostPort := dockertest.HostPort(t, "docker", id, "2222/tcp")
	_, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		t.Fatalf("parse host port %q: %v", hostPort, err)
	}

	tr := &SSHTransport{
		Host:     "testuser@127.0.0.1",
		Port:     port,
		Identity: keyPath,
		ExtraArgs: []string{
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "ConnectTimeout=5",
		},
	}

	// sshd takes a few seconds to generate host keys and come up.
	deadline := time.Now().Add(60 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		if _, err := tr.Run("true"); err == nil {
			lastErr = nil
			break
		} else {
			lastErr = err
			time.Sleep(time.Second)
		}
	}
	if lastErr != nil {
		t.Fatalf("ssh server never became ready: %v", lastErr)
	}

	testTransportSuite(t, tr)
}
