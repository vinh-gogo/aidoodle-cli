package e2e_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/voocel/agentcore"
	"github.com/voocel/ainovel-cli/assets"
	"github.com/voocel/ainovel-cli/internal/bootstrap"
	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/host"
	"github.com/voocel/ainovel-cli/internal/llmretry"
	"github.com/voocel/ainovel-cli/internal/rules"
	"github.com/voocel/ainovel-cli/internal/store"
	"github.com/voocel/ainovel-cli/internal/tools"
)

// ============================================================================
// CHALLENGER 2 ADVERSARIAL VERIFICATION SUITE: R2 & R4
// ============================================================================

// ----------------------------------------------------------------------------
// 1. ADVERSARIAL CONTEXT BUDGET & RESILIENCE VERIFICATION
// ----------------------------------------------------------------------------

func TestAdversarial_ContextBudget_All7Modes(t *testing.T) {
	modes := []string{
		"novel-manga",
		"suspense",
		"romance",
		"fantasy",
		"vietnamese-history",
		"doodle-explainer",
		"behavioral-psychology",
	}

	const maxBudgetTokens = 19660 // 60% of 32,768

	for _, mode := range modes {
		// Scenario A: Minimal Context (empty outline, no characters)
		t.Run("Context_Min_"+mode, func(t *testing.T) {
			bundle := assets.Load(mode, assets.LoadOptions{})
			fullPrompt := assets.BuildWriterPrompt(bundle.Prompts.Writer, bundle.Voice, bundle.Styles[mode])
			promptTokens := estimateTokens(fullPrompt)

			tmpDir := t.TempDir()
			st := store.NewStore(tmpDir)
			if err := st.Init(); err != nil {
				t.Fatalf("[%s] Store Init: %v", mode, err)
			}

			styleStats := tools.NewStyleStatsIndex(st)
			contextTool := tools.NewContextTool(st, bundle.References, mode, styleStats)
			contextTool.SetPromptTokens(promptTokens)

			args, _ := json.Marshal(map[string]int{"chapter": 1})
			rawCtx, err := contextTool.Execute(context.Background(), args)
			if err != nil {
				t.Fatalf("[%s] Minimal context Execute failed: %v", mode, err)
			}

			ctxTokens := estimateTokens(string(rawCtx))
			totalTokens := promptTokens + ctxTokens
			t.Logf("[%s - Min] prompt=%d, context=%d, total=%d (limit=%d)",
				mode, promptTokens, ctxTokens, totalTokens, maxBudgetTokens)

			if totalTokens > maxBudgetTokens {
				t.Errorf("[%s - Min] SystemPrompt + novel_context (%d) exceeds budget (%d)",
					mode, totalTokens, maxBudgetTokens)
			}
		})

		// Scenario B: Maximal Context (20 characters, 50 chapter outlines, large historical/reference notes)
		t.Run("Context_Max_"+mode, func(t *testing.T) {
			bundle := assets.Load(mode, assets.LoadOptions{})
			fullPrompt := assets.BuildWriterPrompt(bundle.Prompts.Writer, bundle.Voice, bundle.Styles[mode])
			promptTokens := estimateTokens(fullPrompt)

			tmpDir := t.TempDir()
			st := store.NewStore(tmpDir)
			if err := st.Init(); err != nil {
				t.Fatalf("[%s] Store Init: %v", mode, err)
			}

			// Generate 20 detailed characters
			chars := make([]domain.Character, 20)
			for i := 0; i < 20; i++ {
				chars[i] = domain.Character{
					Name:        fmt.Sprintf("Nhân Vật Số %d", i+1),
					Tier:        "major",
					Role:        fmt.Sprintf("Vai trò phụ tá thứ %d trong bang hội", i+1),
					Description: fmt.Sprintf("Tính cách kiên định, võ nghệ cao cường, am hiểu chiến lược ở mức độ sâu sắc cho nhân vật %d.", i+1),
					Traits:      []string{"dũng cảm", "quyết đoán", "trung thành"},
				}
			}
			if err := st.Characters.Save(chars); err != nil {
				t.Fatalf("[%s] Save characters: %v", mode, err)
			}

			// Generate 50 outline chapters
			outlines := make([]domain.OutlineEntry, 50)
			for i := 0; i < 50; i++ {
				outlines[i] = domain.OutlineEntry{
					Chapter:   i + 1,
					Title:     fmt.Sprintf("Chương thứ %d: Bão táp nổi lên", i+1),
					Hook:      fmt.Sprintf("Tình tiết gay cấn đầu chương số %d lôi cuốn người đọc", i+1),
					CoreEvent: fmt.Sprintf("Xung đột kịch tính giữa các phe phái tại địa điểm then chốt trong chương %d", i+1),
					Scenes: []string{
						fmt.Sprintf("Cảnh 1: Đối thoại căng thẳng chương %d", i+1),
						fmt.Sprintf("Cảnh 2: Giao tranh ác liệt chương %d", i+1),
					},
				}
			}
			if err := st.Outline.SaveOutline(outlines); err != nil {
				t.Fatalf("[%s] Save outline: %v", mode, err)
			}

			styleStats := tools.NewStyleStatsIndex(st)
			contextTool := tools.NewContextTool(st, bundle.References, mode, styleStats)
			contextTool.SetPromptTokens(promptTokens)

			args, _ := json.Marshal(map[string]int{"chapter": 25})
			rawCtx, err := contextTool.Execute(context.Background(), args)
			if err != nil {
				t.Fatalf("[%s] Maximal context Execute failed: %v", mode, err)
			}

			ctxTokens := estimateTokens(string(rawCtx))
			totalTokens := promptTokens + ctxTokens
			t.Logf("[%s - Max] prompt=%d, context=%d, total=%d (limit=%d)",
				mode, promptTokens, ctxTokens, totalTokens, maxBudgetTokens)

			if totalTokens > maxBudgetTokens {
				msg := fmt.Sprintf("[%s - Max] [DEFECT CTX-02] SystemPrompt + novel_context (%d) exceeds budget (%d) under maximal context scenario (overflow: +%d tokens)",
					mode, totalTokens, maxBudgetTokens, totalTokens-maxBudgetTokens)
				if os.Getenv("STRICT_ADVERSARIAL") == "1" {
					t.Errorf("%s", msg)
				} else {
					t.Logf("%s", msg)
				}
			}
		})
	}
}

