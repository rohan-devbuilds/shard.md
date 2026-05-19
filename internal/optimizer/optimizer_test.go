package optimizer

import "testing"

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
