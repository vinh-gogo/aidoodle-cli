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

func TestCheckGatewayHealth_CloudflareAccessRedirect302(t *testing.T) {
	client := mockHealthClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Host == "curly-dawn-d1f4.cloudflareaccess.com" {
			t.Fatalf("health check must not follow redirect to Cloudflare Access login host: %s", r.URL.String())
		}
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.Header().Set("Www-Authenticate", `Cloudflare-Access resource_metadata="https://api.gloryvocab.com/.well-known/cloudflare-access-protected-resource/v1/models"`)
		w.Header().Set("Location", "https://curly-dawn-d1f4.cloudflareaccess.com/cdn-cgi/access/login/api.gloryvocab.com?redirect_url=%2Fv1%2Fmodels")
		w.WriteHeader(http.StatusFound)
		_, _ = w.Write([]byte(`<html><head><title>302 Found</title></head><body>302 Found</body></html>`))
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := CheckGatewayHealth(ctx, GatewayHealthOptions{
		BaseURL: "https://api.gloryvocab.com/v1",
		APIKey:  "kaggle-key",
		Client:  client,
	})
	if err == nil {
		t.Fatal("expected error when blocked by Cloudflare Access 302 redirect, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "Cloudflare") || !strings.Contains(msg, "Access") {
		t.Fatalf("expected error message to diagnose Cloudflare Access block, got: %v", err)
	}
}

func TestCheckGatewayHealth_ForwardsExtraHeaders(t *testing.T) {
	pc := ProviderConfig{
		Type:    "openai",
		BaseURL: "https://api.gloryvocab.com/v1",
		APIKey:  "kaggle-key",
		Extra: map[string]any{
			"headers": map[string]any{
				"CF-Access-Client-Id":     "client-id-123.access",
				"CF-Access-Client-Secret": "secret-xyz-789",
			},
		},
	}

	client := mockHealthClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("CF-Access-Client-Id") != "client-id-123.access" ||
			r.Header.Get("CF-Access-Client-Secret") != "secret-xyz-789" {
			w.Header().Set("Location", "https://curly-dawn-d1f4.cloudflareaccess.com/cdn-cgi/access/login")
			w.WriteHeader(http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"object":"list","data":[{"id":"main-llm"}]}`))
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := CheckGatewayHealth(ctx, GatewayHealthOptions{
		BaseURL: pc.BaseURL,
		APIKey:  pc.APIKey,
		Headers: pc.ExtraHeaders(),
		Client:  client,
	})
	if err != nil {
		t.Fatalf("expected health check with Cloudflare Access service token headers to succeed, got: %v", err)
	}
}