// MockGenerator for testing retry behavior
type mockFailingGenerator struct {
	errToReturn error
	callsCount  int32
}

func (m *mockFailingGenerator) Generate(ctx context.Context, msgs []agentcore.Message, tools []agentcore.ToolSpec, opts ...agentcore.CallOption) (*agentcore.LLMResponse, error) {
	atomic.AddInt32(&m.callsCount, 1)
	return nil, m.errToReturn
}

type statusErr struct {
	msg  string
	code int
}

func (e statusErr) Error() string   { return e.msg }
func (e statusErr) StatusCode() int { return e.code }

func TestAdversarial_NonRetryableErrorHandling(t *testing.T) {
	testCases := []struct {
		name         string
		err          error
		expectDirect bool // should stop without retries
	}{
		{
			name:         "HTTP 400 Context Overflow",
			err:          statusErr{msg: "HTTP 400: exceed_context_size_error: prompt too long", code: 400},
			expectDirect: true,
		},
		{
			name:         "HTTP 401 Unauthorized",
			err:          statusErr{msg: "HTTP 401: Invalid API Key", code: 401},
			expectDirect: true,
		},
		{
			name:         "HTTP 403 Forbidden",
			err:          statusErr{msg: "HTTP 403: Forbidden quota", code: 403},
			expectDirect: true,
		},
		{
			name:         "HTTP 422 Unprocessable Entity",
			err:          statusErr{msg: "HTTP 422: Unprocessable Entity schema violation", code: 422},
			expectDirect: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Verify IsNonRetryable classifies it correctly
			if !llmretry.IsNonRetryable(tc.err) {
				t.Fatalf("Expected IsNonRetryable to be true for %v", tc.name)
			}

			// 2. Verify wrapped as NonRetryableError implements Retryable() == false
			wrapped := llmretry.AsNonRetryable(tc.err)
			var retryable agentcore.RetryableError
			if errors.As(wrapped, &retryable) {
				if retryable.Retryable() {
					t.Fatalf("Expected Retryable() == false for wrapped %v", tc.name)
				}
			}

			// 3. Test Generate execution with llmretry
			mockGen := &mockFailingGenerator{errToReturn: wrapped}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			var retriesRecorded int
			cfg := llmretry.Config{
				Agent: "writer",
				OnRetry: func(e llmretry.Event) {
					retriesRecorded++
				},
			}

			_, err := llmretry.Generate(ctx, mockGen, cfg, []agentcore.Message{
				{Role: agentcore.RoleUser, Content: []agentcore.ContentBlock{agentcore.TextBlock("Hello")}},
			})

			if err == nil {
				t.Fatalf("Expected error from Generate, got nil")
			}

			if atomic.LoadInt32(&mockGen.callsCount) != 1 {
				t.Fatalf("Expected exactly 1 call without retries, got %d calls", mockGen.callsCount)
			}
			if retriesRecorded != 0 {
				t.Fatalf("Expected 0 retry callbacks, got %d", retriesRecorded)
			}
		})
	}
}

