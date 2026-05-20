package tools

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"shard/internal/permissions"
)

func TestDangerousCommandDetection(t *testing.T) {
	dangerous := []string{"rm -rf .", "del /s *", "format c:", "shutdown /s", "sudo rm file", "powershell Remove-Item file"}
	for _, command := range dangerous {
		if !IsDangerousCommand(command) {
			t.Fatalf("expected dangerous command: %s", command)
		}
	}
	if IsDangerousCommand("go test ./...") {
		t.Fatal("go test should not be dangerous")
	}
}

func TestTruncate(t *testing.T) {
	got := Truncate("abcdef", 3)
	if !strings.Contains(got, "abc") || !strings.Contains(got, "output truncated") {
		t.Fatalf("unexpected truncate output: %q", got)
	}
}

func TestReadFileUsesCleanPathAndPermission(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	if err := os.WriteFile(path, []byte("package main\n"), 0600); err != nil {
		t.Fatal(err)
	}
	fileTools := FileTools{
		Permissions: permissions.NewManager(),
		In:          strings.NewReader("y\n"),
		Out:         &bytes.Buffer{},
	}

	content, err := fileTools.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if content != "package main\n" {
		t.Fatalf("content = %q", content)
	}
}

func TestParseToolRequest(t *testing.T) {
	cases := map[string]string{
		"read_file: main.go":         "read_file",
		"read_project: .":            "read_project",
		"list_dir: .":                "list_dir",
		"run_command: go test ./...": "run_command",
	}
	for body, action := range cases {
		req, ok := ParseToolRequest("```tool_request\n" + body + "\n```")
		if !ok {
			t.Fatalf("expected request for %s", body)
		}
		if req.Action != action {
			t.Fatalf("action = %q, want %q", req.Action, action)
		}
	}
}

func TestParsePlainToolRequest(t *testing.T) {
	req, ok := ParseToolRequest("read_project: .")
	if !ok {
		t.Fatal("expected plain tool request")
	}
	if req.Action != "read_project" || req.Arg != "." {
		t.Fatalf("request = %#v", req)
	}
}

func TestInferToolRequestFromAssistantPlan(t *testing.T) {
	req, ok := InferToolRequest("Let's list the files first.")
	if !ok {
		t.Fatal("expected inferred tool request")
	}
	if req.Action != "list_dir" || req.Arg != "." {
		t.Fatalf("request = %#v", req)
	}
}

func TestInferToolRequestFromUserActionPrompt(t *testing.T) {
	cases := map[string]string{
		"inspect the repo":           "list_dir",
		"run tests now":              "run_command",
		"build the project":          "run_command",
		"read all files and analyze": "read_project",
	}
	for input, action := range cases {
		req, ok := InferToolRequest(input)
		if !ok {
			t.Fatalf("expected request for %q", input)
		}
		if req.Action != action {
			t.Fatalf("%q action = %q, want %q", input, req.Action, action)
		}
	}
}
