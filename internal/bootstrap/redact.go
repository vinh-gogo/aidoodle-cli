package bootstrap

import (
	"regexp"
)

var (
	// reOpenAI matches OpenAI/DeepSeek style secret keys sk-... (at least 16 chars).
	reOpenAI = regexp.MustCompile(`\bsk-[A-Za-z0-9_-]{16,}\b`)

	// reTavily matches Tavily API keys tvly-... (at least 16 chars).
	reTavily = regexp.MustCompile(`\btvly-[A-Za-z0-9_-]{16,}\b`)

	// reBearer matches HTTP Authorization Bearer tokens.
	reBearer = regexp.MustCompile(`(?i)\b(Bearer\s+)[A-Za-z0-9._-]{10,}\b`)

	// reCloudflare matches Cloudflare Tunnel/Access tokens starting with eyJ... (at least 20 chars).
	reCloudflare = regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{20,}\b`)
)

// RedactSecrets scans a string and replaces sensitive tokens (API keys, Bearer tokens, Cloudflare tokens) with [REDACTED].
func RedactSecrets(s string) string {
	if s == "" {
		return ""
	}

	// 1. Redact Bearer tokens: replace the token part with [REDACTED]
	s = reBearer.ReplaceAllString(s, "${1}[REDACTED]")

	// 2. Redact OpenAI / DeepSeek keys
	s = reOpenAI.ReplaceAllString(s, "[REDACTED]")

	// 3. Redact Tavily keys
	s = reTavily.ReplaceAllString(s, "[REDACTED]")

	// 4. Redact Cloudflare tokens
	s = reCloudflare.ReplaceAllString(s, "[REDACTED]")

	return s
}
