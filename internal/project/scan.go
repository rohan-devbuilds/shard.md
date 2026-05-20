package project

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

var ignoredDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	"dist":         true,
	"build":        true,
	"target":       true,
	".next":        true,
	".cache":       true,
	".shard":       true,
}

type Tree struct {
	Root  string
	Files []string
}

func Scan(root string) (Tree, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return Tree{}, err
	}
	files := []string{}
	err = filepath.WalkDir(abs, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if ignoredDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(abs, path)
		if err != nil {
			return nil
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return Tree{}, err
	}
	sort.Strings(files)
	return Tree{Root: abs, Files: files}, nil
}

func (t Tree) Format(limit int) string {
	files := t.Files
	truncated := false
	if limit > 0 && len(files) > limit {
		files = files[:limit]
		truncated = true
	}
	var b strings.Builder
	b.WriteString("Project root: ")
	b.WriteString(t.Root)
	b.WriteString("\n\nFiles:\n")
	for _, file := range files {
		b.WriteString("- ")
		b.WriteString(file)
		b.WriteString("\n")
	}
	if truncated {
		b.WriteString("- ... file tree truncated\n")
	}
	return strings.TrimSpace(b.String())
}
