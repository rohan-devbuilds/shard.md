package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCreatesAllMemoryFiles(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	if err := store.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	for _, category := range Categories {
		path := filepath.Join(dir, ".shard", category+".md")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
		}
		text := string(content)
		if !strings.Contains(text, "category: "+category) {
			t.Fatalf("%s missing category frontmatter: %q", path, text)
		}
		if !strings.Contains(text, "priority: medium") {
			t.Fatalf("%s missing priority frontmatter: %q", path, text)
		}
	}
}
