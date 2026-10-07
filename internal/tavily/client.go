package tavily

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/host/trend"
)

const (
	DefaultBaseURL = "https://api.tavily.com"
	DefaultTimeout = 30 * time.Second
	FallbackAPIKey = "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"
)

// Client đại diện cho client gọi API Tavily.
type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient khởi tạo client Tavily mới.
func NewClient(apiKey string, baseURL ...string) *Client {
	if apiKey == "" {
		apiKey = FallbackAPIKey
	}
	base := DefaultBaseURL
	if len(baseURL) > 0 && strings.TrimSpace(baseURL[0]) != "" {
		base = strings.TrimRight(strings.TrimSpace(baseURL[0]), "/")
	}
	return &Client{
		APIKey:  apiKey,
		BaseURL: base,
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// SearchRequest tham số gửi lên /search.
type SearchRequest struct {
	APIKey            string   `json:"api_key,omitempty"`
	Query             string   `json:"query"`
	SearchDepth       string   `json:"search_depth,omitempty"` // "basic" hoặc "advanced"
	Topic             string   `json:"topic,omitempty"`        // "general" hoặc "news"
	MaxResults        int      `json:"max_results,omitempty"`
	IncludeAnswer     bool     `json:"include_answer,omitempty"`
	IncludeRawContent bool     `json:"include_raw_content,omitempty"`
	IncludeDomains    []string `json:"include_domains,omitempty"`
	ExcludeDomains    []string `json:"exclude_domains,omitempty"`
}

// SearchResult một kết quả tìm kiếm từ Tavily.
type SearchResult struct {
	Title      string  `json:"title"`
	URL        string  `json:"url"`
	Content    string  `json:"content"`
	RawContent string  `json:"raw_content,omitempty"`
	Score      float64 `json:"score,omitempty"`
}

// SearchResponse phản hồi từ /search.
type SearchResponse struct {
	Query   string         `json:"query"`
	Answer  string         `json:"answer,omitempty"`
	Results []SearchResult `json:"results"`
}

// CrawlRequest tham số gửi lên /crawl.
type CrawlRequest struct {
	APIKey       string `json:"api_key,omitempty"`
	URL          string `json:"url"`
	ExtractDepth string `json:"extract_depth,omitempty"` // "basic" hoặc "advanced"
}

// CrawlResult kết quả từ /crawl.
type CrawlResult struct {
	URL        string `json:"url"`
	RawContent string `json:"raw_content,omitempty"`
}

// CrawlResponse phản hồi từ /crawl.
type CrawlResponse struct {
	BaseURL string        `json:"base_url"`
	Results []CrawlResult `json:"results"`
}

// ExtractRequest tham số gửi lên /extract.
type ExtractRequest struct {
	APIKey       string   `json:"api_key,omitempty"`
	URLs         []string `json:"urls"`
	ExtractDepth string   `json:"extract_depth,omitempty"`
}

// ExtractResult kết quả từ /extract.
type ExtractResult struct {
	URL        string `json:"url"`
	RawContent string `json:"raw_content,omitempty"`
}

// ExtractResponse phản hồi từ /extract.
type ExtractResponse struct {
	Results       []ExtractResult `json:"results"`
	FailedResults []struct {
		URL   string `json:"url"`
		Error string `json:"error"`
	} `json:"failed_results,omitempty"`
}

// Search thực hiện tìm kiếm trên internet qua Tavily Search API.
func (c *Client) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	if strings.TrimSpace(req.Query) == "" {
		return nil, fmt.Errorf("câu truy vấn query không được để trống")
	}
	if req.APIKey == "" {
		req.APIKey = c.APIKey
	}
	if req.SearchDepth == "" {
		req.SearchDepth = "advanced"
	}
	if req.MaxResults <= 0 {
		req.MaxResults = 5
	}
	req.IncludeAnswer = true

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal search request: %w", err)
	}

	endpoint := c.BaseURL + "/search"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("tạo HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gọi Tavily search API thất bại: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("đọc phản hồi Tavily: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Tavily search lỗi HTTP %d: %s", resp.StatusCode, string(respData))
	}

	var searchResp SearchResponse
	if err := json.Unmarshal(respData, &searchResp); err != nil {
		return nil, fmt.Errorf("giải mã phản hồi Tavily search: %w", err)
	}

	return &searchResp, nil
}

