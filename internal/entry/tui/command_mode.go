package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/voocel/ainovel-cli/internal/host"
)

// aiModeOption định nghĩa một chế độ làm việc của AI.
type aiModeOption struct {
	ID          string   // "novel-manga", "doodle-explainer", "coming-soon"
	Number      int      // 1, 2, 3
	Title       string   // "Tiểu thuyết / Manga", "Doodle Explainer", ...
	Badge       string   // "📖 Tiểu thuyết & Manga", "🦴 Video TikTok", "⏳ Dự phòng"
	StyleKey    string   // "default", "doodle-explainer", ""
	Available   bool     // true, true, false
	Description string   // Chi tiết nhiệm vụ AI thực hiện trong chế độ này
	Features    []string // Các tính năng nổi bật
}

// modeSelectState lưu trữ trạng thái hiển thị bảng chọn chế độ AI làm việc (/mode).
type modeSelectState struct {
	cursor       int
	options      []aiModeOption
	currentStyle string
	viewport     viewport.Model
	message      string
}

// defaultAIModeOptions trả về 7 chế độ làm việc của AI (5 chế độ Tiểu thuyết chuyên sâu + 2 chế độ Video).
func defaultAIModeOptions() []aiModeOption {
	return []aiModeOption{
		{
			ID:          "novel-manga",
			Number:      1,
			Title:       "Tiểu thuyết / Manga",
			Badge:       "📖 Tiểu thuyết & Truyện tranh",
			StyleKey:    "novel-manga",
			Available:   true,
			Description: "Sáng tác tiểu thuyết dài kỳ, Light Novel, Manga/Webtoon đa thể loại với phân chia hồi/chương chặt chẽ và tâm lý nhân vật sâu sắc.",
			Features: []string{
				"Dàn ý phân hồi & quyển",
				"Quy tắc thế giới đa tầng & nhất quán dài hạn",
				"Chiều sâu tâm lý nhân vật",
			},
		},
		{
			ID:          "suspense",
			Number:      2,
			Title:       "Tiểu thuyết Trinh thám & Kinh dị",
			Badge:       "🕵️ Trinh thám & Kinh dị",
			StyleKey:    "suspense",
			Available:   true,
			Description: "Sáng tác tiểu thuyết trinh thám suy luận công bằng (Fair-Play), tâm lý tội phạm đen tối và kinh dị tâm linh mang bản sắc dân gian.",
			Features: []string{
				"Ma trận Chân tướng Kép & Sổ cái Manh mối (3 manh mối thật : 2 nghi binh)",
				"Đồng hồ đếm ngược sinh tử & Nhịp điệu căng thẳng từng chương",
				"Không khí rùng rợn 5 giác quan & Cấm tuyệt đối sáo ngữ hù dọa",
			},
		},
		{
			ID:          "romance",
			Number:      3,
			Title:       "Tiểu thuyết Ngôn tình & Chữa lành",
			Badge:       "🌸 Ngôn tình & Chữa lành",
			StyleKey:    "romance",
			Available:   true,
			Description: "Sáng tác tiểu thuyết ngôn tình hiện đại, tâm lý tình cảm sâu sắc (Slow-burn) và chữa lành tâm hồn với chất văn tinh tế giàu cảm xúc.",
			Features: []string{
				"Hồ sơ Vết thương Tâm lý (Lie/Ghost) & Cân bằng Sức hút Kép",
				"Thang đo Tiến triển Cảm xúc 6 Cấp độ & Xung đột có lý do chính đáng",
				"Đối thoại giàu ẩn ý (Subtext) & Thanh lọc 100% văn mẫu tổng tài bá đạo",
			},
		},
		{
			ID:          "fantasy",
			Number:      4,
			Title:       "Tiểu thuyết Kỳ ảo & Tiên hiệp",
			Badge:       "🐉 Kỳ ảo & Tiên hiệp",
			StyleKey:    "fantasy",
			Available:   true,
			Description: "Sáng tác tiểu thuyết kỳ ảo sử thi, tiên hiệp tu chân và thăng cấp sức mạnh (Progression) với hệ thống quy tắc cứng và thế giới đa tầng.",
			Features: []string{
				"Hệ thống Phép thuật / Tu luyện Cứng (Quy tắc - Chi phí - Giới hạn)",
				"Bản đồ Thế giới Đa tầng & Sinh thái Thế lực sống động",
				"Chiến đấu chiến thuật 3 pha & Thanh lọc 100% văn phong dịch máy (Convert)",
			},
		},
		{
			ID:          "vietnamese-history",
			Number:      5,
			Title:       "Tiểu thuyết Lịch sử Việt Nam",
			Badge:       "⚔️ Nhân vật Lịch sử",
			StyleKey:    "vietnamese-history",
			Available:   true,
			Description: "Sáng tác tiểu thuyết tái hiện hào khí nhân vật lịch sử Việt Nam, tra cứu chính sử qua Tavily Search và khắc họa nghịch cảnh sinh tử.",
			Features: []string{
				"Khai phá toàn diện chân dung & nội tâm nhân vật lịch sử",
				"Tái hiện bối cảnh trong nước & quốc tế đa tầng",
				"Tra cứu & kiểm chứng dữ liệu sử học qua Tavily Search",
			},
		},
		{
			ID:          "doodle-explainer",
			Number:      6,
			Title:       "Doodle Explainer",
			Badge:       "🦴 Video ngắn TikTok / Shorts",
			StyleKey:    "doodle-explainer",
			Available:   true,
			Description: "Biên kịch video người que đồ đá giải thích kiến thức, thời lượng 5+ phút, nhịp hình ảnh 1:1, giọng kể dí dỏm, tối ưu cho nền tảng video ngắn.",
			Features: []string{
				"Hook 3s giữ chân & Tái Hook 60-90s",
				"Nhịp 1:1 Thoại (LỜI:) và Hình (HÌNH:)",
				"Cơ sở khoa học kiểm chứng (Tavily/NGUỒN:)",
			},
		},
		{
			ID:          "behavioral-psychology",
			Number:      7,
			Title:       "Tâm lý học hành vi",
			Badge:       "🧠 Video Tâm lý học",
			StyleKey:    "behavioral-psychology",
			Available:   true,
			Description: "Biên kịch video giải mã tâm lý học hành vi & bẫy nhận thức (5+ phút, nhịp 1:1 Thoại - Hình), thí nghiệm khoa học và cú hích thực chiến (Nudge).",
			Features: []string{
				"Giải mã nghịch lý & bẫy nhận thức đời thường",
				"Minh họa Não lý trí vs Não cảm xúc trực quan 1:1",
				"Thí nghiệm khoa học có nguồn & Cú hích thực chiến (Nudge)",
			},
		},
	}
}

