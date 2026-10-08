package web

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/voocel/ainovel-cli/assets"
	"github.com/voocel/ainovel-cli/internal/bootstrap"
	"github.com/voocel/ainovel-cli/internal/entry/startup"
	"github.com/voocel/ainovel-cli/internal/host"
	"github.com/voocel/ainovel-cli/internal/host/exp"
	"github.com/voocel/ainovel-cli/internal/store"
	buildversion "github.com/voocel/ainovel-cli/internal/version"
)

// SSEEvent là cấu trúc bản tin phát qua Server-Sent Events.
type SSEEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Server quản lý HTTP Server và kết nối thời gian thực với host.Host.
type Server struct {
	mu        sync.RWMutex
	cfg       bootstrap.Config
	bundle    assets.Bundle
	port      int
	buildInfo buildversion.Info

	host       *host.Host
	hostCancel context.CancelFunc

	// SSE broadcasting
	clientsMu     sync.RWMutex
	clients       map[chan SSEEvent]struct{}
	recentEvents  []host.Event
	maxEvents     int
	streamingText strings.Builder
	streamingMu   sync.RWMutex

	httpServer *http.Server
}

// NewServer khởi tạo một phiên bản Web Server mới.
func NewServer(cfg bootstrap.Config, bundle assets.Bundle, port int, buildInfo buildversion.Info) (*Server, error) {
	if port <= 0 {
		port = 8080
	}

	s := &Server{
		cfg:          cfg,
		bundle:       bundle,
		port:         port,
		buildInfo:    buildInfo,
		clients:      make(map[chan SSEEvent]struct{}),
		recentEvents: make([]host.Event, 0, 200),
		maxEvents:    200,
	}

	// Thử khởi tạo Host với thư mục hiện tại
	h, err := host.New(cfg, bundle, host.WithFileLog("web.log", false,
		slog.String("version", buildInfo.Version),
		slog.String("commit", buildInfo.Commit),
		slog.String("built", buildInfo.Date),
	))
	if err == nil {
		s.attachHost(h)
	} else {
		slog.Warn("Khởi tạo host ban đầu với output dir hiện tại thất bại, sẽ khởi tạo khi bắt đầu dự án", "err", err, "dir", cfg.OutputDir)
	}

	return s, nil
}

// attachHost gắn host mới và kích hoạt vòng lặp lắng nghe sự kiện & streaming.
func (s *Server) attachHost(h *host.Host) {
	if s.hostCancel != nil {
		s.hostCancel()
		s.hostCancel = nil
	}

	s.host = h
	ctx, cancel := context.WithCancel(context.Background())
	s.hostCancel = cancel

	go s.consumeHost(ctx, h)
}

// consumeHost tiêu thụ sự kiện và delta từ Host và phát cho các client SSE.
func (s *Server) consumeHost(ctx context.Context, h *host.Host) {
	for {
		select {
		case ev, ok := <-h.Events():
			if !ok {
				return
			}
			s.recordEvent(ev)
			s.broadcast(SSEEvent{Type: "event", Data: ev})
		case delta, ok := <-h.Stream():
			if !ok {
				continue
			}
			if delta == host.StreamClearSentinel {
				s.clearStreaming()
				s.broadcast(SSEEvent{Type: "clear", Data: ""})
				continue
			}
			if delta != "" {
				s.appendStreaming(delta)
				s.broadcast(SSEEvent{Type: "delta", Data: delta})
			}
		case _, ok := <-h.Done():
			if !ok {
				return
			}
			s.broadcast(SSEEvent{Type: "done", Data: ""})
			snap := h.Snapshot()
			s.broadcast(SSEEvent{Type: "status", Data: snap})
		case <-ctx.Done():
			return
		}
	}
}

// recordEvent lưu sự kiện vào buffer lịch sử.
func (s *Server) recordEvent(ev host.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Nếu sự kiện có ID (lifecycle event), cập nhật bản ghi cũ nếu cùng ID
	if ev.ID != "" {
		for i := len(s.recentEvents) - 1; i >= 0; i-- {
			if s.recentEvents[i].ID == ev.ID {
				s.recentEvents[i] = ev
				return
			}
		}
	}

	if len(s.recentEvents) >= s.maxEvents {
		s.recentEvents = s.recentEvents[1:]
	}
	s.recentEvents = append(s.recentEvents, ev)
}

func (s *Server) appendStreaming(delta string) {
	s.streamingMu.Lock()
	defer s.streamingMu.Unlock()
	s.streamingText.WriteString(delta)
}

