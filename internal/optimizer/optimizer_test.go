package optimizer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"shard/internal/memory"
	"shard/internal/provider"
	"shard/internal/session"
)

func TestParseSectionsMapsSectionsToCategories(t *testing.T) {
	input := `Intro text ignored.

## current
Working on Shard MVP.

## tasks
- Add tests.

## api
Anthropic is the first provider.

## unknown
ignored

## ui
Terminal chat loop.`

	sections := ParseSections(input)

	if sections["current"] != "Working on Shard MVP." {
		t.Fatalf("current = %q", sections["current"])
	}
	if sections["tasks"] != "- Add tests." {
		t.Fatalf("tasks = %q", sections["tasks"])
	}
	if sections["api"] != "Anthropic is the first provider." {
		t.Fatalf("api = %q", sections["api"])
	}
	if sections["ui"] != "Terminal chat loop." {
		t.Fatalf("ui = %q", sections["ui"])
	}
	if _, ok := sections["unknown"]; ok {
		t.Fatal("unknown section should not be parsed")
	}
}

func TestRunWritesParsedSectionsToMemory(t *testing.T) {
	dir := t.TempDir()
	store := memory.NewStore(dir)
	if err := store.Init(); err != nil {
		t.Fatal(err)
	}
	sess := session.New()
	sess.Add("user", "remember current task")
	prov := fakeProvider{response: "## current\nWorking state.\n\n## tasks\n- Next task.\n\n## unknown\nIgnore."}

	summary, err := Run(context.Background(), sess, store, prov)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(summary.UpdatedFiles, ",") != "current.md,tasks.md" {
		t.Fatalf("updated = %v", summary.UpdatedFiles)
	}
	current := mustRead(t, filepath.Join(dir, ".shard", "current.md"))
	tasks := mustRead(t, filepath.Join(dir, ".shard", "tasks.md"))
	if !strings.Contains(current, "Working state.") || !strings.Contains(tasks, "Next task.") {
		t.Fatalf("memory not written\ncurrent=%s\ntasks=%s", current, tasks)
	}
	if strings.Contains(current, "Ignore.") || strings.Contains(tasks, "Ignore.") {
		t.Fatal("unknown section leaked into known files")
	}
}

func TestUpdateMemorySharedPath(t *testing.T) {
	dir := t.TempDir()
	store := memory.NewStore(dir)
	if err := store.Init(); err != nil {
		t.Fatal(err)
	}
	updated, err := UpdateMemory(store, map[string]string{"bugs": "Bug note."})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated) != 1 || updated[0] != "bugs.md" {
		t.Fatalf("updated = %v", updated)
	}
	if !strings.Contains(mustRead(t, filepath.Join(dir, ".shard", "bugs.md")), "Bug note.") {
		t.Fatal("shared update path did not write bugs.md")
	}
}

func TestRunMissingShardFolder(t *testing.T) {
	sess := session.New()
	sess.Add("user", "hello")
	_, err := Run(context.Background(), sess, memory.NewStore(t.TempDir()), fakeProvider{response: "## current\nx"})
	if err == nil || err.Error() != "Memory folder not found. Run `shard init` first." {
		t.Fatalf("Run() error = %v", err)
	}
}

type fakeProvider struct {
	response string
}

func (f fakeProvider) Chat(ctx context.Context, messages []provider.Message) (string, error) {
	return f.response, nil
}

func (f fakeProvider) Name() string {
	return "fake"
}

func (f fakeProvider) Provider() string {
	return "fake"
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