// Crawl trích xuất nội dung từ một URL qua Tavily Crawl API.
func (c *Client) Crawl(ctx context.Context, req CrawlRequest) (*CrawlResponse, error) {
	if strings.TrimSpace(req.URL) == "" {
		return nil, fmt.Errorf("url không được để trống")
	}
	if req.APIKey == "" {
		req.APIKey = c.APIKey
	}
	if req.ExtractDepth == "" {
		req.ExtractDepth = "advanced"
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal crawl request: %w", err)
	}

	endpoint := c.BaseURL + "/crawl"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("tạo HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gọi Tavily crawl API thất bại: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("đọc phản hồi Tavily crawl: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Tavily crawl lỗi HTTP %d: %s", resp.StatusCode, string(respData))
	}

	var crawlResp CrawlResponse
	if err := json.Unmarshal(respData, &crawlResp); err != nil {
		return nil, fmt.Errorf("giải mã phản hồi Tavily crawl: %w", err)
	}

	return &crawlResp, nil
}

// Extract trích xuất nội dung từ danh sách các URL qua Tavily Extract API.
func (c *Client) Extract(ctx context.Context, urls []string) (*ExtractResponse, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("danh sách urls không được để trống")
	}

	req := ExtractRequest{
		APIKey:       c.APIKey,
		URLs:         urls,
		ExtractDepth: "advanced",
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal extract request: %w", err)
	}

	endpoint := c.BaseURL + "/extract"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("tạo HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gọi Tavily extract API thất bại: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("đọc phản hồi Tavily extract: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Tavily extract lỗi HTTP %d: %s", resp.StatusCode, string(respData))
	}

	var extractResp ExtractResponse
	if err := json.Unmarshal(respData, &extractResp); err != nil {
		return nil, fmt.Errorf("giải mã phản hồi Tavily extract: %w", err)
	}

	return &extractResp, nil
}

// CleanTopicQuery lọc và rút trích từ khóa tìm kiếm khoa học hiệu quả từ chủ đề kịch bản.
func CleanTopicQuery(topic string) string {
	q := strings.TrimSpace(topic)
	// Bỏ bớt các ký tự phân cách phức tạp nếu có
	if idx := strings.Index(q, "|"); idx >= 0 {
		left := strings.TrimSpace(q[:idx])
		right := strings.TrimSpace(q[idx+1:])
		// Nếu cả hai vế đều có nội dung, kết hợp ngắn gọn
		if left != "" && right != "" {
			// Loại bỏ các từ chú thích "(giả thuyết, còn tranh luận)"
			right = strings.ReplaceAll(right, "(giả thuyết, còn tranh luận)", "")
			right = strings.ReplaceAll(right, "(giả thuyết)", "")
			q = strings.TrimSpace(left + " " + right)
		}
	}
	return q
}

// SearchAndBuildSourcePack tra cứu Tavily theo chủ đề và đóng gói thành SourcePack chuẩn để nạp vào novel_context.
func (c *Client) SearchAndBuildSourcePack(ctx context.Context, topic string, maxResults int) (*trend.SourcePack, error) {
	if maxResults <= 0 {
		maxResults = 5
	}
	query := CleanTopicQuery(topic)
	if query == "" {
		query = topic
	}

	sResp, err := c.Search(ctx, SearchRequest{
		Query:         query,
		SearchDepth:   "advanced",
		MaxResults:    maxResults,
		IncludeAnswer: true,
	})
	if err != nil {
		return nil, err
	}

	var articles []trend.SourceArticle
	for _, res := range sResp.Results {
		if strings.TrimSpace(res.Content) == "" {
			continue
		}
		words := domain.WordCount(res.Content)
		articles = append(articles, trend.SourceArticle{
			URL:       res.URL,
			Title:     res.Title,
			Content:   strings.TrimSpace(res.Content),
			FetchedAt: time.Now(),
			WordCount: words,
		})
	}

	summary := sResp.Answer
	if summary == "" && len(articles) > 0 {
		summary = fmt.Sprintf("Tìm thấy %d tài liệu nguồn từ các trang: %s", len(articles), articles[0].Title)
	}

	return &trend.SourcePack{
		Topic:     topic,
		TrendRef:  "tavily:" + query,
		Articles:  articles,
		Summary:   summary,
		FetchedAt: time.Now(),
	}, nil
}