func newModeSelectState(rt *host.Host, width, height int) *modeSelectState {
	curStyle := "doodle-explainer"
	if rt != nil {
		if s := rt.Style(); s != "" {
			curStyle = s
		}
	}
	opts := defaultAIModeOptions()
	boxW, boxH := reportModalSize(width, height)
	contentW := paddedModalContentWidth(boxW)
	vp := viewport.New(contentW, boxH-4)

	state := &modeSelectState{
		cursor:       0,
		options:      opts,
		currentStyle: curStyle,
		viewport:     vp,
	}
	for i, opt := range opts {
		if state.isOptionActive(opt) {
			state.cursor = i
			break
		}
	}
	state.refreshViewport(contentW)
	return state
}

func isNovelMode(style string) bool {
	style = strings.ToLower(strings.TrimSpace(style))
	return style == "novel-manga" || style == "novel" || style == "manga" || style == "default"
}

func isSuspenseMode(style string) bool {
	style = strings.ToLower(strings.TrimSpace(style))
	return style == "suspense" || style == "trinh-tham" || style == "mystery" || style == "thriller" || style == "horror"
}

func isRomanceMode(style string) bool {
	style = strings.ToLower(strings.TrimSpace(style))
	return style == "romance" || style == "ngon-tinh" || style == "healing"
}

func isFantasyMode(style string) bool {
	style = strings.ToLower(strings.TrimSpace(style))
	return style == "fantasy" || style == "ky-ao" || style == "tien-hiep" || style == "xianxia"
}

func isPsychologyMode(style string) bool {
	style = strings.ToLower(strings.TrimSpace(style))
	return style == "behavioral-psychology" || style == "psychology" || style == "psychological"
}

func isVietnameseHistoryMode(style string) bool {
	style = strings.ToLower(strings.TrimSpace(style))
	return style == "vietnamese-history" || style == "history" || style == "lich-su"
}

