package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/voocel/ainovel-cli/internal/host"
	"github.com/voocel/ainovel-cli/internal/host/trend"
)

// modeTopicOption đại diện cho một chủ đề hoặc thể loại kịch bản có sẵn trong hệ thống.
type modeTopicOption struct {
	ID          string
	Title       string
	Category    string
	Description string
	Prompt      string
}

// modeSelectState lưu trữ trạng thái hiển thị bảng chọn chủ đề kịch bản (/mode).
type modeSelectState struct {
	cursor   int
	options  []modeTopicOption
	viewport viewport.Model
}

// defaultModeOptions trả về danh sách các chủ đề kịch bản cốt lõi và xu hướng sẵn có.
func defaultModeOptions(bookDir string) []modeTopicOption {
	opts := []modeTopicOption{
		{
			ID:          "hairless-marathon",
			Title:       "Vì sao con người mất gần hết lông?",
			Category:    "Tiến hóa & Sinh học",
			Description: "Siêu năng lực tản nhiệt: Ta là 'vận động viên marathon' săn mồi bằng sức bền và mồ hôi",
			Prompt:      "Vì sao con người mất gần hết lông? | Ta là \"vận động viên marathon\" săn mồi bằng sức bền và mồ hôi (giả thuyết, còn tranh luận)",
		},
		{
			ID:          "inflation-stones",
			Title:       "Bản chất lạm phát và tiền tệ mất giá",
			Category:    "Kinh tế & Xã hội",
			Description: "Củ khoai không đắt lên, vỏ sò mới rẻ đi: Giải mã bí mật tiền tệ qua góc nhìn đồ đá",
			Prompt:      "Giải thích lạm phát và tiền mất giá qua củ khoai và vỏ sò thời đồ đá (series 3 tập)",
		},
		{
			ID:          "ai-hallucination",
			Title:       "Trí tuệ nhân tạo (AI): Tại sao hay bịa chuyện?",
			Category:    "Công nghệ & AI",
			Description: "Hiện tượng ảo giác AI: Vì sao mô hình siêu thông minh nhưng đôi khi nói hươu nói vượn?",
			Prompt:      "Xu hướng trí tuệ nhân tạo (AI): tại sao AI thông minh nhưng hay bịa chuyện? (series 3 tập)",
		},
		{
			ID:          "social-algorithm",
			Title:       "Thuật toán mạng xã hội giữ chân người xem",
			Category:    "Tâm lý học hành vi",
			Description: "Bộ lạc săn bắt hái lượm trong điện thoại: Dopamine, phần thưởng ngẫu nhiên và nỗi sợ bỏ lỡ",
			Prompt:      "Thuật toán mạng xã hội giữ chân người xem như thế nào: giải thích bằng bộ lạc săn bắt hái lượm (series 3 tập)",
		},
		{
			ID:          "stone-age-fear",
			Title:       "Nỗi sợ kỷ đá: Vì sao bạn luôn lo âu vô cớ?",
			Category:    "Tâm lý học tiến hóa",
			Description: "Hạch hạnh nhân và cỗ máy sinh tồn: Người gác cổng đồ đá thức giấc giữa thế giới văn phòng",
			Prompt:      "Nỗi sợ kỷ đá: vì sao bạn luôn lo âu vô cớ và cỗ máy sinh tồn cổ xưa trong não bộ (series 3 tập)",
		},
		{
			ID:          "sleep-stress",
			Title:       "Bí mật giấc ngủ và áp lực sinh tồn",
			Category:    "Sức khỏe & Sinh học",
			Description: "Vì sao tổ tiên ngủ chập chờn canh thú dữ nhưng không bị stress kiệt quệ như con người hiện đại?",
			Prompt:      "Bí mật giấc ngủ: tại sao tổ tiên ngủ chập chờn nhưng không bị stress như người hiện đại? (series 3 tập)",
		},
		{
			ID:          "fire-brain",
			Title:       "Ngọn lửa và sự bùng nổ của não bộ",
			Category:    "Lịch sử tiến hóa",
			Description: "Bữa ăn nấu chín đã giải phóng năng lượng khổng lồ biến vượn trần trụi thành bá chủ Trái Đất",
			Prompt:      "Lửa và não bộ: bữa ăn chín đã biến vượn người thành bá chủ Trái Đất ra sao? (series 3 tập)",
		},
		{
			ID:          "style-doodle",
			Title:       "Doodle Explainer (Phong cách chuẩn TikTok Viral)",
			Category:    "Thể loại phong cách",
			Description: "Người que đồ đá giải thích thế giới hiện đại: Nhịp 3s, đệm hài hước đồng cảm, tối ưu giữ chân",
			Prompt:      "Doodle Explainer: Người que đồ đá giải thích các hiện tượng xã hội và khoa học hiện đại",
		},
		{
			ID:          "style-viet-history",
			Title:       "Lịch sử & Danh nhân Việt Nam qua góc nhìn đồ đá",
			Category:    "Văn hóa & Lịch sử",
			Description: "Kể chuyện lịch sử dân tộc gần gũi, súc tích, đúc kết bài học giá trị cho người trẻ hôm nay",
			Prompt:      "Lịch sử Việt Nam: những bài học thời đại qua lăng kính người que đồ đá dí dỏm",
		},
	}

	// Tự động bổ sung các xu hướng mới nhất nếu hệ thống đã nạp snapshot xu hướng
	if bookDir != "" {
		if snap, err := trend.LoadLatestSnapshot(bookDir); err == nil && snap != nil && len(snap.Items) > 0 {
			for i, item := range snap.Items {
				if i >= 3 {
					break
				}
				opts = append(opts, modeTopicOption{
					ID:          fmt.Sprintf("trend-%d", i+1),
					Title:       item.Title,
					Category:    "Xu hướng thịnh hành (Trend VN)",
					Description: fmt.Sprintf("Xu hướng nóng từ %s: Giải thích bằng góc nhìn người que đồ đá", item.Source),
					Prompt:      fmt.Sprintf("Giải thích xu hướng nóng: %s bằng ẩn dụ người que thời đồ đá (series 3 tập)", item.Title),
				})
			}
		}
	}

	return opts
}

