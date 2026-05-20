package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanIgnoresUnwantedFolders(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "main.go"))
	mustWrite(t, filepath.Join(dir, ".git", "config"))
	mustWrite(t, filepath.Join(dir, "node_modules", "pkg", "index.js"))
	mustWrite(t, filepath.Join(dir, ".shard", "current.md"))

	tree, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree.Files) != 1 || tree.Files[0] != "main.go" {
		t.Fatalf("files = %v, want only main.go", tree.Files)
	}
}

func mustWrite(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
}
