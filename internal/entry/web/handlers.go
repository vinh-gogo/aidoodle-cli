package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/voocel/ainovel-cli/internal/host"
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
	w.Header().Set("Access-Control-Allow-Origin", "*")

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

	chapters, err := ListProjectChapters(dir)
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

	content, err := ReadProjectChapter(dir, ch)
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

	res, err := s.Export(format, dir)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"ok":    false,
			"error": "Xuất file thất bại: " + err.Error(),
		})
		return
	}

	// Nếu client yêu cầu tải trực tiếp file (download=1 hoặc không yêu cầu application/json)
	if r.URL.Query().Get("download") == "1" || !strings.Contains(r.Header.Get("Accept"), "application/json") {
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
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName))
		http.ServeFile(w, r, res.Path)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":           true,
		"path":         res.Path,
		"chapters":     res.Chapters,
		"bytes":        res.Bytes,
		"download_url": fmt.Sprintf("/api/export?format=%s&dir=%s&download=1", format, dir),
	})
}

func (s *Server) handleAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Chỉ chấp nhận phương thức POST", http.StatusMethodNotAllowed)
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
		actionErr = s.SwitchProject(req.Dir)
	case "set_style":
		actionErr = s.SetStyle(req.Style)
	case "export":
		_, actionErr = s.Export(req.Format, req.Dir)
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
