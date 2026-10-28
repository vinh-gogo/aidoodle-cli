package web

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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

// Options chứa các cấu hình khởi tạo máy chủ Web Dashboard.
type Options struct {
	Host        string
	Port        int
	AuthToken   string
	Token       string // Alias for AuthToken
	BaseDir     string
	AllowedHost []string
}

// Server quản lý HTTP Server và kết nối thời gian thực với host.Host.
type Server struct {
	mu           sync.RWMutex
	cfg          bootstrap.Config
	bundle       assets.Bundle
	hostAddr     string
	port         int
	authToken    string
	basePath     string
	allowedHosts []string
	buildInfo    buildversion.Info

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

// HostAddr trả về địa chỉ IP hoặc hostname mà server đang lắng nghe.
func (s *Server) HostAddr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hostAddr
}

// AuthToken trả về token xác thực đã cấu hình (nếu có).
func (s *Server) AuthToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.authToken
}

// NewServerWithOptions khởi tạo một phiên bản Web Server với các tùy chọn bảo mật.
func NewServerWithOptions(cfg bootstrap.Config, bundle assets.Bundle, opts Options, buildInfo buildversion.Info) (*Server, error) {
	port := opts.Port
	if port <= 0 {
		port = 8080
	}
	hostAddr := strings.TrimSpace(opts.Host)
	if hostAddr == "" {
		hostAddr = "127.0.0.1"
	}
	baseDir := strings.TrimSpace(opts.BaseDir)
	if baseDir == "" {
		if cfg.OutputDir != "" {
			parent := filepath.Dir(cfg.OutputDir)
			if parent != "" && parent != "." {
				baseDir = parent
			}
		}
		if baseDir == "" {
			baseDir = "output"
		}
	}

	tok := strings.TrimSpace(opts.AuthToken)
	if tok == "" {
		tok = strings.TrimSpace(opts.Token)
	}

	s := &Server{
		cfg:          cfg,
		bundle:       bundle,
		hostAddr:     hostAddr,
		port:         port,
		authToken:    tok,
		basePath:     baseDir,
		allowedHosts: opts.AllowedHost,
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

// NewServer khởi tạo một phiên bản Web Server mới (tương thích ngược).
func NewServer(cfg bootstrap.Config, bundle assets.Bundle, port int, buildInfo buildversion.Info) (*Server, error) {
	return NewServerWithOptions(cfg, bundle, Options{Port: port}, buildInfo)
}

func (s *Server) baseDir() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.basePath != "" {
		return s.basePath
	}
	return "output"
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
	s.recordEventLocked(ev)
}

func (s *Server) recordEventLocked(ev host.Event) {
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
	safeDir, err := SafeProjectDir(s.basePath, dir)
	if err != nil {
		return fmt.Errorf("không thể chuyển dự án: %w", err)
	}

	fi, err := os.Stat(safeDir)
	if err != nil {
		return fmt.Errorf("thư mục dự án không tồn tại: %w", err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("đường dẫn không phải là thư mục: %s", safeDir)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.host != nil {
		s.host.Close()
		s.host = nil
	}

	s.cfg.OutputDir = safeDir
	if targetStyle := DetectProjectStyle(safeDir); targetStyle != "" {
		s.cfg.Style = targetStyle
	}
	s.cfg.FillDefaults()
	s.bundle = assets.Load(s.cfg.Style, assets.DefaultLoadOptions(s.cfg.OutputDir))

	h, err := host.New(s.cfg, s.bundle, host.WithFileLog("web.log", false,
		slog.String("version", s.buildInfo.Version),
	))
	if err != nil {
		slog.Warn("Mở host cho dự án thất bại", "dir", safeDir, "err", err)
		return fmt.Errorf("mở host cho dự án thất bại: %w", err)
	}
	s.attachHost(h)
	s.clearStreaming()

	// Phát lại replay log gần đây nếu có
	if h != nil {
		if items, err := h.ReplayQueue(20); err == nil {
			for _, it := range items {
				s.recordEventLocked(host.Event{
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
	base := s.basePath
	s.mu.RUnlock()

	safeDir, err := SafeProjectDir(base, dir)
	if err != nil {
		return nil, err
	}
	dir = safeDir

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

// isOriginAllowed kiểm tra Origin có nằm trong danh sách được phép không.
func (s *Server) isOriginAllowed(origin string) bool {
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	hostname := strings.ToLower(u.Hostname())
	if hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" || hostname == "[::1]" {
		return true
	}
	if s.hostAddr != "" && strings.EqualFold(hostname, s.hostAddr) {
		return true
	}
	for _, h := range s.allowedHosts {
		if strings.EqualFold(hostname, h) {
			return true
		}
	}
	return false
}

// securityMiddleware bảo vệ token auth, CSRF, Origin và custom header.
func (s *Server) securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Kiểm tra xác thực Bearer token nếu đã cấu hình authToken (SEC-01)
		if s.authToken != "" {
			authenticated := false

			// Check Authorization: Bearer <token>
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				token := strings.TrimSpace(authHeader[7:])
				if token == s.authToken {
					authenticated = true
				}
			}

			// Check X-Auth-Token: <token>
			if !authenticated && r.Header.Get("X-Auth-Token") == s.authToken {
				authenticated = true
			}

			// Check Cookie: ainovel_token
			if !authenticated {
				if cookie, err := r.Cookie("ainovel_token"); err == nil && cookie.Value == s.authToken {
					authenticated = true
				}
			}

			// Check Query param: ?token=<token>
			if !authenticated {
				if qToken := r.URL.Query().Get("token"); qToken == s.authToken {
					authenticated = true
					http.SetCookie(w, &http.Cookie{
						Name:     "ainovel_token",
						Value:    s.authToken,
						Path:     "/",
						HttpOnly: true,
						SameSite: http.SameSiteStrictMode,
					})
				}
			}

			if !authenticated {
				writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
					"ok":    false,
					"error": "Yêu cầu mã xác thực truy cập (Bearer token)",
				})
				return
			}
		}

		// 2. CSRF & Origin checks cho state-mutating requests (SEC-02)
		isStateMutating := r.Method == http.MethodPost || (r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/export") && r.URL.Query().Get("download") != "1")

		if isStateMutating {
			// Sec-Fetch-Site check
			if secSite := r.Header.Get("Sec-Fetch-Site"); strings.EqualFold(secSite, "cross-site") {
				writeJSON(w, http.StatusForbidden, map[string]interface{}{
					"ok":    false,
					"error": "Yêu cầu cross-site bị từ chối",
				})
				return
			}

			// Origin check
			if origin := r.Header.Get("Origin"); origin != "" {
				if !s.isOriginAllowed(origin) {
					writeJSON(w, http.StatusForbidden, map[string]interface{}{
						"ok":    false,
						"error": "Origin không được phép",
					})
					return
				}
			}

			// Referer check if Origin is empty
			if r.Header.Get("Origin") == "" {
				if ref := r.Header.Get("Referer"); ref != "" {
					if u, err := url.Parse(ref); err == nil && u.Hostname() != "" {
						if !s.isOriginAllowed(u.Scheme + "://" + u.Host) {
							writeJSON(w, http.StatusForbidden, map[string]interface{}{
								"ok":    false,
								"error": "Referer không được phép",
							})
							return
						}
					}
				}
			}

			// Require custom header X-Requested-With: ainovel-web on AJAX POST requests
			if r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/api/") {
				if r.Header.Get("X-Requested-With") != "ainovel-web" {
					writeJSON(w, http.StatusForbidden, map[string]interface{}{
						"ok":    false,
						"error": "Yêu cầu thiếu header bảo mật X-Requested-With: ainovel-web",
					})
					return
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// Handler trả về http.Handler đã gắn toàn bộ routes và security middleware.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	s.registerRoutes(mux)
	return s.securityMiddleware(mux)
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

// RunWithOptions bắt đầu HTTP server với các tùy chọn bảo mật.
func RunWithOptions(cfg bootstrap.Config, bundle assets.Bundle, opts Options, buildInfo buildversion.Info) error {
	srv, err := NewServerWithOptions(cfg, bundle, opts, buildInfo)
	if err != nil {
		return err
	}
	defer srv.Close()

	host := srv.hostAddr
	if host == "" {
		host = "127.0.0.1"
	}
	addr := fmt.Sprintf("%s:%d", host, srv.port)
	srv.httpServer = &http.Server{
		Addr:    addr,
		Handler: srv.Handler(),
	}

	fmt.Println()
	fmt.Println("==================================================================")
	fmt.Printf(" 🚀 aiNovel Studio Web Dashboard đang chạy tại:\n")
	if srv.authToken != "" {
		fmt.Printf("    👉 http://%s:%d/?token=%s\n", host, srv.port, srv.authToken)
		fmt.Printf("    🔑 Token: %s\n", srv.authToken)
	} else {
		fmt.Printf("    👉 http://%s:%d\n", host, srv.port)
	}
	fmt.Println("==================================================================")
	fmt.Println(" Nhấn Ctrl+C trong terminal này để tắt máy chủ Web.")
	fmt.Println()

	return srv.httpServer.ListenAndServe()
}

// Run bắt đầu HTTP server và lắng nghe trên host mặc định (127.0.0.1).
func Run(cfg bootstrap.Config, bundle assets.Bundle, port int, buildInfo buildversion.Info) error {
	return RunWithOptions(cfg, bundle, Options{Port: port}, buildInfo)
}
