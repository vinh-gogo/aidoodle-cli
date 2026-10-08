package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/voocel/agentcore/schema"
	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/tavily"
)

// TavilyCrawlTool cho phép LLM trích xuất nội dung bài viết chi tiết từ một URL cụ thể.
type TavilyCrawlTool struct {
	client *tavily.Client
}

// NewTavilyCrawlTool khởi tạo tool tavily_crawl.
func NewTavilyCrawlTool(client *tavily.Client) *TavilyCrawlTool {
	if client == nil {
		client = tavily.NewClient("")
	}
	return &TavilyCrawlTool{client: client}
}

func (t *TavilyCrawlTool) Name() string { return "tavily_crawl" }
func (t *TavilyCrawlTool) Description() string {
	return "Trích xuất văn bản chi tiết từ một URL cụ thể qua Tavily Crawl/Extract để đào sâu số liệu, luận điểm nghiên cứu khoa học hoặc báo cáo gốc."
}
func (t *TavilyCrawlTool) Label() string { return "Trích xuất URL Tavily" }

func (t *TavilyCrawlTool) ReadOnly(_ json.RawMessage) bool        { return true }
func (t *TavilyCrawlTool) ConcurrencySafe(_ json.RawMessage) bool { return true }

func (t *TavilyCrawlTool) Schema() map[string]any {
	return schema.Object(
		schema.Property("url", schema.String("Đường dẫn URL bài viết cần đọc sâu")).Required(),
	)
}

func (t *TavilyCrawlTool) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var a struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, fmt.Errorf("tham số không hợp lệ: %w", err)
	}

	targetURL := strings.TrimSpace(a.URL)
	if targetURL == "" {
		return nil, fmt.Errorf("url không được để trống")
	}

	// Nếu client chưa có API key hoặc client nil, phản hồi mềm thay vì báo lỗi đứt luồng
	if t.client == nil || strings.TrimSpace(t.client.APIKey) == "" {
		fallbackMap := map[string]any{
			"url":         targetURL,
			"content":     "",
			"word_count":  0,
			"warning":     "Dịch vụ Tavily chưa được cấu hình API key.",
			"instruction": "Không gọi lại tool này. Tiếp tục sáng tác dựa trên tri thức sẵn có.",
		}
		data, _ := json.Marshal(fallbackMap)
		return data, nil
	}

	// Thử extract trước (nhanh và trực tiếp cho bài viết đơn lẻ)
	extResp, err := t.client.Extract(ctx, []string{targetURL})
	var content string
	if err == nil && len(extResp.Results) > 0 && extResp.Results[0].RawContent != "" {
		content = extResp.Results[0].RawContent
	} else {
		// Fallback sang crawl endpoint
		crawlResp, cErr := t.client.Crawl(ctx, tavily.CrawlRequest{
			URL:          targetURL,
			ExtractDepth: "advanced",
		})
		if cErr != nil {
			fallbackMap := map[string]any{
				"url":         targetURL,
				"content":     "",
				"word_count":  0,
				"warning":     fmt.Sprintf("Không thể trích xuất nội dung từ URL %s (chi tiết: %v). Hãy dựa vào kiến thức sẵn có để viết tiếp.", targetURL, cErr),
				"instruction": "Không gọi lại tool này. Tiếp tục viết dựa trên kiến thức sẵn có mà không phụ thuộc vào liên kết này.",
			}
			data, _ := json.Marshal(fallbackMap)
			return data, nil
		}
		if len(crawlResp.Results) > 0 {
			content = crawlResp.Results[0].RawContent
		}
	}

	content = strings.TrimSpace(content)
	words := domain.WordCount(content)

	resultMap := map[string]any{
		"url":        targetURL,
		"content":    content,
		"word_count": words,
		"instruction": "Hãy sử dụng thông tin chi tiết này để làm sáng tỏ các cơ chế khoa học trong kịch bản và trích dẫn URL vào 'NGUỒN:'.",
	}

	data, err := json.Marshal(resultMap)
	if err != nil {
		return nil, fmt.Errorf("marshal result: %w", err)
	}
	return data, nil
}
