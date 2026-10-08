package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

func TestDashCommand_RegistrationAndHighlight(t *testing.T) {
	m := Model{textarea: textarea.New()}
	m.textarea.Focus()

	for _, input := range []string{"/dash", "/dashboard", "/projects", "/history"} {
		m.textarea.SetValue(input)
		m.syncCommandInputHighlight()
		if m.commandToken == "" {
			t.Errorf("Lệnh đã đăng ký %q phải được nhận diện", input)
		}
	}
}

func TestDashCommand_ExecutionOpensModal(t *testing.T) {
	reg := commandRegistryInstance()
	cmd, ok := reg.Find("dash")
	if !ok {
		t.Fatal("Không tìm thấy lệnh /dash trong command registry")
	}

	if !cmd.AutoExecute {
		t.Error("Lệnh /dash phải có AutoExecute = true để người dùng gõ /dash -> Enter là chạy ngay")
	}

	m := Model{
		textarea: textarea.New(),
		width:    100,
		height:   30,
	}
	m.textarea.Focus()

	resModel, _ := cmd.Run(m, nil)
	updated := resModel.(Model)

	if updated.dashSelect == nil {
		t.Fatal("Chạy /dash phải khởi tạo dashSelect state")
	}
	if updated.textarea.Focused() {
		t.Error("Khi mở modal /dash, textarea phải blur")
	}
}

func TestDashState_Navigation(t *testing.T) {
	st := &dashState{
		cursor: 0,
		projects: []dashProjectItem{
			{Title: "Dự án 1", RelPath: "output\\novel-1"},
			{Title: "Dự án 2", RelPath: "output\\novel-2"},
			{Title: "Dự án 3", RelPath: "output\\novel-3"},
		},
	}
	m := Model{
		dashSelect: st,
		width:      100,
		height:     30,
	}

	// 1. Phím KeyDown: 0 -> 1
	res, _ := m.handleDashKey(tea.KeyMsg{Type: tea.KeyDown})
	m = res.(Model)
	if m.dashSelect.cursor != 1 {
		t.Fatalf("Sau khi ấn KeyDown, cursor phải là 1, nhận được %d", m.dashSelect.cursor)
	}

	// 2. Phím 'j': 1 -> 2
	res, _ = m.handleDashKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = res.(Model)
	if m.dashSelect.cursor != 2 {
		t.Fatalf("Sau khi ấn 'j', cursor phải là 2, nhận được %d", m.dashSelect.cursor)
	}

	// 3. Phím 'k': 2 -> 1
	res, _ = m.handleDashKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = res.(Model)
	if m.dashSelect.cursor != 1 {
		t.Fatalf("Sau khi ấn 'k', cursor phải là 1, nhận được %d", m.dashSelect.cursor)
	}

	// 4. Phím KeyUp: 1 -> 0
	res, _ = m.handleDashKey(tea.KeyMsg{Type: tea.KeyUp})
	m = res.(Model)
	if m.dashSelect.cursor != 0 {
		t.Fatalf("Sau khi ấn KeyUp, cursor phải là 0, nhận được %d", m.dashSelect.cursor)
	}

	// 5. Phím KeyUp khi ở 0: cuộn vòng về 2
	res, _ = m.handleDashKey(tea.KeyMsg{Type: tea.KeyUp})
	m = res.(Model)
	if m.dashSelect.cursor != 2 {
		t.Fatalf("Sau khi ấn KeyUp từ 0, cursor phải cuộn vòng đến 2, nhận được %d", m.dashSelect.cursor)
	}

	// 6. Phím KeyDown từ 2: cuộn vòng về 0
	res, _ = m.handleDashKey(tea.KeyMsg{Type: tea.KeyDown})
	m = res.(Model)
	if m.dashSelect.cursor != 0 {
		t.Fatalf("Sau khi ấn KeyDown từ cuối, cursor phải quay về 0, nhận được %d", m.dashSelect.cursor)
	}
}

