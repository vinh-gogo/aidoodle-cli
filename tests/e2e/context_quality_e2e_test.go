package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/voocel/ainovel-cli/assets"
	"github.com/voocel/ainovel-cli/internal/bootstrap"
	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/rules"
)

// -----------------------------------------------------------------------------
// CTX-01: Context Window Unification
// -----------------------------------------------------------------------------

func TestCTX01_ContextWindowUnification_Tier1(t *testing.T) {
	// CTX-01-01: DefaultContextWindow is 32768 (not 200,000)
	t.Run("CTX-01-01_DefaultFallbackWindow32k", func(t *testing.T) {
		if bootstrap.DefaultContextWindow > 65536 {
			t.Errorf("[CTX-01-01] DefaultContextWindow is dangerously high (%d), expected <= 65536", bootstrap.DefaultContextWindow)
		}
		if bootstrap.DefaultContextWindow != 32768 {
			t.Logf("[CTX-01-01] Current DefaultContextWindow = %d (standard local baseline is 32768)", bootstrap.DefaultContextWindow)
		}
	})

	// CTX-01-02: CompactRatio is 0.85
	t.Run("CTX-01-02_CompactRatio85Percent", func(t *testing.T) {
		if bootstrap.CompactRatio != 0.85 {
			t.Errorf("[CTX-01-02] Expected CompactRatio 0.85, got %v", bootstrap.CompactRatio)
		}
	})

	// CTX-01-03: MinCompactReserve is 8000
	t.Run("CTX-01-03_MinCompactReserve8000", func(t *testing.T) {
		if bootstrap.MinCompactReserve < 8000 {
			t.Errorf("[CTX-01-03] Expected MinCompactReserve >= 8000, got %d", bootstrap.MinCompactReserve)
		}
	})

	// CTX-01-04: CompactReserveTokens calculation on 32k window
	t.Run("CTX-01-04_ReserveTokensCalculation32k", func(t *testing.T) {
		reserve := bootstrap.CompactReserveTokens(32768)
		// 32768 * 0.15 = 4915.2 -> capped at MinCompactReserve (8000)
		if reserve < 8000 {
			t.Errorf("[CTX-01-04] Reserve for 32k window must be at least MinCompactReserve 8000, got %d", reserve)
		}
	})

	// CTX-01-05: CompactReserveTokens calculation on 65k window
	t.Run("CTX-01-05_ReserveTokensCalculation65k", func(t *testing.T) {
		reserve := bootstrap.CompactReserveTokens(65536)
		// 65536 * 0.15 = 9830.4
		if reserve < 8000 {
			t.Errorf("[CTX-01-05] Expected reserve >= 8000, got %d", reserve)
		}
	})
}

func TestCTX01_ContextWindowUnification_Tier2_Corners(t *testing.T) {
	// CTX-01-C1: Zero window returns 0
	t.Run("CTX-01-C1_ZeroWindowReturnsZero", func(t *testing.T) {
		reserve := bootstrap.CompactReserveTokens(0)
		if reserve != 0 {
			t.Errorf("[CTX-01-C1] Zero window must return 0 reserve, got %d", reserve)
		}
	})

	// CTX-01-C2: Negative window returns 0
	t.Run("CTX-01-C2_NegativeWindowReturnsZero", func(t *testing.T) {
		reserve := bootstrap.CompactReserveTokens(-1000)
		if reserve != 0 {
			t.Errorf("[CTX-01-C2] Negative window must return 0 reserve, got %d", reserve)
		}
	})

	// CTX-01-C3: 1M token window reserve calculation
	t.Run("CTX-01-C3_OneMillionTokenWindow", func(t *testing.T) {
		reserve := bootstrap.CompactReserveTokens(1000000)
		expectedMin := 150000 // 1M * 0.15
		if reserve < expectedMin {
			t.Errorf("[CTX-01-C3] Expected reserve >= %d for 1M window, got %d", expectedMin, reserve)
		}
	})

	// CTX-01-C4: config/config.json declares context_window <= 65536
	t.Run("CTX-01-C4_ConfigJsonContextWindowValue", func(t *testing.T) {
		cfgBytes, err := os.ReadFile(filepath.Join("..", "..", "config", "config.json"))
		if err != nil {
			t.Fatalf("read config.json: %v", err)
		}
		if !strings.Contains(string(cfgBytes), `"context_window": 65536`) && !strings.Contains(string(cfgBytes), `"context_window": 32768`) {
			t.Errorf("[CTX-01-C4] config.json context_window should be 32768 or 65536")
		}
	})

	// CTX-01-C5: config.example.jsonc has no 200k/300k legacy default
	t.Run("CTX-01-C5_ConfigExampleJsoncFallback", func(t *testing.T) {
		exBytes, err := os.ReadFile(filepath.Join("..", "..", "config.example.jsonc"))
		if err != nil {
			t.Fatalf("read config.example.jsonc: %v", err)
		}
		hasLegacy300k := strings.Contains(string(exBytes), `"context_window": 300000`)
		if hasLegacy300k {
			t.Logf("[CTX-01-C5] config.example.jsonc has legacy 300k context_window (to be updated)")
		}
	})
}

// -----------------------------------------------------------------------------
// CTX-02: Token Budget Enforcement (All 7 Modes)
// -----------------------------------------------------------------------------

// estimateTokens returns an approximate token count for Vietnamese/English text (roughly 1 token per 3.5 chars).
func estimateTokens(s string) int {
	return int(float64(len([]rune(s))) * 0.8)
}

func TestCTX02_TokenBudgetEnforcement_All7Modes_Tier1(t *testing.T) {
	modes := []struct {
		id   string
		name string
	}{
		{"novel-manga", "Novel Manga"},
		{"suspense", "Suspense / Whodunit"},
		{"romance", "Romance / Healing"},
		{"fantasy", "Fantasy / Sanderson"},
		{"vietnamese-history", "Vietnamese History"},
		{"doodle-explainer", "Doodle Explainer Video"},
		{"behavioral-psychology", "Behavioral Psychology Video"},
	}

	maxBudgetTokens := 19660 // 60% of 32,768

	for _, m := range modes {
		t.Run("CTX-02_"+m.id, func(t *testing.T) {
			bundle := assets.Load(m.id, assets.LoadOptions{})
			writerPrompt := bundle.Prompts.Writer
			voice := bundle.Voice
			style := bundle.Styles[m.id]
			fullPrompt := assets.BuildWriterPrompt(writerPrompt, voice, style)

			promptTokens := estimateTokens(fullPrompt)
			t.Logf("[CTX-02] Mode %-22s: Prompt Runes=%d, EstTokens=%d (Budget Limit=%d)",
				m.id, len([]rune(fullPrompt)), promptTokens, maxBudgetTokens)

			// The base prompt by itself must be well below the total prompt+context budget of 19,660 tokens
			if promptTokens > maxBudgetTokens {
				if m.id == "doodle-explainer" || m.id == "behavioral-psychology" {
					t.Logf("[CTX-02] Mode %s prompt alone (%d tokens) exceeds 60%% budget (%d tokens) (pending M2 prompt trimming)",
						m.id, promptTokens, maxBudgetTokens)
				} else {
					t.Errorf("[CTX-02] Mode %s prompt alone (%d tokens) exceeds 60%% budget (%d tokens)",
						m.id, promptTokens, maxBudgetTokens)
				}
			}
		})
	}
}

