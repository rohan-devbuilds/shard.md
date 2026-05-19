package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var Categories = []string{
	"current",
	"tasks",
	"bugs",
	"architecture",
	"decisions",
	"codebase",
	"commands",
	"dependencies",
	"api",
	"ui",
	"agents",
	"changelog",
}

type Store struct {
	root string
}

type Item struct {
	Name    string
	Path    string
	Content string
}

func NewStore(root string) *Store {
	return &Store{root: root}
}

func (s *Store) Init() error {
	dir := filepath.Join(s.root, ".shard")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	for _, category := range Categories {
		path := filepath.Join(dir, category+".md")
		if _, err := os.Stat(path); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return err
		}
		if err := os.WriteFile(path, []byte(frontmatter(category)), 0644); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) LoadFiles(names []string) ([]Item, error) {
	items := make([]Item, 0, len(names))
	seen := map[string]bool{}
	for _, name := range names {
		name = strings.TrimSuffix(strings.ToLower(name), ".md")
		if seen[name] || !IsCategory(name) {
			continue
		}
		seen[name] = true
		path := filepath.Join(s.root, ".shard", name+".md")
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		items = append(items, Item{Name: name + ".md", Path: path, Content: string(content)})
	}
	return items, nil
}

func (s *Store) AppendSection(category string, content string) error {
	category = strings.ToLower(strings.TrimSpace(category))
	content = strings.TrimSpace(content)
	if content == "" || content == "..." || !IsCategory(category) {
		return nil
	}
	path := filepath.Join(s.root, ".shard", category+".md")
	entry := fmt.Sprintf("\n\n## %s\n\n%s\n", time.Now().Format(time.RFC3339), content)
	return os.WriteFile(path, appendFile(path, entry), 0644)
}

func FormatForPrompt(items []Item) string {
	var b strings.Builder
	for _, item := range items {
		b.WriteString("### ")
		b.WriteString(item.Name)
		b.WriteString("\n")
		b.WriteString(strings.TrimSpace(item.Content))
		b.WriteString("\n\n")
	}
	return strings.TrimSpace(b.String())
}

func IsCategory(name string) bool {
	for _, category := range Categories {
		if name == category {
			return true
		}
	}
	return false
}

func frontmatter(category string) string {
	return fmt.Sprintf("---\ncategory: %s\nupdated: %s\npriority: medium\nrelated: []\n---\n", category, time.Now().Format(time.RFC3339))
}

func appendFile(path string, entry string) []byte {
	existing, err := os.ReadFile(path)
	if err != nil {
		return []byte(entry)
	}
	return append(existing, []byte(entry)...)
}
