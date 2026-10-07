package tavily

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCleanTopicQuery(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    `Vì sao con người mất gần hết lông? | Ta là "vận động viên marathon" săn mồi bằng sức bền và mồ hôi (giả thuyết, còn tranh luận)`,
			expected: `Vì sao con người mất gần hết lông? Ta là "vận động viên marathon" săn mồi bằng sức bền và mồ hôi`,
		},
		{
			input:    "Tại sao lại có tiền tệ",
			expected: "Tại sao lại có tiền tệ",
		},
	}

	for _, tc := range tests {
		got := CleanTopicQuery(tc.input)
		if got != tc.expected {
			t.Errorf("CleanTopicQuery(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestTavilyClient_Search(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		var req SearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp := SearchResponse{
			Query:  req.Query,
			Answer: "Tóm tắt khoa học về tiến hóa mất lông của con người.",
			Results: []SearchResult{
				{
					Title:   "Vì sao con người mất lông",
					URL:     "https://vietnamplus.vn/bai-viet-1",
					Content: "Nội dung nghiên cứu từ đại học...",
					Score:   0.95,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-key", server.URL)
	ctx := context.Background()

	res, err := client.Search(ctx, SearchRequest{
		Query: "con người mất lông",
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(res.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res.Results))
	}
	if res.Results[0].Title != "Vì sao con người mất lông" {
		t.Errorf("unexpected title: %s", res.Results[0].Title)
	}

	// Test SearchAndBuildSourcePack
	sp, err := client.SearchAndBuildSourcePack(ctx, "Vì sao con người mất lông", 3)
	if err != nil {
		t.Fatalf("SearchAndBuildSourcePack failed: %v", err)
	}
	if len(sp.Articles) != 1 {
		t.Fatalf("expected 1 article in source pack, got %d", len(sp.Articles))
	}
	if sp.Summary != "Tóm tắt khoa học về tiến hóa mất lông của con người." {
		t.Errorf("unexpected summary: %s", sp.Summary)
	}
}

func TestTavilyClient_Crawl(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/crawl" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		resp := CrawlResponse{
			BaseURL: "https://example.com/test",
			Results: []CrawlResult{
				{
					URL:        "https://example.com/test",
					RawContent: "Chi tiết bài báo khoa học...",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-key", server.URL)
	ctx := context.Background()

	res, err := client.Crawl(ctx, CrawlRequest{
		URL: "https://example.com/test",
	})
	if err != nil {
		t.Fatalf("Crawl failed: %v", err)
	}
	if len(res.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res.Results))
	}
}

func TestTavilyClient_Extract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/extract" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		resp := ExtractResponse{
			Results: []ExtractResult{
				{
					URL:        "https://example.com/test",
					RawContent: "Nội dung trích xuất sâu...",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-key", server.URL)
	ctx := context.Background()

	res, err := client.Extract(ctx, []string{"https://example.com/test"})
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	if len(res.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res.Results))
	}
}