func (s *Server) clearStreaming() {
	s.streamingMu.Lock()
	defer s.streamingMu.Unlock()
	s.streamingText.Reset()
}

func (s *Server) getStreamingText() string {
	s.streamingMu.RLock()
	defer s.streamingMu.RUnlock()
	return s.streamingText.String()
}

// broadcast gửi bản tin SSE đến tất cả client đang kết nối.
func (s *Server) broadcast(ev SSEEvent) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	for ch := range s.clients {
		select {
		case ch <- ev:
		default:
			// client bị đầy buffer hoặc nghẽn mạng, bỏ qua tránh block
		}
	}
}

// addClient đăng ký một client SSE mới.
func (s *Server) addClient(ch chan SSEEvent) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	s.clients[ch] = struct{}{}
}

// removeClient hủy đăng ký client SSE.
func (s *Server) removeClient(ch chan SSEEvent) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()
	delete(s.clients, ch)
}

// StartProject khởi chạy một dự án sáng tác mới hoặc viết trên dự án hiện tại.
func (s *Server) StartProject(prompt, style string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cleanPrompt, err := startup.PrepareQuick(prompt)
	if err != nil {
		return err
	}

	// Nếu style chỉ định khác style hiện tại
	if style != "" && style != s.cfg.Style {
		s.cfg.Style = style
	}

	// Kiểm tra nếu cần tạo thư mục mới (nếu thư mục hiện tại đã có chương hoàn thành)
	needNewDir := false
	if s.host != nil {
		snap := s.host.Snapshot()
		if snap.CompletedCount > 0 {
			needNewDir = true
		}
	}

	if needNewDir || s.host == nil {
		if needNewDir {
			s.cfg.OutputDir = NextOutputDir(s.cfg.OutputDir, time.Now())
		}
		s.cfg.FillDefaults()
		s.bundle = assets.Load(s.cfg.Style, assets.DefaultLoadOptions(s.cfg.OutputDir))

		if s.host != nil {
			s.host.Close()
			s.host = nil
		}

		h, err := host.New(s.cfg, s.bundle, host.WithFileLog("web.log", false,
			slog.String("version", s.buildInfo.Version),
		))
		if err != nil {
			return fmt.Errorf("khởi tạo host mới thất bại: %w", err)
		}
		s.attachHost(h)
	}

	if err := s.host.PrepareUserRules(cleanPrompt); err != nil {
		return fmt.Errorf("chuẩn bị quy tắc thất bại: %w", err)
	}

	if err := s.host.StartPrepared(cleanPrompt); err != nil {
		return fmt.Errorf("khởi động sáng tác thất bại: %w", err)
	}

	s.clearStreaming()
	return nil
}

// SwitchProject chuyển sang một dự án khác trong thư mục output.
func (s *Server) SwitchProject(dir string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.host != nil {
		s.host.Close()
		s.host = nil
	}

	s.cfg.OutputDir = dir
	if targetStyle := DetectProjectStyle(dir); targetStyle != "" {
		s.cfg.Style = targetStyle
	}
	s.cfg.FillDefaults()
	s.bundle = assets.Load(s.cfg.Style, assets.DefaultLoadOptions(s.cfg.OutputDir))

	h, err := host.New(s.cfg, s.bundle, host.WithFileLog("web.log", false,
		slog.String("version", s.buildInfo.Version),
	))
	if err != nil {
		slog.Warn("Mở host cho dự án thất bại, dự án có thể đã hoàn thành", "dir", dir, "err", err)
	} else {
		s.attachHost(h)
	}
	s.clearStreaming()

	// Phát lại replay log gần đây nếu có
	if h != nil {
		if items, err := h.ReplayQueue(20); err == nil {
			for _, it := range items {
				s.recordEvent(host.Event{
					Time:     it.Time,
					Category: it.Category,
					Summary:  it.Summary,
					Level:    "info",
				})
			}
		}
	}

	return nil
}

// NewProject tạo dự án mới hoàn toàn trong thư mục output timestamped mới.
func (s *Server) NewProject(prompt, style string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.host != nil {
		s.host.Close()
		s.host = nil
	}

	s.cfg.OutputDir = NextOutputDir(s.cfg.OutputDir, time.Now())
	if style != "" {
		s.cfg.Style = style
	}
	s.cfg.FillDefaults()
	s.bundle = assets.Load(s.cfg.Style, assets.DefaultLoadOptions(s.cfg.OutputDir))

	h, err := host.New(s.cfg, s.bundle, host.WithFileLog("web.log", false,
		slog.String("version", s.buildInfo.Version),
	))
	if err != nil {
		return fmt.Errorf("tạo dự án mới thất bại: %w", err)
	}
	s.attachHost(h)
	s.clearStreaming()

	if strings.TrimSpace(prompt) != "" {
		cleanPrompt, err := startup.PrepareQuick(prompt)
		if err != nil {
			return err
		}
		if err := h.PrepareUserRules(cleanPrompt); err != nil {
			return err
		}
		if err := h.StartPrepared(cleanPrompt); err != nil {
			return err
		}
	}

	return nil
}

