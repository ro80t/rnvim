package nvimsetup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ro80t/rnvim/internal/dockertest"
	"github.com/ro80t/rnvim/internal/transport"
)

// TestEnsure_InstallStrategy_Docker exercises the real "auto" -> install
// path end to end: apk-installing nvim inside a live alpine container and
// copying a local config into it.
func TestEnsure_InstallStrategy_Docker(t *testing.T) {
	dockertest.RequireDocker(t)

	id := dockertest.RunContainer(t, "docker", nil, "alpine:3.20", "sleep", "600")
	ct := transport.NewContainerTransport("docker", id)

	configDir := t.TempDir()
	const sentinel = "-- rnvim test sentinel"
	if err := os.WriteFile(filepath.Join(configDir, "init.lua"), []byte(sentinel), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := Ensure(ct, Options{Strategy: "install", LocalConfig: configDir})
	if err != nil {
		t.Fatalf("Ensure: %v", err)
	}
	if res.Exec != "nvim" {
		t.Fatalf("Exec = %q, want %q", res.Exec, "nvim")
	}

	out, err := ct.Run("nvim --version")
	if err != nil || !strings.Contains(out, "NVIM") {
		t.Fatalf("nvim --version = %q, err=%v", out, err)
	}

	out, err = ct.Run("cat /root/.config/nvim/init.lua")
	if err != nil || !strings.Contains(out, sentinel) {
		t.Fatalf("config was not copied: out=%q err=%v", out, err)
	}
}