func newModeSelectState(rt *host.Host, width, height int) *modeSelectState {
	bookDir := ""
	if rt != nil {
		bookDir = rt.Dir()
	}
	opts := defaultModeOptions(bookDir)
	boxW, boxH := reportModalSize(width, height)
	contentW := paddedModalContentWidth(boxW)
	vp := viewport.New(contentW, boxH-4)

	state := &modeSelectState{
		cursor:   0,
		options:  opts,
		viewport: vp,
	}
	state.refreshViewport(contentW)
	return state
}

func (s *modeSelectState) refreshViewport(contentW int) {
	titleStyle := lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	numStyle := lipgloss.NewStyle().Foreground(colorAccent2).Bold(true)
	catStyle := lipgloss.NewStyle().Foreground(colorMuted)
	descStyle := lipgloss.NewStyle().Foreground(bodyTextColor)
	selTitleStyle := lipgloss.NewStyle().Foreground(colorAccent2).Bold(true)
	selArrowStyle := lipgloss.NewStyle().Foreground(colorAccent).Bold(true)

	var b strings.Builder
	b.WriteString(titleStyle.Render("DANH SÁCH CHỦ ĐỀ & PHONG CÁCH KỊCH BẢN HIỆN CÓ"))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorDim).Render("Dùng ↑/↓ hoặc j/k để chọn · Phím 1-9 để chọn nhanh · Enter để nạp chủ đề · Esc để đóng"))
	b.WriteString("\n\n")

	for i, opt := range s.options {
		isSelected := i == s.cursor
		if isSelected {
			b.WriteString(selArrowStyle.Render("▶ ") + numStyle.Render(fmt.Sprintf("[%d] ", i+1)) + selTitleStyle.Render(opt.Title))
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().Foreground(colorAccent).Render("    🏷️  " + opt.Category + "  ·  💡 " + opt.Description))
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(colorDim).Render("  ") + catStyle.Render(fmt.Sprintf("[%d] ", i+1)) + descStyle.Render(opt.Title))
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().Foreground(colorDim).Render("    " + opt.Category + " · " + opt.Description))
		}
		b.WriteString("\n\n")
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
		"Lựa chọn chủ đề kịch bản (/mode)",
		"  ↑↓/jk Chọn · 1-9 Chọn nhanh · Enter Nạp chủ đề · Esc Đóng",
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
		state.cursor = (state.cursor - 1 + numOpts) % numOpts
		boxW, _ := reportModalSize(m.width, m.height)
		state.refreshViewport(paddedModalContentWidth(boxW))
		return m, nil

	case tea.KeyDown:
		state.cursor = (state.cursor + 1) % numOpts
		boxW, _ := reportModalSize(m.width, m.height)
		state.refreshViewport(paddedModalContentWidth(boxW))
		return m, nil

	case tea.KeyEnter:
		if state.cursor >= 0 && state.cursor < numOpts {
			chosen := state.options[state.cursor]
			m.modeSelect = nil
			m.textarea.SetValue(chosen.Prompt)
			m.applyEvent(host.Event{
				Time:     time.Now(),
				Category: "SYSTEM",
				Level:    "info",
				Summary:  fmt.Sprintf("Đã nạp chủ đề: %s", chosen.Title),
			})
			m.refreshEventViewport()
			return m, m.textarea.Focus()
		}
		m.modeSelect = nil
		return m, m.textarea.Focus()

	case tea.KeyRunes:
		r := msg.Runes[0]
		switch r {
		case 'k', 'K':
			state.cursor = (state.cursor - 1 + numOpts) % numOpts
			boxW, _ := reportModalSize(m.width, m.height)
			state.refreshViewport(paddedModalContentWidth(boxW))
			return m, nil
		case 'j', 'J':
			state.cursor = (state.cursor + 1) % numOpts
			boxW, _ := reportModalSize(m.width, m.height)
			state.refreshViewport(paddedModalContentWidth(boxW))
			return m, nil
		case 'q', 'Q':
			m.modeSelect = nil
			return m, m.textarea.Focus()
		default:
			// Phím số 1-9 để chọn nhanh
			if r >= '1' && r <= '9' {
				idx := int(r - '1')
				if idx < numOpts {
					state.cursor = idx
					chosen := state.options[idx]
					m.modeSelect = nil
					m.textarea.SetValue(chosen.Prompt)
					m.applyEvent(host.Event{
						Time:     time.Now(),
						Category: "SYSTEM",
						Level:    "info",
						Summary:  fmt.Sprintf("Đã nạp chủ đề: %s", chosen.Title),
					})
					m.refreshEventViewport()
					return m, m.textarea.Focus()
				}
			}
		}
	}

	return m, nil
}
