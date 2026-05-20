package permissions

import (
	"bytes"
	"strings"
	"testing"
)

func TestManagerAllowDenyAlways(t *testing.T) {
	manager := NewManager()

	ok, err := manager.Approve(ReadFile, "main.go", strings.NewReader("y\n"), &bytes.Buffer{})
	if err != nil || !ok {
		t.Fatalf("Approve yes = %v, %v", ok, err)
	}
	if manager.Status(ReadFile) != "ask" {
		t.Fatalf("yes should not always allow")
	}

	ok, err = manager.Approve(ReadFile, "main.go", strings.NewReader("n\n"), &bytes.Buffer{})
	if err != nil || ok {
		t.Fatalf("Approve no = %v, %v", ok, err)
	}
	if manager.DeniedCount(ReadFile) != 1 {
		t.Fatalf("DeniedCount = %d, want 1", manager.DeniedCount(ReadFile))
	}

	ok, err = manager.Approve(ReadFile, "main.go", strings.NewReader("a\n"), &bytes.Buffer{})
	if err != nil || !ok {
		t.Fatalf("Approve always = %v, %v", ok, err)
	}
	if manager.Status(ReadFile) != "allowed" {
		t.Fatalf("Status = %q, want allowed", manager.Status(ReadFile))
	}
}