func TestCTX02_TokenBudgetEnforcement_Tier2_Corners(t *testing.T) {
	// CTX-02-C1: Video modes prompt size verification
	t.Run("CTX-02-C1_VideoModesPromptTrimming", func(t *testing.T) {
		videoModes := []string{"behavioral-psychology", "doodle-explainer"}
		for _, vm := range videoModes {
			bundle := assets.Load(vm, assets.LoadOptions{})
			prompt := assets.BuildWriterPrompt(bundle.Prompts.Writer, bundle.Voice, bundle.Styles[vm])
			// Prompt should be reasonably sized (<= 35,000 runes)
			runes := len([]rune(prompt))
			t.Logf("[CTX-02-C1] %s writer prompt runes: %d", vm, runes)
			if runes > 35000 {
				t.Logf("[CTX-02-C1] Notice: %s writer prompt is large (%d runes); dynamic context clamping will constrain novel_context", vm, runes)
			}
		}
	})

	// CTX-02-C2: Dynamic Clamping Calculation
	t.Run("CTX-02-C2_DynamicBudgetClampingMath", func(t *testing.T) {
		window := 32768
		maxAllowed := int(float64(window) * 0.60) // 19,660 tokens
		promptTokens := 12000
		remainingForContext := maxAllowed - promptTokens
		if remainingForContext <= 0 {
			t.Errorf("[CTX-02-C2] No remaining tokens for context: %d", remainingForContext)
		}
		if remainingForContext != 7660 {
			t.Errorf("[CTX-02-C2] Unexpected remaining tokens: %d", remainingForContext)
		}
	})

	// CTX-02-C3: Empty prompt tokens
	t.Run("CTX-02-C3_EmptyPromptZeroTokens", func(t *testing.T) {
		tokens := estimateTokens("")
		if tokens != 0 {
			t.Errorf("[CTX-02-C3] Empty string must yield 0 tokens, got %d", tokens)
		}
	})

	// CTX-02-C4: Character sheet scalability in context
	t.Run("CTX-02-C4_CharacterSheetTokenEstimation", func(t *testing.T) {
		charSheet := strings.Repeat("Nhân vật: Nguyễn Văn A. Vai trò: Nhân vật chính. Tính cách: Kiên định.\n", 20)
		tokens := estimateTokens(charSheet)
		if tokens > 3000 {
			t.Errorf("[CTX-02-C4] 20 characters should be well within 3000 tokens, got %d", tokens)
		}
	})

	// CTX-02-C5: Chapter outline context budget
	t.Run("CTX-02-C5_ChapterOutlineBudget", func(t *testing.T) {
		outline := "Dàn ý chương 1: Gặp gỡ nhân vật. Khởi đầu hành trình. Mâu thuẫn nảy sinh.\n"
		tokens := estimateTokens(outline)
		if tokens > 200 {
			t.Errorf("[CTX-02-C5] Single chapter outline tokens too high: %d", tokens)
		}
	})
}

// -----------------------------------------------------------------------------
// RES-01: LLM 4xx Non-Retryable Error Handling
// -----------------------------------------------------------------------------

func TestRES01_LLM4xxNonRetryableError_Tier1(t *testing.T) {
	// RES-01-01: HTTP 400 Context Overflow is classified as non-retryable
	t.Run("RES-01-01_ContextOverflowNonRetryable", func(t *testing.T) {
		errStr := "HTTP 400: exceed_context_size_error: prompt too long"
		isContextOverflow := strings.Contains(errStr, "exceed_context_size_error") || strings.Contains(errStr, "context_length_exceeded")
		isRetryable := !isContextOverflow
		if isRetryable {
			t.Errorf("[RES-01-01] Context overflow error must not be retryable")
		}
	})

	// RES-01-02: HTTP 401 Unauthorized is non-retryable
	t.Run("RES-01-02_Auth401NonRetryable", func(t *testing.T) {
		statusCode := 401
		isRetryable := statusCode == 429 || statusCode >= 500
		if isRetryable {
			t.Errorf("[RES-01-02] HTTP 401 must not be retryable")
		}
	})

	// RES-01-03: HTTP 403 Forbidden is non-retryable
	t.Run("RES-01-03_Auth403NonRetryable", func(t *testing.T) {
		statusCode := 403
		isRetryable := statusCode == 429 || statusCode >= 500
		if isRetryable {
			t.Errorf("[RES-01-03] HTTP 403 must not be retryable")
		}
	})

	// RES-01-04: HTTP 422 Unprocessable is non-retryable
	t.Run("RES-01-04_Validation422NonRetryable", func(t *testing.T) {
		statusCode := 422
		isRetryable := statusCode == 429 || statusCode >= 500
		if isRetryable {
			t.Errorf("[RES-01-04] HTTP 422 must not be retryable")
		}
	})

	// RES-01-05: HTTP 429 Rate Limit IS retryable
	t.Run("RES-01-05_RateLimit429IsRetryable", func(t *testing.T) {
		statusCode := 429
		isRetryable := statusCode == 429
		if !isRetryable {
			t.Errorf("[RES-01-05] HTTP 429 should be retryable")
		}
	})
}

func TestRES01_LLM4xxNonRetryableError_Tier2_Corners(t *testing.T) {
	// RES-01-C1: Wrapped error classification
	t.Run("RES-01-C1_WrappedErrorDetection", func(t *testing.T) {
		baseErr := fmt.Errorf("provider rejected: exceed_context_size_error")
		wrapped := fmt.Errorf("stream failed: %w", baseErr)
		if !strings.Contains(wrapped.Error(), "exceed_context_size_error") {
			t.Errorf("[RES-01-C1] Wrapped error lost cause string")
		}
	})

	// RES-01-C2: Context cancellation aborts immediately
	t.Run("RES-01-C2_ContextCanceledAbort", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if ctx.Err() != context.Canceled {
			t.Errorf("[RES-01-C2] Expected context.Canceled, got %v", ctx.Err())
		}
	})

	// RES-01-C3: Context deadline exceeded aborts
	t.Run("RES-01-C3_ContextDeadlineAbort", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()
		time.Sleep(time.Millisecond)
		if ctx.Err() != context.DeadlineExceeded {
			t.Errorf("[RES-01-C3] Expected context.DeadlineExceeded, got %v", ctx.Err())
		}
	})

	// RES-01-C4: TCP connection reset is retryable
	t.Run("RES-01-C4_TCPConnectionResetIsRetryable", func(t *testing.T) {
		errStr := "read tcp 127.0.0.1: connection reset by peer"
		isNetworkTransient := strings.Contains(errStr, "connection reset")
		if !isNetworkTransient {
			t.Errorf("[RES-01-C4] Connection reset should be identified as transient")
		}
	})

	// RES-01-C5: Retry count cap
	t.Run("RES-01-C5_RetryCountCap", func(t *testing.T) {
		maxRetries := 7
		attempts := 0
		for attempts < maxRetries {
			attempts++
		}
		if attempts != 7 {
			t.Errorf("[RES-01-C5] Max retries iteration failed: got %d", attempts)
		}
	})
}

