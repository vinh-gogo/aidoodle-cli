package bootstrap

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func mockHealthClient(h http.HandlerFunc) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			rec := httptest.NewRecorder()
			h(rec, r)
			return rec.Result(), nil
		}),
	}
}

func TestCheckGatewayHealth_SuccessModels(t *testing.T) {
	client := mockHealthClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/models" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data": [{"id": "main-llm"}]}`))
			return
		}
		http.NotFound(w, r)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := CheckGatewayHealth(ctx, GatewayHealthOptions{
		BaseURL: "http://gateway.local",
		APIKey:  "test-key",
		Client:  client,
	})
	if err != nil {
		t.Fatalf("expected healthy, got error: %v", err)
	}
}

func TestCheckGatewayHealth_FallbackHealth(t *testing.T) {
	client := mockHealthClient(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/models" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == "/health" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status": "ok"}`))
			return
		}
		http.NotFound(w, r)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := CheckGatewayHealth(ctx, GatewayHealthOptions{
		BaseURL: "http://gateway.local",
		Client:  client,
	})
	if err != nil {
		t.Fatalf("expected healthy on fallback /health, got error: %v", err)
	}
}

func TestCheckGatewayHealth_CloudflareTunnel502(t *testing.T) {
	client := mockHealthClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><body>Error 502: Bad Gateway Cloudflare</body></html>`))
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := CheckGatewayHealth(ctx, GatewayHealthOptions{
		BaseURL: "http://gateway.local",
		Client:  client,
	})
	if err == nil {
		t.Fatal("expected error on 502 Cloudflare tunnel, got nil")
	}
	if !strings.Contains(err.Error(), "Cloudflare Tunnel") && !strings.Contains(err.Error(), "502") {
		t.Fatalf("expected error message to mention Cloudflare Tunnel 502, got: %v", err)
	}
}

func TestCheckGatewayHealth_AuthFailure401(t *testing.T) {
	client := mockHealthClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": "invalid_api_key"}`))
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := CheckGatewayHealth(ctx, GatewayHealthOptions{
		BaseURL: "http://gateway.local",
		APIKey:  "wrong-key",
		Client:  client,
	})
	if err == nil {
		t.Fatal("expected error on 401, got nil")
	}
	if !strings.Contains(err.Error(), "xác thực thất bại") && !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected error message to mention auth failure, got: %v", err)
	}
}

func TestCheckGatewayHealth_EmptyOrMockBypassed(t *testing.T) {
	ctx := context.Background()
	if err := CheckGatewayHealth(ctx, GatewayHealthOptions{BaseURL: ""}); err != nil {
		t.Fatalf("empty baseURL should be bypassed, got %v", err)
	}
	if err := CheckGatewayHealth(ctx, GatewayHealthOptions{BaseURL: "mock://test"}); err != nil {
		t.Fatalf("mock baseURL should be bypassed, got %v", err)
	}
}
