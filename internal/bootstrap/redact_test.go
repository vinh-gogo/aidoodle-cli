package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRedactSecrets_OpenAI(t *testing.T) {
	input := "Failed to connect to LLM with key sk-1234567890abcdef1234567890 and error code 401"
	expected := "Failed to connect to LLM with key [REDACTED] and error code 401"
	got := RedactSecrets(input)
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
	if strings.Contains(got, "sk-1234567890abcdef1234567890") {
		t.Fatalf("secret key leaked in output: %s", got)
	}
}

func TestRedactSecrets_Tavily(t *testing.T) {
	input := "Tavily search failed with key tvly-abcdef1234567890abcdef123456"
	expected := "Tavily search failed with key [REDACTED]"
	got := RedactSecrets(input)
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
	if strings.Contains(got, "tvly-abcdef1234567890abcdef123456") {
		t.Fatalf("tavily key leaked in output: %s", got)
	}
}

func TestRedactSecrets_Bearer(t *testing.T) {
	input := "Authorization: Bearer secret-auth-token-12345678"
	expected := "Authorization: Bearer [REDACTED]"
	got := RedactSecrets(input)
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
	if strings.Contains(got, "secret-auth-token-12345678") {
		t.Fatalf("bearer token leaked in output: %s", got)
	}
}

func TestRedactSecrets_Cloudflare(t *testing.T) {
	input := "Cloudflare tunnel token: eyJhY2NvdW50X2lkIjoiMTIzNDU2Nzg5MCJ9"
	expected := "Cloudflare tunnel token: [REDACTED]"
	got := RedactSecrets(input)
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
	if strings.Contains(got, "eyJhY2NvdW50X2lkIjoiMTIzNDU2Nzg5MCJ9") {
		t.Fatalf("cloudflare token leaked in output: %s", got)
	}
}

func TestRedactSecrets_WriteStartupError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)

	secretMsg := "Crash with api key sk-secret1234567890abcdef123456 and tvly-searchkey1234567890"
	logPath := WriteStartupError(secretMsg)
	if logPath == "" {
		t.Fatalf("WriteStartupError returned empty path")
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read error log %s: %v", logPath, err)
	}

	content := string(data)
	if strings.Contains(content, "sk-secret1234567890abcdef123456") {
		t.Fatalf("found raw openai secret in last-error.log: %s", content)
	}
	if strings.Contains(content, "tvly-searchkey1234567890") {
		t.Fatalf("found raw tavily secret in last-error.log: %s", content)
	}
	if !strings.Contains(content, "[REDACTED]") {
		t.Fatalf("expected [REDACTED] in last-error.log: %s", content)
	}

	expectedDir := filepath.Clean(tmpDir)
	actualDir := filepath.Clean(filepath.Dir(logPath))
	if actualDir != expectedDir {
		t.Logf("log path created at %s, temp dir was %s", logPath, tmpDir)
	}
}
