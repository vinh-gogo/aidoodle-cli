package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/host"
	"github.com/voocel/ainovel-cli/internal/store"
)

func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/stream", s.handleStream)
	mux.HandleFunc("/api/chapters", s.handleChapters)
	mux.HandleFunc("/api/chapter", s.handleChapter)
	mux.HandleFunc("/api/projects", s.handleProjects)
	mux.HandleFunc("/api/export", s.handleExport)
	mux.HandleFunc("/api/action", s.handleAction)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		http.NotFound(w, r)
		return
	}

	data, err := IndexHTML()
	if err != nil {
		http.Error(w, "Không thể tải giao diện: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	dir := s.cfg.OutputDir
	style := s.cfg.Style
	styles := s.bundle.Styles

	var snap *host.UISnapshot
	if s.host != nil {
		sp := s.host.Snapshot()
		snap = &sp
	} else {
		snap = loadStoreSnapshot(dir)
	}

	recent := make([]host.Event, len(s.recentEvents))
	copy(recent, s.recentEvents)
	s.mu.RUnlock()

	streaming := s.getStreamingText()
	projects := ScanProjects("", dir)

	resp := map[string]interface{}{
		"ok":             true,
		"dir":            dir,
		"style":          style,
		"styles":         styles,
		"snapshot":       snap,
		"recent_events":  recent,
		"streaming_text": streaming,
		"projects":       projects,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Máy chủ không hỗ trợ streaming", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Chỉ cho phép origin được ủy quyền (loại bỏ wildcard "*")
	origin := r.Header.Get("Origin")
	if origin != "" && s.isOriginAllowed(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
	}

	// Đăng ký client
	ch := make(chan SSEEvent, 64)
	s.addClient(ch)
	defer s.removeClient(ch)

	// Gửi snapshot ban đầu
	s.mu.RLock()
	if s.host != nil {
		snap := s.host.Snapshot()
		snapBytes, _ := json.Marshal(snap)
		fmt.Fprintf(w, "event: status\ndata: %s\n\n", string(snapBytes))
		flusher.Flush()
	}
	s.mu.RUnlock()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case ev := <-ch:
			dataBytes, err := json.Marshal(ev.Data)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Type, string(dataBytes))
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprintf(w, "event: ping\ndata: {}\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) handleChapters(w http.ResponseWriter, r *http.Request) {
	dir := r.URL.Query().Get("dir")
	if dir == "" {
		s.mu.RLock()
		dir = s.cfg.OutputDir
		s.mu.RUnlock()
	}

	safeDir, err := SafeProjectDir(s.baseDir(), dir)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"ok":    false,
			"error": "Đường dẫn không hợp lệ: " + err.Error(),
		})
		return
	}

	chapters, err := ListProjectChapters(safeDir)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, chapters)
}

func (s *Server) handleChapter(w http.ResponseWriter, r *http.Request) {
	chStr := r.URL.Query().Get("ch")
	ch, err := strconv.Atoi(chStr)
	if err != nil || ch <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"ok":    false,
			"error": "Tham số ch không hợp lệ",
		})
		return
	}

	dir := r.URL.Query().Get("dir")
	if dir == "" {
		s.mu.RLock()
		dir = s.cfg.OutputDir
		s.mu.RUnlock()
	}

	safeDir, err := SafeProjectDir(s.baseDir(), dir)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"ok":    false,
			"error": "Đường dẫn không hợp lệ: " + err.Error(),
		})
		return
	}

	content, err := ReadProjectChapter(safeDir, ch)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]interface{}{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"chapter": ch,
		"content": content,
	})
}

func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	currentDir := s.cfg.OutputDir
	s.mu.RUnlock()

	projects := ScanProjects("", currentDir)
	writeJSON(w, http.StatusOK, projects)
}

type actionRequest struct {
	Action  string `json:"action"`
	Prompt  string `json:"prompt"`
	Style   string `json:"style"`
	Dir     string `json:"dir"`
	Message string `json:"message"`
	Format  string `json:"format"`
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "word"
	}
	dir := r.URL.Query().Get("dir")
	if dir == "" {
		s.mu.RLock()
		dir = s.cfg.OutputDir
		s.mu.RUnlock()
	}

	safeDir, err := SafeProjectDir(s.baseDir(), dir)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"ok":    false,
			"error": "Đường dẫn không hợp lệ: " + err.Error(),
		})
		return
	}

	res, err := s.Export(format, safeDir)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"ok":    false,
			"error": "Xuất file thất bại: " + err.Error(),
		})
		return
	}

	// Nếu client yêu cầu tải trực tiếp file (download=1)
	if r.URL.Query().Get("download") == "1" {
		cleanResPath := filepath.Clean(res.Path)
		rel, relErr := filepath.Rel(safeDir, cleanResPath)
		if relErr != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			writeJSON(w, http.StatusForbidden, map[string]interface{}{
				"ok":    false,
				"error": "Truy cập file tải về bị từ chối",
			})
			return
		}

		fileName := filepath.Base(res.Path)
		contentType := "application/octet-stream"
		switch strings.ToLower(format) {
		case "word", "docx":
			contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		case "epub":
			contentType = "application/epub+zip"
		case "txt":
			contentType = "text/plain; charset=utf-8"
		}
		data, readErr := os.ReadFile(res.Path)
		if readErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"ok":    false,
				"error": "Đọc file đã xuất thất bại: " + readErr.Error(),
			})
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		asciiFallback := sanitizeASCIIFileName(fileName)
		encodedFileName := url.PathEscape(fileName)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q; filename*=UTF-8''%s", asciiFallback, encodedFileName))
		w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}

	fileName := filepath.Base(res.Path)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":           true,
		"filename":     fileName,
		"path":         res.Path,
		"chapters":     res.Chapters,
		"bytes":        res.Bytes,
		"download_url": fmt.Sprintf("/api/export?format=%s&dir=%s&download=1", format, url.QueryEscape(safeDir)),
	})
}

