package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/voocel/ainovel-cli/assets"
	"github.com/voocel/ainovel-cli/internal/bootstrap"
	buildversion "github.com/voocel/ainovel-cli/internal/version"
)

// setupAdversarialServer creates an isolated web.Server for adversarial testing.
func setupAdversarialServer(t *testing.T, opts Options) (*Server, string) {
	t.Helper()
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "output", "proj-alpha")
	if err := os.MkdirAll(filepath.Join(outDir, "chapters"), 0755); err != nil {
		t.Fatalf("setup outDir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "chapters", "01.md"), []byte("# Chap 1\nContent"), 0644); err != nil {
		t.Fatalf("write ch1: %v", err)
	}

	opts.BaseDir = filepath.Join(tmpDir, "output")

	cfg := bootstrap.Config{
		OutputDir: outDir,
		Style:     "novel-manga",
		Provider:  "mock",
	}
	bundle := assets.Bundle{
		Styles: map[string]string{"novel-manga": "Manga"},
	}
	info := buildversion.Info{Version: "v1.0.0-challenger"}

	srv, err := NewServerWithOptions(cfg, bundle, opts, info)
	if err != nil {
		t.Fatalf("NewServerWithOptions: %v", err)
	}
	return srv, tmpDir
}

// TestAdversarial_SafeProjectDir tests double dot, encoded, absolute, UNC, null byte, trailing slash.
func TestAdversarial_SafeProjectDir(t *testing.T) {
	srv, tmpDir := setupAdversarialServer(t, Options{})
	defer srv.Close()
	baseDir := filepath.Join(tmpDir, "output")

	// 1. Direct SafeProjectDir unit attacks
	attackVectors := []struct {
		name        string
		path        string
		expectError bool
	}{
		{"DoubleDotSimple", "..", true},
		{"DoubleDotDeep", "../..", true},
		{"DoubleDotSystem32", "../../../../Windows/System32", true},
		{"DoubleDotEtcPasswd", "../../etc/passwd", true},
		{"DoubleDotSubdir", "proj-alpha/../../secret", true},
		{"AbsoluteLinuxRoot", "/etc/passwd", true},
		{"AbsoluteWinC", `C:\Windows`, true},
		{"AbsoluteWinSys32", `C:\Windows\System32\drivers\etc\hosts`, true},
		{"UNCPathBackslash", `\\server\share`, true},
		{"UNCPathIP", `\\127.0.0.1\c$\Windows`, true},
		{"UNCPathForward", `//server/share`, true},
		{"NullByteInjection", "proj-alpha\x00/../../etc", true},
		{"ValidSubdir", "proj-alpha", false},
		{"ValidSubdirTrailingSlash", "proj-alpha/", false},
		{"ValidSubdirMultipleTrailingSlashes", "proj-alpha///", false},
	}

	for _, tc := range attackVectors {
		t.Run("SafeProjectDir_"+tc.name, func(t *testing.T) {
			res, err := SafeProjectDir(baseDir, tc.path)
			if tc.expectError {
				if err == nil {
					t.Fatalf("SECURITY VIOLATION: expected error for vector %q, but got path %q", tc.path, res)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error for valid vector %q: %v", tc.path, err)
				}
				expectedPrefix := filepath.Clean(baseDir)
				if !filepath.IsAbs(res) || filepath.Clean(res) != filepath.Join(expectedPrefix, "proj-alpha") {
					t.Fatalf("expected resolved path within baseDir, got %q", res)
				}
			}
		})
	}
}

