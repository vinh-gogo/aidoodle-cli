package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/voocel/ainovel-cli/assets"
	"github.com/voocel/ainovel-cli/internal/bootstrap"
	buildversion "github.com/voocel/ainovel-cli/internal/version"
)

func TestProjectHelpers(t *testing.T) {
	// 1. Test DetectStyleBadge
	if badge := DetectStyleBadge("vietnamese-history"); !strings.Contains(badge, "Lịch sử Việt Nam") {
		t.Errorf("expected history badge, got %s", badge)
	}
	if badge := DetectStyleBadge("doodle-explainer"); !strings.Contains(badge, "TikTok") {
		t.Errorf("expected doodle badge, got %s", badge)
	}

	// 2. Test NextOutputDir
	now := time.Date(2026, 10, 8, 14, 30, 0, 0, time.UTC)
	next := NextOutputDir("output/novel-20261008-1037", now)
	if !strings.HasPrefix(filepath.ToSlash(next), "output/novel-20261008-1430") {
		t.Errorf("expected output/novel-20261008-1430, got %s", next)
	}

	// 3. Test ScanProjects in a temp dir
	tmpDir := t.TempDir()
	p1 := filepath.Join(tmpDir, "novel-20261008-1000")
	os.MkdirAll(filepath.Join(p1, "meta"), 0755)
	os.WriteFile(filepath.Join(p1, "meta", "run.json"), []byte(`{"style": "vietnamese-history", "started_at": "2026-10-08T10:00:00Z"}`), 0644)
	os.WriteFile(filepath.Join(p1, "meta", "book.json"), []byte(`{"title": "Test Book"}`), 0644)

	projects := ScanProjects(tmpDir, p1)
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Title != "Test Book" {
		t.Errorf("expected 'Test Book', got %s", projects[0].Title)
	}
	if !projects[0].IsCurrent {
		t.Errorf("expected project to be current")
	}
}

func TestWebEndpoints(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := bootstrap.Config{
		OutputDir: tmpDir,
		Style:     "vietnamese-history",
		Provider:  "mock",
		ModelName: "test-model",
	}
	bundle := assets.Load("vietnamese-history", assets.DefaultLoadOptions(tmpDir))
	buildInfo := buildversion.Info{Version: "test-v1"}

	srv, err := NewServer(cfg, bundle, 8088, buildInfo)
	if err != nil {
		t.Fatalf("NewServer error: %v", err)
	}
	defer srv.Close()

	mux := http.NewServeMux()
	srv.registerRoutes(mux)

	// 1. Test Index
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET / returned %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "aiNovel Studio") {
		t.Errorf("body missing aiNovel Studio")
	}

	// 2. Test /api/status
	req = httptest.NewRequest("GET", "/api/status", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/status returned %d", w.Code)
	}
	var statusResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &statusResp); err != nil {
		t.Fatalf("status JSON unmarshal error: %v", err)
	}
	if statusResp["ok"] != true {
		t.Errorf("expected ok=true in status")
	}

	// 3. Test /api/projects
	req = httptest.NewRequest("GET", "/api/projects", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/projects returned %d", w.Code)
	}

	// 4. Test /api/action with invalid action
	req = httptest.NewRequest("POST", "/api/action", strings.NewReader(`{"action": "unknown"}`))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown action, got %d", w.Code)
	}
}