func (s *modeSelectState) isOptionActive(opt aiModeOption) bool {
	if !opt.Available {
		return false
	}
	switch opt.StyleKey {
	case "novel-manga":
		return isNovelMode(s.currentStyle)
	case "suspense":
		return isSuspenseMode(s.currentStyle)
	case "romance":
		return isRomanceMode(s.currentStyle)
	case "fantasy":
		return isFantasyMode(s.currentStyle)
	case "behavioral-psychology":
		return isPsychologyMode(s.currentStyle)
	case "vietnamese-history":
		return isVietnameseHistoryMode(s.currentStyle)
	default:
		return s.currentStyle == opt.StyleKey
	}
}

func (s *modeSelectState) refreshViewport(contentW int) {
	titleStyle := lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	numStyle := lipgloss.NewStyle().Foreground(colorAccent2).Bold(true)
	badgeStyle := lipgloss.NewStyle().Foreground(colorAccent)
	descStyle := lipgloss.NewStyle().Foreground(bodyTextColor)
	featStyle := lipgloss.NewStyle().Foreground(colorDim)
	selTitleStyle := lipgloss.NewStyle().Foreground(colorAccent2).Bold(true)
	selArrowStyle := lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	activeBadgeStyle := lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
	disabledBadgeStyle := lipgloss.NewStyle().Foreground(colorDim).Italic(true)
	warnStyle := lipgloss.NewStyle().Foreground(colorReview).Bold(true)

	var b strings.Builder
	b.WriteString(titleStyle.Render("CHỌN CHẾ ĐỘ LÀM VIỆC CỦA AI (/mode)"))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorDim).Render("Dùng ↑/↓ hoặc j/k để chọn · Phím 1-7 chọn nhanh · Enter để kích hoạt · Esc để đóng"))
	b.WriteString("\n")

	for i, opt := range s.options {
		isSelected := i == s.cursor
		isActive := s.isOptionActive(opt)

		statusText := ""
		if isActive {
			statusText = activeBadgeStyle.Render("  ✓ [ĐANG KÍCH HOẠT]")
		} else if !opt.Available {
			statusText = disabledBadgeStyle.Render("  (Chưa kích hoạt)")
		}

		if isSelected {
			b.WriteString(selArrowStyle.Render("▶ ") + numStyle.Render(fmt.Sprintf("[%d] ", opt.Number)) + selTitleStyle.Render(opt.Title) + statusText)
			b.WriteString("\n")
			b.WriteString("    " + badgeStyle.Render(opt.Badge) + "  ·  " + descStyle.Render(opt.Description))
			b.WriteString("\n")
			b.WriteString("    " + featStyle.Render("Tính năng: "+strings.Join(opt.Features, "  •  ")))
			b.WriteString("\n")
		} else {
			dimTitleStyle := lipgloss.NewStyle().Foreground(colorMuted).Bold(true)
			b.WriteString(lipgloss.NewStyle().Foreground(colorDim).Render("  ") + featStyle.Render(fmt.Sprintf("[%d] ", opt.Number)) + dimTitleStyle.Render(opt.Title) + statusText)
			b.WriteString("\n")
			b.WriteString("    " + featStyle.Render(opt.Badge+" · "+opt.Description))
			b.WriteString("\n")
		}
	}

	if s.message != "" {
		b.WriteString(warnStyle.Render("⚠️  " + s.message))
		b.WriteString("\n")
	}

	s.viewport.SetContent(b.String())

	// Tự động cuộn viewport để mục đang chọn luôn hiển thị đầy đủ
	if s.viewport.Height > 0 {
		itemTop := 2 + s.cursor*2
		itemBottom := itemTop + 2
		if s.cursor == 0 {
			s.viewport.GotoTop()
		} else if itemTop < s.viewport.YOffset {
			s.viewport.SetYOffset(itemTop)
		} else if itemBottom >= s.viewport.YOffset+s.viewport.Height {
			s.viewport.SetYOffset(itemBottom - s.viewport.Height + 1)
		}
	}
}

func renderModeSelectModal(width, height int, state *modeSelectState) string {
	if state == nil {
		return ""
	}

	boxW, boxH := reportModalSize(width, height)
	contentW := paddedModalContentWidth(boxW)

	if state.viewport.Width != contentW {
		state.viewport.Width = contentW
	}
	if state.viewport.Height != boxH-4 {
		state.viewport.Height = boxH - 4
	}

	state.refreshViewport(contentW)

	modal := renderPaddedModalFrame(
		boxW,
		boxH,
		"Chế độ làm việc AI (/mode)",
		"  ↑↓/jk Chọn · 1-7 Chọn nhanh · Enter Kích hoạt · Esc Đóng",
		strings.Split(state.viewport.View(), "\n"),
	)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, modal)
}

