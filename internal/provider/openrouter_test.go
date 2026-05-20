package provider

import (
	"encoding/json"
	"testing"
)

func TestOpenRouterContentTextString(t *testing.T) {
	raw := json.RawMessage(`"hello"`)
	if got := openRouterContentText(raw); got != "hello" {
		t.Fatalf("content = %q", got)
	}
}

func TestOpenRouterContentTextParts(t *testing.T) {
	raw := json.RawMessage(`[{"type":"text","text":"hello "},{"type":"text","text":"world"}]`)
	if got := openRouterContentText(raw); got != "hello world" {
		t.Fatalf("content = %q", got)
	}
}

func TestOpenRouterContentTextObject(t *testing.T) {
	raw := json.RawMessage(`{"output_text":"hello"}`)
	if got := openRouterContentText(raw); got != "hello" {
		t.Fatalf("content = %q", got)
	}
}

func TestOpenRouterContentTextGenericParts(t *testing.T) {
	raw := json.RawMessage(`[{"type":"output_text","content":"hello"},{"text":" world"}]`)
	if got := openRouterContentText(raw); got != "hello world" {
		t.Fatalf("content = %q", got)
	}
}
