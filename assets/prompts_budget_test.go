package assets_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/voocel/ainovel-cli/assets"
	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/store"
	"github.com/voocel/ainovel-cli/internal/tools"
)

// estimateTokens returns an approximate token count for Vietnamese/English text (roughly 0.8 tokens per rune).
func estimateTokens(s string) int {
	return int(float64(len([]rune(s))) * 0.8)
}

func TestPromptsBudget_All7ModesWithinLimit(t *testing.T) {
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

	const maxBudgetTokens = 19660 // 60% of 32,768 DefaultContextWindow
	const maxBasePromptTokens = 10000

	for _, m := range modes {
		t.Run("Budget_"+m.id, func(t *testing.T) {
			bundle := assets.Load(m.id, assets.LoadOptions{})
			writerPrompt := bundle.Prompts.Writer
			voice := bundle.Voice
			style := bundle.Styles[m.id]
			fullSystemPrompt := assets.BuildWriterPrompt(writerPrompt, voice, style)

			basePromptTokens := estimateTokens(writerPrompt)
			promptTokens := estimateTokens(fullSystemPrompt)
			t.Logf("[%s] BasePrompt tokens=%d | FullSystemPrompt runes=%d, tokens=%d",
				m.id, basePromptTokens, len([]rune(fullSystemPrompt)), promptTokens)

			// CTX-02: For trimmed video modes, base prompt size must stay <= 10,000 tokens
			if m.id == "behavioral-psychology" || m.id == "doodle-explainer" {
				if basePromptTokens > maxBasePromptTokens {
					t.Errorf("[%s] Base prompt alone (%d tokens) exceeds max base prompt limit (%d tokens)",
						m.id, basePromptTokens, maxBasePromptTokens)
				}
			}

			// Instantiate real store and context tool with real reference bundle
			tmpDir := t.TempDir()
			st := store.NewStore(tmpDir)
			if err := st.Init(); err != nil {
				t.Fatalf("[%s] Store Init: %v", m.id, err)
			}

			// Populate realistic foundation and progress
			if err := st.Progress.Save(&domain.Progress{
				TotalChapters:     10,
				CompletedChapters: []int{},
				Phase:             domain.PhaseWriting,
			}); err != nil {
				t.Fatalf("[%s] Progress Save: %v", m.id, err)
			}

			if err := st.Outline.SaveOutline([]domain.OutlineEntry{
				{
					Chapter:   1,
					Title:     "Chương 1 Khởi Đầu",
					Hook:      "Mở màn kịch tính lôi cuốn người xem",
					CoreEvent: "Biến cố then chốt thay đổi số phận",
					Scenes:    []string{"Cảnh 1: Giới thiệu bối cảnh", "Cảnh 2: Xung đột nảy sinh"},
				},
			}); err != nil {
				t.Fatalf("[%s] Outline Save: %v", m.id, err)
			}

			if err := st.Characters.Save([]domain.Character{
				{
					Name:        "Nhân vật chính",
					Tier:        "core",
					Role:        "Protagonist",
					Description: "Kiên định, thông minh, tìm kiếm chân lý và vượt qua nghịch cảnh",
					Traits:      []string{"kiên định", "thông minh"},
				},
			}); err != nil {
				t.Fatalf("[%s] Characters Save: %v", m.id, err)
			}

			styleStats := tools.NewStyleStatsIndex(st)
			contextTool := tools.NewContextTool(st, bundle.References, m.id, styleStats)
			contextTool.SetPromptTokens(promptTokens)

			// Execute novel_context for chapter 1
			args, _ := json.Marshal(map[string]int{"chapter": 1})
			rawCtx, err := contextTool.Execute(context.Background(), args)
			if err != nil {
				t.Fatalf("[%s] ContextTool.Execute: %v", m.id, err)
			}

			novelContextStr := string(rawCtx)
			contextTokens := estimateTokens(novelContextStr)
			totalTokens := promptTokens + contextTokens

			t.Logf("[%s] NovelContext runes=%d, tokens=%d | Total = %d tokens (Budget Limit = %d)",
				m.id, len([]rune(novelContextStr)), contextTokens, totalTokens, maxBudgetTokens)

			if totalTokens > maxBudgetTokens {
				t.Errorf("[%s] Combined SystemPrompt + novel_context (%d tokens) exceeds 60%% 32k budget (%d tokens)",
					m.id, totalTokens, maxBudgetTokens)
			}
		})
	}
}

func TestPromptsBudget_DynamicClampingScale(t *testing.T) {
	// Verify that setting high prompt tokens triggers reference scaling proportionally
	tmpDir := t.TempDir()
	st := store.NewStore(tmpDir)
	if err := st.Init(); err != nil {
		t.Fatalf("Store Init: %v", err)
	}

	bundle := assets.Load("behavioral-psychology", assets.LoadOptions{})
	styleStats := tools.NewStyleStatsIndex(st)
	contextTool := tools.NewContextTool(st, bundle.References, "behavioral-psychology", styleStats)

	// Set large prompt tokens (12,000 tokens)
	contextTool.SetPromptTokens(12000)

	args, _ := json.Marshal(map[string]int{"chapter": 1})
	rawCtx, err := contextTool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("ContextTool.Execute: %v", err)
	}

	ctxTokens := estimateTokens(string(rawCtx))
	total := 12000 + ctxTokens
	t.Logf("Clamping test with 12k prompt: context=%d tokens, total=%d tokens (max=%d)",
		ctxTokens, total, tools.MaxContextBudgetTokens)

	if total > tools.MaxContextBudgetTokens {
		t.Errorf("Clamping failed: total %d > limit %d", total, tools.MaxContextBudgetTokens)
	}
}

func TestPromptsBudget_All7Modes_MaximalContext(t *testing.T) {
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

	const maxBudgetTokens = tools.MaxContextBudgetTokens // 19,660

	for _, m := range modes {
		t.Run("MaximalBudget_"+m.id, func(t *testing.T) {
			bundle := assets.Load(m.id, assets.LoadOptions{})
			fullSystemPrompt := assets.BuildWriterPrompt(bundle.Prompts.Writer, bundle.Voice, bundle.Styles[m.id])
			promptTokens := estimateTokens(fullSystemPrompt)

			tmpDir := t.TempDir()
			st := store.NewStore(tmpDir)
			if err := st.Init(); err != nil {
				t.Fatalf("[%s] Store Init: %v", m.id, err)
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
				t.Fatalf("[%s] Save characters: %v", m.id, err)
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
				t.Fatalf("[%s] Save outline: %v", m.id, err)
			}

			styleStats := tools.NewStyleStatsIndex(st)
			contextTool := tools.NewContextTool(st, bundle.References, m.id, styleStats)
			contextTool.SetPromptTokens(promptTokens)

			// Execute for mid-book chapter 25
			args, _ := json.Marshal(map[string]int{"chapter": 25})
			rawCtx, err := contextTool.Execute(context.Background(), args)
			if err != nil {
				t.Fatalf("[%s] ContextTool.Execute: %v", m.id, err)
			}

			novelContextStr := string(rawCtx)
			contextTokens := estimateTokens(novelContextStr)
			totalTokens := promptTokens + contextTokens

			t.Logf("[%s - Maximal] prompt=%d, context=%d, total=%d (limit=%d)",
				m.id, promptTokens, contextTokens, totalTokens, maxBudgetTokens)

			if totalTokens > maxBudgetTokens {
				t.Errorf("[%s - Maximal] Combined SystemPrompt + novel_context (%d tokens) exceeds 60%% 32k budget (%d tokens)",
					m.id, totalTokens, maxBudgetTokens)
			}
		})
	}
}
