package host

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/voocel/ainovel-cli/assets"
	"github.com/voocel/ainovel-cli/internal/bootstrap"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestPrepareUserRules_FailsFastOnCloudflareAccessBeforeLLMRetry(t *testing.T) {
	var chatCalls atomic.Int32
	origTransport := http.DefaultTransport
	http.DefaultTransport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if strings.HasSuffix(r.URL.Path, "/chat/completions") {
			chatCalls.Add(1)
		}
		rec := httptest.NewRecorder()
		rec.Header().Set("Content-Type", "text/html; charset=UTF-8")
		rec.Header().Set("Www-Authenticate", `Cloudflare-Access resource_metadata="https://api.gloryvocab.com/.well-known/cloudflare-access-protected-resource/v1/models"`)
		rec.Header().Set("Location", "https://curly-dawn-d1f4.cloudflareaccess.com/cdn-cgi/access/login/api.gloryvocab.com")
		rec.WriteHeader(http.StatusFound)
		_, _ = rec.Write([]byte(`<html><head><title>302 Found</title></head></html>`))
		resp := rec.Result()
		resp.Request = r
		return resp, nil
	})
	t.Cleanup(func() {
		http.DefaultTransport = origTransport
	})

	dir := t.TempDir()
	cfg := bootstrap.Config{
		OutputDir: dir,
		Provider:  "kaggle-llama",
		ModelName: "main-llm",
		Providers: map[string]bootstrap.ProviderConfig{
			"kaggle-llama": {
				Type:    "openai",
				BaseURL: "https://api.gloryvocab.com/v1",
				APIKey:  "kaggle-key",
			},
		},
	}
	bundle := assets.Load("novel-manga", assets.DefaultLoadOptions(dir))

	h, err := New(cfg, bundle)
	if err != nil {
		t.Fatalf("New host: %v", err)
	}
	defer h.Close()

	start := time.Now()
	err = h.PrepareUserRules("hi")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected PrepareUserRules to fail fast when gateway is blocked by Cloudflare Access, got nil")
	}
	if !strings.Contains(err.Error(), "Cloudflare") || !strings.Contains(err.Error(), "Access") {
		t.Fatalf("expected error to mention Cloudflare Access, got: %v", err)
	}
	if chatCalls.Load() != 0 {
		t.Fatalf("expected 0 /chat/completions calls when pre-flight gateway check fails, got %d", chatCalls.Load())
	}
	if elapsed > 2*time.Second {
		t.Fatalf("expected fail-fast (<2s), took %s", elapsed)
	}
}

func TestPrepareUserRules_NoRetryStormWhenChatReturnsHTML(t *testing.T) {
	var chatCalls atomic.Int32
	origTransport := http.DefaultTransport
	http.DefaultTransport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		if strings.HasSuffix(r.URL.Path, "/models") {
			rec.Header().Set("Content-Type", "application/json")
			rec.WriteHeader(http.StatusOK)
			_, _ = rec.Write([]byte(`{"object":"list","data":[{"id":"main-llm"}]}`))
			resp := rec.Result()
			resp.Request = r
			return resp, nil
		}
		if strings.HasSuffix(r.URL.Path, "/chat/completions") {
			chatCalls.Add(1)
			rec.Header().Set("Content-Type", "text/html; charset=UTF-8")
			rec.WriteHeader(http.StatusOK)
			_, _ = rec.Write([]byte(`<!DOCTYPE html><html><body>Cloudflare Access Login</body></html>`))
			resp := rec.Result()
			resp.Request = r
			return resp, nil
		}
		rec.WriteHeader(http.StatusNotFound)
		resp := rec.Result()
		resp.Request = r
		return resp, nil
	})
	t.Cleanup(func() {
		http.DefaultTransport = origTransport
	})

	dir := t.TempDir()
	cfg := bootstrap.Config{
		OutputDir: dir,
		Provider:  "kaggle-llama",
		ModelName: "main-llm",
		Providers: map[string]bootstrap.ProviderConfig{
			"kaggle-llama": {
				Type:    "openai",
				BaseURL: "https://api.gloryvocab.com/v1",
				APIKey:  "kaggle-key",
			},
		},
	}
	bundle := assets.Load("novel-manga", assets.DefaultLoadOptions(dir))

	h, err := New(cfg, bundle)
	if err != nil {
		t.Fatalf("New host: %v", err)
	}
	defer h.Close()

	start := time.Now()
	_ = h.PrepareUserRules("hi")
	elapsed := time.Since(start)

	if chatCalls.Load() != 1 {
		t.Fatalf("expected exactly 1 /chat/completions call (no retry storm on HTML response), got %d", chatCalls.Load())
	}
	if elapsed > 2*time.Second {
		t.Fatalf("expected non-retryable HTML decode error to finish immediately (<2s), took %s", elapsed)
	}
}
