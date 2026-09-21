package devcontainer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestStripJSONC(t *testing.T) {
	input := []byte(`{
  // a line comment
  "a": 1, /* a block
  comment */ "b": "text // not a comment", "c": "text /* not a comment */"
}`)
	var doc map[string]any
	if err := json.Unmarshal(stripJSONC(input), &doc); err != nil {
		t.Fatalf("stripJSONC output does not parse as JSON: %v\n%s", err, stripJSONC(input))
	}
	if doc["a"].(float64) != 1 {
		t.Errorf(`a = %v, want 1`, doc["a"])
	}
	if doc["b"] != "text // not a comment" {
		t.Errorf(`b = %q, want a literal "//" inside the string to survive`, doc["b"])
	}
	if doc["c"] != "text /* not a comment */" {
		t.Errorf(`c = %q, want a literal "/* */" inside the string to survive`, doc["c"])
	}
}

func TestEnsureFeature_AddsWhenMissing(t *testing.T) {
	dir := t.TempDir()
	mustMkdir(t, filepath.Join(dir, ".devcontainer"))
	path := filepath.Join(dir, ".devcontainer", "devcontainer.json")
	mustWrite(t, path, `{
  // base image
  "image": "mcr.microsoft.com/devcontainers/base:ubuntu"
}`)

	const feature = "ghcr.io/devcontainers-extra/features/neovim:1"
	if err := ensureFeature(dir, feature); err != nil {
		t.Fatalf("ensureFeature: %v", err)
	}

	var doc struct {
		Features map[string]json.RawMessage `json:"features"`
	}
	raw := mustRead(t, path)
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("rewritten file is not valid JSON: %v\n%s", err, raw)
	}
	if _, ok := doc.Features[feature]; !ok {
		t.Fatalf("feature %q was not added; features = %v", feature, doc.Features)
	}
}

func TestEnsureFeature_NoopWhenPresent(t *testing.T) {
	dir := t.TempDir()
	mustMkdir(t, filepath.Join(dir, ".devcontainer"))
	path := filepath.Join(dir, ".devcontainer", "devcontainer.json")
	const feature = "ghcr.io/devcontainers-extra/features/neovim:1"
	original := `{
  // already has the feature
  "features": {
    "` + feature + `": {}
  }
}`
	mustWrite(t, path, original)

	if err := ensureFeature(dir, feature); err != nil {
		t.Fatalf("ensureFeature: %v", err)
	}

	got := string(mustRead(t, path))
	if got != original {
		t.Fatalf("file was rewritten even though the feature was already present:\n--- got ---\n%s\n--- want ---\n%s", got, original)
	}
}

func TestEnsureFeature_AlternatePath(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, ".devcontainer.json"), `{"image": "alpine"}`)

	const feature = "ghcr.io/devcontainers-extra/features/neovim:1"
	if err := ensureFeature(dir, feature); err != nil {
		t.Fatalf("ensureFeature: %v", err)
	}
}

func TestEnsureFeature_NoDevcontainerJSON(t *testing.T) {
	dir := t.TempDir()
	if err := ensureFeature(dir, "some/feature:1"); err == nil {
		t.Fatal("expected an error when no devcontainer.json exists")
	}
}

func mustMkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
