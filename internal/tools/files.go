package tools

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"shard/internal/permissions"
)

type FileTools struct {
	Permissions *permissions.Manager
	In          io.Reader
	Out         io.Writer
}

const MaxProjectFileBytes = 100 * 1024

func (t *FileTools) ReadFile(path string) (string, error) {
	clean, err := cleanPath(path)
	if err != nil {
		return "", err
	}
	ok, err := t.Permissions.Approve(permissions.ReadFile, clean, t.In, t.Out)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("permission denied for read file: %s", clean)
	}
	content, err := os.ReadFile(clean)
	if err != nil {
		return "", err
	}
	return Truncate(string(content), MaxOutputBytes), nil
}

func (t *FileTools) ReadProjectFiles(root string, paths []string) (string, error) {
	cleanRoot, err := cleanPath(root)
	if err != nil {
		return "", err
	}
	ok, err := t.Permissions.Approve(permissions.ReadFile, fmt.Sprintf("all indexed project files under %s (%d files)", cleanRoot, len(paths)), t.In, t.Out)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("permission denied for project file read")
	}

	var b strings.Builder
	for _, rel := range paths {
		cleanRel, err := cleanPath(rel)
		if err != nil {
			continue
		}
		full := filepath.Join(cleanRoot, cleanRel)
		info, err := os.Stat(full)
		if err != nil || info.IsDir() {
			continue
		}
		if info.Size() > MaxProjectFileBytes {
			fmt.Fprintf(&b, "\n--- %s ---\n[skipped: file larger than %d bytes]\n", filepath.ToSlash(cleanRel), MaxProjectFileBytes)
			continue
		}
		content, err := os.ReadFile(full)
		if err != nil {
			fmt.Fprintf(&b, "\n--- %s ---\n[error: %v]\n", filepath.ToSlash(cleanRel), err)
			continue
		}
		if looksBinary(content) {
			fmt.Fprintf(&b, "\n--- %s ---\n[skipped: binary file]\n", filepath.ToSlash(cleanRel))
			continue
		}
		fmt.Fprintf(&b, "\n--- %s ---\n%s\n", filepath.ToSlash(cleanRel), string(content))
	}
	return Truncate(strings.TrimSpace(b.String()), MaxOutputBytes*4), nil
}

func (t *FileTools) WriteFile(path string, content string) error {
	clean, err := cleanPath(path)
	if err != nil {
		return err
	}
	ok, err := t.Permissions.Approve(permissions.WriteFile, clean, t.In, t.Out)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("permission denied for write file: %s", clean)
	}
	return os.WriteFile(clean, []byte(content), 0644)
}

func (t *FileTools) ListDir(path string) (string, error) {
	clean, err := cleanPath(path)
	if err != nil {
		return "", err
	}
	entries, err := os.ReadDir(clean)
	if err != nil {
		return "", err
	}
	lines := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		lines = append(lines, name)
	}
	return Truncate(strings.Join(lines, "\n"), MaxOutputBytes), nil
}

func cleanPath(path string) (string, error) {
	path = strings.TrimSpace(strings.Trim(path, `"'`))
	if path == "" {
		return "", fmt.Errorf("path is required")
	}
	clean := filepath.Clean(path)
	if strings.Contains(clean, "\x00") {
		return "", fmt.Errorf("invalid path")
	}
	return clean, nil
}

func looksBinary(content []byte) bool {
	limit := len(content)
	if limit > 8000 {
		limit = 8000
	}
	for i := 0; i < limit; i++ {
		if content[i] == 0 {
			return true
		}
	}
	return false
}
