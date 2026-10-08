package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/voocel/ainovel-cli/internal/host"
)

// dashProjectItem lưu trữ thông tin tóm tắt của một dự án trong thư mục output.
type dashProjectItem struct {
	DirName      string    // e.g. "novel-20261008-1006"
	RelPath      string    // e.g. "output\novel-20261008-1006"
	FullPath     string    // đường dẫn đầy đủ
	Title        string    // Tiêu đề sách/kịch bản
	Prompt       string    // StartPrompt dự phòng nếu chưa có tiêu đề
	Style        string    // "behavioral-psychology", "doodle-explainer", ...
	StyleBadge   string    // "[🧠 Tâm lý học hành vi]", "[🦴 Video TikTok]", ...
	ProgressText string    // "✓ Đã xong (3/3)", "⏳ Đang viết (1/3)", ...
	Phase        string    // "complete", "writing", "init"
	TotalCh      int
	CompletedCh  int
	ModTime      time.Time
	IsCurrent    bool // true nếu là dự án đang mở hiện tại
}

// dashState quản lý trạng thái modal danh sách outputs (/dash).
type dashState struct {
	cursor   int
	projects []dashProjectItem
	viewport viewport.Model
	message  string
}

// detectStyleBadge tạo tag hiển thị phong cách trực quan cho từng mode.
func detectStyleBadge(style string) string {
	switch strings.ToLower(style) {
	case "behavioral-psychology", "psychology":
		return "[🧠 Tâm lý học hành vi]"
	case "doodle-explainer":
		return "[🦴 Video TikTok]"
	case "vietnamese-history":
		return "[⚔️ Lịch sử Việt Nam]"
	case "novel-manga", "novel", "manga", "default", "":
		return "[📖 Tiểu thuyết / Manga]"
	default:
		return fmt.Sprintf("[🏷️ %s]", style)
	}
}

// detectProjectStyle trích xuất style của một dự án từ đĩa.
func detectProjectStyle(dir string) string {
	// 1. Thử đọc meta/run.json
	runJSONPath := filepath.Join(dir, "meta", "run.json")
	if data, err := os.ReadFile(runJSONPath); err == nil {
		var meta struct {
			Style string `json:"style"`
		}
		if err := json.Unmarshal(data, &meta); err == nil && strings.TrimSpace(meta.Style) != "" {
			return strings.TrimSpace(meta.Style)
		}
	}

	// 2. Thử đọc meta/user_rules.json
	userRulesPath := filepath.Join(dir, "meta", "user_rules.json")
	if data, err := os.ReadFile(userRulesPath); err == nil {
		var r struct {
			Structured struct {
				Genre string `json:"genre"`
			} `json:"structured"`
		}
		if err := json.Unmarshal(data, &r); err == nil && strings.TrimSpace(r.Structured.Genre) != "" {
			return strings.TrimSpace(r.Structured.Genre)
		}
	}

	return ""
}