// AdvanceOneChapter tiến hành viết chương tiếp theo (chế độ review).
func (s *Server) AdvanceOneChapter() error {
	s.mu.RLock()
	h := s.host
	s.mu.RUnlock()

	if h == nil {
		return fmt.Errorf("chưa có phiên sáng tác nào đang hoạt động")
	}
	return h.AdvanceOneChapter()
}

// Steer gửi chỉ đạo hoặc can thiệp của người dùng.
func (s *Server) Steer(message string) error {
	s.mu.RLock()
	h := s.host
	s.mu.RUnlock()

	if h == nil {
		return fmt.Errorf("chưa có phiên sáng tác nào đang hoạt động")
	}

	snap := h.Snapshot()
	if snap.IsRunning {
		return h.Steer(message)
	}
	return h.Continue(message)
}

// Pause tạm dừng sáng tác.
func (s *Server) Pause() error {
	s.mu.RLock()
	h := s.host
	s.mu.RUnlock()

	if h == nil {
		return fmt.Errorf("chưa có phiên sáng tác nào đang hoạt động")
	}
	_ = h.Abort()
	return nil
}

// Resume tiếp tục sáng tác.
func (s *Server) Resume() error {
	s.mu.RLock()
	h := s.host
	s.mu.RUnlock()

	if h == nil {
		return fmt.Errorf("chưa có phiên sáng tác nào đang hoạt động")
	}
	_, err := h.Resume()
	return err
}

// SetStyle thay đổi chế độ AI làm việc.
func (s *Server) SetStyle(style string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.bundle.Styles[style]; !ok {
		return fmt.Errorf("chế độ %q không tồn tại", style)
	}
	s.cfg.Style = style
	s.bundle = assets.Load(s.cfg.Style, assets.DefaultLoadOptions(s.cfg.OutputDir))
	return nil
}

// Export xuất bản thảo sang định dạng chỉ định (word, docx, txt, epub, video).
func (s *Server) Export(format, dir string) (*exp.Result, error) {
	s.mu.RLock()
	if dir == "" {
		dir = s.cfg.OutputDir
	}
	s.mu.RUnlock()

	st := store.NewStore(dir)
	f := exp.Format(strings.ToLower(format))
	if f == "" || f == "word" || f == "docx" {
		f = exp.FormatWord
	}

	res, err := exp.Run(context.Background(), exp.Deps{Store: st}, exp.Options{
		Format:    f,
		Overwrite: true,
	})
	if err != nil {
		return nil, err
	}

	// Phát log sự kiện ra web
	ev := host.Event{
		Time:     time.Now(),
		Category: "SYSTEM",
		Summary:  fmt.Sprintf("✓ Đã xuất %s (%d chương) đến %s", string(f), res.Chapters, res.Path),
		Level:    "info",
	}
	s.recordEvent(ev)
	s.broadcast(SSEEvent{
		Type: "event",
		Data: ev,
	})

	return res, nil
}

// Close giải phóng tài nguyên server và host.
func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.hostCancel != nil {
		s.hostCancel()
	}
	if s.host != nil {
		s.host.Close()
		s.host = nil
	}
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		s.httpServer.Shutdown(ctx)
	}
	return nil
}

// Run bắt đầu HTTP server và lắng nghe trên port được cấu hình.
func Run(cfg bootstrap.Config, bundle assets.Bundle, port int, buildInfo buildversion.Info) error {
	srv, err := NewServer(cfg, bundle, port, buildInfo)
	if err != nil {
		return err
	}
	defer srv.Close()

	mux := http.NewServeMux()
	srv.registerRoutes(mux)

	addr := fmt.Sprintf(":%d", srv.port)
	srv.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	fmt.Println()
	fmt.Println("==================================================================")
	fmt.Printf(" 🚀 aiNovel Studio Web Dashboard đang chạy tại:\n")
	fmt.Printf("    👉 http://localhost:%d\n", srv.port)
	fmt.Println("==================================================================")
	fmt.Println(" Nhấn Ctrl+C trong terminal này để tắt máy chủ Web.")
	fmt.Println()

	return srv.httpServer.ListenAndServe()
}