func (m Model) handleModeSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.modeSelect == nil {
		return m, nil
	}

	state := m.modeSelect
	numOpts := len(state.options)

	switch msg.Type {
	case tea.KeyEsc:
		m.modeSelect = nil
		return m, m.textarea.Focus()

	case tea.KeyUp:
		state.message = ""
		state.cursor = (state.cursor - 1 + numOpts) % numOpts
		boxW, _ := reportModalSize(m.width, m.height)
		state.refreshViewport(paddedModalContentWidth(boxW))
		return m, nil

	case tea.KeyDown:
		state.message = ""
		state.cursor = (state.cursor + 1) % numOpts
		boxW, _ := reportModalSize(m.width, m.height)
		state.refreshViewport(paddedModalContentWidth(boxW))
		return m, nil

	case tea.KeyEnter:
		if state.cursor >= 0 && state.cursor < numOpts {
			chosen := state.options[state.cursor]
			if !chosen.Available {
				state.message = fmt.Sprintf("Chế độ [%d] %s đang được nghiên cứu ở bước tiếp theo!", chosen.Number, chosen.Title)
				boxW, _ := reportModalSize(m.width, m.height)
				state.refreshViewport(paddedModalContentWidth(boxW))
				return m, nil
			}

			m.modeSelect = nil
			if m.runtime != nil {
				if err := m.runtime.SetStyle(chosen.StyleKey); err != nil {
					m.applyEvent(host.Event{
						Time:     time.Now(),
						Category: "ERROR",
						Level:    "error",
						Summary:  fmt.Sprintf("Chuyển đổi chế độ thất bại: %v", err),
					})
				} else {
					m.applyEvent(host.Event{
						Time:     time.Now(),
						Category: "SYSTEM",
						Level:    "info",
						Summary:  fmt.Sprintf("Đã chuyển sang chế độ làm việc: %s", chosen.Title),
					})
				}
			}
			m.refreshEventViewport()
			return m, tea.Batch(fetchSnapshot(m.runtime), m.textarea.Focus())
		}
		m.modeSelect = nil
		return m, m.textarea.Focus()

	case tea.KeyRunes:
		r := msg.Runes[0]
		switch r {
		case 'k', 'K':
			state.message = ""
			state.cursor = (state.cursor - 1 + numOpts) % numOpts
			boxW, _ := reportModalSize(m.width, m.height)
			state.refreshViewport(paddedModalContentWidth(boxW))
			return m, nil
		case 'j', 'J':
			state.message = ""
			state.cursor = (state.cursor + 1) % numOpts
			boxW, _ := reportModalSize(m.width, m.height)
			state.refreshViewport(paddedModalContentWidth(boxW))
			return m, nil
		case 'q', 'Q':
			m.modeSelect = nil
			return m, m.textarea.Focus()
		default:
			// Phím số 1-7 để chọn nhanh
			if r >= '1' && r <= '7' {
				idx := int(r - '1')
				if idx < numOpts {
					chosen := state.options[idx]
					if !chosen.Available {
						state.cursor = idx
						state.message = fmt.Sprintf("Chế độ [%d] %s đang được nghiên cứu ở bước tiếp theo, vui lòng chọn chế độ 1 hoặc 2!", chosen.Number, chosen.Title)
						boxW, _ := reportModalSize(m.width, m.height)
						state.refreshViewport(paddedModalContentWidth(boxW))
						return m, nil
					}

					m.modeSelect = nil
					if m.runtime != nil {
						if err := m.runtime.SetStyle(chosen.StyleKey); err != nil {
							m.applyEvent(host.Event{
								Time:     time.Now(),
								Category: "ERROR",
								Level:    "error",
								Summary:  fmt.Sprintf("Chuyển đổi chế độ thất bại: %v", err),
							})
						} else {
							m.applyEvent(host.Event{
								Time:     time.Now(),
								Category: "SYSTEM",
								Level:    "info",
								Summary:  fmt.Sprintf("Đã chuyển sang chế độ làm việc: %s", chosen.Title),
							})
						}
					}
					m.refreshEventViewport()
					return m, tea.Batch(fetchSnapshot(m.runtime), m.textarea.Focus())
				}
			}
		}
	}

	return m, nil
}
