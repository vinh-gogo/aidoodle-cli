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

// ExpandNextArcTool 将当前已完成弧之后的骨架展开为详细章节。
type ExpandNextArcTool struct {
	store *store.Store
}

func NewExpandNextArcTool(store *store.Store) *ExpandNextArcTool {
	return &ExpandNextArcTool{store: store}
}

func (t *ExpandNextArcTool) Name() string  { return "expand_next_arc" }
func (t *ExpandNextArcTool) Label() string { return "Mở rộng hồi tiếp theo" }
func (t *ExpandNextArcTool) Description() string {
	return "Khai triển hồi khung xương kế tiếp sau hồi đã hoàn thành hiện tại. Quyển/hồi mục tiêu do hệ thống xác định dựa trên tiến độ và dàn ý; chỉ cần nộp title, goal và chapters đã được hiệu chỉnh theo các sự thật đã hoàn thành."
}

func (t *ExpandNextArcTool) ReadOnly(json.RawMessage) bool        { return false }
func (t *ExpandNextArcTool) ConcurrencySafe(json.RawMessage) bool { return false }
func (t *ExpandNextArcTool) StrictSchema() bool                   { return true }

func (t *ExpandNextArcTool) Schema() map[string]any {
	chapter := schema.Object(
		schema.Property("title", schema.String("Tiêu đề chương")).Required(),
		schema.Property("core_event", schema.String("Sự kiện cốt lõi của chương này")).Required(),
		schema.Property("hook", schema.String("Móc câu cuối chương")).Required(),
		schema.Property("scenes", schema.Array("Các cảnh dự kiến; nếu không có thì để mảng rỗng", schema.String(""))).Required(),
	)
	return schema.Object(
		schema.Property("title", schema.String("Tiêu đề hồi đã hiệu chỉnh theo các sự thật đã hoàn thành")).Required(),
		schema.Property("goal", schema.String("Mục tiêu hồi đã hiệu chỉnh theo các sự thật đã hoàn thành")).Required(),
		schema.Property("chapters", schema.Array("Kế hoạch chương chi tiết của hồi này", chapter)).Required(),
	)
}

func (t *ExpandNextArcTool) Execute(_ context.Context, args json.RawMessage) (json.RawMessage, error) {
	var expansion domain.ArcExpansion
	if err := json.Unmarshal(args, &expansion); err != nil {
		return nil, fmt.Errorf("invalid args: %w: %w", errs.ErrToolArgs, err)
	}
	position, err := t.store.ExpandNextArc(expansion)
	if err != nil {
		return nil, fmt.Errorf("expand next arc: %w: %w", errs.ErrStoreWrite, err)
	}
	if err := consumeWriterFeedback(t.store); err != nil {
		return nil, err
	}
	if _, err := t.store.Checkpoints.AppendArtifact(domain.ArcScope(position.Volume, position.Arc), t.Name(), "layered_outline.json"); err != nil {
		return nil, fmt.Errorf("checkpoint %s: %w: %w", t.Name(), errs.ErrStoreWrite, err)
	}
	return json.Marshal(map[string]any{
		"saved":    true,
		"type":     t.Name(),
		"volume":   position.Volume,
		"arc":      position.Arc,
		"title":    expansion.Title,
		"goal":     expansion.Goal,
		"chapters": len(expansion.Chapters),
	})
}
