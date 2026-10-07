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
