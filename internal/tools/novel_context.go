package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/voocel/agentcore/schema"
	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/store"
	"github.com/voocel/ainovel-cli/internal/utils"
)

// References 嵌入的参考资料。
type References struct {
	// V0
	ChapterGuide      string
	HookTechniques    string
	QualityChecklist  string
	OutlineTemplate   string
	CharacterTemplate string
	ChapterTemplate   string
	// V1
	Consistency      string
	ContentExpansion string
	DialogueWriting  string
	// V2
	StyleReference   string // 风格补充参考（可为空）
	LongformPlanning string // 通用长篇规划参考
	Differentiation  string // 通用差异化设计参考
	ArcTemplates     string // 题材弧型模板（按 style 加载，可为空）
	AntiAITone       string // 去 AI 味判据库（writer/editor 共用，全程注入）
	// Doodle explainer
	DoodleVisual      string // Ngôn ngữ hình ảnh doodle que (writer luôn nhận)
	FactGrounding     string // Neo dữ kiện cho chủ đề trend (writer + architect luôn nhận)
	TiktokSafety      string // Tiêu chuẩn an toàn nội dung TikTok & pháp lý VN (writer + architect luôn nhận)
	HumorRelatability string // Chuyên gia gây cười & soi chiếu đồng cảm nhân sinh (writer + architect luôn nhận)
	// Character & conflict references
	CharacterBuilding string // Xây dựng nhân vật đa chiều, cửa vào độc giả
	AdversityConflict string // Xung đột kép & khoảnh khắc đánh đổi sinh tử
}

// ContextTool 组装当前章节所需上下文。
type ContextTool struct {
	store        *store.Store
	refs         References
	style        string
	styleStats   *StyleStatsIndex
	promptTokens int
}

type contextReads struct {
	warnings []string
	seen     map[string]struct{}
	err      error
}

func (r *contextReads) warn(scope string, err error) {
	if err == nil || os.IsNotExist(err) {
		return
	}
	msg := fmt.Sprintf("%s đọc thất bại: %v", scope, err)
	if r.seen == nil {
		r.seen = make(map[string]struct{})
	}
	if _, ok := r.seen[msg]; ok {
		return
	}
	r.seen[msg] = struct{}{}
	r.warnings = append(r.warnings, msg)
}

func (r *contextReads) require(scope string, err error) {
	if r.err != nil || err == nil || os.IsNotExist(err) || errors.Is(err, store.ErrOutlineChapterNotFound) {
		return
	}
	r.err = fmt.Errorf("%s đọc thất bại: %w", scope, err)
}

func (r *contextReads) fail(err error) {
	if r.err == nil {
		r.err = err
	}
}

// NewContextTool 创建上下文工具。styleStats 必须与 commit_chapter 共享，
// 否则重写章节后上下文会继续读取旧统计。
// user_rules 由 buildUserRules 直接读本书快照（meta/user_rules.json）注入，不再依赖加载选项。
func NewContextTool(
	store *store.Store,
	refs References,
	style string,
	styleStats *StyleStatsIndex,
) *ContextTool {
	if styleStats == nil {
		panic("tools: NewContextTool requires StyleStatsIndex")
	}
	return &ContextTool{store: store, refs: refs, style: style, styleStats: styleStats}
}

func (t *ContextTool) Name() string { return "novel_context" }
func (t *ContextTool) Description() string {
	return "Lấy trạng thái hiện tại của tiểu thuyết và ngữ cảnh sáng tác. " +
		"Không truyền chapter: trả về progress_status (các trường tiến độ như phase/flow/next_chapter/pending_rewrites) + tổng quan kế hoạch rút gọn, dùng để xác định bước tiếp theo cần làm; " +
		"Architect truyện dài có thể truyền volume + arc để đọc tập trung một hồi chỉ định: hồi đã khai triển có chi tiết chương, hồi khung xương có title/goal/estimated_chapters. " +
		"Truyền chapter=N: trả thêm ngữ cảnh viết của chương đó như tóm tắt diễn biến trước đó, phục bút, trạng thái nhân vật, quy tắc văn phong"
}
func (t *ContextTool) Label() string { return "Tải ngữ cảnh" }

// 纯读工具，可被并发调度。
func (t *ContextTool) ReadOnly(_ json.RawMessage) bool        { return true }
func (t *ContextTool) ConcurrencySafe(_ json.RawMessage) bool { return true }

func (t *ContextTool) Schema() map[string]any {
	return schema.Object(
		schema.Property("chapter", schema.Int("Số chương. Không truyền thì trả về trạng thái tiến độ và thiết lập nền tảng (dành cho Architect); truyền vào thì trả thêm ngữ cảnh viết của chương đó (dành cho Writer/Editor)")),
		schema.Property("volume", schema.Int("Tùy chọn cho Architect truyện dài: số thứ tự quyển cần đọc tập trung; hồi đã khai triển trả về chi tiết chương, hồi khung xương trả về mục tiêu kế hoạch; phải truyền cùng arc, không được dùng cùng chapter")),
		schema.Property("arc", schema.Int("Tùy chọn cho Architect truyện dài: số thứ tự hồi trong quyển cần đọc tập trung; hồi đã khai triển trả về chi tiết chương, hồi khung xương trả về mục tiêu kế hoạch; phải truyền cùng volume, không được dùng cùng chapter")),
	)
}

