package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/voocel/ainovel-cli/internal/tavily"
)

func TestTavilySearchTool(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			http.NotFound(w, r)
			return
		}
		resp := tavily.SearchResponse{
			Query:  "con người mất lông",
			Answer: "Tóm tắt khoa học về việc tản nhiệt.",
			Results: []tavily.SearchResult{
				{
					Title:   "Bài báo nghiên cứu",
					URL:     "https://vietnamplus.vn/bai-1",
					Content: "Nội dung bài viết về tuyến mồ hôi...",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := tavily.NewClient("test-key", server.URL)
	tool := NewTavilySearchTool(client)

	if tool.Name() != "tavily_search" {
		t.Errorf("unexpected name: %s", tool.Name())
	}

	args, _ := json.Marshal(map[string]any{
		"query": "con người mất lông",
	})
	resRaw, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	var res map[string]any
	if err := json.Unmarshal(resRaw, &res); err != nil {
		t.Fatalf("unmarshal result failed: %v", err)
	}

	if res["answer"] != "Tóm tắt khoa học về việc tản nhiệt." {
		t.Errorf("unexpected answer: %v", res["answer"])
	}
	results, ok := res["results"].([]any)
	if !ok || len(results) != 1 {
		t.Fatalf("expected 1 result, got %v", res["results"])
	}
}

func TestTavilyCrawlTool(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/extract" {
			resp := tavily.ExtractResponse{
				Results: []tavily.ExtractResult{
					{
						URL:        "https://example.com/test",
						RawContent: "Toàn bộ bài báo khoa học chi tiết về gen MC1R...",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := tavily.NewClient("test-key", server.URL)
	tool := NewTavilyCrawlTool(client)

	if tool.Name() != "tavily_crawl" {
		t.Errorf("unexpected name: %s", tool.Name())
	}

	args, _ := json.Marshal(map[string]any{
		"url": "https://example.com/test",
	})
	resRaw, err := tool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	var res map[string]any
	if err := json.Unmarshal(resRaw, &res); err != nil {
		t.Fatalf("unmarshal result failed: %v", err)
	}

	if res["url"] != "https://example.com/test" {
		t.Errorf("unexpected url: %v", res["url"])
	}
	content, ok := res["content"].(string)
	if !ok || len(content) == 0 {
		t.Fatalf("expected non-empty content, got %v", res["content"])
	}
}

func TestTavilySearchTool_GracefulFallback(t *testing.T) {
	// 1. Trường hợp client không có API key
	emptyTool := NewTavilySearchTool(tavily.NewClient("", "http://127.0.0.1:9999"))
	args, _ := json.Marshal(map[string]any{"query": "nghịch lý lựa chọn"})
	resRaw, err := emptyTool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute không được trả về error khi thiếu API key: %v", err)
	}
	var res map[string]any
	_ = json.Unmarshal(resRaw, &res)
	if res["total_found"] != float64(0) {
		t.Errorf("expected total_found 0, got %v", res["total_found"])
	}

	// 2. Trường hợp server lỗi HTTP 401
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"detail":"Invalid API Key"}`, http.StatusUnauthorized)
	}))
	defer errorServer.Close()

	errClient := tavily.NewClient("invalid-key", errorServer.URL)
	errTool := NewTavilySearchTool(errClient)
	resRaw2, err2 := errTool.Execute(context.Background(), args)
	if err2 != nil {
		t.Fatalf("Execute không được trả về error khi API trả về 401: %v", err2)
	}
	var res2 map[string]any
	_ = json.Unmarshal(resRaw2, &res2)
	if res2["total_found"] != float64(0) {
		t.Errorf("expected total_found 0 on 401 error, got %v", res2["total_found"])
	}
}

func TestTavilyCrawlTool_GracefulFallback(t *testing.T) {
	// Trường hợp server lỗi HTTP 401
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"detail":"Invalid API Key"}`, http.StatusUnauthorized)
	}))
	defer errorServer.Close()

	errClient := tavily.NewClient("invalid-key", errorServer.URL)
	errTool := NewTavilyCrawlTool(errClient)
	args, _ := json.Marshal(map[string]any{"url": "https://example.com/test"})
	resRaw, err := errTool.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("Execute không được trả về error khi crawl gặp lỗi: %v", err)
	}
	var res map[string]any
	_ = json.Unmarshal(resRaw, &res)
	if res["word_count"] != float64(0) {
		t.Errorf("expected word_count 0 on error, got %v", res["word_count"])
	}
}