// scanOutputProjects quét toàn bộ các thư mục dự án trong output/ và trích xuất siêu dữ liệu.
func scanOutputProjects(currentDir string) []dashProjectItem {
	baseDir := "output"
	if currentDir != "" {
		parent := filepath.Dir(currentDir)
		if parent != "" && parent != "." {
			baseDir = parent
		}
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil && baseDir != "output" {
		baseDir = "output"
		entries, err = os.ReadDir(baseDir)
	}
	if err != nil {
		return nil
	}

	currentClean := ""
	if currentDir != "" {
		currentClean = filepath.Clean(currentDir)
	}

	var items []dashProjectItem
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := entry.Name()
		projectPath := filepath.Join(baseDir, dirName)

		// Chỉ nhận các thư mục novel hoặc thư mục có meta/ bên trong
		metaDir := filepath.Join(projectPath, "meta")
		if !strings.HasPrefix(dirName, "novel") {
			if fi, err := os.Stat(metaDir); err != nil || !fi.IsDir() {
				continue
			}
		}

		info, err := entry.Info()
		modTime := time.Now()
		if err == nil {
			modTime = info.ModTime()
		}

		item := dashProjectItem{
			DirName:  dirName,
			RelPath:  filepath.Join("output", dirName),
			FullPath: projectPath,
			ModTime:  modTime,
		}

		if currentClean != "" && filepath.Clean(projectPath) == currentClean {
			item.IsCurrent = true
		}

		// 1. Đọc meta/run.json
		runJSONPath := filepath.Join(projectPath, "meta", "run.json")
		var runMeta struct {
			StartedAt   string `json:"started_at"`
			Style       string `json:"style"`
			StartPrompt string `json:"start_prompt"`
		}
		if data, err := os.ReadFile(runJSONPath); err == nil {
			if err := json.Unmarshal(data, &runMeta); err == nil {
				item.Style = runMeta.Style
				item.Prompt = runMeta.StartPrompt
				if t, err := time.Parse(time.RFC3339, runMeta.StartedAt); err == nil {
					item.ModTime = t
				}
			}
		}
		if item.Style == "" {
			item.Style = detectProjectStyle(projectPath)
		}
		item.StyleBadge = detectStyleBadge(item.Style)

		// 2. Tìm Title dự án
		// Thử meta/book.json
		bookJSONPath := filepath.Join(projectPath, "meta", "book.json")
		if data, err := os.ReadFile(bookJSONPath); err == nil {
			var b struct {
				Title string `json:"title"`
			}
			if err := json.Unmarshal(data, &b); err == nil && strings.TrimSpace(b.Title) != "" {
				item.Title = strings.TrimSpace(b.Title)
			}
		}
		// Thử book.json ở thư mục gốc
		if item.Title == "" {
			if data, err := os.ReadFile(filepath.Join(projectPath, "book.json")); err == nil {
				var b struct {
					Title string `json:"title"`
				}
				if err := json.Unmarshal(data, &b); err == nil && strings.TrimSpace(b.Title) != "" {
					item.Title = strings.TrimSpace(b.Title)
				}
			}
		}
		// Thử summaries/01.json
		if item.Title == "" {
			if data, err := os.ReadFile(filepath.Join(projectPath, "summaries", "01.json")); err == nil {
				var s struct {
					Title string `json:"title"`
				}
				if err := json.Unmarshal(data, &s); err == nil && strings.TrimSpace(s.Title) != "" {
					item.Title = strings.TrimSpace(s.Title)
				}
			}
		}
		// Dự phòng: StartPrompt
		if item.Title == "" && item.Prompt != "" {
			item.Title = item.Prompt
			runes := []rune(item.Title)
			if len(runes) > 55 {
				item.Title = string(runes[:52]) + "..."
			}
		}
		if item.Title == "" {
			item.Title = dirName
		}

		// 3. Đọc meta/progress.json
		progJSONPath := filepath.Join(projectPath, "meta", "progress.json")
		if data, err := os.ReadFile(progJSONPath); err == nil {
			var prog struct {
				Phase             string `json:"phase"`
				TotalChapters     int    `json:"total_chapters"`
				CompletedChapters []int  `json:"completed_chapters"`
			}
			if err := json.Unmarshal(data, &prog); err == nil {
				item.Phase = prog.Phase
				item.TotalCh = prog.TotalChapters
				item.CompletedCh = len(prog.CompletedChapters)
				if prog.Phase == "complete" || (prog.TotalChapters > 0 && len(prog.CompletedChapters) >= prog.TotalChapters) {
					item.ProgressText = fmt.Sprintf("✓ Đã xong (%d/%d)", item.CompletedCh, item.TotalCh)
				} else if len(prog.CompletedChapters) > 0 {
					item.ProgressText = fmt.Sprintf("⏳ Đang viết (%d/%d)", item.CompletedCh, item.TotalCh)
				} else {
					item.ProgressText = "🌱 Khởi tạo"
				}
			}
		}
		if item.ProgressText == "" {
			if chEntries, err := os.ReadDir(filepath.Join(projectPath, "chapters")); err == nil {
				cnt := 0
				for _, ce := range chEntries {
					if !ce.IsDir() && strings.HasSuffix(ce.Name(), ".md") {
						cnt++
					}
				}
				if cnt > 0 {
					item.ProgressText = fmt.Sprintf("✓ %d chương", cnt)
				} else {
					item.ProgressText = "🌱 Trống"
				}
			} else {
				item.ProgressText = "🌱 Trống"
			}
		}

		items = append(items, item)
	}

	// Sắp xếp: Mới nhất lên đầu (theo tên thư mục timestamp hoặc ModTime)
	sort.Slice(items, func(i, j int) bool {
		if strings.HasPrefix(items[i].DirName, "novel-") && strings.HasPrefix(items[j].DirName, "novel-") {
			return items[i].DirName > items[j].DirName
		}
		return items[i].ModTime.After(items[j].ModTime)
	})

	return items
}

