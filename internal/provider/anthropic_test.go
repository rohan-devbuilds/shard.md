package provider

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnthropicAPIKeyReadsDotEnv(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	dir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Fatal(err)
		}
	})
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("ANTHROPIC_API_KEY=test-key\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	if got := anthropicAPIKey(); got != "test-key" {
		t.Fatalf("anthropicAPIKey() = %q, want test-key", got)
	}
}

func TestOpenRouterAPIKeyReadsDotEnv(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "")
	dir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Fatal(err)
		}
	})
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("OPENROUTER_API_KEY=or-key\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	if got := envValue("OPENROUTER_API_KEY"); got != "or-key" {
		t.Fatalf("envValue(OPENROUTER_API_KEY) = %q, want or-key", got)
	}
}

func TestNewProviderSelectsOpenRouter(t *testing.T) {
	prov, err := New("openrouter", "openai/gpt-4o-mini")
	if err != nil {
		t.Fatal(err)
	}
	if got := prov.Name(); got != "openai/gpt-4o-mini" {
		t.Fatalf("Name() = %q", got)
	}
}

func TestNewProviderUsesDotEnvModel(t *testing.T) {
	t.Setenv("SHARD_PROVIDER", "")
	t.Setenv("OPENROUTER_MODEL", "")
	dir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Fatal(err)
		}
	})
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("SHARD_PROVIDER=openrouter\nOPENROUTER_MODEL=meta-llama/llama-3.1-8b-instruct\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	prov, err := New("", "")
	if err != nil {
		t.Fatal(err)
	}
	if got := prov.Name(); got != "meta-llama/llama-3.1-8b-instruct" {
		t.Fatalf("Name() = %q", got)
	}
}

func TestAnthropicAPIKeyPrefersEnvironment(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "env-key")

	if got := anthropicAPIKey(); got != "env-key" {
		t.Fatalf("anthropicAPIKey() = %q, want env-key", got)
	}
}
