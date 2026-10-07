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
			StyleKey:    "default",
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
			ID:          "coming-soon",
			Number:      3,
			Title:       "Chế độ mở rộng (Đang nghiên cứu & suy nghĩ tiếp)",
			Badge:       "⏳ Sắp ra mắt",
			StyleKey:    "",
			Available:   false,
			Description: "Dành cho các chế độ AI tiếp theo (như Podcast đối thoại, Kịch bản phóng sự, Video tài liệu chuyên sâu...). Hiện để ngỏ để cấu hình ở bước tiếp theo.",
			Features: []string{
				"Khung cấu hình mở rộng trong tương lai",
				"Đang chờ ý tưởng và kịch bản thiết kế tiếp",
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
			if opt.StyleKey == curStyle || (opt.StyleKey == "default" && curStyle != "doodle-explainer") {
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

func (s *modeSelectState) isOptionActive(opt aiModeOption) bool {
	if !opt.Available {
		return false
	}
	if opt.StyleKey == "doodle-explainer" {
		return s.currentStyle == "doodle-explainer"
	}
	// "default" hoặc các phong cách tiểu thuyết khác
	return s.currentStyle != "doodle-explainer"
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
				_ = m.runtime.SetStyle(chosen.StyleKey)
			}
			m.applyEvent(host.Event{
				Time:     time.Now(),
				Category: "SYSTEM",
				Level:    "info",
				Summary:  fmt.Sprintf("Đã chuyển sang chế độ làm việc: %s", chosen.Title),
			})
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
			// Phím số 1-3 để chọn nhanh
			if r >= '1' && r <= '3' {
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
						_ = m.runtime.SetStyle(chosen.StyleKey)
					}
					m.applyEvent(host.Event{
						Time:     time.Now(),
						Category: "SYSTEM",
						Level:    "info",
						Summary:  fmt.Sprintf("Đã chuyển sang chế độ làm việc: %s", chosen.Title),
					})
					m.refreshEventViewport()
					return m, tea.Batch(fetchSnapshot(m.runtime), m.textarea.Focus())
				}
			}
		}
	}

	return m, nil
}