func (s *Server) handleAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Chỉ chấp nhận phương thức POST", http.StatusMethodNotAllowed)
		return
	}

	// Yêu cầu custom header X-Requested-With: ainovel-web (SEC-02-04)
	if r.Header.Get("X-Requested-With") != "ainovel-web" {
		writeJSON(w, http.StatusForbidden, map[string]interface{}{
			"ok":    false,
			"error": "Yêu cầu thiếu header bảo mật X-Requested-With: ainovel-web",
		})
		return
	}

	// Kiểm tra Sec-Fetch-Site (SEC-02-03)
	if secSite := r.Header.Get("Sec-Fetch-Site"); strings.EqualFold(secSite, "cross-site") {
		writeJSON(w, http.StatusForbidden, map[string]interface{}{
			"ok":    false,
			"error": "Yêu cầu cross-site bị từ chối",
		})
		return
	}

	// Kiểm tra Origin (SEC-02-02)
	if origin := r.Header.Get("Origin"); origin != "" && !s.isOriginAllowed(origin) {
		writeJSON(w, http.StatusForbidden, map[string]interface{}{
			"ok":    false,
			"error": "Origin không được phép",
		})
		return
	}

	var req actionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"ok":    false,
			"error": "Dữ liệu JSON không hợp lệ: " + err.Error(),
		})
		return
	}

	var actionErr error
	switch req.Action {
	case "start":
		actionErr = s.StartProject(req.Prompt, req.Style)
	case "steer":
		actionErr = s.Steer(req.Message)
	case "next":
		actionErr = s.AdvanceOneChapter()
	case "pause":
		actionErr = s.Pause()
	case "resume":
		actionErr = s.Resume()
	case "new":
		actionErr = s.NewProject(req.Prompt, req.Style)
	case "switch":
		safeDir, err := SafeProjectDir(s.baseDir(), req.Dir)
		if err != nil {
			actionErr = fmt.Errorf("đường dẫn dự án không hợp lệ: %w", err)
		} else {
			actionErr = s.SwitchProject(safeDir)
		}
	case "set_style":
		actionErr = s.SetStyle(req.Style)
	case "export":
		safeDir := req.Dir
		if safeDir == "" {
			s.mu.RLock()
			safeDir = s.cfg.OutputDir
			s.mu.RUnlock()
		}
		safeDir, err := SafeProjectDir(s.baseDir(), safeDir)
		if err != nil {
			actionErr = fmt.Errorf("đường dẫn xuất file không hợp lệ: %w", err)
		} else {
			_, actionErr = s.Export(req.Format, safeDir)
		}
	default:
		actionErr = fmt.Errorf("hành động không xác định: %s", req.Action)
	}

	if actionErr != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"ok":    false,
			"error": actionErr.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "Thao tác thành công",
	})
}

func writeJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(data)
}

func sanitizeASCIIFileName(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 32 && r < 127 && r != '"' && r != '\\' && r != ';' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	res := strings.TrimSpace(b.String())
	if res == "" || res == ".docx" || res == ".epub" || res == ".txt" {
		return "novel" + filepath.Ext(s)
	}
	return res
}

func loadStoreSnapshot(dir string) *host.UISnapshot {
	st := store.NewStore(dir)
	book, _ := st.Book.Load()
	progress, _ := st.Progress.Load()
	outline, _ := st.Outline.LoadOutline()

	title := ""
	if book != nil {
		title = book.Title
	}
	if title == "" {
		title = filepath.Base(dir)
	}

	snap := &host.UISnapshot{
		BookTitle:    title,
		RuntimeState: "idle",
		StatusLabel:  "READY",
	}

	if outline != nil {
		snap.TotalChapters = len(outline)
	}

	if progress != nil {
		snap.Phase = string(progress.Phase)
		snap.CompletedCount = len(progress.CompletedChapters)
		snap.TotalWordCount = progress.TotalWordCount
		snap.CurrentChapter = progress.CurrentChapter
		if progress.TotalChapters > 0 {
			snap.TotalChapters = progress.TotalChapters
		}
		if progress.Phase == domain.PhaseComplete || (snap.TotalChapters > 0 && snap.CompletedCount >= snap.TotalChapters) {
			snap.StatusLabel = "COMPLETE"
			snap.RuntimeState = "completed"
		}
	}

	return snap
}