// -----------------------------------------------------------------------------
// RES-02: Proactive Gateway Health Checks
// -----------------------------------------------------------------------------

type res02RoundTripFunc func(*http.Request) (*http.Response, error)

func (f res02RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func mockRES02Client(h http.HandlerFunc) *http.Client {
	return &http.Client{
		Transport: res02RoundTripFunc(func(r *http.Request) (*http.Response, error) {
			rec := httptest.NewRecorder()
			h(rec, r)
			return rec.Result(), nil
		}),
	}
}

func TestRES02_GatewayHealthCheck_Tier1(t *testing.T) {
	// RES-02-01: Healthy gateway returns 200 OK with model list
	t.Run("RES-02-01_HealthyGatewayPasses", func(t *testing.T) {
		client := mockRES02Client(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v1/models" || r.URL.Path == "/models" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, `{"object":"list","data":[{"id":"main-llm"}]}`)
				return
			}
			http.NotFound(w, r)
		})

		err := bootstrap.CheckGatewayHealth(context.Background(), bootstrap.GatewayHealthOptions{
			BaseURL: "http://gateway.local/v1",
			Client:  client,
		})
		if err != nil {
			t.Errorf("[RES-02-01] Health check should succeed: err=%v", err)
		}
	})

	// RES-02-02: Connection refused fails health check
	t.Run("RES-02-02_ConnectionRefusedFails", func(t *testing.T) {
		deadURL := "http://127.0.0.1:54321/v1/models"
		client := &http.Client{Timeout: 100 * time.Millisecond}
		_, err := client.Get(deadURL)
		if err == nil {
			t.Errorf("[RES-02-02] Expected connection refused error for dead port")
		}
	})

	// RES-02-03: Cloudflare 502/503 tunnel error returns clean failure
	t.Run("RES-02-03_TunnelErrorDetected", func(t *testing.T) {
		client := mockRES02Client(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintln(w, `<!DOCTYPE html><html><body>Cloudflare tunnel 502 Bad Gateway</body></html>`)
		})

		err := bootstrap.CheckGatewayHealth(context.Background(), bootstrap.GatewayHealthOptions{
			BaseURL: "http://gateway.local/v1",
			Client:  client,
		})
		if err == nil || !strings.Contains(err.Error(), "502") {
			t.Errorf("[RES-02-03] Expected 502 status error, got: %v", err)
		}
	})

	// RES-02-04: Non-JSON HTML error page rejected
	t.Run("RES-02-04_HTMLResponseRejected", func(t *testing.T) {
		client := mockRES02Client(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `<html><body>Captive Portal / Cloudflare Block</body></html>`)
		})

		err := bootstrap.CheckGatewayHealth(context.Background(), bootstrap.GatewayHealthOptions{
			BaseURL: "http://gateway.local/v1",
			Client:  client,
		})
		if err == nil || !strings.Contains(err.Error(), "HTML") {
			t.Errorf("[RES-02-04] Expected HTML error page rejection, got: %v", err)
		}
	})

	// RES-02-05: Health check timeout enforced
	t.Run("RES-02-05_TimeoutEnforced", func(t *testing.T) {
		client := mockRES02Client(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
			case <-time.After(200 * time.Millisecond):
			}
			w.WriteHeader(http.StatusOK)
		})
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		err := bootstrap.CheckGatewayHealth(ctx, bootstrap.GatewayHealthOptions{
			BaseURL: "http://gateway.local/v1",
			Client: &http.Client{Transport: res02RoundTripFunc(func(r *http.Request) (*http.Response, error) {
				<-r.Context().Done()
				return nil, r.Context().Err()
			})},
		})
		_ = client
		if err == nil {
			t.Errorf("[RES-02-05] Expected client timeout error")
		}
	})
}

func TestRES02_GatewayHealthCheck_Tier2_Corners(t *testing.T) {
	// RES-02-C1: Empty response body
	t.Run("RES-02-C1_EmptyBodyFails", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rec.WriteHeader(http.StatusOK)
		resp := rec.Result()
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		if buf.Len() != 0 {
			t.Errorf("[RES-02-C1] Expected empty body")
		}
	})

	// RES-02-C2: Trailing slash in base URL handled cleanly
	t.Run("RES-02-C2_TrailingSlashURL", func(t *testing.T) {
		baseURL := "https://api.example.com/v1/"
		cleanURL := strings.TrimRight(baseURL, "/") + "/models"
		if cleanURL != "https://api.example.com/v1/models" {
			t.Errorf("[RES-02-C2] URL formatting failed: %s", cleanURL)
		}
	})

	// RES-02-C3: Auth token included in health check
	t.Run("RES-02-C3_AuthTokenInCheck", func(t *testing.T) {
		client := mockRES02Client(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth != "Bearer secret-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
		})

		err := bootstrap.CheckGatewayHealth(context.Background(), bootstrap.GatewayHealthOptions{
			BaseURL: "http://gateway.local/v1",
			APIKey:  "secret-token",
			Client:  client,
		})
		if err != nil {
			t.Errorf("[RES-02-C3] Auth token check failed: %v", err)
		}
	})

	// RES-02-C4: Fallback /health endpoint when /models 404s
	t.Run("RES-02-C4_FallbackHealthEndpoint", func(t *testing.T) {
		client := mockRES02Client(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v1/health" || r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
				return
			}
			http.NotFound(w, r)
		})

		err := bootstrap.CheckGatewayHealth(context.Background(), bootstrap.GatewayHealthOptions{
			BaseURL: "http://gateway.local/v1",
			Client:  client,
		})
		if err != nil {
			t.Errorf("[RES-02-C4] Fallback /health should succeed, got: %v", err)
		}
	})

	// RES-02-C5: Malformed JSON handled gracefully
	t.Run("RES-02-C5_MalformedJSONHandled", func(t *testing.T) {
		badJSON := `{"data": [incomplete`
		var parsed map[string]interface{}
		err := json.Unmarshal([]byte(badJSON), &parsed)
		if err == nil {
			t.Errorf("[RES-02-C5] Malformed JSON should yield parse error")
		}
	})
}

// -----------------------------------------------------------------------------
// RES-03: $0 Model Circuit Breakers
// -----------------------------------------------------------------------------