// dashModalSize tính toán kích thước modal dashboard tối ưu cho việc hiển thị bảng danh sách.
func dashModalSize(termW, termH int) (int, int) {
	w := termW - 8
	if w > 115 {
		w = 115
	}
	if w < 76 {
		w = 76
	}
	h := termH * 85 / 100
	if h < 20 {
		h = termH - 2
	}
	return w, h
}

func newDashState(rt *host.Host, width, height int) *dashState {
	curDir := ""
	if rt != nil {
		curDir = rt.Dir()
	}
	projects := scanOutputProjects(curDir)
	boxW, boxH := dashModalSize(width, height)
	contentW := paddedModalContentWidth(boxW)
	vp := viewport.New(contentW, boxH-4)

	cursor := 0
	for i, p := range projects {
		if p.IsCurrent {
			cursor = i
			break
		}
	}

	state := &dashState{
		cursor:   cursor,
		projects: projects,
		viewport: vp,
	}
	state.refreshViewport(contentW)
	return state
}

func (s *dashState) refreshViewport(contentW int) {
	if len(s.projects) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Italic(true)
		s.viewport.SetContent(emptyStyle.Render("Không tìm thấy dự án nào trong thư mục output."))
		return
	}

	var b strings.Builder
	for i, p := range s.projects {
		isSelected := i == s.cursor

		titleStyle := lipgloss.NewStyle().Bold(true)
		badgeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		progressStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		if strings.Contains(p.ProgressText, "Đang viết") {
			progressStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
		}

		cursorPrefix := "   "
		if isSelected {
			cursorPrefix = lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true).Render(" ▸ ")
			titleStyle = titleStyle.Foreground(lipgloss.Color("51")).Underline(true)
		} else {
			titleStyle = titleStyle.Foreground(lipgloss.Color("255"))
		}

		currentTag := ""
		if p.IsCurrent {
			currentTag = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true).Render(" ⭐ [Đang mở]")
		}

		numStr := fmt.Sprintf("[%d] ", i+1)
		displayTitle := p.Title
		titleRunes := []rune(displayTitle)
		maxTitleLen := contentW - 40
		if maxTitleLen < 25 {
			maxTitleLen = 25
		}
		if len(titleRunes) > maxTitleLen {
			displayTitle = string(titleRunes[:maxTitleLen-3]) + "..."
		}

		line1 := fmt.Sprintf("%s%s%s  %s",
			cursorPrefix,
			numStr,
			badgeStyle.Render(p.StyleBadge),
			titleStyle.Render(displayTitle),
		)
		if p.IsCurrent {
			line1 += " " + currentTag
		}
		b.WriteString(line1 + "\n")

		pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
		timeStr := p.ModTime.Format("02/01 15:04")
		line2 := fmt.Sprintf("     %s   %s   (Cập nhật: %s)",
			pathStyle.Render(p.RelPath),
			progressStyle.Render(p.ProgressText),
			timeStr,
		)
		b.WriteString(line2 + "\n")

		if i < len(s.projects)-1 {
			b.WriteString("\n")
		}
	}

	s.viewport.SetContent(b.String())

	// Tự động cuộn viewport để con trỏ luôn hiển thị
	itemHeight := 3 // Mỗi item gồm 2 dòng thông tin + 1 dòng ngắt
	itemTop := s.cursor * itemHeight
	itemBottom := itemTop + 1
	if itemTop < s.viewport.YOffset {
		s.viewport.SetYOffset(itemTop)
	} else if itemBottom >= s.viewport.YOffset+s.viewport.Height {
		s.viewport.SetYOffset(itemBottom - s.viewport.Height + 1)
	}
}

