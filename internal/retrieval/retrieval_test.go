package retrieval

import "testing"

func TestSelectFilesForBugPrompt(t *testing.T) {
	got := SelectFiles("fix crash error", "", "max")
	want := []string{"current", "bugs", "commands", "codebase"}
	assertFiles(t, got, want)
}

func TestSelectFilesForAPIPrompt(t *testing.T) {
	got := SelectFiles("anthropic provider model api", "", "max")
	want := []string{"current", "api", "dependencies"}
	assertFiles(t, got, want)
}

func TestSelectFilesForUIPrompt(t *testing.T) {
	got := SelectFiles("terminal ui theme", "", "max")
	want := []string{"current", "ui"}
	assertFiles(t, got, want)
}

func TestSelectFilesAlwaysIncludesCurrent(t *testing.T) {
	got := SelectFiles("hello there", "", "medium")
	assertFiles(t, got, []string{"current"})
}

func assertFiles(t *testing.T, got []string, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d; got %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("file[%d] = %q, want %q; got %v", i, got[i], want[i], got)
		}
	}
}