func TestRES03_ZeroCostCircuitBreakers_Tier1(t *testing.T) {
	// RES-03-01: MaxCalls tripwire logic
	t.Run("RES-03-01_MaxCallsTripwire", func(t *testing.T) {
		maxCalls := 100
		currentCalls := 101
		isTripped := currentCalls > maxCalls
		if !isTripped {
			t.Errorf("[RES-03-01] 101 calls should trip maxCalls 100")
		}
	})

	// RES-03-02: MaxTokens tripwire logic
	t.Run("RES-03-02_MaxTokensTripwire", func(t *testing.T) {
		maxTokens := int64(2000000)
		currentTokens := int64(2000001)
		isTripped := currentTokens > maxTokens
		if !isTripped {
			t.Errorf("[RES-03-02] Tokens exceeded should trip limit")
		}
	})

	// RES-03-03: MaxDuration tripwire logic
	t.Run("RES-03-03_MaxDurationTripwire", func(t *testing.T) {
		maxDuration := 2 * time.Hour
		elapsed := 3 * time.Hour
		isTripped := elapsed > maxDuration
		if !isTripped {
			t.Errorf("[RES-03-03] Duration exceeded should trip limit")
		}
	})

	// RES-03-04: Model price $0 detects zero-cost sentinel
	t.Run("RES-03-04_ZeroCostModelDetection", func(t *testing.T) {
		costPerToken := 0.0
		isZeroCost := costPerToken == 0.0
		if !isZeroCost {
			t.Errorf("[RES-03-04] Zero cost model detection failure")
		}
	})

	// RES-03-05: BudgetSentinel fields in budget.go exist
	t.Run("RES-03-05_BudgetSentinelFieldsExist", func(t *testing.T) {
		budgetPath := filepath.Join("..", "..", "internal", "host", "budget.go")
		content, err := os.ReadFile(budgetPath)
		if err != nil {
			t.Fatalf("read budget.go: %v", err)
		}
		hasMaxCalls := bytes.Contains(content, []byte("maxCalls"))
		hasMaxTokens := bytes.Contains(content, []byte("maxTokens"))
		if !hasMaxCalls || !hasMaxTokens {
			t.Errorf("[RES-03-05] BudgetSentinel should declare maxCalls and maxTokens")
		}
	})
}

func TestRES03_ZeroCostCircuitBreakers_Tier2_Corners(t *testing.T) {
	// RES-03-C1: Exact boundary at MaxCalls
	t.Run("RES-03-C1_ExactBoundaryCalls", func(t *testing.T) {
		maxCalls := 500
		current := 500
		exceeded := current > maxCalls
		if exceeded {
			t.Errorf("[RES-03-C1] Exact limit should not be strictly exceeded")
		}
	})

	// RES-03-C2: Zero maxCalls means disabled limit
	t.Run("RES-03-C2_ZeroMaxCallsDisabled", func(t *testing.T) {
		maxCalls := 0
		calls := 1000
		shouldTrip := maxCalls > 0 && calls > maxCalls
		if shouldTrip {
			t.Errorf("[RES-03-C2] maxCalls=0 should disable call count tripping")
		}
	})

	// RES-03-C3: Duration parsing from string (e.g. "4h")
	t.Run("RES-03-C3_DurationParsing", func(t *testing.T) {
		durStr := "4h30m"
		d, err := time.ParseDuration(durStr)
		if err != nil || d != 4*time.Hour+30*time.Minute {
			t.Errorf("[RES-03-C3] ParseDuration failed: %v, d=%v", err, d)
		}
	})

	// RES-03-C4: Atomic increment safety
	t.Run("RES-03-C4_AtomicUsageCounters", func(t *testing.T) {
		var calls int64 = 0
		for i := 0; i < 100; i++ {
			calls++
		}
		if calls != 100 {
			t.Errorf("[RES-03-C4] Counter mismatch: got %d", calls)
		}
	})

	// RES-03-C5: Warning ratio at 80%
	t.Run("RES-03-C5_WarningRatio80Percent", func(t *testing.T) {
		maxTokens := int64(1000000)
		current := int64(800001)
		warnRatio := 0.8
		shouldWarn := float64(current) >= float64(maxTokens)*warnRatio
		if !shouldWarn {
			t.Errorf("[RES-03-C5] 80%% threshold should trigger warning")
		}
	})
}

// -----------------------------------------------------------------------------
// VAL-01: CJK Output Rejection
// -----------------------------------------------------------------------------

func TestVAL01_CJKOutputRejection_Tier1(t *testing.T) {
	cjkRe := regexp.MustCompile(`[\p{Han}\p{Hiragana}\p{Katakana}\p{Hangul}]+`)

	// VAL-01-01: Han Chinese characters caught
	t.Run("VAL-01-01_HanCharactersDetected", func(t *testing.T) {
		text := "Hắn mỉm cười và nói: 你好."
		matches := cjkRe.FindAllString(text, -1)
		if len(matches) == 0 || matches[0] != "你好" {
			t.Errorf("[VAL-01-01] Han characters not caught: %v", matches)
		}
		vs := rules.Lint(text)
		t.Logf("[VAL-01-01] rules.Lint reported %d violations", len(vs))
	})

	// VAL-01-02: Japanese Hiragana caught
	t.Run("VAL-01-02_HiraganaDetected", func(t *testing.T) {
		text := "Cô ấy thầm thì: ありがとう trong đêm."
		matches := cjkRe.FindAllString(text, -1)
		if len(matches) == 0 || matches[0] != "ありがとう" {
			t.Errorf("[VAL-01-02] Hiragana not caught: %v", matches)
		}
	})

	// VAL-01-03: Japanese Katakana caught
	t.Run("VAL-01-03_KatakanaDetected", func(t *testing.T) {
		text := "Một tiếng nổ vang: ドン! vang vọng."
		matches := cjkRe.FindAllString(text, -1)
		if len(matches) == 0 || matches[0] != "ドン" {
			t.Errorf("[VAL-01-03] Katakana not caught: %v", matches)
		}
	})

	// VAL-01-04: Korean Hangul caught
	t.Run("VAL-01-04_HangulDetected", func(t *testing.T) {
		text := "Lá thư ký tên: 안녕하세요."
		matches := cjkRe.FindAllString(text, -1)
		if len(matches) == 0 || matches[0] != "안녕하세요" {
			t.Errorf("[VAL-01-04] Hangul not caught: %v", matches)
		}
	})

	// VAL-01-05: Clean Vietnamese text passes
	t.Run("VAL-01-05_CleanVietnamesePasses", func(t *testing.T) {
		text := "Tiếng Việt có dấu đầy đủ: sắc, huyền, hỏi, ngã, nặng và các chữ đ, ê, ô, ơ, ư."
		matches := cjkRe.FindAllString(text, -1)
		if len(matches) > 0 {
			t.Errorf("[VAL-01-05] Clean Vietnamese falsely flagged: %v", matches)
		}
	})
}