func renderDashModal(width, height int, state *dashState) string {
	if state == nil {
		return ""
	}

	boxW, boxH := dashModalSize(width, height)
	contentW := paddedModalContentWidth(boxW)

	if state.viewport.Width != contentW {
		state.viewport.Width = contentW
	}
	if state.viewport.Height != boxH-4 {
		state.viewport.Height = boxH - 4
	}

	state.refreshViewport(contentW)

	subtitle := fmt.Sprintf("  Tổng cộng: %d dự án outputs · ↑↓/jk Di chuyển · Enter Mở dự án · Esc Đóng", len(state.projects))
	if state.message != "" {
		subtitle = fmt.Sprintf("  ⚠️ %s", state.message)
	}

	modal := renderPaddedModalFrame(
		boxW,
		boxH,
		"Danh sách outputs (/dash)",
		subtitle,
		strings.Split(state.viewport.View(), "\n"),
	)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, modal)
}

func (m Model) handleDashKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.dashSelect == nil {
		return m, nil
	}

	state := m.dashSelect
	numProjects := len(state.projects)

	switch msg.Type {
	case tea.KeyEsc:
		m.dashSelect = nil
		return m, m.textarea.Focus()

	case tea.KeyUp:
		state.message = ""
		if numProjects > 0 {
			state.cursor = (state.cursor - 1 + numProjects) % numProjects
			boxW, _ := dashModalSize(m.width, m.height)
			state.refreshViewport(paddedModalContentWidth(boxW))
		}
		return m, nil

	case tea.KeyDown:
		state.message = ""
		if numProjects > 0 {
			state.cursor = (state.cursor + 1) % numProjects
			boxW, _ := dashModalSize(m.width, m.height)
			state.refreshViewport(paddedModalContentWidth(boxW))
		}
		return m, nil

	case tea.KeyEnter:
		if numProjects > 0 && state.cursor >= 0 && state.cursor < numProjects {
			chosen := state.projects[state.cursor]
			if chosen.IsCurrent {
				state.message = fmt.Sprintf("Dự án %s (%s) đang được mở sẵn.", chosen.Title, chosen.RelPath)
				boxW, _ := dashModalSize(m.width, m.height)
				state.refreshViewport(paddedModalContentWidth(boxW))
				return m, nil
			}

			// Chuyển sang dự án đã chọn
			m.dashSelect = nil
			m.switchOutputDir = chosen.FullPath
			return m, tea.Quit
		}
		m.dashSelect = nil
		return m, m.textarea.Focus()

	case tea.KeyRunes:
		r := msg.Runes[0]
		switch r {
		case 'k', 'K':
			state.message = ""
			if numProjects > 0 {
				state.cursor = (state.cursor - 1 + numProjects) % numProjects
				boxW, _ := dashModalSize(m.width, m.height)
				state.refreshViewport(paddedModalContentWidth(boxW))
			}
			return m, nil
		case 'j', 'J':
			state.message = ""
			if numProjects > 0 {
				state.cursor = (state.cursor + 1) % numProjects
				boxW, _ := dashModalSize(m.width, m.height)
				state.refreshViewport(paddedModalContentWidth(boxW))
			}
			return m, nil
		case 'q', 'Q':
			m.dashSelect = nil
			return m, m.textarea.Focus()
		}
	}

	return m, nil
}
