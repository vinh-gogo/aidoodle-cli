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

// defaultAIModeOptions trả về 3 chế độ làm việc của AI theo yêu cầu hệ thống.
func defaultAIModeOptions() []aiModeOption {
	return []aiModeOption{
		{
			ID:          "novel-manga",
			Number:      1,
			Title:       "Tiểu thuyết / Manga",
			Badge:       "📖 Tiểu thuyết & Truyện tranh",
			StyleKey:    "novel-manga",
			Available:   true,
			Description: "Sáng tác tiểu thuyết dài kỳ, truyện tranh (Manga/Webtoon) với phân chia hồi/chương chặt chẽ, phát triển tâm lý nhân vật và quy tắc thế giới chuyên sâu.",
			Features: []string{
				"Dàn ý phân hồi & quyển",
				"Quy tắc thế giới đa tầng & nhất quán dài hạn",
				"Chiều sâu tâm lý nhân vật",
			},
		},
		{
			ID:          "doodle-explainer",
			Number:      2,
			Title:       "Doodle Explainer",
			Badge:       "🦴 Video ngắn TikTok / Shorts",
			StyleKey:    "doodle-explainer",
			Available:   true,
			Description: "Biên kịch video người que đồ đá giải thích kiến thức, thời lượng 5+ phút, nhịp hình ảnh 1:1, giọng kể dí dỏm, phản trực giác, tối ưu cho nền tảng video ngắn.",
			Features: []string{
				"Hook 3s giữ chân & Tái Hook 60-90s",
				"Nhịp 1:1 Thoại (LỜI:) và Hình (HÌNH:)",
				"Cơ sở khoa học kiểm chứng (Tavily/NGUỒN:)",
			},
		},
		{
			ID:          "behavioral-psychology",
			Number:      3,
			Title:       "Tâm lý học hành vi",
			Badge:       "🧠 Video Tâm lý học",
			StyleKey:    "behavioral-psychology",
			Available:   true,
			Description: "Biên kịch video giải mã tâm lý học hành vi & bẫy nhận thức (5+ phút, nhịp 1:1 Thoại - Hình), giải mã sự phi lý trí đời thường, thí nghiệm khoa học chuẩn xác và cú hích thực chiến (Nudge).",
			Features: []string{
				"Giải mã nghịch lý & bẫy nhận thức đời thường",
				"Minh họa Não lý trí vs Não cảm xúc trực quan 1:1",
				"Thí nghiệm khoa học có nguồn & Cú hích thực chiến (Nudge)",
			},
		},
		{
			ID:          "vietnamese-history",
			Number:      4,
			Title:       "Tiểu thuyết Lịch sử Việt Nam",
			Badge:       "⚔️ Nhân vật Lịch sử",
			StyleKey:    "vietnamese-history",
			Available:   true,
			Description: "Sáng tác tiểu thuyết tái hiện hào khí nhân vật lịch sử Việt Nam, tra cứu chính sử qua Tavily Search, khai phá toàn diện chân dung nhân vật, bối cảnh trong nước & quốc tế, và những nghịch cảnh sinh tử.",
			Features: []string{
				"Khai phá toàn diện chân dung & nội tâm nhân vật",
				"Tái hiện bối cảnh trong nước & quốc tế đa tầng",
				"Đối diện hiểm nguy & quyết định sinh tử bi tráng",
				"Tra cứu & kiểm chứng dữ liệu sử học qua Tavily Search",
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

	// Đặt cursor mặc định vào chế độ hiện tại
	initialCursor := 0
	for i, opt := range opts {
		if opt.Available {
			if opt.StyleKey == curStyle || 
			   (opt.StyleKey == "novel-manga" && isNovelMode(curStyle)) ||
			   (opt.StyleKey == "behavioral-psychology" && isPsychologyMode(curStyle)) ||
			   (opt.StyleKey == "vietnamese-history" && isVietnameseHistoryMode(curStyle)) {
				initialCursor = i
				break
			}
		}
	}

	state := &modeSelectState{
		cursor:       initialCursor,
		options:      opts,
		currentStyle: curStyle,
		viewport:     vp,
	}
	state.refreshViewport(contentW)
	return state
}

func isNovelMode(style string) bool {
	style = strings.ToLower(strings.TrimSpace(style))
	return style == "novel-manga" || style == "novel" || style == "manga" || style == "fantasy" || style == "romance" || style == "suspense" || style == "default"
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
	if opt.StyleKey == "novel-manga" {
		return isNovelMode(s.currentStyle)
	}
	if opt.StyleKey == "behavioral-psychology" {
		return isPsychologyMode(s.currentStyle)
	}
	if opt.StyleKey == "vietnamese-history" {
		return isVietnameseHistoryMode(s.currentStyle)
	}
	return s.currentStyle == opt.StyleKey
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
	b.WriteString(lipgloss.NewStyle().Foreground(colorDim).Render("Dùng ↑/↓ hoặc j/k để chọn · Phím 1-3 chọn nhanh · Enter để kích hoạt · Esc để đóng"))
	b.WriteString("\n\n")

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
		} else {
			dimTitleStyle := lipgloss.NewStyle().Foreground(colorMuted).Bold(true)
			b.WriteString(lipgloss.NewStyle().Foreground(colorDim).Render("  ") + featStyle.Render(fmt.Sprintf("[%d] ", opt.Number)) + dimTitleStyle.Render(opt.Title) + statusText)
			b.WriteString("\n")
			b.WriteString("    " + featStyle.Render(opt.Badge+" · "+opt.Description))
		}
		b.WriteString("\n\n")
	}

	if s.message != "" {
		b.WriteString(warnStyle.Render("⚠️  " + s.message))
		b.WriteString("\n")
	}

	s.viewport.SetContent(b.String())

	// Tự động cuộn viewport để mục đang chọn luôn hiển thị đầy đủ
	if s.viewport.Height > 0 {
		itemTop := 3 + s.cursor*3
		itemBottom := itemTop + 1
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
		"  ↑↓/jk Chọn · 1-3 Chọn nhanh · Enter Kích hoạt · Esc Đóng",
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
			// Phím số 1-4 để chọn nhanh
			if r >= '1' && r <= '4' {
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