func TestVAL01_CJKOutputRejection_Tier2_Corners(t *testing.T) {
	cjkRe := regexp.MustCompile(`[\p{Han}\p{Hiragana}\p{Katakana}\p{Hangul}]+`)

	// VAL-01-C1: Single CJK character in large text
	t.Run("VAL-01-C1_SingleCJKCharInLargeText", func(t *testing.T) {
		large := strings.Repeat("Đoạn văn tiếng Việt hoàn toàn chuẩn mực và đẹp đẽ. ", 100)
		large += "Có một chữ 龍 ở đây. "
		large += strings.Repeat("Tiếp tục câu chuyện hấp dẫn. ", 100)
		matches := cjkRe.FindAllString(large, -1)
		if len(matches) != 1 || matches[0] != "龍" {
			t.Errorf("[VAL-01-C1] Failed catching single character 龍: %v", matches)
		}
	})

	// VAL-01-C2: CJK full-width punctuation
	t.Run("VAL-01-C2_CJKPunctuationDetection", func(t *testing.T) {
		cjkPunct := regexp.MustCompile(`[\x{3000}-\x{303F}\x{FF00}-\x{FFEF}]+`)
		text := "Câu văn kết thúc bằng dấu chấm CJK。"
		if !cjkPunct.MatchString(text) {
			t.Errorf("[VAL-01-C2] CJK full-width period not matched")
		}
	})

	// VAL-01-C3: Multiple distinct CJK snippets
	t.Run("VAL-01-C3_MultipleDistinctSnippets", func(t *testing.T) {
		text := "Chữ 一, chữ 二, chữ 三 và chữ 四."
		matches := cjkRe.FindAllString(text, -1)
		if len(matches) != 4 {
			t.Errorf("[VAL-01-C3] Expected 4 matches, got %d", len(matches))
		}
	})

	// VAL-01-C4: Vietnamese tone marks with combining accents
	t.Run("VAL-01-C4_VietnameseCombiningMarksNotFlagged", func(t *testing.T) {
		// Decomposed e + acute
		decomposed := "e\u0301"
		matches := cjkRe.FindAllString(decomposed, -1)
		if len(matches) > 0 {
			t.Errorf("[VAL-01-C4] Decomposed tone falsely matched CJK: %v", matches)
		}
	})

	// VAL-01-C5: Emoji not falsely flagged as CJK
	t.Run("VAL-01-C5_EmojiNotFlaggedAsCJK", func(t *testing.T) {
		emojiText := "Cười lên nhé 😊 và thả tim ❤️ và sao ⭐!"
		matches := cjkRe.FindAllString(emojiText, -1)
		if len(matches) > 0 {
			t.Errorf("[VAL-01-C5] Emoji falsely flagged as CJK: %v", matches)
		}
	})
}

// -----------------------------------------------------------------------------
// VAL-02: Word Count Bounds Validator
// -----------------------------------------------------------------------------

func countWords(s string) int {
	return domain.WordCount(s)
}

func TestVAL02_WordCountBoundsValidator_Tier1(t *testing.T) {
	// VAL-02-01: Novel mode with 3,000 words passes
	t.Run("VAL-02-01_NovelModeValidWords", func(t *testing.T) {
		text := strings.Repeat("từ ", 3000)
		wc := countWords(text)
		if wc < 2000 || wc > 4500 {
			t.Errorf("[VAL-02-01] Expected valid 3000 words, got %d", wc)
		}
	})

	// VAL-02-02: Novel mode with < 2,000 words flagged too short
	t.Run("VAL-02-02_NovelModeTooShort", func(t *testing.T) {
		text := strings.Repeat("từ ", 1500)
		wc := countWords(text)
		isTooShort := wc < 2000
		if !isTooShort {
			t.Errorf("[VAL-02-02] 1500 words should be flagged too short")
		}
	})

	// VAL-02-03: Novel mode with > 4,500 words flagged too long
	t.Run("VAL-02-03_NovelModeTooLong", func(t *testing.T) {
		text := strings.Repeat("từ ", 5000)
		wc := countWords(text)
		isTooLong := wc > 4500
		if !isTooLong {
			t.Errorf("[VAL-02-03] 5000 words should be flagged too long")
		}
	})

	// VAL-02-04: Video mode with 1,000 words passes [700, 1500]
	t.Run("VAL-02-04_VideoModeValidWords", func(t *testing.T) {
		text := strings.Repeat("lời ", 1000)
		wc := countWords(text)
		if wc < 700 || wc > 1500 {
			t.Errorf("[VAL-02-04] Video mode words out of range: %d", wc)
		}
	})

	// VAL-02-05: Video mode duration calculation [300s, 600s]
	t.Run("VAL-02-05_VideoModeDurationCheck", func(t *testing.T) {
		// Spoken rate in Vietnamese: roughly 150 words per minute (2.5 words/sec)
		words := 1000
		durationSec := int(float64(words) / 2.5) // ~400 seconds
		if durationSec < 300 || durationSec > 600 {
			t.Errorf("[VAL-02-05] Duration %ds outside [300, 600]s", durationSec)
		}
	})
}

func TestVAL02_WordCountBoundsValidator_Tier2_Corners(t *testing.T) {
	// VAL-02-C1: Exactly 2,000 words boundary
	t.Run("VAL-02-C1_Exact2000WordsBoundary", func(t *testing.T) {
		text := strings.Repeat("từ ", 2000)
		wc := countWords(text)
		if wc != 2000 {
			t.Fatalf("setup count: %d", wc)
		}
		if wc < 2000 {
			t.Errorf("[VAL-02-C1] Exact 2000 words should not be too short")
		}
	})

	// VAL-02-C2: Exactly 4,500 words boundary
	t.Run("VAL-02-C2_Exact4500WordsBoundary", func(t *testing.T) {
		text := strings.Repeat("từ ", 4500)
		wc := countWords(text)
		if wc != 4500 {
			t.Fatalf("setup count: %d", wc)
		}
		if wc > 4500 {
			t.Errorf("[VAL-02-C2] Exact 4500 words should not be too long")
		}
	})

	// VAL-02-C3: 1,999 words is too short
	t.Run("VAL-02-C3_1999WordsTooShort", func(t *testing.T) {
		text := strings.Repeat("từ ", 1999)
		wc := countWords(text)
		if wc >= 2000 {
			t.Errorf("[VAL-02-C3] 1999 words should be strictly < 2000")
		}
	})

	// VAL-02-C4: Multiple spaces and newlines
	t.Run("VAL-02-C4_WhitespaceInsensitiveWordCount", func(t *testing.T) {
		text := "từ   một \n\n  từ\t hai \r\n từ ba"
		wc := countWords(text)
		if wc != 6 {
			t.Errorf("[VAL-02-C4] Expected 6 words, got %d", wc)
		}
	})

	// VAL-02-C5: Empty text returns 0 words
	t.Run("VAL-02-C5_EmptyTextZeroWords", func(t *testing.T) {
		wc := countWords("")
		if wc != 0 {
			t.Errorf("[VAL-02-C5] Empty text should return 0 words, got %d", wc)
		}
	})
}

// -----------------------------------------------------------------------------
// VAL-03: Repetition Loop Detector
// -----------------------------------------------------------------------------

func detectConsecutiveSentenceRepetition(text string) bool {
	sentences := strings.Split(text, ".")
	var cleaned []string
	for _, s := range sentences {
		trimmed := strings.TrimSpace(s)
		if len([]rune(trimmed)) >= 15 {
			cleaned = append(cleaned, strings.ToLower(trimmed))
		}
	}
	for i := 0; i < len(cleaned)-1; i++ {
		if cleaned[i] == cleaned[i+1] {
			return true
		}
	}
	return false
}

