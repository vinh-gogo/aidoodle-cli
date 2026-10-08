package tui

import (
	"reflect"
	"testing"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewCommand_RegisteredAndAutoExecute(t *testing.T) {
	spec, ok := commandRegistryInstance().Find("new")
	if !ok {
		t.Fatal("Lệnh /new chưa được đăng ký trong commandRegistry")
	}
	if spec.Usage != "/new" {
		t.Fatalf("spec.Usage = %q, kỳ vọng /new", spec.Usage)
	}
	if !spec.AutoExecute {
		t.Fatal("spec.AutoExecute phải bằng true để nhấn Enter thực thi ngay phiên mới")
	}
}

func TestNewCommand_EnterExecutesImmediately(t *testing.T) {
	m := Model{
		textarea: textarea.New(),
	}
	m.textarea.Focus()
	m.textarea.SetValue("/new")
	m.updateCommandPalette()

	if !m.compActive {
		t.Fatal("Command palette phải active khi gõ /new")
	}
	item, ok := m.selectedCommandItem()
	if !ok || item.Name != "new" {
		t.Fatalf("Lệnh được chọn trong palette phải là new, nhận: %+v", item)
	}

	updated, cmd, handled := m.handleCommandPaletteKey(tea.KeyMsg{Type: tea.KeyEnter})
	if !handled {
		t.Fatal("handleCommandPaletteKey phải xử lý phím Enter")
	}

	nextModel, ok := updated.(Model)
	if !ok {
		t.Fatalf("Model trả về kiểu %T, kỳ vọng Model", updated)
	}

	if !nextModel.restartRequested {
		t.Fatal("restartRequested phải bằng true sau khi nhấn Enter trên /new")
	}
	if nextModel.restartPrompt != "" {
		t.Fatalf("restartPrompt phải rỗng khi chỉ gõ /new, nhận: %q", nextModel.restartPrompt)
	}
	if cmd == nil {
		t.Fatal("Cmd trả về không được nil (phải là tea.Quit)")
	}
	// Kiểm tra cmd có phải tea.Quit không
	if reflect.ValueOf(cmd).Pointer() != reflect.ValueOf(tea.Quit).Pointer() {
		t.Fatalf("Cmd phải là tea.Quit")
	}
}

func TestNewCommand_WithTopicArgument(t *testing.T) {
	spec, ok := commandRegistryInstance().Find("new")
	if !ok {
		t.Fatal("Lệnh /new chưa được đăng ký")
	}
	m := Model{}
	resModel, cmd := spec.Run(m, []string{"chủ", "đề", "mới"})
	updated := resModel.(Model)

	if !updated.restartRequested {
		t.Fatal("restartRequested phải bằng true")
	}
	if updated.restartPrompt != "chủ đề mới" {
		t.Fatalf("restartPrompt = %q, kỳ vọng 'chủ đề mới'", updated.restartPrompt)
	}
	if reflect.ValueOf(cmd).Pointer() != reflect.ValueOf(tea.Quit).Pointer() {
		t.Fatal("Cmd phải là tea.Quit")
	}
}
