package bootstrap

import (
	"strings"
	"testing"
)

// TestAdversarial_RedactSecrets tests redaction across diverse token patterns.
func TestAdversarial_RedactSecrets(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		token    string
		expected string
	}{
		{
			name:     "OpenAIStandard",
			input:    "Error calling OpenAI with key sk-proj-1234567890abcdef123456 failed",
			token:    "sk-proj-1234567890abcdef123456",
			expected: "Error calling OpenAI with key [REDACTED] failed",
		},
		{
			name:     "AnthropicDeepSeekKey",
			input:    "Authorization failed with sk-ant-api03-abcdefghijklmnop123456",
			token:    "sk-ant-api03-abcdefghijklmnop123456",
			expected: "Authorization failed with [REDACTED]",
		},
		{
			name:     "TavilyKeyStandard",
			input:    "Tavily search client initialized with tvly-prod-1234567890abcdef123456",
			token:    "tvly-prod-1234567890abcdef123456",
			expected: "Tavily search client initialized with [REDACTED]",
		},
		{
			name:     "BearerTokenMixedCase",
			input:    "Request header Authorization: Bearer secret-jwt-session-token-998877",
			token:    "secret-jwt-session-token-998877",
			expected: "Request header Authorization: Bearer [REDACTED]",
		},
		{
			name:     "BearerTokenLower",
			input:    "header bearer custom-auth-token-1234567890 rejected",
			token:    "custom-auth-token-1234567890",
			expected: "header bearer [REDACTED] rejected",
		},
		{
			name:     "CloudflareTokenEyJ",
			input:    "Cloudflare tunnel credential eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.token123456789",
			token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expected: "[REDACTED]",
		},
		{
			name:     "MultipleTokensInOneLog",
			input:    "Bearer my-secret-token-123456 and sk-proj-1234567890abcdef123456 with tvly-prod-1234567890abcdef123456",
			token:    "my-secret-token-123456",
			expected: "Bearer [REDACTED] and [REDACTED] with [REDACTED]",
		},
	}

	for _, tc := range testCases {
		t.Run("Redact_"+tc.name, func(t *testing.T) {
			output := RedactSecrets(tc.input)
			if strings.Contains(output, tc.token) {
				t.Fatalf("SECURITY FLAW: Token %q leaked in redacted output: %s", tc.token, output)
			}
			if !strings.Contains(output, "[REDACTED]") {
				t.Fatalf("Expected [REDACTED] marker in output, got %s", output)
			}
		})
	}
}
