package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/voocel/ainovel-cli/assets"
	"github.com/voocel/ainovel-cli/internal/bootstrap"
	buildversion "github.com/voocel/ainovel-cli/internal/version"
)

func createSecurityTestServer(t *testing.T, opts Options) (*Server, string) {
	t.Helper()
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "output", "proj1")
	if err := os.MkdirAll(filepath.Join(outDir, "chapters"), 0755); err != nil {
		t.Fatalf("mkdir outDir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "chapters", "01.md"), []byte("# Chap 1\nContent"), 0644); err != nil {
		t.Fatalf("write chap: %v", err)
	}

	cfg := bootstrap.Config{
		OutputDir: outDir,
		Style:     "vietnamese-history",
		Provider:  "mock",
	}
	bundle := assets.Bundle{}
	info := buildversion.Info{Version: "v1.0.0-test"}

	srv, err := NewServerWithOptions(cfg, bundle, opts, info)
	if err != nil {
		t.Fatalf("NewServerWithOptions: %v", err)
	}
	return srv, tmpDir
}

func TestSecurity_SafeProjectDir(t *testing.T) {
	tmpDir := t.TempDir()
	baseDir := filepath.Join(tmpDir, "output")
	p1 := filepath.Join(baseDir, "novel-1")
	if err := os.MkdirAll(p1, 0755); err != nil {
		t.Fatalf("mkdir p1: %v", err)
	}

	// 1. Safe path inside baseDir
	resolved, err := SafeProjectDir(baseDir, "novel-1")
	if err != nil {
		t.Fatalf("expected valid safe dir, got err: %v", err)
	}
	if resolved != p1 {
		t.Errorf("expected %q, got %q", p1, resolved)
	}

	// 2. Traversal using ..
	_, err = SafeProjectDir(baseDir, "../secret")
	if err == nil {
		t.Errorf("expected error for traversal, got nil")
	}

	// 3. Absolute path outside baseDir
	outsidePath := filepath.Join(tmpDir, "outside")
	os.MkdirAll(outsidePath, 0755)
	_, err = SafeProjectDir(baseDir, outsidePath)
	if err == nil {
		t.Errorf("expected error for outside absolute path, got nil")
	}

	// 4. Empty dir defaults to baseDir
	resolved, err = SafeProjectDir(baseDir, "")
	if err != nil {
		t.Fatalf("expected baseDir for empty, got err: %v", err)
	}
	if resolved != baseDir {
		t.Errorf("expected %q, got %q", baseDir, resolved)
	}
}

func TestSecurity_HostBinding(t *testing.T) {
	srv, _ := createSecurityTestServer(t, Options{})
	if srv.HostAddr() != "127.0.0.1" {
		t.Errorf("expected default host 127.0.0.1, got %q", srv.HostAddr())
	}

	srvCustom, _ := createSecurityTestServer(t, Options{Host: "0.0.0.0"})
	if srvCustom.HostAddr() != "0.0.0.0" {
		t.Errorf("expected custom host 0.0.0.0, got %q", srvCustom.HostAddr())
	}
}

func TestSecurity_AuthenticationMiddleware(t *testing.T) {
	const secretToken = "my-secret-security-token-12345"
	srv, _ := createSecurityTestServer(t, Options{Token: secretToken})
	handler := srv.Handler()

	// 1. Unauthenticated request to /api/status -> 401
	{
		req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized, got %d", w.Code)
		}
	}

	// 2. Request with Bearer token -> 200
	{
		req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
		req.Header.Set("Authorization", "Bearer "+secretToken)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200 OK with Bearer token, got %d", w.Code)
		}
	}

	// 3. Request with ?token= query param -> 200 and sets cookie
	var sessionCookie *http.Cookie
	{
		req := httptest.NewRequest(http.MethodGet, "/?token="+secretToken, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200 OK with query token, got %d", w.Code)
		}
		for _, c := range w.Result().Cookies() {
			if c.Name == "ainovel_token" && c.Value == secretToken {
				sessionCookie = c
				break
			}
		}
		if sessionCookie == nil {
			t.Errorf("expected session cookie ainovel_token to be set")
		}
	}

	// 4. Request with established session cookie -> 200
	if sessionCookie != nil {
		req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
		req.AddCookie(sessionCookie)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200 OK with session cookie, got %d", w.Code)
		}
	}

	// 5. Request with invalid Bearer token -> 401
	{
		req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
		req.Header.Set("Authorization", "Bearer wrong-token")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 with wrong token, got %d", w.Code)
		}
	}
}

func TestSecurity_CSRFAndOriginValidation(t *testing.T) {
	srv, _ := createSecurityTestServer(t, Options{})
	handler := srv.Handler()

	bodyData, _ := json.Marshal(map[string]string{"action": "pause"})

	// 1. Same-origin request with X-Requested-With -> 200
	{
		req := httptest.NewRequest(http.MethodPost, "/api/action", bytes.NewReader(bodyData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "http://127.0.0.1:8080")
		req.Header.Set("X-Requested-With", "ainovel-web")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code == http.StatusForbidden {
			t.Errorf("expected action to be allowed, got 403 Forbidden")
		}
	}

	// 2. Untrusted Origin -> 403 Forbidden
	{
		req := httptest.NewRequest(http.MethodPost, "/api/action", bytes.NewReader(bodyData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "http://malicious.evil.com")
		req.Header.Set("X-Requested-With", "ainovel-web")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden for evil Origin, got %d", w.Code)
		}
	}

	// 3. Sec-Fetch-Site: cross-site -> 403 Forbidden
	{
		req := httptest.NewRequest(http.MethodPost, "/api/action", bytes.NewReader(bodyData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Sec-Fetch-Site", "cross-site")
		req.Header.Set("X-Requested-With", "ainovel-web")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden for cross-site fetch, got %d", w.Code)
		}
	}

	// 4. Missing X-Requested-With -> 403 Forbidden
	{
		req := httptest.NewRequest(http.MethodPost, "/api/action", bytes.NewReader(bodyData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "http://127.0.0.1:8080")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden for missing X-Requested-With, got %d", w.Code)
		}
	}
}

func TestSecurity_NoWildcardCORSOnStream(t *testing.T) {
	srv, _ := createSecurityTestServer(t, Options{})
	handler := srv.Handler()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/api/stream", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	cors := w.Header().Get("Access-Control-Allow-Origin")
	if cors == "*" {
		t.Errorf("stream endpoint must not contain wildcard Access-Control-Allow-Origin: *")
	}
}

func TestSecurity_PathTraversalEndpoints(t *testing.T) {
	srv, _ := createSecurityTestServer(t, Options{})
	handler := srv.Handler()

	// 1. /api/chapter with path traversal dir param -> 400
	{
		req := httptest.NewRequest(http.MethodGet, "/api/chapter?ch=1&dir=../../../../Windows/System32", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request for traversal on /api/chapter, got %d", w.Code)
		}
	}

	// 2. /api/chapters with path traversal dir param -> 400
	{
		req := httptest.NewRequest(http.MethodGet, "/api/chapters?dir=../../../../Windows/System32", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request for traversal on /api/chapters, got %d", w.Code)
		}
	}

	// 3. /api/export with path traversal dir param -> 400
	{
		req := httptest.NewRequest(http.MethodGet, "/api/export?format=epub&dir=../../../../Windows/System32", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request for traversal on /api/export, got %d", w.Code)
		}
	}
}
