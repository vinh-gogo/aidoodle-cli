package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

func TestModeCommand_RegistrationAndHighlight(t *testing.T) {
	m := Model{textarea: textarea.New()}
	m.textarea.Focus()

	for _, input := range []string{"/mode", "/topic", "/topics"} {
		m.textarea.SetValue(input)
		m.syncCommandInputHighlight()
		if m.commandToken == "" {
			t.Errorf("Lệnh đã đăng ký %q phải được nhận diện", input)
		}
	}
}

func TestModeCommand_ExecutionOpensModal(t *testing.T) {
	reg := commandRegistryInstance()
	cmd, ok := reg.Find("mode")
	if !ok {
		t.Fatal("Không tìm thấy lệnh /mode trong command registry")
	}

	m := Model{
		textarea: textarea.New(),
		width:    100,
		height:   30,
	}
	m.textarea.Focus()

	resModel, _ := cmd.Run(m, nil)
	updated := resModel.(Model)

	if updated.modeSelect == nil {
		t.Fatal("Chạy /mode phải khởi tạo modeSelect state")
	}
	if len(updated.modeSelect.options) != 3 {
		t.Fatalf("modeSelect phải có đúng 3 tùy chọn chế độ, nhận được %d", len(updated.modeSelect.options))
	}
	if updated.modeSelect.options[0].Title != "Tiểu thuyết / Manga" {
		t.Errorf("Tùy chọn 1 phải là Tiểu thuyết / Manga, nhận: %s", updated.modeSelect.options[0].Title)
	}
	if updated.modeSelect.options[1].Title != "Doodle Explainer" {
		t.Errorf("Tùy chọn 2 phải là Doodle Explainer, nhận: %s", updated.modeSelect.options[1].Title)
	}
	if updated.modeSelect.options[2].Available {
		t.Error("Tùy chọn 3 (chưa có/sắp ra mắt) không được đặt Available = true")
	}
	if updated.textarea.Focused() {
		t.Fatal("Textarea phải bị blur khi mở modal lựa chọn chế độ")
	}
}

func TestModeSelect_Navigation(t *testing.T) {
	m := Model{
		textarea:   textarea.New(),
		width:      100,
		height:     30,
		modeSelect: newModeSelectState(nil, 100, 30),
	}
	numOpts := len(m.modeSelect.options)
	m.modeSelect.cursor = 0

	// Down arrow
	res, _ := m.handleModeSelectKey(tea.KeyMsg{Type: tea.KeyDown})
	m = res.(Model)
	if m.modeSelect.cursor != 1 {
		t.Fatalf("Sau khi ấn KeyDown, cursor phải là 1, nhận được %d", m.modeSelect.cursor)
	}

	// 'j' key
	res, _ = m.handleModeSelectKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = res.(Model)
	if m.modeSelect.cursor != 2 {
		t.Fatalf("Sau khi ấn 'j', cursor phải là 2, nhận được %d", m.modeSelect.cursor)
	}

	// 'k' key
	res, _ = m.handleModeSelectKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = res.(Model)
	if m.modeSelect.cursor != 1 {
		t.Fatalf("Sau khi ấn 'k', cursor phải là 1, nhận được %d", m.modeSelect.cursor)
	}

	// Up arrow
	res, _ = m.handleModeSelectKey(tea.KeyMsg{Type: tea.KeyUp})
	m = res.(Model)
	if m.modeSelect.cursor != 0 {
		t.Fatalf("Sau khi ấn KeyUp, cursor phải là 0, nhận được %d", m.modeSelect.cursor)
	}

	// Up arrow wrap-around to end
	res, _ = m.handleModeSelectKey(tea.KeyMsg{Type: tea.KeyUp})
	m = res.(Model)
	if m.modeSelect.cursor != numOpts-1 {
		t.Fatalf("Sau khi ấn KeyUp từ 0, cursor phải cuộn vòng đến %d, nhận được %d", numOpts-1, m.modeSelect.cursor)
	}

	// Down arrow wrap-around back to 0
	res, _ = m.handleModeSelectKey(tea.KeyMsg{Type: tea.KeyDown})
	m = res.(Model)
	if m.modeSelect.cursor != 0 {
		t.Fatalf("Sau khi ấn KeyDown từ cuối, cursor phải quay về 0, nhận được %d", m.modeSelect.cursor)
	}
}