func (t *ContextTool) Execute(_ context.Context, args json.RawMessage) (json.RawMessage, error) {
	var a struct {
		Chapter int `json:"chapter"`
		Volume  int `json:"volume"`
		Arc     int `json:"arc"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}
	if a.Chapter < 0 || a.Volume < 0 || a.Arc < 0 {
		return nil, fmt.Errorf("chapter, volume and arc must be >= 0")
	}
	if a.Chapter > 0 && (a.Volume > 0 || a.Arc > 0) {
		return nil, fmt.Errorf("chapter cannot be combined with volume or arc")
	}
	if (a.Volume > 0) != (a.Arc > 0) {
		return nil, fmt.Errorf("volume and arc must be provided together")
	}

	result := make(map[string]any)
	reads := &contextReads{}

	if a.Chapter > 0 {
		// Writer 路径：加载全量基础数据 + 章节上下文
		t.buildBaseContext(result, reads)
		seed := newChapterContextEnvelope()
		state := t.prepareChapterContext(a.Chapter, &seed, reads)
		seed.apply(result)
		t.buildChapterContext(result, state, reads)
		// episodic 是已写入正文的备忘，不是待写素材。
		if epi, ok := result["episodic_memory"].(map[string]any); ok && len(epi) > 0 {
			epi["_usage"] = "Vùng chứa này là ghi chú sự kiện đã được viết vào chính văn (để đối chiếu tính nhất quán và mạch nối); lặp lại nguyên văn những nội dung này trong chính văn chương mới là lỗi trùng lặp"
		}
	} else {
		// Architect 路径：只返回状态 + 结构化数据，不加载全量原文
		t.buildProgressStatus(result, reads)
		t.buildArchitectContext(result, reads, a.Volume, a.Arc)
	}

	// 注入 working_memory.user_rules（canonical 路径）。架构师路径原本没有 working_memory，
	// 由 buildUserRules 按需新建只装 user_rules 的容器。快照缺失时退到内置默认，
	// 始终输出稳定结构，避免 LLM 看到 user_rules=null 走异常分支。
	if a.Chapter > 0 {
		t.buildSimulationProfile(result, "working_memory", reads)
	} else {
		t.buildSimulationProfile(result, "planning_memory", reads)
	}

	userRules := t.buildUserRules(result, reads)
	if a.Chapter > 0 {
		t.buildRuleViolations(result, a.Chapter, userRules, reads)
	}

	if reads.err != nil {
		return nil, reads.err
	}
	if len(reads.warnings) > 0 {
		result["_warnings"] = reads.warnings
	}

	promptTokens := t.PromptTokens()
	availableTokens := MaxContextBudgetTokens - promptTokens
	if availableTokens < 500 {
		availableTokens = 500
	}

	if a.Chapter > 0 {
		// Target story budget: at most 55% of availableTokens when tight
		maxStoryBudget := int(float64(availableTokens) * 0.55)
		if maxStoryBudget < 1500 {
			maxStoryBudget = 1500
		}
		if availableTokens >= 8000 && maxStoryBudget < 4000 {
			maxStoryBudget = 4000
		}

		// Remove uncalibrated references to measure actual story tokens
		if refPack, ok := result["reference_pack"].(map[string]any); ok {
			delete(refPack, "references")
		}

		// Measure story context tokens before reference pack
		previewData, _ := json.Marshal(result)
		storyTokens := EstimateTokens(string(previewData))

		if storyTokens > maxStoryBudget {
			t.clampStoryContext(result, maxStoryBudget, a.Chapter)
			previewData, _ = json.Marshal(result)
			storyTokens = EstimateTokens(string(previewData))
		}

		// If still over budget, clamp more aggressively
		if storyTokens > maxStoryBudget {
			t.clampStoryContext(result, 1200, a.Chapter)
			previewData, _ = json.Marshal(result)
			storyTokens = EstimateTokens(string(previewData))
		}

		refBudget := availableTokens - storyTokens - 250
		if refBudget < 100 {
			refBudget = 100
		}

		// Inject references with dynamic budget
		t.injectWriterReferences(result, a.Chapter, refBudget)
	}

	// 工具层只做与任务相关的语义选择；上下文体积由各 Worker 按实际模型窗口管理。
	result["_loading_summary"] = buildLoadingSummary(result, a.Chapter)

	// Final Safety Check: Guarantee total tokens <= MaxContextBudgetTokens across all 7 modes
	finalData, _ := json.Marshal(result)
	totalTokens := promptTokens + EstimateTokens(string(finalData))
	if totalTokens > MaxContextBudgetTokens {
		excess := totalTokens - MaxContextBudgetTokens + 25
		excessRunes := int(float64(excess)/0.8) + 1
		if refPack, ok := result["reference_pack"].(map[string]any); ok {
			if refs, ok := refPack["references"].(map[string]string); ok {
				keys := make([]string, 0, len(refs))
				for k := range refs {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					if excessRunes <= 0 {
						break
					}
					v := refs[k]
					vRunes := []rune(v)
					if len(vRunes) > excessRunes+20 {
						refs[k] = string(vRunes[:len(vRunes)-excessRunes]) + "..."
						excessRunes = 0
						break
					} else if len(vRunes) > 20 {
						cut := len(vRunes) - 20
						refs[k] = string(vRunes[:20]) + "..."
						excessRunes -= cut
					}
				}
			}
		}
		result["_loading_summary"] = buildLoadingSummary(result, a.Chapter)
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal context payload: %w", err)
	}
	return data, nil
}

// buildLoadingSummary 从已组装的 result 中统计各项数据量，生成一行可读摘要。
func buildLoadingSummary(result map[string]any, chapter int) string {
	var parts []string
	working, _ := result["working_memory"].(map[string]any)
	episodic, _ := result["episodic_memory"].(map[string]any)
	planning, _ := result["planning_memory"].(map[string]any)
	foundation, _ := result["foundation_memory"].(map[string]any)
	referencePack, _ := result["reference_pack"].(map[string]any)

	if chapter > 0 {
		parts = append(parts, fmt.Sprintf("ch=%d", chapter))
		if tier, ok := episodic["planning_tier"].(domain.PlanningTier); ok && tier != "" {
			parts = append(parts, fmt.Sprintf("tier=%s", tier))
		}
	} else {
		parts = append(parts, "architect")
		if tier, ok := planning["planning_tier"].(domain.PlanningTier); ok && tier != "" {
			parts = append(parts, fmt.Sprintf("tier=%s", tier))
		}
	}

	if pos, ok := episodic["position"].(map[string]any); ok {
		parts = append(parts, fmt.Sprintf("V%dA%d", pos["volume"], pos["arc"]))
	}

	var items []string

	if n := firstSliceLen(episodic["character_snapshots"], foundation["character_snapshots"]); n > 0 {
		items = append(items, fmt.Sprintf("nhân vật:%d(ảnh chụp)", n))
	} else if n := firstSliceLen(episodic["characters"], foundation["characters"]); n > 0 {
		items = append(items, fmt.Sprintf("nhân vật:%d", n))
	}

	if len(working) > 0 {
		items = append(items, fmt.Sprintf("bộ nhớ làm việc:%d", len(working)))
	}
	if len(episodic) > 0 {
		items = append(items, fmt.Sprintf("bộ nhớ tình tiết:%d", len(episodic)))
	}
	if len(planning) > 0 {
		items = append(items, fmt.Sprintf("bộ nhớ kế hoạch:%d", len(planning)))
	}
	if len(foundation) > 0 {
		items = append(items, fmt.Sprintf("bộ nhớ nền tảng:%d", len(foundation)))
	}

	if n := firstSliceLen(working["volume_summaries"], planning["volume_summaries"]); n > 0 {
		items = append(items, fmt.Sprintf("tóm tắt quyển:%d", n))
	}
	if n := firstSliceLen(working["arc_summaries"], planning["arc_summaries"]); n > 0 {
		items = append(items, fmt.Sprintf("tóm tắt hồi:%d", n))
	}
	if n := sliceLen(working["recent_summaries"]); n > 0 {
		items = append(items, fmt.Sprintf("tóm tắt chương:%d", n))
	}

	if n := sliceLen(planning["layered_outline"]); n > 0 {
		items = append(items, fmt.Sprintf("dàn ý phân tầng:%d quyển", n))
	}

	if n := sliceLen(working["timeline"]); n > 0 {
		items = append(items, fmt.Sprintf("dòng thời gian:%d", n))
	}
	if n := firstSliceLen(episodic["foreshadow_ledger"], foundation["foreshadow_ledger"]); n > 0 {
		items = append(items, fmt.Sprintf("phục bút:%d", n))
	}
	if n := sliceLen(episodic["relationship_state"]); n > 0 {
		items = append(items, fmt.Sprintf("quan hệ:%d", n))
	}
	if n := sliceLen(episodic["recent_state_changes"]); n > 0 {
		items = append(items, fmt.Sprintf("biến đổi trạng thái:%d", n))
	}
	if _, ok := working["previous_tail"]; ok {
		items = append(items, "đoạn cuối chương trước:ok")
	}
	if _, ok := working["source_pack"]; ok {
		items = append(items, "tài liệu nguồn (source_pack):ok")
	}
	if _, ok := referencePack["style_rules"]; ok {
		items = append(items, "quy tắc văn phong:ok")
	}
	if n := sliceLen(episodic["related_chapters"]); n > 0 {
		items = append(items, fmt.Sprintf("chương liên quan:%d", n))
	}
	if selected, ok := result["selected_memory"].(map[string]any); ok && len(selected) > 0 {
		if n := sliceLen(selected["story_threads"]); n > 0 {
			items = append(items, fmt.Sprintf("gợi lại mạch truyện:%d", n))
		}
		if n := sliceLen(selected["review_lessons"]); n > 0 {
			items = append(items, fmt.Sprintf("gợi lại thẩm định:%d", n))
		}
	}

	if refs, ok := referencePack["references"].(map[string]string); ok && len(refs) > 0 {
		items = append(items, fmt.Sprintf("tham khảo:%d mục", len(refs)))
	}
	if len(referencePack) > 0 {
		items = append(items, fmt.Sprintf("gói tham khảo:%d", len(referencePack)))
	}
	if _, ok := result["memory_policy"]; ok {
		items = append(items, "chính sách bộ nhớ:ok")
	}
	if _, ok := working["simulation_profile"]; ok {
		items = append(items, "hồ sơ mô phỏng văn phong:ok")
	} else if _, ok := planning["simulation_profile"]; ok {
		items = append(items, "hồ sơ mô phỏng văn phong:ok")
	}
	if _, ok := planning["trend_brief"]; ok {
		items = append(items, "tóm tắt xu hướng (trend_brief):ok")
	}
	if _, ok := planning["source_pack"]; ok {
		items = append(items, "tài liệu nguồn (source_pack):ok")
	}
	if warnings, ok := result["_warnings"].([]string); ok && len(warnings) > 0 {
		items = append(items, fmt.Sprintf("cảnh báo:%d", len(warnings)))
	}
	if len(items) > 0 {
		parts = append(parts, strings.Join(items, " "))
	}
	return strings.Join(parts, " | ")
}

// sliceLen 对 any 类型尝试取 slice 长度。
func sliceLen(v any) int {
	switch s := v.(type) {
	case []domain.ChapterSummary:
		return len(s)
	case []domain.ArcSummary:
		return len(s)
	case []domain.VolumeSummary:
		return len(s)
	case []domain.CharacterSnapshot:
		return len(s)
	case []domain.TimelineEvent:
		return len(s)
	case []domain.ForeshadowEntry:
		return len(s)
	case []domain.RelationshipEntry:
		return len(s)
	case []domain.StateChange:
		return len(s)
	case []domain.VolumeOutline:
		return len(s)
	case []domain.Character:
		return len(s)
	case []domain.RelatedChapter:
		return len(s)
	case []domain.RecallItem:
		return len(s)
	case []planningVolumeOutline:
		return len(s)
	default:
		return 0
	}
}

func firstSliceLen(values ...any) int {
	for _, value := range values {
		if n := sliceLen(value); n > 0 {
			return n
		}
	}
	return 0
}

// loadFilteredCharacters 按 Tier 和场景出场过滤角色。
// core/important 始终返回；secondary/decorative 只在当前章节大纲提及时返回。
func (t *ContextTool) loadFilteredCharacters(result map[string]any, chapter int, reads *contextReads) {
	chars, err := t.store.Characters.Load()
	if err != nil {
		reads.require("characters", err)
		return
	}
	if len(chars) == 0 {
		return
	}

	// 获取当前章节大纲的场景描述，用于匹配次要角色
	entry, err := t.store.Outline.GetChapterOutline(chapter)
	if err != nil {
		reads.require("current_chapter_outline", err)
		result["characters"] = chars
		return
	}
	if entry == nil {
		result["characters"] = chars
		return
	}
	sceneText := strings.Join(entry.Scenes, " ") + " " + entry.CoreEvent + " " + entry.Title

	var filtered []domain.Character
	for _, c := range chars {
		switch c.Tier {
		case "secondary", "decorative":
			if matchCharacter(sceneText, c) {
				filtered = append(filtered, c)
			}
		default: // core, important, 或未设置
			filtered = append(filtered, c)
		}
	}
	result["characters"] = filtered
}

// matchCharacter 检查场景文本中是否包含角色的正式名或任一别名。
func matchCharacter(text string, c domain.Character) bool {
	if strings.Contains(text, c.Name) {
		return true
	}
	for _, alias := range c.Aliases {
		if strings.Contains(text, alias) {
			return true
		}
	}
	return false
}

// loadLayeredSummaries 分层摘要加载：卷摘要 + 当前卷弧摘要 + 弧内章摘要。
func (t *ContextTool) loadLayeredSummaries(result map[string]any, chapter, summaryWindow int, reads *contextReads) {
	vol, arc, err := t.store.Outline.LocateChapter(chapter)
	if err != nil {
		reads.require("layered_outline_position", err)
		return
	}

	// 1. 已完成卷的卷摘要
	if volSummaries, err := t.store.Summaries.LoadAllVolumeSummaries(); err == nil && len(volSummaries) > 0 {
		result["volume_summaries"] = volSummaries
	} else {
		reads.require("volume_summaries", err)
	}

	// 2. 当前卷内已完成弧的弧摘要（不含当前弧）
	if arcSummaries, err := t.store.Summaries.LoadArcSummaries(vol); err == nil && len(arcSummaries) > 0 {
		var prior []domain.ArcSummary
		for _, s := range arcSummaries {
			if s.Arc < arc {
				prior = append(prior, s)
			}
		}
		if len(prior) > 0 {
			result["arc_summaries"] = prior
		}
	} else {
		reads.require("arc_summaries", err)
	}

	// 3. 当前弧内最近 N 章的章摘要
	if summaries, err := t.store.Summaries.LoadRecentSummaries(chapter, summaryWindow); err == nil && len(summaries) > 0 {
		result["recent_summaries"] = summaries
	} else {
		reads.require("recent_summaries", err)
	}
}

// loadLayeredCharacters Layered 模式下的角色加载：优先用最近快照，回退到原始设定 + Tier 过滤。
func (t *ContextTool) loadLayeredCharacters(result map[string]any, chapter int, reads *contextReads) {
	snapshots, err := t.store.Characters.LoadLatestSnapshots()
	if err == nil && len(snapshots) > 0 {
		result["character_snapshots"] = snapshots
		// 同时保留原始设定中的 core/important 角色（快照可能不含新登场角色）
		t.loadFilteredCharacters(result, chapter, reads)
		return
	}
	reads.require("character_snapshots", err)
	// 无快照时回退到原始设定
	t.loadFilteredCharacters(result, chapter, reads)
}

func isNovelOrHistoryStyle(style string) bool {
	s := strings.ToLower(strings.TrimSpace(style))
	switch s {
	case "novel-manga", "novel", "manga",
		"vietnamese-history", "history", "lich-su",
		"suspense", "trinh-tham", "mystery", "thriller", "horror",
		"romance", "ngon-tinh", "healing",
		"fantasy", "ky-ao", "tien-hiep", "xianxia":
		return true
	default:
		return false
	}
}

func isSpecializedNovelGenre(style string) bool {
	s := strings.ToLower(strings.TrimSpace(style))
	switch s {
	case "vietnamese-history", "history", "lich-su",
		"suspense", "trinh-tham", "mystery", "thriller", "horror",
		"romance", "ngon-tinh", "healing",
		"fantasy", "ky-ao", "tien-hiep", "xianxia":
		return true
	default:
		return false
	}
}

// compactReference cắt bỏ các khối ví dụ kịch bản mẫu quá dài (vốn đã có trong system prompt)
// và giới hạn độ dài mỗi tài liệu tham khảo để novel_context không vượt ngưỡng 32K token.
func compactReference(raw string, maxRunes int) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || maxRunes <= 0 {
		return raw
	}
	for _, marker := range []string{
		"\n## Ví dụ hoàn chỉnh",
		"\n## Ví dụ kịch bản",
		"\n## Bản mẫu hoàn chỉnh",
		"\n## Ví dụ thực chiến hoàn chỉnh",
	} {
		if idx := strings.Index(raw, marker); idx > 0 {
			raw = strings.TrimSpace(raw[:idx])
		}
	}
	runes := []rune(raw)
	if len(runes) <= maxRunes {
		return raw
	}
	const suffix = "\n...(đã rút gọn tham khảo)"
	suffixRunes := len([]rune(suffix))
	targetRunes := maxRunes - suffixRunes
	if targetRunes < 20 {
		targetRunes = maxRunes
		if len(runes) <= targetRunes {
			return raw
		}
		return string(runes[:targetRunes])
	}
	clipped := string(runes[:targetRunes])
	if lastNL := strings.LastIndexByte(clipped, '\n'); lastNL > len(clipped)/2 {
		clipped = clipped[:lastNL]
	}
	return strings.TrimSpace(clipped) + suffix
}

// MaxContextBudgetTokens là giới hạn tổng token tối đa cho SystemPrompt + novel_context (60% của 32,768 = 19,660).
const MaxContextBudgetTokens = 19660

// EstimateTokens ước tính số token cho văn bản tiếng Việt/tiếng Anh (khoảng 0.8 token/rune).
func EstimateTokens(s string) int {
	return int(float64(len([]rune(s))) * 0.8)
}

func (t *ContextTool) SetPromptTokens(tokens int) {
	t.promptTokens = tokens
}

func (t *ContextTool) PromptTokens() int {
	if t.promptTokens > 0 {
		return t.promptTokens
	}
	s := strings.ToLower(strings.TrimSpace(t.style))
	switch {
	case strings.Contains(s, "psychology"):
		return 7500
	case strings.Contains(s, "doodle"):
		return 7000
	case strings.Contains(s, "history"):
		return 6000
	default:
		return 5800
	}
}

type refEntry struct {
	key      string
	content  string
	maxRunes int
}

// clampStoryContext dynamically trims characters and state when remaining context budget is constrained.
func (t *ContextTool) clampStoryContext(result map[string]any, maxStoryTokens int, chapter int) {
	if maxStoryTokens <= 0 {
		return
	}

	// 1. Clamp characters in episodic_memory
	if episodic, ok := result["episodic_memory"].(map[string]any); ok {
		sceneText := ""
		if working, ok := result["working_memory"].(map[string]any); ok {
			switch entry := working["current_chapter_outline"].(type) {
			case *domain.OutlineEntry:
				if entry != nil {
					sceneText = strings.Join(entry.Scenes, " ") + " " + entry.CoreEvent + " " + entry.Title
				}
			case domain.OutlineEntry:
				sceneText = strings.Join(entry.Scenes, " ") + " " + entry.CoreEvent + " " + entry.Title
			}
		}

		if rawChars, exists := episodic["characters"]; exists {
			var chars []domain.Character
			switch v := rawChars.(type) {
			case []domain.Character:
				chars = v
			}

			if len(chars) > 0 {
				var active []domain.Character
				var inactive []domain.Character
				for _, c := range chars {
					if sceneText != "" && matchCharacter(sceneText, c) {
						active = append(active, c)
					} else {
						inactive = append(inactive, c)
					}
				}

				maxAllowedChars := 6
				if maxStoryTokens < 2000 {
					maxAllowedChars = 3
				} else if maxStoryTokens < 3000 {
					maxAllowedChars = 4
				}

				var retained []domain.Character
				retained = append(retained, active...)
				for _, c := range inactive {
					if len(retained) >= maxAllowedChars {
						break
					}
					retained = append(retained, c)
				}
				if len(retained) > maxAllowedChars {
					retained = retained[:maxAllowedChars]
				}

				maxDescRunes := 120
				if maxStoryTokens < 2000 {
					maxDescRunes = 60
				} else if maxStoryTokens < 3000 {
					maxDescRunes = 80
				}

				for i := range retained {
					c := retained[i]
					runes := []rune(c.Description)
					if len(runes) > maxDescRunes {
						c.Description = string(runes[:maxDescRunes]) + "..."
					}
					if len(c.Traits) > 3 {
						c.Traits = c.Traits[:3]
					}
					retained[i] = c
				}
				episodic["characters"] = retained
			}
		}

		if rawSnaps, exists := episodic["character_snapshots"]; exists {
			if snaps, ok := rawSnaps.([]domain.CharacterSnapshot); ok && len(snaps) > 4 {
				maxSnaps := 4
				if maxStoryTokens < 2500 {
					maxSnaps = 2
				}
				if len(snaps) > maxSnaps {
					episodic["character_snapshots"] = snaps[:maxSnaps]
				}
			}
		}

		if rawCast, exists := episodic["recent_cast"]; exists {
			if cast, ok := rawCast.([]map[string]any); ok && len(cast) > 3 {
				maxCast := 3
				if maxStoryTokens < 2500 {
					maxCast = 1
				}
				episodic["recent_cast"] = cast[:maxCast]
			}
		}

		if rawRel, exists := episodic["related_chapters"]; exists {
			if rels, ok := rawRel.([]domain.RelatedChapter); ok && len(rels) > 2 {
				episodic["related_chapters"] = rels[:2]
			}
		}
	}

	// 2. Clamp outline window and summaries in working_memory
	if working, ok := result["working_memory"].(map[string]any); ok {
		if rawOutline, exists := working["outline_window"]; exists {
			if entries, ok := rawOutline.([]domain.OutlineEntry); ok && len(entries) > 3 {
				maxOutline := 3
				if maxStoryTokens < 2500 {
					maxOutline = 2
				}
				start := max(0, len(entries)-maxOutline)
				working["outline_window"] = entries[start:]
			}
		}
		if rawSummaries, exists := working["recent_summaries"]; exists {
			if summaries, ok := rawSummaries.([]domain.ChapterSummary); ok && len(summaries) > 2 {
				maxSummaries := 2
				if maxStoryTokens < 2500 {
					maxSummaries = 1
				}
				start := max(0, len(summaries)-maxSummaries)
				working["recent_summaries"] = summaries[start:]
			}
		}
		if rawTimeline, exists := working["timeline"]; exists {
			if events, ok := rawTimeline.([]domain.TimelineEvent); ok && len(events) > 3 {
				working["timeline"] = events[len(events)-2:]
			}
		}
		if prevTail, exists := working["previous_tail"].(string); exists {
			runes := []rune(prevTail)
			maxTail := 400
			if maxStoryTokens < 2500 {
				maxTail = 200
			}
			if len(runes) > maxTail {
				working["previous_tail"] = string(runes[len(runes)-maxTail:])
			}
		}
	}

	// 3. Clamp foundation_memory (for architect path or if present)
	if foundation, ok := result["foundation_memory"].(map[string]any); ok {
		if rawChars, exists := foundation["characters"]; exists {
			if chars, ok := rawChars.([]domain.Character); ok && len(chars) > 6 {
				maxChars := 6
				if maxStoryTokens < 3000 {
					maxChars = 4
				}
				foundation["characters"] = chars[:maxChars]
			}
		}
	}
}

// injectWriterReferences chèn hoặc cập nhật references vào reference_pack theo ngân sách token thực tế.
func (t *ContextTool) injectWriterReferences(result map[string]any, chapter int, refBudgetTokens int) {
	refPack, ok := result["reference_pack"].(map[string]any)
	if !ok {
		refPack = make(map[string]any)
		result["reference_pack"] = refPack
	}
	refPack["references"] = t.writerReferencesWithBudget(chapter, refBudgetTokens)
}

// writerReferences 返回写作参考资料。章节 1 返回全量，后续章节裁剪掉不再需要的模板。
// Áp dụng dynamic reference budget clamping để bảo đảm SystemPrompt + novel_context <= 19,660 tokens.
func (t *ContextTool) writerReferences(chapter int) map[string]string {
	promptTokens := t.PromptTokens()
	remainingTokens := MaxContextBudgetTokens - promptTokens
	refBudgetTokens := remainingTokens - 2600
	if refBudgetTokens < 200 {
		refBudgetTokens = 200
	}
	return t.writerReferencesWithBudget(chapter, refBudgetTokens)
}

// writerReferencesWithBudget scales reference materials to strictly fit within refBudgetTokens.
func (t *ContextTool) writerReferencesWithBudget(chapter int, refBudgetTokens int) map[string]string {
	var entries []refEntry
	add := func(k, v string, maxRunes int) {
		if v != "" {
			entries = append(entries, refEntry{key: k, content: v, maxRunes: maxRunes})
		}
	}
	// 渐进式加载：始终保留核心参考，前 3 章额外加载完整写作指南
	add("consistency", t.refs.Consistency, 1200)
	add("hook_techniques", t.refs.HookTechniques, 1200)
	add("quality_checklist", t.refs.QualityChecklist, 1200)
	add("anti_ai_tone", t.refs.AntiAITone, 1200) // 去 AI 味判据全程注入，不随章节裁剪
	add("character_building", t.refs.CharacterBuilding, 1200)
	add("adversity_conflict", t.refs.AdversityConflict, 1200)
	if !isNovelOrHistoryStyle(t.style) {
		add("doodle_visual_language", t.refs.DoodleVisual, 1000)
		add("fact_grounding", t.refs.FactGrounding, 1000)
		add("tiktok_content_safety", t.refs.TiktokSafety, 1000)
		add("humor_relatability", t.refs.HumorRelatability, 1000)
	}
	if chapter <= 3 {
		add("chapter_guide", t.refs.ChapterGuide, 1000)
		add("dialogue_writing", t.refs.DialogueWriting, 1000)
	}
	if chapter <= 3 || isSpecializedNovelGenre(t.style) {
		add("style_reference", t.refs.StyleReference, 1000)
	}

	// 仅首章加载的补充参考
	if chapter <= 1 {
		add("chapter_template", t.refs.ChapterTemplate, 1000)
		add("content_expansion", t.refs.ContentExpansion, 900)
	}

	if refBudgetTokens < 100 {
		refBudgetTokens = 100
	}
	refBudgetRunes := int(float64(refBudgetTokens) / 0.8)

	totalRequestedRunes := 0
	for _, e := range entries {
		totalRequestedRunes += e.maxRunes
	}

	scale := 1.0
	if totalRequestedRunes > refBudgetRunes && totalRequestedRunes > 0 {
		scale = float64(refBudgetRunes) / float64(totalRequestedRunes)
	}

	minLimit := 100
	if len(entries) > 0 && refBudgetRunes < len(entries)*minLimit {
		minLimit = refBudgetRunes / len(entries)
		if minLimit < 25 {
			minLimit = 25
		}
	}

	refs := make(map[string]string, len(entries))
	for _, e := range entries {
		limit := e.maxRunes
		if scale < 1.0 {
			limit = int(float64(e.maxRunes) * scale)
			if limit < minLimit {
				limit = minLimit
			}
		}
		refs[e.key] = compactReference(e.content, limit)
	}
	return refs
}

func (t *ContextTool) architectReferences() map[string]string {
	refs := map[string]string{}
	add := func(k, v string, maxRunes int) {
		if v != "" {
			refs[k] = compactReference(v, maxRunes)
		}
	}
	add("outline_template", t.refs.OutlineTemplate, 1200)
	add("character_template", t.refs.CharacterTemplate, 1000)
	add("longform_planning", t.refs.LongformPlanning, 1000)
	add("differentiation", t.refs.Differentiation, 1000)
	add("character_building", t.refs.CharacterBuilding, 1000)
	add("adversity_conflict", t.refs.AdversityConflict, 1000)
	add("style_reference", t.refs.StyleReference, 1000)
	add("arc_templates", t.refs.ArcTemplates, 1000)
	add("anti_ai_tone", t.refs.AntiAITone, 1200) // architect 大纲去 AI 腔；亦兜 editor 走 Chapter=0 路径
	if !isNovelOrHistoryStyle(t.style) {
		add("fact_grounding", t.refs.FactGrounding, 1000)
		add("tiktok_content_safety", t.refs.TiktokSafety, 1000)
		add("humor_relatability", t.refs.HumorRelatability, 1000)
	}
	return refs
}

// foundationStatus 检查基础设定的完备性，返回缺失项列表。
// 与 save_foundation 工具共用 store.FoundationMissing 判定逻辑，保证 LLM 从
// novel_context 看到的 ready/missing 与 save_foundation 返回的 foundation_ready
// 永远一致（长篇 compass 必需项等细节不会漂移）。
func (t *ContextTool) foundationStatus() (map[string]any, error) {
	missing, err := t.store.FoundationMissing()
	if err != nil {
		return nil, err
	}
	status := map[string]any{"ready": len(missing) == 0}
	if len(missing) > 0 {
		status["missing"] = missing
	}
	if len(missing) == 1 && missing[0] == "foundation_audit" {
		fingerprint, err := t.store.FoundationFingerprint()
		if err != nil {
			return nil, err
		}
		status["fingerprint"] = fingerprint
	}
	if audit, err := t.store.Outline.LoadFoundationAudit(); err != nil {
		return nil, err
	} else if audit != nil && !audit.Ready {
		status["last_audit"] = audit
	}
	return status, nil
}

// buildRelatedChapters 根据结构化数据反查与当前章相关的历史章节。
// 从伏笔、角色出场、状态变化、关系四个维度推荐，去重后最多返回 5 条。
// 所有数据通过参数传入，不做额外 IO。
func (t *ContextTool) buildRelatedChapters(
	chapter int,
	entry *domain.OutlineEntry,
	foreshadow []domain.ForeshadowEntry,
	relationships []domain.RelationshipEntry,
	stateChanges []domain.StateChange,
	reads *contextReads,
) []domain.RelatedChapter {
	const recentWindow = 10
	const maxResults = 5

	seen := make(map[int]struct{})
	var results []domain.RelatedChapter
	add := func(ch int, reason string) {
		if ch <= 0 || ch >= chapter {
			return
		}
		// 最近几章太近，不推荐
		if ch > chapter-recentWindow {
			return
		}
		if _, ok := seen[ch]; ok {
			return
		}
		seen[ch] = struct{}{}
		results = append(results, domain.RelatedChapter{Chapter: ch, Reason: reason})
	}

	// 拼接大纲文本用于关键词匹配
	outlineText := entry.Title + " " + entry.CoreEvent
	for _, s := range entry.Scenes {
		outlineText += " " + s
	}

	// 1. 伏笔反查：活跃伏笔的描述是否与当前章大纲相关
	for _, f := range foreshadow {
		if strings.Contains(outlineText, f.ID) || containsAny(outlineText, strings.Fields(f.Description)) {
			add(f.PlantedAt, fmt.Sprintf("Chương cài phục bút %s(%s)", f.ID, utils.TruncateRunes(f.Description, 15)))
		}
		if len(results) >= maxResults {
			break
		}
	}

	// 2. 角色出场反查：批量单次遍历，IO 从 O(角色数×章节数) 降为 O(章节数)
	chars, err := t.store.Characters.Load()
	if err != nil {
		reads.warn("related_chapters.characters", err)
	}
	outlineChars := matchOutlineCharacters(outlineText, chars)
	if len(outlineChars) > 0 {
		appearances, err := t.store.Summaries.FindCharacterAppearances(outlineChars, chapter, recentWindow)
		if err != nil {
			reads.warn("related_chapters.summaries", err)
		}
		for _, name := range outlineChars {
			if len(results) >= maxResults {
				break
			}
			if ch, ok := appearances[name]; ok {
				add(ch, fmt.Sprintf("Chương nhân vật '%s' xuất hiện gần nhất", name))
			}
		}
	}

	// 3. 状态变化反查：在已加载的 slice 上操作，零 IO
	for _, name := range outlineChars {
		if len(results) >= maxResults {
			break
		}
		ch := findLastStateChange(stateChanges, name, chapter)
		if ch > 0 && ch <= chapter-recentWindow {
			add(ch, fmt.Sprintf("Chương '%s' biến đổi trạng thái", name))
		}
	}

	// 4. 关系反查：当前章涉及的角色对之间关系最后变化
	if len(relationships) > 0 && len(outlineChars) >= 2 {
		charSet := make(map[string]struct{}, len(outlineChars))
		for _, c := range outlineChars {
			charSet[c] = struct{}{}
		}
		for _, r := range relationships {
			if len(results) >= maxResults {
				break
			}
			_, aIn := charSet[r.CharacterA]
			_, bIn := charSet[r.CharacterB]
			if aIn && bIn {
				add(r.Chapter, fmt.Sprintf("Quan hệ %s-%s thay đổi", r.CharacterA, r.CharacterB))
			}
		}
	}

	return results
}

// findLastStateChange 在已加载的状态变化列表中查找实体最近一次变化的章节号。
func findLastStateChange(changes []domain.StateChange, entity string, currentChapter int) int {
	for i := len(changes) - 1; i >= 0; i-- {
		if changes[i].Entity == entity && changes[i].Chapter < currentChapter {
			return changes[i].Chapter
		}
	}
	return 0
}

// matchOutlineCharacters 从大纲文本中匹配出场角色名。
func matchOutlineCharacters(text string, chars []domain.Character) []string {
	var matched []string
	for _, c := range chars {
		if strings.Contains(text, c.Name) {
			matched = append(matched, c.Name)
			continue
		}
		for _, alias := range c.Aliases {
			if strings.Contains(text, alias) {
				matched = append(matched, c.Name)
				break
			}
		}
	}
	return matched
}

// containsAny 检查 text 是否包含 words 中的任一词（至少 2 字才匹配，避免噪音）。
func containsAny(text string, words []string) bool {
	for _, w := range words {
		if len([]rune(w)) >= 2 && strings.Contains(text, w) {
			return true
		}
	}
	return false
}

func (t *ContextTool) selectStoryThreads(state contextBuildState) []domain.RecallItem {
	if state.currentEntry == nil {
		return nil
	}
	if len(state.foreshadow) < storyThreadRecallThreshold {
		return nil
	}

	const maxThreads = 5
	var items []domain.RecallItem
	seen := make(map[string]struct{})
	picked := make(map[string]struct{}) // 已选中的伏笔 ID，供账龄回填去重
	add := func(item domain.RecallItem) {
		key := item.Kind + "|" + item.Key + "|" + item.Summary
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		picked[item.Key] = struct{}{}
		items = append(items, item)
	}

	// 1. 相关性召回：与当前章 focus 词重叠的伏笔。
	focusTerms := recallFocusTerms(state.currentEntry, state.chapterPlan)
	focusText := strings.Join(focusTerms, " ")
	for _, entry := range state.foreshadow {
		if !matchesRecallTerms(entry.ID+" "+entry.Description, focusTerms) && !strings.Contains(focusText, entry.ID) {
			continue
		}
		add(domain.RecallItem{
			Kind:    "story_thread",
			Key:     entry.ID,
			Chapter: entry.PlantedAt,
			Reason:  "Chương hiện tại có thể cần nối tiếp phục bút đã có",
			Summary: fmt.Sprintf("Phục bút “%s” được cài ở chương %d: %s", entry.ID, entry.PlantedAt, utils.TruncateRunes(entry.Description, 30)),
		})
		if len(items) >= maxThreads {
			return items
		}
	}

	// 2. 账龄回填：与当前章无关、但久挂未回收的伏笔（最旧优先），补足剩余名额。
	//    补的是相关性召回天然的盲区——独自悬挂太久、却没在本章撞上关键词的那根线。
	for _, entry := range agingForeshadow(state.foreshadow, state.chapter, picked) {
		add(domain.RecallItem{
			Kind:    "story_thread",
			Key:     entry.ID,
			Chapter: entry.PlantedAt,
			Reason:  "Phục bút treo lâu chưa thu hồi, cần chú ý kịp thời đẩy tiến hoặc thu hồi",
			Summary: fmt.Sprintf("Phục bút “%s” được cài ở chương %d, đã %d chương chưa thu hồi: %s", entry.ID, entry.PlantedAt, state.chapter-entry.PlantedAt, utils.TruncateRunes(entry.Description, 30)),
		})
		if len(items) >= maxThreads {
			break
		}
	}

	return items
}

// agingForeshadow 返回账龄 ≥ foreshadowAgingChapters 的未回收伏笔，按最旧优先排序，
// 跳过 picked 中已被相关性召回选中的。入参 all 已是 active（未回收）列表，故无需再过滤状态。
func agingForeshadow(all []domain.ForeshadowEntry, chapter int, picked map[string]struct{}) []domain.ForeshadowEntry {
	var aging []domain.ForeshadowEntry
	for _, e := range all {
		if _, ok := picked[e.ID]; ok {
			continue
		}
		if e.PlantedAt <= 0 || chapter-e.PlantedAt < foreshadowAgingChapters {
			continue
		}
		aging = append(aging, e)
	}
	sort.SliceStable(aging, func(i, j int) bool {
		return aging[i].PlantedAt < aging[j].PlantedAt
	})
	return aging
}

func (t *ContextTool) selectReviewLessons(chapter int, reads *contextReads) []domain.RecallItem {
	if chapter <= 1 {
		return nil
	}

	var items []domain.RecallItem
	seen := make(map[string]struct{})
	add := func(item domain.RecallItem) {
		key := item.Summary
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		items = append(items, item)
	}

	appendReview := func(review *domain.ReviewEntry) bool {
		if review == nil {
			return false
		}
		for i, miss := range review.ContractMisses {
			add(domain.RecallItem{
				Kind:    "review_lesson",
				Key:     fmt.Sprintf("review-%d-contract-%d", review.Chapter, i),
				Chapter: review.Chapter,
				Reason:  "Lần thẩm định gần nhất chỉ ra thiếu sót về contract",
				Summary: fmt.Sprintf("Chương %d thiếu sót contract: %s", review.Chapter, miss),
			})
			if len(items) >= 3 {
				return true
			}
		}
		for i, issue := range review.Issues {
			switch issue.Severity {
			case "", "warning", "error", "critical":
				add(domain.RecallItem{
					Kind:    "review_lesson",
					Key:     fmt.Sprintf("review-%d-issue-%d", review.Chapter, i),
					Chapter: review.Chapter,
					Reason:  "Lần thẩm định gần nhất chỉ ra vấn đề cần tránh lặp lại",
					Summary: fmt.Sprintf("Nhắc nhở thẩm định chương %d: %s", review.Chapter, utils.TruncateRunes(issue.Description, 36)),
				})
			}
			if len(items) >= 3 {
				return true
			}
		}
		return false
	}

	for ch := chapter - 1; ch >= max(chapter-3, 1); ch-- {
		review, err := t.store.World.LoadReview(ch)
		if err != nil {
			reads.warn("review", err)
			continue
		}
		if appendReview(review) {
			return items
		}
	}

	globalReview, err := t.store.World.LoadLastReview(chapter - 1)
	if err != nil {
		reads.warn("global_review", err)
	} else if appendReview(globalReview) {
		return items
	}
	return items
}

func recallFocusTerms(entry *domain.OutlineEntry, plan *domain.ChapterPlan) []string {
	if entry == nil {
		return nil
	}
	var terms []string
	add := func(v string) {
		v = strings.TrimSpace(v)
		if v != "" {
			terms = append(terms, v)
		}
	}

	add(entry.Title)
	add(entry.CoreEvent)
	add(entry.Hook)
	for _, scene := range entry.Scenes {
		add(scene)
	}
	if plan != nil {
		add(plan.Goal)
		add(plan.Hook)
		for _, point := range plan.Contract.PayoffPoints {
			add(point)
		}
		add(plan.Contract.HookGoal)
	}
	return terms
}

func matchesRecallTerms(text string, terms []string) bool {
	for _, term := range terms {
		term = strings.TrimSpace(term)
		if len([]rune(term)) < 2 {
			continue
		}
		if strings.Contains(text, term) || strings.Contains(term, text) {
			return true
		}
		if hasMeaningfulOverlap(term, text) {
			return true
		}
	}
	return false
}

func hasMeaningfulOverlap(a, b string) bool {
	ar := []rune(strings.TrimSpace(a))
	br := []rune(strings.TrimSpace(b))
	if len(ar) < 5 || len(br) < 5 {
		return false
	}
	shorter := len(ar)
	if len(br) < shorter {
		shorter = len(br)
	}
	threshold := 5
	switch {
	case shorter >= 12:
		threshold = 7
	case shorter >= 9:
		threshold = 6
	}
	return longestCommonSubstringRunes(ar, br) >= threshold
}

const storyThreadRecallThreshold = 6
const storyThreadRecallMinSelected = 2

// foreshadowAgingChapters：一条伏笔自埋设起超过这么多章仍未回收，视为"久挂"。
// 这类伏笔即使与当前章关键词无关，也回填进 story_threads，避免长篇里被彻底遗忘
// （相关性召回天然只看见与本章相关的线，看不见独自悬挂太久的那根）。
// 账龄是纯代码派生的事实（当前章 - 埋设章），只陈述"已挂 N 章未回收"，不下指令。
const foreshadowAgingChapters = 30

func longestCommonSubstringRunes(a, b []rune) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	prev := make([]int, len(b)+1)
	best := 0
	for i := 1; i <= len(a); i++ {
		curr := make([]int, len(b)+1)
		for j := 1; j <= len(b); j++ {
			if a[i-1] != b[j-1] {
				continue
			}
			curr[j] = prev[j-1] + 1
			if curr[j] > best {
				best = curr[j]
			}
		}
		prev = curr
	}
	return best
}

// truncateRunes 截断字符串到指定 rune 数。