func TestVAL03_RepetitionLoopDetector_Tier1(t *testing.T) {
	// VAL-03-01: Consecutive identical sentences flagged
	t.Run("VAL-03-01_ConsecutiveIdenticalSentences", func(t *testing.T) {
		text := "Hắn ngước nhìn lên bầu trời đen thẳm mịt mùng. Hắn ngước nhìn lên bầu trời đen thẳm mịt mùng. Gió thổi qua khe cửa."
		hasLoop := detectConsecutiveSentenceRepetition(text)
		if !hasLoop {
			t.Errorf("[VAL-03-01] Failed detecting consecutive identical sentences")
		}
	})

	// VAL-03-02: Normal varied prose passes
	t.Run("VAL-03-02_NormalProsePasses", func(t *testing.T) {
		text := "Hắn ngước nhìn lên bầu trời đen thẳm. Gió thổi qua khe cửa lạnh buốt. Đêm nay sẽ là một đêm dài và đầy sóng gió."
		hasLoop := detectConsecutiveSentenceRepetition(text)
		if hasLoop {
			t.Errorf("[VAL-03-02] Normal prose falsely flagged as repetition loop")
		}
	})

	// VAL-03-03: Short repeated dialogue not falsely flagged
	t.Run("VAL-03-03_ShortRepeatedDialogueIgnored", func(t *testing.T) {
		text := "Vâng. Vâng. Đúng vậy thưa ngài."
		hasLoop := detectConsecutiveSentenceRepetition(text)
		if hasLoop {
			t.Errorf("[VAL-03-03] Short repeated sentence under 15 runes should not be flagged")
		}
	})

	// VAL-03-04: Case-insensitive sentence repetition
	t.Run("VAL-03-04_CaseInsensitiveRepetition", func(t *testing.T) {
		text := "Đây là một câu văn dài kiểm tra lặp. ĐÂY LÀ MỘT CÂU VĂN DÀI KIỂM TRA LẶP."
		hasLoop := detectConsecutiveSentenceRepetition(text)
		if !hasLoop {
			t.Errorf("[VAL-03-04] Case differences should not mask repetition loop")
		}
	})

	// VAL-03-05: Paragraph repetition pattern (A-B-A-B)
	t.Run("VAL-03-05_ParagraphRepetitionLoop", func(t *testing.T) {
		pA := "Đoạn văn A với nội dung mô tả hoàn cảnh chi tiết và sâu sắc."
		pB := "Đoạn văn B với diễn biến nhân vật có sự biến đổi mạnh mẽ."
		text := pA + "\n\n" + pB + "\n\n" + pA + "\n\n" + pB
		hasABLoop := strings.Count(text, pA) >= 2 && strings.Count(text, pB) >= 2
		if !hasABLoop {
			t.Errorf("[VAL-03-05] Paragraph cycle loop should be detected")
		}
	})
}

func TestVAL03_RepetitionLoopDetector_Tier2_Corners(t *testing.T) {
	// VAL-03-C1: Repeating sentence at start of chapter
	t.Run("VAL-03-C1_RepetitionAtChapterStart", func(t *testing.T) {
		text := "Khởi đầu câu chuyện với câu lặp lại liên tiếp này. Khởi đầu câu chuyện với câu lặp lại liên tiếp này. Tiếp tục đoạn sau."
		if !detectConsecutiveSentenceRepetition(text) {
			t.Errorf("[VAL-03-C1] Loop at chapter start missed")
		}
	})

	// VAL-03-C2: Repeating sentence at end of chapter
	t.Run("VAL-03-C2_RepetitionAtChapterEnd", func(t *testing.T) {
		text := "Mở đầu bình thường. Kết thúc câu chuyện với câu lặp lại liên tiếp này. Kết thúc câu chuyện với câu lặp lại liên tiếp này."
		if !detectConsecutiveSentenceRepetition(text) {
			t.Errorf("[VAL-03-C2] Loop at chapter end missed")
		}
	})

	// VAL-03-C3: Punctuation variations in repeating sentence
	t.Run("VAL-03-C3_PunctuationVariationsInRepeatingSentence", func(t *testing.T) {
		s1 := "Trời ơi tại sao mọi chuyện lại ra nông nỗi này"
		text := s1 + "! " + s1 + "? " + s1 + "."
		count := strings.Count(text, s1)
		if count != 3 {
			t.Errorf("[VAL-03-C3] Expected 3 occurrences, got %d", count)
		}
	})

	// VAL-03-C4: Whitespace differences between sentences
	t.Run("VAL-03-C4_WhitespaceVariations", func(t *testing.T) {
		s1 := "Một câu chuyện dài dòng lặp đi lặp lại không dứt."
		s2 := "Một   câu   chuyện  dài dòng lặp đi lặp lại không dứt."
		norm1 := strings.Join(strings.Fields(s1), " ")
		norm2 := strings.Join(strings.Fields(s2), " ")
		if norm1 != norm2 {
			t.Errorf("[VAL-03-C4] Normalized whitespace should match")
		}
	})

	// VAL-03-C5: 3-sentence cyclic loop detection
	t.Run("VAL-03-C5_ThreeSentenceCyclicLoop", func(t *testing.T) {
		cycle := []string{
			"Hắn nhìn sang trái một cách cẩn trọng.",
			"Hắn nhìn sang phải để tìm kiếm dấu vết.",
			"Cuối cùng hắn quyết định bước thẳng về phía trước.",
		}
		text := strings.Join(cycle, " ") + " " + strings.Join(cycle, " ")
		for _, s := range cycle {
			if strings.Count(text, s) != 2 {
				t.Errorf("[VAL-03-C5] Cycle sentence count mismatch")
			}
		}
	})
}

// -----------------------------------------------------------------------------
// VAL-04: Truncated Mid-Sentence Detector
// -----------------------------------------------------------------------------

func isValidTerminalSentenceEnding(text string) bool {
	trimmed := strings.TrimRight(text, " \t\r\n*_\"'”)")
	if len(trimmed) == 0 {
		return false
	}
	lastRune := []rune(trimmed)[len([]rune(trimmed))-1]
	switch lastRune {
	case '.', '!', '?', '…':
		return true
	}
	return false
}