// TestAdversarial_HTTPPathTraversalEndpoints tests HTTP 400 rejection on all traversal endpoints.
func TestAdversarial_HTTPPathTraversalEndpoints(t *testing.T) {
	srv, _ := setupAdversarialServer(t, Options{})
	defer srv.Close()
	handler := srv.Handler()

	traversalPayloads := []string{
		"..",
		"../..",
		"../../../../Windows/System32",
		"/etc/passwd",
		`C:\Windows`,
		`\\server\share`,
		"%2e%2e",
		"%2e%2e%2f",
		"%252e%252e",
		"proj-alpha%00/../../",
	}

	for _, payload := range traversalPayloads {
		// 1. GET /api/chapter
		t.Run("HTTP_Chapter_"+payload, func(t *testing.T) {
			url := fmt.Sprintf("/api/chapter?ch=1&dir=%s", payload)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("SECURITY FLAW: GET /api/chapter?dir=%s returned status %d, expected 400 Bad Request", payload, w.Code)
			}
		})

		// 2. GET /api/chapters
		t.Run("HTTP_Chapters_"+payload, func(t *testing.T) {
			url := fmt.Sprintf("/api/chapters?dir=%s", payload)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("SECURITY FLAW: GET /api/chapters?dir=%s returned status %d, expected 400 Bad Request", payload, w.Code)
			}
		})

		// 3. GET /api/export
		t.Run("HTTP_Export_"+payload, func(t *testing.T) {
			url := fmt.Sprintf("/api/export?format=epub&dir=%s", payload)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("SECURITY FLAW: GET /api/export?dir=%s returned status %d, expected 400 Bad Request", payload, w.Code)
			}
		})

		// 4. POST /api/action (action: "switch")
		t.Run("HTTP_ActionSwitch_"+payload, func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{
				"action": "switch",
				"dir":    payload,
			})
			req := httptest.NewRequest(http.MethodPost, "/api/action", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Origin", "http://127.0.0.1:8080")
			req.Header.Set("X-Requested-With", "ainovel-web")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("SECURITY FLAW: POST /api/action switch with dir=%s returned status %d, expected 400 Bad Request", payload, w.Code)
			}
		})

		// 5. POST /api/action (action: "export")
		t.Run("HTTP_ActionExport_"+payload, func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{
				"action": "export",
				"format": "epub",
				"dir":    payload,
			})
			req := httptest.NewRequest(http.MethodPost, "/api/action", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Origin", "http://127.0.0.1:8080")
			req.Header.Set("X-Requested-With", "ainovel-web")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("SECURITY FLAW: POST /api/action export with dir=%s returned status %d, expected 400 Bad Request", payload, w.Code)
			}
		})
	}
}

// TestAdversarial_OriginAndCSRFProtection tests unauthorized Origins, missing headers, cross-origin -> HTTP 403.
func TestAdversarial_OriginAndCSRFProtection(t *testing.T) {
	srv, _ := setupAdversarialServer(t, Options{})
	defer srv.Close()
	handler := srv.Handler()

	bodyData, _ := json.Marshal(map[string]string{"action": "pause"})

	unauthorizedOrigins := []string{
		"http://evil.com",
		"https://attacker.org",
		"null",
		"http://127.0.0.1.attacker.com",
		"http://localhost.evil.com",
		"http://evil-127.0.0.1:8080",
		"https://sub.domain.evil",
		"http://192.168.1.100:8080",
	}

	for _, origin := range unauthorizedOrigins {
		t.Run("UnauthorizedOrigin_"+origin, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/action", bytes.NewReader(bodyData))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Origin", origin)
			req.Header.Set("X-Requested-With", "ainovel-web")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != http.StatusForbidden {
				t.Fatalf("CSRF VULNERABILITY: Origin %q was not rejected with 403 Forbidden, got status %d", origin, w.Code)
			}
		})
	}

	// Missing X-Requested-With header
	t.Run("MissingXRequestedWith", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/action", bytes.NewReader(bodyData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "http://127.0.0.1:8080")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("CSRF VULNERABILITY: Missing X-Requested-With was not rejected with 403, got %d", w.Code)
		}
	})

	// Wrong X-Requested-With header
	t.Run("InvalidXRequestedWith", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/action", bytes.NewReader(bodyData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "http://127.0.0.1:8080")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("CSRF VULNERABILITY: Invalid X-Requested-With was not rejected with 403, got %d", w.Code)
		}
	})

	// Sec-Fetch-Site: cross-site
	t.Run("CrossSiteSecFetchSite", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/action", bytes.NewReader(bodyData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Origin", "http://127.0.0.1:8080")
		req.Header.Set("Sec-Fetch-Site", "cross-site")
		req.Header.Set("X-Requested-With", "ainovel-web")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("CSRF VULNERABILITY: Sec-Fetch-Site: cross-site was not rejected with 403, got %d", w.Code)
		}
	})
}

// TestAdversarial_BearerAuth tests missing/invalid Bearer token -> HTTP 401.
func TestAdversarial_BearerAuth(t *testing.T) {
	const validToken = "super-secret-auth-token-xyz987"
	srv, _ := setupAdversarialServer(t, Options{AuthToken: validToken})
	defer srv.Close()
	handler := srv.Handler()

	invalidCases := []struct {
		name   string
		header string
	}{
		{"NoHeader", ""},
		{"InvalidToken", "Bearer wrong-token-value"},
		{"EmptyBearer", "Bearer "},
		{"ShortToken", "Bearer abc"},
		{"BasicAuth", "Basic dXNlcjpwYXNz"},
		{"GarbageHeader", "Token super-secret-auth-token-xyz987"},
		{"LowercaseBearerWrong", "bearer wrong"},
	}

	for _, tc := range invalidCases {
		t.Run("BearerAuth_"+tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("AUTH VULNERABILITY: Header %q allowed access or returned %d instead of 401 Unauthorized", tc.header, w.Code)
			}
		})
	}

	// Valid Bearer token must succeed (HTTP 200)
	t.Run("BearerAuth_ValidToken", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("Valid Bearer token was rejected with %d", w.Code)
		}
	})
}