func TestDashState_DismissEscAndQ(t *testing.T) {
	// Kiểm tra Esc
	{
		m := Model{
			dashSelect: &dashState{
				projects: []dashProjectItem{{Title: "Dự án 1"}},
			},
			textarea: textarea.New(),
		}
		res, _ := m.handleDashKey(tea.KeyMsg{Type: tea.KeyEsc})
		m = res.(Model)
		if m.dashSelect != nil {
			t.Fatal("Sau khi ấn Esc, dashSelect modal phải đóng (nil)")
		}
	}

	// Kiểm tra 'q'
	{
		m := Model{
			dashSelect: &dashState{
				projects: []dashProjectItem{{Title: "Dự án 1"}},
			},
			textarea: textarea.New(),
		}
		res, _ := m.handleDashKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		m = res.(Model)
		if m.dashSelect != nil {
			t.Fatal("Sau khi ấn 'q', dashSelect modal phải đóng (nil)")
		}
	}
}

func TestDashState_SelectProject(t *testing.T) {
	// Chọn dự án khác -> Đặt switchOutputDir và quit
	{
		st := &dashState{
			cursor: 1,
			projects: []dashProjectItem{
				{Title: "Dự án 1", FullPath: "D:\\test\\novel-1", IsCurrent: true},
				{Title: "Dự án 2", FullPath: "D:\\test\\novel-2", IsCurrent: false},
			},
		}
		m := Model{
			dashSelect: st,
			textarea:   textarea.New(),
		}

		res, cmd := m.handleDashKey(tea.KeyMsg{Type: tea.KeyEnter})
		m = res.(Model)
		if m.dashSelect != nil {
			t.Fatal("Sau khi chọn dự án, dashSelect phải được xóa")
		}
		if m.switchOutputDir != "D:\\test\\novel-2" {
			t.Fatalf("switchOutputDir mong đợi D:\\test\\novel-2, nhận được %s", m.switchOutputDir)
		}
		if cmd == nil {
			t.Fatal("handleDashKey phải trả về tea.Quit cmd để chuyển dự án")
		}
	}

	// Chọn dự án đang mở sẵn -> Báo thông tin, không quit
	{
		st := &dashState{
			cursor: 0,
			projects: []dashProjectItem{
				{Title: "Dự án 1", FullPath: "D:\\test\\novel-1", IsCurrent: true, RelPath: "output\\novel-1"},
			},
		}
		m := Model{
			dashSelect: st,
			width:      100,
			height:     30,
			textarea:   textarea.New(),
		}

		res, cmd := m.handleDashKey(tea.KeyMsg{Type: tea.KeyEnter})
		m = res.(Model)
		if m.dashSelect == nil {
			t.Fatal("Chọn dự án đang mở không được đóng modal")
		}
		if m.switchOutputDir != "" {
			t.Fatalf("Không được set switchOutputDir khi chọn dự án hiện tại, nhận được %s", m.switchOutputDir)
		}
		if cmd != nil {
			t.Fatal("Không được gửi command quit khi chọn dự án đang mở")
		}
		if !strings.Contains(m.dashSelect.message, "đang được mở sẵn") {
			t.Fatalf("Thông báo phải chứa 'đang được mở sẵn', nhận: %s", m.dashSelect.message)
		}
	}
}

func TestDetectStyleBadge(t *testing.T) {
	cases := []struct {
		style    string
		expected string
	}{
		{"behavioral-psychology", "[🧠 Tâm lý học hành vi]"},
		{"psychology", "[🧠 Tâm lý học hành vi]"},
		{"doodle-explainer", "[🦴 Video TikTok]"},
		{"vietnamese-history", "[⚔️ Lịch sử Việt Nam]"},
		{"novel-manga", "[📖 Tiểu thuyết / Manga]"},
		{"", "[📖 Tiểu thuyết / Manga]"},
		{"custom-mode", "[🏷️ custom-mode]"},
	}

	for _, tc := range cases {
		badge := detectStyleBadge(tc.style)
		if badge != tc.expected {
			t.Errorf("detectStyleBadge(%q) = %q, mong đợi %q", tc.style, badge, tc.expected)
		}
	}
}

