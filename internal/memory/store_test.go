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

func TestAppendSectionsPreservesFrontmatterAndUpdatesTimestamp(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	if err := store.Init(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, ".shard", "current.md")
	beforeBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	before := string(beforeBytes)
	beforeUpdated := frontmatterValue(before, "updated:")

	updated, err := store.AppendSections(map[string]string{
		"current": "Durable update.",
		"unknown": "Ignore me.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated) != 1 || updated[0] != "current.md" {
		t.Fatalf("updated = %v", updated)
	}

	afterBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	after := string(afterBytes)
	if strings.Count(after, "---") != 2 {
		t.Fatalf("frontmatter duplicated or damaged:\n%s", after)
	}
	if !strings.Contains(after, "Durable update.") {
		t.Fatalf("missing appended content:\n%s", after)
	}
	if strings.Contains(after, "Ignore me.") {
		t.Fatalf("unknown section was written:\n%s", after)
	}
	afterUpdated := frontmatterValue(after, "updated:")
	if beforeUpdated == "" || afterUpdated == "" || beforeUpdated == afterUpdated {
		t.Fatalf("updated timestamp did not change: before=%q after=%q", beforeUpdated, afterUpdated)
	}
}

func TestValidateMissingShardFolder(t *testing.T) {
	err := NewStore(t.TempDir()).Validate()
	if err == nil || err.Error() != "Memory folder not found. Run `shard init` first." {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestAppendSectionsSkipsNoDurablePlaceholders(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	if err := store.Init(); err != nil {
		t.Fatal(err)
	}
	updated, err := store.AppendSections(map[string]string{
		"current": "No durable project information has been provided yet.",
		"tasks":   "None",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated) != 0 {
		t.Fatalf("updated = %v, want none", updated)
	}
}

func frontmatterValue(content string, key string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, key) {
			return strings.TrimSpace(strings.TrimPrefix(line, key))
		}
	}
	return ""
}
