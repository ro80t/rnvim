package transport

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testTransportSuite runs the same assertions against any live Transport
// (container or ssh), so both implementations are exercised identically.
func testTransportSuite(t *testing.T, tr Transport) {
	t.Helper()

	t.Run("Home", func(t *testing.T) {
		home, err := tr.Home()
		if err != nil {
			t.Fatalf("Home: %v", err)
		}
		if home == "" {
			t.Fatal("Home returned an empty string")
		}
	})

	t.Run("Uname", func(t *testing.T) {
		osName, arch, err := tr.Uname()
		if err != nil {
			t.Fatalf("Uname: %v", err)
		}
		if osName != "Linux" {
			t.Fatalf("osName = %q, want Linux", osName)
		}
		if arch == "" {
			t.Fatal("arch is empty")
		}
	})

	t.Run("Run", func(t *testing.T) {
		out, err := tr.Run("echo hello-rnvim")
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if !strings.Contains(out, "hello-rnvim") {
			t.Fatalf("Run output = %q, want it to contain hello-rnvim", out)
		}
	})

	t.Run("MkdirAll", func(t *testing.T) {
		if err := tr.MkdirAll("/tmp/rnvim-test/nested/dir"); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		out, err := tr.Run("test -d /tmp/rnvim-test/nested/dir && echo yes")
		if err != nil || !strings.Contains(out, "yes") {
			t.Fatalf("directory was not created: out=%q err=%v", out, err)
		}
	})

	t.Run("CopyTo file", func(t *testing.T) {
		local := filepath.Join(t.TempDir(), "hello.txt")
		if err := os.WriteFile(local, []byte("hello-file-content"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := tr.CopyTo(local, "/tmp/rnvim-test/copied/hello.txt"); err != nil {
			t.Fatalf("CopyTo: %v", err)
		}
		out, err := tr.Run("cat /tmp/rnvim-test/copied/hello.txt")
		if err != nil || !strings.Contains(out, "hello-file-content") {
			t.Fatalf("copied file content = %q, err=%v", out, err)
		}
	})

	t.Run("CopyTo directory", func(t *testing.T) {
		localDir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(localDir, "sub"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(localDir, "sub", "nested.txt"), []byte("nested-content"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := tr.CopyTo(localDir, "/tmp/rnvim-test/copieddir"); err != nil {
			t.Fatalf("CopyTo: %v", err)
		}
		out, err := tr.Run("cat /tmp/rnvim-test/copieddir/sub/nested.txt")
		if err != nil || !strings.Contains(out, "nested-content") {
			t.Fatalf("nested file content = %q, err=%v", out, err)
		}
	})
}
