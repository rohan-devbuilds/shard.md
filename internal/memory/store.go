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

func (s *Store) Dir() string {
	return filepath.Join(s.root, ".shard")
}

func (s *Store) Init() error {
	dir := s.Dir()
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

func (s *Store) Validate() error {
	info, err := os.Stat(s.Dir())
	if os.IsNotExist(err) {
		return fmt.Errorf("Memory folder not found. Run `shard init` first.")
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("memory path is not a directory: %s", s.Dir())
	}
	for _, category := range Categories {
		path := filepath.Join(s.Dir(), category+".md")
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("memory file missing: %s", filepath.Base(path))
			}
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
		path := filepath.Join(s.Dir(), name+".md")
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		items = append(items, Item{Name: name + ".md", Path: path, Content: string(content)})
	}
	return items, nil
}

func (s *Store) AppendSection(category string, content string) error {
	_, err := s.AppendSections(map[string]string{category: content})
	return err
}

func (s *Store) AppendSections(sections map[string]string) ([]string, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	updated := []string{}
	for _, category := range Categories {
		content := sections[category]
		changed, err := s.appendKnownSection(category, content)
		if err != nil {
			return nil, err
		}
		if changed {
			updated = append(updated, category+".md")
		}
	}
	return updated, nil
}

func (s *Store) appendKnownSection(category string, content string) (bool, error) {
	category = strings.ToLower(strings.TrimSpace(category))
	content = strings.TrimSpace(content)
	if !IsCategory(category) || !isDurableContent(content) {
		return false, nil
	}
	path := filepath.Join(s.Dir(), category+".md")
	now := time.Now().Format(time.RFC3339Nano)
	entry := fmt.Sprintf("\n\n## %s\n\n%s\n", now, content)
	existing, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	updated := updateFrontmatterTimestamp(string(existing), now)
	updated += entry
	return true, os.WriteFile(path, []byte(updated), 0644)
}

func isDurableContent(content string) bool {
	normalized := strings.ToLower(strings.TrimSpace(content))
	if normalized == "" || normalized == "..." || normalized == "none" || normalized == "n/a" {
		return false
	}
	noDurablePhrases := []string{
		"no durable",
		"no relevant",
		"nothing durable",
		"not applicable",
		"no updates",
	}
	for _, phrase := range noDurablePhrases {
		if strings.Contains(normalized, phrase) {
			return false
		}
	}
	return true
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
	return fmt.Sprintf("---\ncategory: %s\nupdated: %s\npriority: medium\nrelated: []\n---\n", category, time.Now().Format(time.RFC3339Nano))
}

func updateFrontmatterTimestamp(content string, timestamp string) string {
	lines := strings.SplitAfter(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return content
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end < 0 {
		return content
	}
	for i := 1; i < end; i++ {
		if strings.HasPrefix(lines[i], "updated:") {
			lines[i] = "updated: " + timestamp + "\n"
			return strings.Join(lines, "")
		}
	}
	withUpdated := append([]string{}, lines[:end]...)
	withUpdated = append(withUpdated, "updated: "+timestamp+"\n")
	withUpdated = append(withUpdated, lines[end:]...)
	return strings.Join(withUpdated, "")
}
