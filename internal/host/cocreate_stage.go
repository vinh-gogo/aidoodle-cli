package host

import (
	"fmt"
	"strings"

	"github.com/voocel/ainovel-cli/internal/store"
	"github.com/voocel/ainovel-cli/internal/utils"
)

// buildStoryStateSummary 组装一段精简的故事现状摘要，供阶段共创助手了解"已经写了什么"。
// 复用 store 访问点，只取规划方向所需的高层事实（进度 / 罗盘 / 最近卷 / 主要人物 / 活跃伏笔）；
// 不拉正文、不喂 novel_context 的全量 JSON——共创是对话，要的是可读概览，不是写作上下文。
// 任一项缺失都跳过（best-effort），返回空串表示尚无可用进度。
func buildStoryStateSummary(s *store.Store) string {
	if s == nil {
		return ""
	}
	var b strings.Builder
	var warnings []string
	warn := func(scope string, err error) {
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s đọc thất bại: %v", scope, err))
		}
	}

	if book, err := s.Book.Load(); book != nil {
		fmt.Fprintf(&b, "- Tên tác phẩm: 《%s》\n", book.Title)
	} else {
		warn("book", err)
	}

	if progress, err := s.Progress.Load(); progress != nil {
		fmt.Fprintf(&b, "- Tiến độ: Đã hoàn thành %d chương", len(progress.CompletedChapters))
		if progress.Layered {
			outline, outlineErr := s.Outline.LoadOutline()
			if outlineErr != nil {
				warn("outline", outlineErr)
			} else if len(outline) > 0 {
				fmt.Fprintf(&b, " / Hiện đã chi tiết hóa %d chương (phần sau quy hoạch động theo hồi)", len(outline))
			}
		} else if progress.TotalChapters > 0 {
			fmt.Fprintf(&b, " / Quy hoạch %d chương", progress.TotalChapters)
		}
		fmt.Fprintf(&b, ", khoảng %d từ, chương tiếp theo là chương %d\n", progress.TotalWordCount, progress.NextChapter())
		if progress.Layered && progress.CurrentVolume > 0 {
			fmt.Fprintf(&b, "- Vị trí hiện tại: Quyển %d Hồi %d\n", progress.CurrentVolume, progress.CurrentArc)
		}
	} else {
		warn("progress", err)
	}

	if compass, err := s.Outline.LoadCompass(); compass != nil {
		if dir := strings.TrimSpace(compass.EndingDirection); dir != "" {
			fmt.Fprintf(&b, "- Định hướng kết cục: %s\n", dir)
		}
		if compass.EstimatedScale != "" {
			fmt.Fprintf(&b, "- Quy mô ước tính: %s\n", compass.EstimatedScale)
		}
		if len(compass.OpenThreads) > 0 {
			fmt.Fprintf(&b, "- Tuyến dài hạn đang hoạt động: %s\n", strings.Join(compass.OpenThreads, "; "))
		}
	} else {
		warn("story_compass", err)
	}

	// Tóm tắt quyển gần nhất, để trợ lý biết câu chuyện vừa đi tới đâu
	if vols, err := s.Summaries.LoadAllVolumeSummaries(); len(vols) > 0 {
		last := vols[len(vols)-1]
		fmt.Fprintf(&b, "- Gần đây 《%s》: %s\n", last.Title, utils.TruncateRunes(last.Summary, 200))
	} else {
		warn("volume_summaries", err)
	}

	// Nhân vật chính (core/important), tối đa 8 người
	if chars, err := s.Characters.Load(); len(chars) > 0 {
		var names []string
		for _, c := range chars {
			if c.Tier == "secondary" || c.Tier == "decorative" {
				continue
			}
			line := c.Name
			if role := strings.TrimSpace(c.Role); role != "" {
				line += " (" + role + ")"
			}
			names = append(names, line)
			if len(names) >= 8 {
				break
			}
		}
		if len(names) > 0 {
			fmt.Fprintf(&b, "- Nhân vật chính: %s\n", strings.Join(names, ", "))
		}
	} else {
		warn("characters", err)
	}

	// Phục bút chưa thu hồi, tối đa 6 mục
	if fs, err := s.World.LoadActiveForeshadow(); len(fs) > 0 {
		var items []string
		for _, f := range fs {
			items = append(items, utils.TruncateRunes(f.Description, 40))
			if len(items) >= 6 {
				break
			}
		}
		fmt.Fprintf(&b, "- Phục bút chưa thu hồi: %s\n", strings.Join(items, "; "))
	} else {
		warn("foreshadow", err)
	}

	if len(warnings) > 0 {
		fmt.Fprintf(&b, "- Cảnh báo dữ liệu: %s\n", strings.Join(warnings, "; "))
	}

	return strings.TrimSpace(b.String())
}

// stageSystemPrompt lắp ráp system prompt hoàn chỉnh cho đồng sáng tạo giai đoạn: prompt giai đoạn + tóm tắt trạng thái câu chuyện hiện tại.
// Phần tóm tắt được đính kèm ở cuối dưới dạng phụ lục dữ liệu (ngăn cách bằng đường kẻ phân cách và quy cách định dạng).
func stageSystemPrompt(s *store.Store) string {
	prompt := stageCoCreateSystemPrompt
	if summary := buildStoryStateSummary(s); summary != "" {
		prompt += "\n\n---\n## Trạng thái câu chuyện hiện tại\n(Dưới đây là tóm tắt khách quan về nội dung đã viết, dùng để tham khảo khi quy hoạch phần tiếp theo, không chép lại nguyên văn trong <draft>)\n" + summary
	}
	return prompt
}