func TestModeSelect_SelectNovelManga(t *testing.T) {
	m := Model{
		textarea:   textarea.New(),
		width:      100,
		height:     30,
		modeSelect: newModeSelectState(nil, 100, 30),
	}
	m.modeSelect.cursor = 0 // Option 1: Tiểu thuyết / Manga

	res, _ := m.handleModeSelectKey(tea.KeyMsg{Type: tea.KeyEnter})
	m = res.(Model)

	if m.modeSelect != nil {
		t.Fatal("Sau khi chọn chế độ 1, modal phải đóng")
	}
	if !m.textarea.Focused() {
		t.Fatal("Textarea phải được focus sau khi chọn chế độ")
	}
}

func TestModeSelect_SelectDoodleExplainerQuickKey(t *testing.T) {
	m := Model{
		textarea:   textarea.New(),
		width:      100,
		height:     30,
		modeSelect: newModeSelectState(nil, 100, 30),
	}

	// Phím '2' chọn Doodle Explainer
	res, _ := m.handleModeSelectKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m = res.(Model)

	if m.modeSelect != nil {
		t.Fatal("Sau khi ấn '2', modal phải đóng")
	}
	if !m.textarea.Focused() {
		t.Fatal("Textarea phải được focus sau khi chọn")
	}
}

func TestModeSelect_SelectComingSoonShowsWarning(t *testing.T) {
	m := Model{
		textarea:   textarea.New(),
		width:      100,
		height:     30,
		modeSelect: newModeSelectState(nil, 100, 30),
	}

	// Chọn phím 3 (chưa có)
	res, _ := m.handleModeSelectKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = res.(Model)

	if m.modeSelect == nil {
		t.Fatal("Khi chọn chế độ chưa sẵn sàng, modal không được đóng")
	}
	if m.modeSelect.message == "" {
		t.Fatal("Phải hiển thị cảnh báo giải thích chế độ 3 đang được nghiên cứu")
	}
	if !strings.Contains(m.modeSelect.message, "bước tiếp theo") {
		t.Errorf("Nội dung thông báo không khớp mong đợi: %s", m.modeSelect.message)
	}
}

func TestModeSelect_DismissEscAndQ(t *testing.T) {
	t.Run("Dismiss with Esc", func(t *testing.T) {
		m := Model{
			textarea:   textarea.New(),
			width:      100,
			height:     30,
			modeSelect: newModeSelectState(nil, 100, 30),
		}

		res, _ := m.handleModeSelectKey(tea.KeyMsg{Type: tea.KeyEsc})
		m = res.(Model)

		if m.modeSelect != nil {
			t.Fatal("Sau khi ấn Esc, modeSelect modal phải đóng")
		}
	})

	t.Run("Dismiss with q", func(t *testing.T) {
		m := Model{
			textarea:   textarea.New(),
			width:      100,
			height:     30,
			modeSelect: newModeSelectState(nil, 100, 30),
		}

		res, _ := m.handleModeSelectKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		m = res.(Model)

		if m.modeSelect != nil {
			t.Fatal("Sau khi ấn 'q', modeSelect modal phải đóng")
		}
	})
}

func TestModeSelect_RenderModal(t *testing.T) {
	st := newModeSelectState(nil, 100, 30)
	view := renderModeSelectModal(100, 30, st)

	if view == "" {
		t.Fatal("renderModeSelectModal không được trả về rỗng khi state != nil")
	}
	if !strings.Contains(view, "CHỌN CHẾ ĐỘ LÀM VIỆC CỦA AI (/mode)") {
		t.Fatal("View phải chứa tiêu đề modal")
	}
	if !strings.Contains(view, "Tiểu thuyết / Manga") {
		t.Fatal("View phải chứa lựa chọn Tiểu thuyết / Manga")
	}
	if !strings.Contains(view, "Doodle Explainer") {
		t.Fatal("View phải chứa lựa chọn Doodle Explainer")
	}

	nilView := renderModeSelectModal(100, 30, nil)
	if nilView != "" {
		t.Fatalf("renderModeSelectModal(state=nil) phải trả về chuỗi rỗng, nhận được %q", nilView)
	}
}