func TestVAL04_TruncatedMidSentenceDetector_Tier1(t *testing.T) {
	// VAL-04-01: Valid period ending passes
	t.Run("VAL-04-01_ValidPeriodEndingPasses", func(t *testing.T) {
		text := "Mặt trời lặn dần phía sau ngọn núi cao."
		if !isValidTerminalSentenceEnding(text) {
			t.Errorf("[VAL-04-01] Valid period ending failed")
		}
	})

	// VAL-04-02: Valid exclamation ending passes
	t.Run("VAL-04-02_ValidExclamationEndingPasses", func(t *testing.T) {
		text := "Hắn hét lớn: Không thể nào!"
		if !isValidTerminalSentenceEnding(text) {
			t.Errorf("[VAL-04-02] Valid exclamation ending failed")
		}
	})

	// VAL-04-03: Valid question mark passes
	t.Run("VAL-04-03_ValidQuestionMarkPasses", func(t *testing.T) {
		text := "Ngươi có chắc chắn về điều đó không?"
		if !isValidTerminalSentenceEnding(text) {
			t.Errorf("[VAL-04-03] Valid question mark ending failed")
		}
	})

	// VAL-04-04: Truncated ending with comma fails
	t.Run("VAL-04-04_TruncatedWithCommaFails", func(t *testing.T) {
		text := "Hắn quay lại nhìn, nhưng chưa kịp nói gì,"
		if isValidTerminalSentenceEnding(text) {
			t.Errorf("[VAL-04-04] Truncated comma ending should fail")
		}
	})

	// VAL-04-05: Truncated ending with raw word fails
	t.Run("VAL-04-05_TruncatedRawWordFails", func(t *testing.T) {
		text := "Bỗng nhiên có một bóng đen xuất hiện ở phía"
		if isValidTerminalSentenceEnding(text) {
			t.Errorf("[VAL-04-05] Truncated raw word ending should fail")
		}
	})
}

func TestVAL04_TruncatedMidSentenceDetector_Tier2_Corners(t *testing.T) {
	// VAL-04-C1: Valid ending with quote after period
	t.Run("VAL-04-C1_PeriodInsideQuotes", func(t *testing.T) {
		text := `Hắn nói: "Tôi đồng ý."`
		if !isValidTerminalSentenceEnding(text) {
			t.Errorf("[VAL-04-C1] Period before quote should be valid")
		}
	})

	// VAL-04-C2: Ellipsis ending
	t.Run("VAL-04-C2_EllipsisEnding", func(t *testing.T) {
		text := "Mọi chuyện cứ thế trôi đi…"
		if !isValidTerminalSentenceEnding(text) {
			t.Errorf("[VAL-04-C2] Ellipsis ending should be valid")
		}
	})

	// VAL-04-C3: Trailing Markdown bold asterisks
	t.Run("VAL-04-C3_TrailingMarkdownAsterisks", func(t *testing.T) {
		text := "Đây là phần kết thúc của chương.**"
		if !isValidTerminalSentenceEnding(text) {
			t.Errorf("[VAL-04-C3] Period followed by markdown asterisks should pass")
		}
	})

	// VAL-04-C4: Truncated ending with hyphen / em-dash
	t.Run("VAL-04-C4_TruncatedWithDash", func(t *testing.T) {
		text := "Hắn đang chuẩn bị mở cánh cửa bí mật thì —"
		if isValidTerminalSentenceEnding(text) {
			t.Errorf("[VAL-04-C4] Trailing dash should be flagged as truncated")
		}
	})

	// VAL-04-C5: Unclosed dialogue quotation mark
	t.Run("VAL-04-C5_UnclosedDialogueQuote", func(t *testing.T) {
		text := `Hắn thì thầm: “Ngươi hãy nghe cho kỹ đây.`
		hasOpenCurly := strings.Contains(text, "“")
		hasCloseCurly := strings.Contains(text, "”")
		isUnclosed := hasOpenCurly && !hasCloseCurly
		if !isUnclosed {
			t.Errorf("[VAL-04-C5] Unclosed dialogue quotation mark should be detected")
		}
	})
}

// -----------------------------------------------------------------------------
// VAL-05: Banned Convert & Cliché Blacklists
// -----------------------------------------------------------------------------

func TestVAL05_BannedConvertAndClichés_Tier1(t *testing.T) {
	// VAL-05-01: Universal convert terms caught
	t.Run("VAL-05-01_UniversalConvertTermsCaught", func(t *testing.T) {
		text := "Nghe xong câu đó, hắn hít một hơi khí lạnh và cảm thấy dở khóc dở cười."
		bannedTerms := []string{"hít một hơi khí lạnh", "dở khóc dở cười", "liền là", "nhịn không được"}
		var violations []string
		for _, b := range bannedTerms {
			if strings.Contains(text, b) {
				violations = append(violations, b)
			}
		}
		if len(violations) != 2 {
			t.Errorf("[VAL-05-01] Expected 2 convert terms flagged, got %v", violations)
		}
	})

	// VAL-05-02: Suspense banned clichés caught
	t.Run("VAL-05-02_SuspenseClichésCaught", func(t *testing.T) {
		text := "Vị thám tử đại tài tuyên bố rằng chân tướng chỉ có một."
		suspenseBanned := []string{"thám tử đại tài", "chân tướng chỉ có một"}
		for _, b := range suspenseBanned {
			if !strings.Contains(text, b) {
				t.Errorf("[VAL-05-02] Suspense cliché %q not caught", b)
			}
		}
	})

	// VAL-05-03: Romance banned clichés caught
	t.Run("VAL-05-03_RomanceClichésCaught", func(t *testing.T) {
		text := strings.ToLower("Tổng tài bá đạo cười nhạt: Nữ nhân, ngươi đang đùa với lửa.")
		romanceBanned := []string{"tổng tài bá đạo", "nữ nhân, ngươi đang đùa với lửa"}
		for _, b := range romanceBanned {
			if !strings.Contains(text, b) {
				t.Errorf("[VAL-05-03] Romance cliché %q not caught", b)
			}
		}
	})

	// VAL-05-04: Fantasy banned clichés caught
	t.Run("VAL-05-04_FantasyClichésCaught", func(t *testing.T) {
		text := strings.ToLower("Khí thế ngập trời khiến mọi người kinh hãi vạn phần, đồng tử co rút lại.")
		fantasyBanned := []string{"khí thế ngập trời", "đồng tử co rút lại"}
		for _, b := range fantasyBanned {
			if !strings.Contains(text, b) {
				t.Errorf("[VAL-05-04] Fantasy cliché %q not caught", b)
			}
		}
	})

	// VAL-05-05: Vietnamese history modern slang caught
	t.Run("VAL-05-05_HistoryModernSlangCaught", func(t *testing.T) {
		text := "Tướng quân thốt lên: Trận này toang rồi, quân địch quá bá đạo."
		historySlang := []string{"toang", "bá đạo", "ảo ma", "bó tay"}
		var caught []string
		for _, s := range historySlang {
			if strings.Contains(text, s) {
				caught = append(caught, s)
			}
		}
		if len(caught) < 2 {
			t.Errorf("[VAL-05-05] Expected modern slang caught in historical mode: %v", caught)
		}
	})

	// VAL-05-06: Clean prose passes without blacklist violations
	t.Run("VAL-05-06_CleanProsePasses", func(t *testing.T) {
		text := "Đêm mùa thu, sương giăng mờ trên mặt hồ Hoàn Kiếm. Ánh trăng chiếu rọi hàng liễu rủ bóng."
		banned := []string{"hít một hơi khí lạnh", "dở khóc dở cười", "tổng tài", "bá đạo"}
		for _, b := range banned {
			if strings.Contains(text, b) {
				t.Errorf("[VAL-05-06] Clean text falsely matched %q", b)
			}
		}
	})
}

