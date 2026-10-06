package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/voocel/agentcore/schema"
	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/errs"
	"github.com/voocel/ainovel-cli/internal/store"
)

// PlanChapterTool 保存章节构思，Agent 自主决定规划粒度。
type PlanChapterTool struct {
	store *store.Store
}

func NewPlanChapterTool(store *store.Store) *PlanChapterTool {
	return &PlanChapterTool{store: store}
}

func (t *PlanChapterTool) Name() string { return "plan_chapter" }
func (t *PlanChapterTool) Description() string {
	return "Lưu cấu tứ viết chương. Agent tự quyết định mức độ chi tiết của kế hoạch, không bắt buộc chia cảnh"
}
func (t *PlanChapterTool) Label() string { return "Lập kế hoạch chương" }

// 写工具，禁止并发。
func (t *PlanChapterTool) ReadOnly(_ json.RawMessage) bool        { return false }
func (t *PlanChapterTool) ConcurrencySafe(_ json.RawMessage) bool { return false }

func (t *PlanChapterTool) Schema() map[string]any {
	return schema.Object(
		schema.Property("chapter", schema.Int("Số chương")).Required(),
		schema.Property("title", schema.String("Tiêu đề chương tạm định; sau khi viết có thể điều chỉnh theo chính văn")).Required(),
		schema.Property("goal", schema.String("Mục tiêu của chương này")).Required(),
		schema.Property("conflict", schema.String("Xung đột cốt lõi")).Required(),
		schema.Property("hook", schema.String("Móc câu cuối chương")).Required(),
		schema.Property("emotion_arc", schema.String("Đường cong cảm xúc")),
		schema.Property("notes", schema.String("Ghi chú tự do (bất cứ điều gì bạn thấy cần ghi nhớ khi viết)")),
		schema.Property("required_beats", schema.Array("Các mục tiến triển chương này bắt buộc phải hoàn thành", schema.String(""))),
		schema.Property("forbidden_moves", schema.Array("Các diễn tiến chương này tuyệt đối không được xảy ra", schema.String(""))),
		schema.Property("continuity_checks", schema.Array("Các điểm liên tục cần đối chiếu đặc biệt trong chương này", schema.String(""))),
		schema.Property("evaluation_focus", schema.Array("Các mục Editor cần kiểm tra trọng điểm", schema.String(""))),
		schema.Property("emotion_target", schema.String("Tùy chọn: cảm xúc chính mà chương này muốn độc giả cảm nhận")),
		schema.Property("payoff_points", schema.Array("Tùy chọn: các điểm tình tiết cần hồi đáp hoặc thực hiện trong chương then chốt", schema.String(""))),
		schema.Property("hook_goal", schema.String("Tùy chọn: ham muốn đọc tiếp hoặc mục tiêu hồi hộp mà cuối chương muốn thúc đẩy")),
	)
}

func (t *PlanChapterTool) Execute(_ context.Context, args json.RawMessage) (json.RawMessage, error) {
	plan, err := decodeChapterPlanArgs(args)
	if err != nil {
		return nil, fmt.Errorf("invalid args: %w: %w", errs.ErrToolArgs, err)
	}
	if plan.Chapter <= 0 {
		return nil, fmt.Errorf("chapter must be > 0: %w", errs.ErrToolArgs)
	}
	completed, err := t.store.Progress.IsChapterCompleted(plan.Chapter)
	if err != nil {
		return nil, fmt.Errorf("load progress: %w: %w", errs.ErrStoreRead, err)
	}
	if completed {
		return json.Marshal(map[string]any{
			"chapter":   plan.Chapter,
			"skipped":   true,
			"completed": true,
			"reason":    fmt.Sprintf("Chương %d đã nộp hoàn tất, không thể lập kế hoạch lại", plan.Chapter),
		})
	}
	if err := t.store.Progress.ValidateChapterWork(plan.Chapter); err != nil {
		return nil, err
	}
	if err := EnsureChapterExpanded(t.store, plan.Chapter); err != nil {
		return nil, err
	}

	if err := t.store.Drafts.SaveChapterPlan(plan); err != nil {
		return nil, fmt.Errorf("save chapter plan: %w", err)
	}
	if err := t.store.Progress.StartChapter(plan.Chapter); err != nil {
		return nil, fmt.Errorf("mark chapter in progress: %w", err)
	}

	if _, err := t.store.Checkpoints.AppendArtifact(
		domain.ChapterScope(plan.Chapter), "plan",
		fmt.Sprintf("drafts/%02d.plan.json", plan.Chapter),
	); err != nil {
		return nil, fmt.Errorf("checkpoint chapter plan: %w", err)
	}

	return json.Marshal(map[string]any{
		"planned":   true,
		"chapter":   plan.Chapter,
		"next_step": "Gọi ngay draft_chapter(chapter=số chương này, content=chuỗi chính văn đầy đủ) để ghi chính văn, đừng lập kế hoạch lại cùng một chương",
	})
}

func decodeChapterPlanArgs(args json.RawMessage) (domain.ChapterPlan, error) {
	var a struct {
		Chapter          int      `json:"chapter"`
		Title            string   `json:"title"`
		Goal             string   `json:"goal"`
		Conflict         string   `json:"conflict"`
		Hook             string   `json:"hook"`
		EmotionArc       string   `json:"emotion_arc"`
		Notes            string   `json:"notes"`
		RequiredBeats    []string `json:"required_beats"`
		ForbiddenMoves   []string `json:"forbidden_moves"`
		ContinuityChecks []string `json:"continuity_checks"`
		EvaluationFocus  []string `json:"evaluation_focus"`
		EmotionTarget    string   `json:"emotion_target"`
		PayoffPoints     []string `json:"payoff_points"`
		HookGoal         string   `json:"hook_goal"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return domain.ChapterPlan{}, err
	}

	return domain.ChapterPlan{
		Chapter:    a.Chapter,
		Title:      a.Title,
		Goal:       a.Goal,
		Conflict:   a.Conflict,
		Hook:       a.Hook,
		EmotionArc: a.EmotionArc,
		Notes:      a.Notes,
		Contract: domain.ChapterContract{
			RequiredBeats:    a.RequiredBeats,
			ForbiddenMoves:   a.ForbiddenMoves,
			ContinuityChecks: a.ContinuityChecks,
			EvaluationFocus:  a.EvaluationFocus,
			EmotionTarget:    a.EmotionTarget,
			PayoffPoints:     a.PayoffPoints,
			HookGoal:         a.HookGoal,
		},
	}, nil
}