func TestDashState_RenderModal(t *testing.T) {
	st := &dashState{
		cursor: 0,
		projects: []dashProjectItem{
			{
				Title:        "Vùng Xám: Giải Mã Ranh Giới Thiện - Ác",
				DirName:      "novel-20261008-1006",
				RelPath:      "output\\novel-20261008-1006",
				StyleBadge:   "[🧠 Tâm lý học hành vi]",
				ProgressText: "✓ Đã xong (3/3)",
				ModTime:      time.Now(),
				IsCurrent:    true,
			},
			{
				Title:        "Bí Ẩn Trì Hoãn",
				DirName:      "novel-20261008-0954",
				RelPath:      "output\\novel-20261008-0954",
				StyleBadge:   "[🦴 Video TikTok]",
				ProgressText: "⏳ Đang viết (1/3)",
				ModTime:      time.Now(),
			},
		},
	}

	view := renderDashModal(100, 30, st)
	if view == "" {
		t.Fatal("renderDashModal không được trả về chuỗi rỗng")
	}
	if !strings.Contains(view, "Danh sách outputs (/dash)") {
		t.Error("Modal phải hiển thị tiêu đề 'Danh sách outputs (/dash)'")
	}
	if !strings.Contains(view, "Vùng Xám") {
		t.Error("Modal phải hiển thị tên tác phẩm 'Vùng Xám'")
	}
	if !strings.Contains(view, "[🧠 Tâm lý học hành vi]") {
		t.Error("Modal phải hiển thị tag mode '[🧠 Tâm lý học hành vi]'")
	}
	if !strings.Contains(view, "output\\novel-20261008-1006") {
		t.Error("Modal phải hiển thị đường dẫn output")
	}
	if !strings.Contains(view, "✓ Đã xong (3/3)") {
		t.Error("Modal phải hiển thị trạng thái hoàn thành")
	}
	if !strings.Contains(view, "[Đang mở]") {
		t.Error("Modal phải đánh dấu dự án [Đang mở]")
	}

	nilView := renderDashModal(100, 30, nil)
	if nilView != "" {
		t.Fatalf("renderDashModal(nil) phải trả về chuỗi rỗng, nhận được %q", nilView)
	}
}

func TestScanOutputProjects(t *testing.T) {
	tempDir := t.TempDir()
	outDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Tạo 2 dự án mock
	proj1 := filepath.Join(outDir, "novel-20261008-01")
	_ = os.MkdirAll(filepath.Join(proj1, "meta"), 0755)
	_ = os.WriteFile(filepath.Join(proj1, "meta", "book.json"), []byte(`{"title":"Tác phẩm 1"}`), 0644)
	_ = os.WriteFile(filepath.Join(proj1, "meta", "run.json"), []byte(`{"style":"behavioral-psychology"}`), 0644)

	proj2 := filepath.Join(outDir, "novel-20261008-02")
	_ = os.MkdirAll(filepath.Join(proj2, "meta"), 0755)
	_ = os.WriteFile(filepath.Join(proj2, "meta", "book.json"), []byte(`{"title":"Tác phẩm 2"}`), 0644)
	_ = os.WriteFile(filepath.Join(proj2, "meta", "run.json"), []byte(`{"style":"doodle-explainer"}`), 0644)

	items := scanOutputProjects(proj1)
	if len(items) != 2 {
		t.Fatalf("scanOutputProjects phải phát hiện được 2 dự án, nhận được %d", len(items))
	}

	foundCurrent := false
	for _, it := range items {
		if it.IsCurrent {
			foundCurrent = true
			if it.DirName != "novel-20261008-01" {
				t.Errorf("Dự án hiện tại phải là novel-20261008-01, nhận: %s", it.DirName)
			}
		}
	}
	if !foundCurrent {
		t.Error("scanOutputProjects phải đánh dấu IsCurrent = true cho proj1")
	}
}
