package tui

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/voocel/ainovel-cli/assets"
	"github.com/voocel/ainovel-cli/internal/bootstrap"
	"github.com/voocel/ainovel-cli/internal/host"
	buildversion "github.com/voocel/ainovel-cli/internal/version"
)

var timestampSuffixRe = regexp.MustCompile(`^(.+)-\d{8}-\d{4}(?:-\d+)?$`)

// nextOutputDir tạo đường dẫn thư mục mới theo kiểu output/novel-YYYYMMDD-HHMM
// dựa trên thư mục output hiện tại mà không xóa bỏ dữ liệu dự án cũ.
func nextOutputDir(currentDir string, now time.Time) string {
	if currentDir == "" {
		currentDir = filepath.Join("output", "novel")
	}
	parent := filepath.Dir(currentDir)
	base := filepath.Base(currentDir)

	prefix := base
	if m := timestampSuffixRe.FindStringSubmatch(base); len(m) > 1 {
		prefix = m[1]
	}

	ts := now.Format("20060102-1504")
	candidate := filepath.Join(parent, fmt.Sprintf("%s-%s", prefix, ts))
	if _, err := os.Stat(candidate); os.IsNotExist(err) {
		return candidate
	}

	for i := 1; i <= 99; i++ {
		alt := filepath.Join(parent, fmt.Sprintf("%s-%s-%02d", prefix, ts, i))
		if _, err := os.Stat(alt); os.IsNotExist(err) {
			return alt
		}
	}
	return filepath.Join(parent, fmt.Sprintf("%s-%s-%s", prefix, ts, now.Format("05")))
}

func wipeDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if entry.Name() == ".ainovel.lock" {
			continue
		}
		err = os.RemoveAll(filepath.Join(dir, entry.Name()))
		if err != nil {
			return err
		}
	}
	return nil
}

// Run 启动 TUI。
// 启动模式分层约定：
// 1. 快速模式、共创模式属于“启动编排”；
// 2. 正式创作会话进入 host.Host；
// 3. 未来若新增“续写已有小说”等共享模式，统一落到 internal/entry/startup。
func Run(cfg bootstrap.Config, bundle assets.Bundle, build buildversion.Info) error {
	var initialPrompt string
	var newProjectDirHint string
	for {
		rt, err := host.New(cfg, bundle, host.WithFileLog("tui.log", false,
			slog.String("version", build.Version),
			slog.String("commit", build.Commit),
			slog.String("built", build.Date),
		))
		if err != nil {
			return err
		}

		m := NewModel(rt, build.Version, initialPrompt)
		m.disableUpdateCheck = cfg.DisableUpdateCheck
		if newProjectDirHint != "" {
			m.applyEvent(host.Event{
				Time:     time.Now(),
				Category: "SYSTEM",
				Level:    "info",
				Summary:  fmt.Sprintf("Đã khởi tạo dự án mới tại: %s", newProjectDirHint),
			})
			newProjectDirHint = ""
		}
		if logErr := rt.FileLogError(); logErr != nil {
			logWarning := fmt.Errorf("nhật ký tệp không khả dụng, tiếp tục dùng nhật ký terminal: %w", logErr)
			m.err = logWarning
			m.applyEvent(host.Event{
				Time: time.Now(), Category: "SYSTEM", Level: "warn",
				Summary: logWarning.Error(), Detail: logWarning.Error(),
			})
		}
		// 不在启动时全局开启鼠标上报：欢迎页用不到鼠标，关闭上报可保留终端原生
		// 拖拽选中复制。进入创作工作台（modeRunning）时再由 enterRunning 打开上报，
		// 以支持点击切面板 / 滚轮 / 拖拽侧边栏。
		p := tea.NewProgram(m, tea.WithAltScreen())
		finalModel, err := p.Run()
		rt.Close()

		if err != nil {
			return err
		}

		if m, ok := finalModel.(Model); ok && m.restartRequested {
			cfg.OutputDir = nextOutputDir(cfg.OutputDir, time.Now())
			if rt != nil && rt.Style() != "" {
				cfg.Style = rt.Style()
			}
			bundle = assets.Load(cfg.Style, assets.DefaultLoadOptions(cfg.OutputDir))
			initialPrompt = m.restartPrompt
			newProjectDirHint = cfg.OutputDir
			continue
		}
		break
	}
	return nil
}
