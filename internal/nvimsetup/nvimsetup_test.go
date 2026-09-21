package nvimsetup

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestEnsure_UnknownStrategy(t *testing.T) {
	_, err := Ensure(&fakeTransport{}, Options{Strategy: "bogus"})
	if err == nil || !strings.Contains(err.Error(), "unknown strategy") {
		t.Fatalf("err = %v, want an unknown-strategy error", err)
	}
}

func TestEnsure_InstallStrategy_FailsWithoutFallback(t *testing.T) {
	installErr := errors.New("apt-get: no network")
	fake := &fakeTransport{runFunc: func(string) (string, error) { return "", installErr }}

	_, err := Ensure(fake, Options{Strategy: "install"})
	if err == nil || !errors.Is(err, installErr) {
		t.Fatalf("err = %v, want it to wrap %v", err, installErr)
	}
}

func TestEnsure_AutoStrategy_FallsBackToPush(t *testing.T) {
	fake := &fakeTransport{
		// Only the install script should fail; push()'s own "chmod +x" call
		// (via the same Run method) must still succeed.
		runFunc: func(script string) (string, error) {
			if strings.Contains(script, "apt-get") {
				return "", errors.New("no supported package manager")
			}
			return "", nil
		},
		unameOS:   unameOS[runtime.GOOS],
		unameArch: unameArch[runtime.GOARCH],
	}

	dir := t.TempDir()
	nvimBin := writeFakeNvimInstall(t, dir)

	res, err := Ensure(fake, Options{Strategy: "auto", LocalNvimBin: nvimBin})
	if err != nil {
		t.Fatalf("Ensure: %v", err)
	}
	if res.Exec != remoteBase+"/bin/nvim" {
		t.Fatalf("Exec = %q, want the pushed binary path", res.Exec)
	}
}

func TestCheckArchMatch(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		fake := &fakeTransport{unameOS: unameOS[runtime.GOOS], unameArch: unameArch[runtime.GOARCH]}
		if err := checkArchMatch(fake); err != nil {
			t.Fatalf("checkArchMatch: %v", err)
		}
	})

	t.Run("mismatch", func(t *testing.T) {
		fake := &fakeTransport{unameOS: "Plan9", unameArch: "mips"}
		err := checkArchMatch(fake)
		if err == nil || !strings.Contains(err.Error(), "must match architectures") {
			t.Fatalf("err = %v, want an architecture-mismatch error", err)
		}
	})

	t.Run("uname failure", func(t *testing.T) {
		fake := &fakeTransport{unameErr: errors.New("no such command: uname")}
		if err := checkArchMatch(fake); err == nil {
			t.Fatal("expected an error when Uname fails")
		}
	})
}

func TestPush_NoRuntimeDir(t *testing.T) {
	fake := &fakeTransport{unameOS: unameOS[runtime.GOOS], unameArch: unameArch[runtime.GOARCH]}

	dir := t.TempDir()
	binDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	nvimBin := filepath.Join(binDir, "nvim")
	if err := os.WriteFile(nvimBin, []byte("#!/bin/sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	// no sibling <dir>/share/nvim, so push should refuse.

	_, err := push(fake, Options{LocalNvimBin: nvimBin})
	if err == nil || !strings.Contains(err.Error(), "runtime dir") {
		t.Fatalf("err = %v, want a runtime-dir-not-found error", err)
	}
}

func TestPush_Success(t *testing.T) {
	fake := &fakeTransport{unameOS: unameOS[runtime.GOOS], unameArch: unameArch[runtime.GOARCH]}
	dir := t.TempDir()
	nvimBin := writeFakeNvimInstall(t, dir)

	configDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(configDir, "init.lua"), []byte("-- sentinel"), 0o644); err != nil {
		t.Fatal(err)
	}

	var copied []string
	fake.copyFunc = func(local, remote string) error {
		copied = append(copied, local+" -> "+remote)
		return nil
	}

	res, err := push(fake, Options{LocalNvimBin: nvimBin, LocalConfig: configDir})
	if err != nil {
		t.Fatalf("push: %v", err)
	}
	if res.Exec != remoteBase+"/bin/nvim" {
		t.Errorf("Exec = %q, want %q", res.Exec, remoteBase+"/bin/nvim")
	}
	if res.Env["VIMRUNTIME"] != remoteBase+"/share/nvim/runtime" {
		t.Errorf("VIMRUNTIME = %q", res.Env["VIMRUNTIME"])
	}
	if res.Env["XDG_CONFIG_HOME"] != remoteBase+"/xdgconfig" {
		t.Errorf("XDG_CONFIG_HOME = %q", res.Env["XDG_CONFIG_HOME"])
	}
	if len(copied) != 3 {
		t.Fatalf("CopyTo called %d times, want 3 (binary, runtime, config): %v", len(copied), copied)
	}
}

// writeFakeNvimInstall creates <dir>/bin/nvim and <dir>/share/nvim/runtime,
// mimicking the layout push() expects next to a real nvim binary.
func writeFakeNvimInstall(t *testing.T, dir string) string {
	t.Helper()
	binDir := filepath.Join(dir, "bin")
	runtimeDir := filepath.Join(dir, "share", "nvim", "runtime")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(runtimeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	nvimBin := filepath.Join(binDir, "nvim")
	if err := os.WriteFile(nvimBin, []byte("#!/bin/sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runtimeDir, "marker.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	return nvimBin
}