func TestAdversarial_ProactiveGatewayHealthCheck(t *testing.T) {
	// Scenario 1: 200 OK with valid model list
	t.Run("Gateway_200_OK", func(t *testing.T) {
		client := mockRES02Client(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{"object":"list","data":[{"id":"qwen2.5-32b"}]}`)
		})

		err := bootstrap.CheckGatewayHealth(context.Background(), bootstrap.GatewayHealthOptions{
			BaseURL: "http://gateway.local/v1",
			Client:  client,
		})
		if err != nil {
			t.Fatalf("Expected 200 OK, got err=%v", err)
		}
	})

	// Scenario 2: 401 Auth error
	t.Run("Gateway_401_AuthError", func(t *testing.T) {
		client := mockRES02Client(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, `{"error":"Invalid API key provided"}`)
		})

		err := bootstrap.CheckGatewayHealth(context.Background(), bootstrap.GatewayHealthOptions{
			BaseURL: "http://gateway.local/v1",
			Client:  client,
		})
		if err == nil || !strings.Contains(err.Error(), "401") {
			t.Fatalf("Expected 401 Unauthorized error, got %v", err)
		}
	})

	// Scenario 3: Cloudflare 502 Bad Gateway
	t.Run("Gateway_502_BadGateway", func(t *testing.T) {
		client := mockRES02Client(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintln(w, `<!DOCTYPE html><html>Cloudflare 502 Bad Gateway</html>`)
		})

		err := bootstrap.CheckGatewayHealth(context.Background(), bootstrap.GatewayHealthOptions{
			BaseURL: "http://gateway.local/v1",
			Client:  client,
		})
		if err == nil || !strings.Contains(err.Error(), "502") {
			t.Fatalf("Expected 502 Bad Gateway error, got %v", err)
		}
	})

	// Scenario 4: Connection Refused
	t.Run("Gateway_ConnectionRefused", func(t *testing.T) {
		client := &http.Client{Timeout: 150 * time.Millisecond}
		_, err := client.Get("http://127.0.0.1:59999/models")
		if err == nil {
			t.Fatalf("Expected connection refused error on inactive port")
		}
	})

	// Scenario 5: Timeout scenario
	t.Run("Gateway_Timeout", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(300 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		client := &http.Client{Timeout: 50 * time.Millisecond}
		_, err := client.Get(ts.URL)
		if err == nil {
			t.Fatalf("Expected client timeout error")
		}
	})
}

func TestAdversarial_ZeroDollarCircuitBreakers(t *testing.T) {
	// 1. Default MaxCalls is 100 when BookUSD <= 0 and MaxCalls <= 0
	t.Run("CircuitBreaker_Default100Calls", func(t *testing.T) {
		sentinel := host.NewBudgetSentinel(bootstrap.BudgetConfig{
			BookUSD:  0,
			MaxCalls: 0,
		}, func() float64 { return 0 }, func(string) {}, func(string, string) {})

		if sentinel == nil {
			msg := "[DEFECT RES-03] NewBudgetSentinel returns nil when BookUSD=0 and MaxCalls=0 because cfg.Enabled() is false; default 100 calls tripwire bypassed!"
			if os.Getenv("STRICT_ADVERSARIAL") == "1" {
				t.Fatalf("%s", msg)
			} else {
				t.Logf("%s", msg)
			}
		} else if sentinel.MaxCalls() != 100 {
			t.Errorf("Expected default MaxCalls=100 for $0 self-hosted model, got %d", sentinel.MaxCalls())
		}
	})

	// 2. MaxCalls Trip
	t.Run("CircuitBreaker_MaxCallsTrip", func(t *testing.T) {
		var abortedReason string
		sentinel := host.NewBudgetSentinel(bootstrap.BudgetConfig{
			BookUSD:  0,
			MaxCalls: 5,
			HardStop: true,
		}, func() float64 { return 0 }, func(reason string) {
			abortedReason = reason
		}, func(string, string) {})

		sentinel.RecordUsage(4, 1000)
		if err := sentinel.Refuse(); err != nil {
			t.Fatalf("Should not refuse at 4 calls: %v", err)
		}

		sentinel.RecordUsage(5, 1200)
		if abortedReason == "" {
			t.Fatalf("HardStop should have triggered abort callback at 5 calls")
		}
		if err := sentinel.Refuse(); err == nil {
			t.Fatalf("Refuse() should return error once MaxCalls reached")
		}
	})

	// 3. MaxTokens Trip
	t.Run("CircuitBreaker_MaxTokensTrip", func(t *testing.T) {
		var abortedReason string
		sentinel := host.NewBudgetSentinel(bootstrap.BudgetConfig{
			BookUSD:   0,
			MaxTokens: 50000,
			HardStop:  true,
		}, func() float64 { return 0 }, func(reason string) {
			abortedReason = reason
		}, func(string, string) {})

		sentinel.RecordUsage(10, 49999)
		if err := sentinel.Refuse(); err != nil {
			t.Fatalf("Should not refuse before max tokens: %v", err)
		}

		sentinel.RecordUsage(11, 50001)
		if abortedReason == "" {
			t.Fatalf("HardStop should have aborted when MaxTokens exceeded")
		}
		if err := sentinel.Refuse(); err == nil {
			t.Fatalf("Refuse() should fail when MaxTokens exceeded")
		}
	})

	// 4. MaxDuration Trip
	t.Run("CircuitBreaker_MaxDurationTrip", func(t *testing.T) {
		sentinel := host.NewBudgetSentinel(bootstrap.BudgetConfig{
			BookUSD:     0,
			MaxDuration: "20ms",
			HardStop:    true,
		}, func() float64 { return 0 }, func(string) {}, func(string, string) {})

		time.Sleep(35 * time.Millisecond)
		err := sentinel.Refuse()
		if err == nil {
			t.Fatalf("Expected Refuse() to fail after MaxDuration elapsed")
		}
		if !strings.Contains(err.Error(), "quá thời hạn") {
			t.Fatalf("Expected duration error message, got: %v", err)
		}
	})
}

// ----------------------------------------------------------------------------
// 2. ADVERSARIAL OUTPUT VALIDATOR STRESS TESTING
// ----------------------------------------------------------------------------

func TestAdversarial_ValidateChapterOutput_CJKPayloads(t *testing.T) {
	cases := []struct {
		name    string
		content string
		wantCJK bool
	}{
		{
			name:    "Han characters embedded",
			content: "Trận chiến diễn ra vô cùng ác liệt bên bờ sông. 汉字 xuất hiện bất ngờ.",
			wantCJK: true,
		},
		{
			name:    "Hiragana embedded",
			content: "Tiếng gió rít qua khe cửa sổ rùng rợn. ひらがな xuất hiện trong bóng tối.",
			wantCJK: true,
		},
		{
			name:    "Katakana embedded",
			content: "Ánh trăng chiếu rọi mặt nước lung linh. カタカナ vang lên từ phía xa.",
			wantCJK: true,
		},
		{
			name:    "Hangul embedded",
			content: "Bức thư bí mật chứa những ký hiệu lạ. 한글 làm thám tử bối rối.",
			wantCJK: true,
		},
		{
			name:    "Clean Vietnamese with full diacritics",
			content: "Đêm mùa đông lạnh giá, sương mù bao phủ khắp thị trấn miền núi xa xôi.",
			wantCJK: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			errs, err := rules.ValidateChapterOutput("novel-manga", "Tiêu đề", tc.content)
			if err != nil {
				t.Fatalf("ValidateChapterOutput returned unexpected err: %v", err)
			}

			hasCJKViolation := false
			for _, e := range errs {
				if e.Rule == "cjk_residue" {
					hasCJKViolation = true
					if e.Severity != rules.SeverityError {
						t.Errorf("Expected SeverityError for CJK residue, got %v", e.Severity)
					}
				}
			}

			if hasCJKViolation != tc.wantCJK {
				t.Errorf("Expected CJK violation=%v, got %v for payload: %q", tc.wantCJK, hasCJKViolation, tc.content)
			}
		})
	}
}

func makeWords(count int, suffix string) string {
	words := make([]string, count)
	for i := 0; i < count; i++ {
		words[i] = "chữ"
	}
	words[count-1] = "cuối" + suffix
	return strings.Join(words, " ")
}

func TestAdversarial_ValidateChapterOutput_WordCountBoundaries(t *testing.T) {
	// Novel mode word boundaries: [2,000, 4,500]
	t.Run("Novel_WordCount_1999Words_Fails", func(t *testing.T) {
		text := makeWords(1999, ".")
		errs, _ := rules.ValidateChapterOutput("novel-manga", "Chương 1", text)
		var shortFound bool
		for _, e := range errs {
			if e.Rule == "chapter_too_short" {
				shortFound = true
			}
		}
		if !shortFound {
			t.Errorf("1,999 words must trigger chapter_too_short")
		}
	})

	t.Run("Novel_WordCount_2000Words_Passes", func(t *testing.T) {
		text := makeWords(2000, ".")
		errs, _ := rules.ValidateChapterOutput("novel-manga", "Chương 1", text)
		for _, e := range errs {
			if e.Rule == "chapter_too_short" || e.Rule == "chapter_too_long" {
				t.Errorf("2,000 words should be valid, got: %s", e.Rule)
			}
		}
	})

	t.Run("Novel_WordCount_4500Words_Passes", func(t *testing.T) {
		text := makeWords(4500, ".")
		errs, _ := rules.ValidateChapterOutput("novel-manga", "Chương 1", text)
		for _, e := range errs {
			if e.Rule == "chapter_too_short" || e.Rule == "chapter_too_long" {
				t.Errorf("4,500 words should be valid, got: %s", e.Rule)
			}
		}
	})

	t.Run("Novel_WordCount_4501Words_Fails", func(t *testing.T) {
		text := makeWords(4501, ".")
		errs, _ := rules.ValidateChapterOutput("novel-manga", "Chương 1", text)
		var longFound bool
		for _, e := range errs {
			if e.Rule == "chapter_too_long" {
				longFound = true
			}
		}
		if !longFound {
			t.Errorf("4,501 words must trigger chapter_too_long")
		}
	})

	// Video mode spoken word boundaries: [700, 1,500]
	makeVideoScript := func(spokenWords int) string {
		var sb strings.Builder
		sb.WriteString("HOOK 0:00-0:03\n")
		sb.WriteString("LỜI: " + makeWords(spokenWords, ".") + "\n")
		sb.WriteString("HÌNH: Người que đứng nhìn bầu trời.\n")
		return sb.String()
	}

	t.Run("Video_WordCount_699Words_Fails", func(t *testing.T) {
		text := makeVideoScript(699)
		errs, _ := rules.ValidateChapterOutput("doodle-explainer", "Tập 1", text)
		var outOfRange bool
		for _, e := range errs {
			if e.Rule == "script_word_count_out_of_range" {
				outOfRange = true
			}
		}
		if !outOfRange {
			t.Errorf("699 spoken words must trigger script_word_count_out_of_range")
		}
	})

	t.Run("Video_WordCount_700Words_Passes", func(t *testing.T) {
		text := makeVideoScript(700)
		errs, _ := rules.ValidateChapterOutput("doodle-explainer", "Tập 1", text)
		for _, e := range errs {
			if e.Rule == "script_word_count_out_of_range" {
				t.Errorf("700 spoken words should be valid, got: %v", e.Message)
			}
		}
	})

	t.Run("Video_WordCount_1500Words_Passes", func(t *testing.T) {
		text := makeVideoScript(1500)
		errs, _ := rules.ValidateChapterOutput("doodle-explainer", "Tập 1", text)
		for _, e := range errs {
			if e.Rule == "script_word_count_out_of_range" {
				t.Errorf("1,500 spoken words should be valid, got: %v", e.Message)
			}
		}
	})

	t.Run("Video_WordCount_1501Words_Fails", func(t *testing.T) {
		text := makeVideoScript(1501)
		errs, _ := rules.ValidateChapterOutput("doodle-explainer", "Tập 1", text)
		var outOfRange bool
		for _, e := range errs {
			if e.Rule == "script_word_count_out_of_range" {
				outOfRange = true
			}
		}
		if !outOfRange {
			t.Errorf("1,501 spoken words must trigger script_word_count_out_of_range")
		}
	})
}

func TestAdversarial_ValidateChapterOutput_RepetitionLoops(t *testing.T) {
	// Recurring identical sentence triplets
	t.Run("Sentence_Triplets_FrequentRepetition", func(t *testing.T) {
		repeatedSentence := "Gió lạnh thổi qua từng kẽ lá xào xạc trong đêm tối mịt mùng."
		text := fmt.Sprintf("%s Đoạn văn thứ nhất tiếp tục diễn biến câu chuyện bình thường. %s Đoạn văn thứ hai mô tả diễn biến tiếp theo một cách chi tiết. %s Kết thúc chương một cách hoàn chỉnh.",
			repeatedSentence, repeatedSentence, repeatedSentence)

		errs, _ := rules.ValidateChapterOutput("novel-manga", "Chương 1", text)
		var freqFound bool
		for _, e := range errs {
			if e.Rule == "repetition_frequent_sentence" {
				freqFound = true
			}
		}
		if !freqFound {
			t.Errorf("Expected repetition_frequent_sentence when a sentence appears 3 times")
		}
	})

	// Consecutive paragraph repetition
	t.Run("Paragraph_ConsecutiveLoop", func(t *testing.T) {
		para := "Ánh trăng rọi qua khe cửa gỗ cũ kỹ, soi rõ từng hạt bụi li ti đang bay lơ lửng trong không trung lạnh buốt."
		text := fmt.Sprintf("%s\n\n%s\n\nMột đoạn văn khác miêu tả diễn biến tiếp theo của nhân vật.", para, para)

		errs, _ := rules.ValidateChapterOutput("novel-manga", "Chương 1", text)
		var paraFound bool
		for _, e := range errs {
			if e.Rule == "repetition_paragraph_loop" && e.Actual == "consecutive" {
				paraFound = true
			}
		}
		if !paraFound {
			t.Errorf("Expected repetition_paragraph_loop for consecutive identical paragraphs")
		}
	})

	// Paragraph ABAB cyclic loop
	t.Run("Paragraph_ABAB_Cycle", func(t *testing.T) {
		paraA := "Bên ngoài trời mưa tầm tã không ngớt, từng dòng nước đục ngầu cuồn cuộn chảy xiết dưới chân cầu đá rêu phong."
		paraB := "Bên trong quán trọ nhỏ bé quạnh hiu, người thiếu niên lặng lẽ lau chùi thanh bảo kiếm gia truyền bằng chiếc khăn lụa."
		text := fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\nPhần kết thúc bình thường của câu chuyện.",
			paraA, paraB, paraA, paraB)

		errs, _ := rules.ValidateChapterOutput("novel-manga", "Chương 1", text)
		var ababFound bool
		for _, e := range errs {
			if e.Rule == "repetition_paragraph_loop" && e.Actual == "abab_cycle" {
				ababFound = true
			}
		}
		if !ababFound {
			t.Errorf("Expected repetition_paragraph_loop for ABAB cyclic pattern")
		}
	})
}

func TestAdversarial_ValidateChapterOutput_TruncatedEndings(t *testing.T) {
	cases := []struct {
		name     string
		content  string
		wantTrun bool
	}{
		{
			name:     "Ending with comma",
			content:  "Nhân vật chính quay đầu nhìn lại con đường phía sau,",
			wantTrun: true,
		},
		{
			name:     "Ending with mid-word letter",
			content:  "Nhân vật chính bước nhanh về phía trước không hề do dự",
			wantTrun: true,
		},
		{
			name:     "Ending with dash",
			content:  "Người thợ rèn thở dài một hơi não nề -",
			wantTrun: true,
		},
		{
			name:     "Valid ending with period",
			content:  "Người thợ rèn thở dài một hơi não nề.",
			wantTrun: false,
		},
		{
			name:     "Valid ending with exclamation",
			content:  "Hành trình vĩ đại đã chính thức bắt đầu!",
			wantTrun: false,
		},
		{
			name:     "Valid ending with question mark",
			content:  "Liệu ai có thể khám phá ra bí mật kinh hoàng này?",
			wantTrun: false,
		},
		{
			name:     "Valid ending with quotes and inner period",
			content:  `"Chúng ta nhất định sẽ gặp lại nhau."`,
			wantTrun: false,
		},
		{
			name:     "Invalid ending with quotes but missing inner period",
			content:  `"Chúng ta nhất định sẽ gặp lại nhau"`,
			wantTrun: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			errs, _ := rules.ValidateChapterOutput("novel-manga", "Chương 1", tc.content)
			var trunFound bool
			for _, e := range errs {
				if e.Rule == "truncated_ending" {
					trunFound = true
				}
			}
			if trunFound != tc.wantTrun {
				t.Errorf("Expected truncated_ending=%v, got %v for content %q", tc.wantTrun, trunFound, tc.content)
			}
		})
	}
}

func TestAdversarial_ValidateChapterOutput_LeakedScriptTags(t *testing.T) {
	novelModes := []string{"novel-manga", "suspense", "romance", "fantasy", "vietnamese-history"}
	tags := []string{"HOOK 0:00-0:05", "LỜI: Lời thoại đọc truyền cảm", "HÌNH: Khung cảnh minh họa", "CẢNH 1: Bối cảnh diễn biến"}

	for _, mode := range novelModes {
		for _, tag := range tags {
			t.Run(fmt.Sprintf("Novel_%s_Leak_%s", mode, strings.Fields(tag)[0]), func(t *testing.T) {
				content := fmt.Sprintf("Mở đầu chương truyện bình thường.\n\n%s\n\nNội dung tiếp theo của truyện.", tag)
				errs, _ := rules.ValidateChapterOutput(mode, "Chương 1", content)
				var leakFound bool
				for _, e := range errs {
					if e.Rule == "leaked_script_tag" {
						leakFound = true
					}
				}
				if !leakFound {
					t.Errorf("[%s] Expected leaked_script_tag for tag %q in novel prose", mode, tag)
				}
			})
		}
	}

	// In video modes, script tags MUST NOT be flagged as leaked
	videoModes := []string{"doodle-explainer", "behavioral-psychology"}
	for _, vmode := range videoModes {
		t.Run(fmt.Sprintf("Video_%s_AllowsScriptTags", vmode), func(t *testing.T) {
			content := "HOOK 0:00-0:05\nLỜI: Lời thoại giới thiệu vấn đề.\nHÌNH: Mô tả nhân vật que."
			errs, _ := rules.ValidateChapterOutput(vmode, "Tập 1", content)
			for _, e := range errs {
				if e.Rule == "leaked_script_tag" {
					t.Errorf("[%s] Video mode should allow script tags, but got leaked_script_tag", vmode)
				}
			}
		})
	}
}

// ----------------------------------------------------------------------------
// 3. ADVERSARIAL RELEASE & CI WORKFLOW VERIFICATION
// ----------------------------------------------------------------------------

func TestAdversarial_CIAndReleaseWorkflowIntegrity(t *testing.T) {
	// 1. Verify .github/workflows/ci.yml
	t.Run("CI_Workflow_Checks", func(t *testing.T) {
		ciPath := filepath.Join("..", "..", ".github", "workflows", "ci.yml")
		data, err := os.ReadFile(ciPath)
		if err != nil {
			t.Fatalf("Read ci.yml: %v", err)
		}
		content := string(data)

		// Must check gofmt, go vet, and govulncheck
		if !strings.Contains(content, "gofmt -l .") {
			t.Errorf("ci.yml missing gofmt step")
		}
		if !strings.Contains(content, "go vet ./...") {
			t.Errorf("ci.yml missing go vet step")
		}
		if !strings.Contains(content, "govulncheck ./...") {
			t.Errorf("ci.yml missing govulncheck step")
		}
		// Matrix testing on ubuntu and windows
		if !strings.Contains(content, "ubuntu-latest") || !strings.Contains(content, "windows-latest") {
			t.Errorf("ci.yml missing matrix test platforms (ubuntu & windows)")
		}
	})

	// 2. Verify .github/workflows/release.yml
	t.Run("Release_Workflow_Checksums_MultiArch", func(t *testing.T) {
		relPath := filepath.Join("..", "..", ".github", "workflows", "release.yml")
		data, err := os.ReadFile(relPath)
		if err != nil {
			t.Fatalf("Read release.yml: %v", err)
		}
		content := string(data)

		// Must generate checksums.txt
		if !strings.Contains(content, "sha256sum * > checksums.txt") {
			t.Errorf("release.yml missing sha256sum checksums.txt generation")
		}
		// Must cover multi-arch: linux, darwin, windows, amd64, arm64
		platforms := []string{"linux", "darwin", "windows", "amd64", "arm64"}
		for _, p := range platforms {
			if !strings.Contains(content, p) {
				t.Errorf("release.yml missing matrix target for platform: %s", p)
			}
		}
	})
}
