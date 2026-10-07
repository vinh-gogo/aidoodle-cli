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
	if len(updated.modeSelect.options) == 0 {
		t.Fatal("modeSelect phải nạp ít nhất một tùy chọn chủ đề")
	}
	if updated.modeSelect.cursor != 0 {
		t.Fatalf("cursor ban đầu phải là 0, nhận được %d", updated.modeSelect.cursor)
	}
	if updated.textarea.Focused() {
		t.Fatal("Textarea phải bị blur khi mở modal lựa chọn chủ đề")
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
	if numOpts < 2 {
		t.Fatalf("Cần ít nhất 2 tùy chọn để kiểm tra điều hướng, có %d", numOpts)
	}

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

func TestModeSelect_SelectEnter(t *testing.T) {
	m := Model{
		textarea:   textarea.New(),
		width:      100,
		height:     30,
		modeSelect: newModeSelectState(nil, 100, 30),
	}
	expectedPrompt := m.modeSelect.options[0].Prompt

	res, _ := m.handleModeSelectKey(tea.KeyMsg{Type: tea.KeyEnter})
	m = res.(Model)

	if m.modeSelect != nil {
		t.Fatal("Sau khi ấn Enter, modeSelect modal phải được đóng")
	}
	if m.textarea.Value() != expectedPrompt {
		t.Fatalf("Textarea value không đúng.\nNhận được: %q\nKỳ vọng: %q", m.textarea.Value(), expectedPrompt)
	}
	if !m.textarea.Focused() {
		t.Fatal("Textarea phải được focus sau khi chọn chủ đề")
	}
}

func TestModeSelect_QuickSelectNumber(t *testing.T) {
	m := Model{
		textarea:   textarea.New(),
		width:      100,
		height:     30,
		modeSelect: newModeSelectState(nil, 100, 30),
	}
	expectedPrompt := m.modeSelect.options[1].Prompt // Phím '2' chọn phần tử thứ 2 (index 1)

	res, _ := m.handleModeSelectKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m = res.(Model)

	if m.modeSelect != nil {
		t.Fatal("Sau khi ấn '2', modeSelect modal phải được đóng")
	}
	if m.textarea.Value() != expectedPrompt {
		t.Fatalf("Textarea value không đúng khi chọn nhanh bằng phím 2.\nNhận được: %q\nKỳ vọng: %q", m.textarea.Value(), expectedPrompt)
	}
	if !m.textarea.Focused() {
		t.Fatal("Textarea phải được focus sau khi chọn số")
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
		m.textarea.SetValue("văn bản gốc")

		res, _ := m.handleModeSelectKey(tea.KeyMsg{Type: tea.KeyEsc})
		m = res.(Model)

		if m.modeSelect != nil {
			t.Fatal("Sau khi ấn Esc, modeSelect modal phải đóng")
		}
		if m.textarea.Value() != "văn bản gốc" {
			t.Fatalf("Textarea không được thay đổi khi hủy bằng Esc, nhận được: %q", m.textarea.Value())
		}
	})

	t.Run("Dismiss with q", func(t *testing.T) {
		m := Model{
			textarea:   textarea.New(),
			width:      100,
			height:     30,
			modeSelect: newModeSelectState(nil, 100, 30),
		}
		m.textarea.SetValue("văn bản gốc")

		res, _ := m.handleModeSelectKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		m = res.(Model)

		if m.modeSelect != nil {
			t.Fatal("Sau khi ấn 'q', modeSelect modal phải đóng")
		}
		if m.textarea.Value() != "văn bản gốc" {
			t.Fatalf("Textarea không được thay đổi khi hủy bằng 'q', nhận được: %q", m.textarea.Value())
		}
	})
}

func TestModeSelect_RenderModal(t *testing.T) {
	st := newModeSelectState(nil, 100, 30)
	view := renderModeSelectModal(100, 30, st)

	if view == "" {
		t.Fatal("renderModeSelectModal không được trả về rỗng khi state != nil")
	}
	if !strings.Contains(view, "Lựa chọn chủ đề kịch bản (/mode)") {
		t.Fatal("View phải chứa tiêu đề modal")
	}
	if !strings.Contains(view, "Vì sao con người mất gần hết lông?") {
		t.Fatal("View phải chứa tên chủ đề mẫu")
	}

	nilView := renderModeSelectModal(100, 30, nil)
	if nilView != "" {
		t.Fatalf("renderModeSelectModal(state=nil) phải trả về chuỗi rỗng, nhận được %q", nilView)
	}
}