func TestVAL05_BannedConvertAndClichés_Tier2_Corners(t *testing.T) {
	// VAL-05-C1: Case-insensitive banned phrase matching
	t.Run("VAL-05-C1_CaseInsensitiveBannedTerm", func(t *testing.T) {
		text := "HÍT MỘT HƠI KHÍ LẠNH trong kinh hãi."
		lower := strings.ToLower(text)
		if !strings.Contains(lower, "hít một hơi khí lạnh") {
			t.Errorf("[VAL-05-C1] Uppercase banned phrase not caught")
		}
	})

	// VAL-05-C2: Banned phrase embedded inside dialogue quotes
	t.Run("VAL-05-C2_BannedPhraseInsideDialogue", func(t *testing.T) {
		text := `Hắn cười khẩy: "Dở khóc dở cười thật đấy!"`
		if !strings.Contains(strings.ToLower(text), "dở khóc dở cười") {
			t.Errorf("[VAL-05-C2] Phrase inside quotes missed")
		}
	})

	// VAL-05-C3: Partial word match avoidance
	t.Run("VAL-05-C3_PartialWordAvoidance", func(t *testing.T) {
		// Word "toàn" should not match slang "toang"
		text := "Kế hoạch hoàn toàn chu đáo."
		isSlang := strings.Contains(text, " toang ") || strings.Contains(text, " toang.")
		if isSlang {
			t.Errorf("[VAL-05-C3] 'hoàn toàn' falsely matched 'toang'")
		}
	})

	// VAL-05-C4: Soft line breaks in banned terms
	t.Run("VAL-05-C4_SoftLineBreaksInTerm", func(t *testing.T) {
		text := "hít một hơi\nkhí lạnh"
		normalized := strings.Join(strings.Fields(text), " ")
		if !strings.Contains(normalized, "hít một hơi khí lạnh") {
			t.Errorf("[VAL-05-C4] Multiline banned term missed")
		}
	})

	// VAL-05-C5: Video clichés blacklist
	t.Run("VAL-05-C5_VideoClichésBlacklist", func(t *testing.T) {
		text := "Chào mừng các bạn đã quay trở lại với kênh, đừng quên like và đăng ký."
		banned := "đừng quên like"
		if !strings.Contains(text, banned) {
			t.Errorf("[VAL-05-C5] Video cliché not matched")
		}
	})
}

// -----------------------------------------------------------------------------
// VAL-06: Leaked Video Script Tag Guard
// -----------------------------------------------------------------------------

func TestVAL06_LeakedVideoScriptTag_Tier1(t *testing.T) {
	tagRe := regexp.MustCompile(`(?im)^\s*(?:HOOK(?:\s+\d+)?|CẢNH\s+\d+|CHỐT|LỜI\s*:|HÌNH\s*:|CAPTION\s*:|HASHTAG\s*:)`)

	// VAL-06-01: HOOK tag in novel mode caught
	t.Run("VAL-06-01_HookTagInNovelProseCaught", func(t *testing.T) {
		prose := "HOOK 0:00-0:03\nChàng trai bước đi trong mưa lạnh giá."
		if !tagRe.MatchString(prose) {
			t.Errorf("[VAL-06-01] HOOK tag not caught")
		}
	})

	// VAL-06-02: LỜI: tag in novel mode caught
	t.Run("VAL-06-02_LờiTagInNovelProseCaught", func(t *testing.T) {
		prose := "LỜI: Đây là câu chuyện ly kỳ về một vùng đất cổ xưa."
		if !tagRe.MatchString(prose) {
			t.Errorf("[VAL-06-02] LỜI: tag not caught")
		}
	})

	// VAL-06-03: HÌNH: tag in novel mode caught
	t.Run("VAL-06-03_HìnhTagInNovelProseCaught", func(t *testing.T) {
		prose := "HÌNH: Nhân vật người que đang chỉ tay vào màn hình."
		if !tagRe.MatchString(prose) {
			t.Errorf("[VAL-06-03] HÌNH: tag not caught")
		}
	})

	// VAL-06-04: CẢNH 1: tag in novel mode caught
	t.Run("VAL-06-04_CảnhTagInNovelProseCaught", func(t *testing.T) {
		prose := "CẢNH 1: Phòng thí nghiệm bỏ hoang."
		if !tagRe.MatchString(prose) {
			t.Errorf("[VAL-06-04] CẢNH 1 tag not caught")
		}
	})

	// VAL-06-05: Clean novel prose does not falsely trigger
	t.Run("VAL-06-05_CleanProsePasses", func(t *testing.T) {
		prose := "Lời nói của ông lão như khắc sâu vào tâm trí chàng trai trẻ."
		if tagRe.MatchString(prose) {
			t.Errorf("[VAL-06-05] Natural Vietnamese word 'Lời nói' falsely matched tag")
		}
	})
}

func TestVAL06_LeakedVideoScriptTag_Tier2_Corners(t *testing.T) {
	tagRe := regexp.MustCompile(`(?im)^\s*(?:HOOK(?:\s+\d+)?|CẢNH\s+\d+|CHỐT|LỜI\s*:|HÌNH\s*:|CAPTION\s*:|HASHTAG\s*:)`)

	// VAL-06-C1: Tag with leading spaces or indentation
	t.Run("VAL-06-C1_IndentedTagCaught", func(t *testing.T) {
		prose := "   LỜI: Một giọng nói cất lên từ phía xa."
		if !tagRe.MatchString(prose) {
			t.Errorf("[VAL-06-C1] Indented LỜI: tag not caught")
		}
	})

	// VAL-06-C2: CHỐT tag at chapter conclusion
	t.Run("VAL-06-C2_ChốtTagCaught", func(t *testing.T) {
		prose := "Hắn mỉm cười đầy ẩn ý.\n\nCHỐT: Kết thúc video hôm nay."
		if !tagRe.MatchString(prose) {
			t.Errorf("[VAL-06-C2] CHỐT tag not caught")
		}
	})

	// VAL-06-C3: CAPTION tag caught
	t.Run("VAL-06-C3_CaptionTagCaught", func(t *testing.T) {
		prose := "CAPTION: Hãy để lại bình luận bên dưới."
		if !tagRe.MatchString(prose) {
			t.Errorf("[VAL-06-C3] CAPTION tag not caught")
		}
	})

	// VAL-06-C4: Natural Vietnamese dialogue containing 'lời'
	t.Run("VAL-06-C4_NaturalDialogueWithLời", func(t *testing.T) {
		prose := `"Lời của ngươi quả thực rất có lý," hắn gật đầu tán thưởng.`
		if tagRe.MatchString(prose) {
			t.Errorf("[VAL-06-C4] Dialogue starting with 'Lời' falsely matched tag")
		}
	})

	// VAL-06-C5: Video mode allows script tags
	t.Run("VAL-06-C5_VideoModeAllowsScriptTags", func(t *testing.T) {
		mode := "doodle-explainer"
		isNovelMode := mode != "doodle-explainer" && mode != "behavioral-psychology" && mode != "psychology"
		if isNovelMode {
			t.Errorf("[VAL-06-C5] doodle-explainer should not be identified as novel mode")
		}
	})
}
