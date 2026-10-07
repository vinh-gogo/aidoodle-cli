package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/voocel/agentcore/schema"
	"github.com/voocel/ainovel-cli/internal/tavily"
)

// TavilySearchTool cho phép LLM (Architect, Writer) tra cứu dữ liệu khoa học và nguồn tin cậy trên internet.
type TavilySearchTool struct {
	client *tavily.Client
}

// NewTavilySearchTool khởi tạo tool tavily_search.
func NewTavilySearchTool(client *tavily.Client) *TavilySearchTool {
	if client == nil {
		client = tavily.NewClient("")
	}
	return &TavilySearchTool{client: client}
}

func (t *TavilySearchTool) Name() string { return "tavily_search" }
func (t *TavilySearchTool) Description() string {
	return "Tra cứu internet qua Tavily Search để lấy cơ sở khoa học, bài báo từ nguồn uy tín (VietnamPlus, Dân trí, KhoaHoc.tv, Nature...), số liệu thực tế và các luồng tranh luận học thuật. Kết quả trả về gồm câu trả lời tổng hợp và danh sách bài viết để bạn trích dẫn cụ thể vào phần 'NGUỒN:' và ghi các điểm chưa ngã ngũ vào 'CẦN KIỂM CHỨNG:' ở chân kịch bản."
}
func (t *TavilySearchTool) Label() string { return "Tìm kiếm Tavily" }

func (t *TavilySearchTool) ReadOnly(_ json.RawMessage) bool        { return true }
func (t *TavilySearchTool) ConcurrencySafe(_ json.RawMessage) bool { return true }

func (t *TavilySearchTool) Schema() map[string]any {
	return schema.Object(
		schema.Property("query", schema.String("Từ khóa hoặc câu hỏi cần tra cứu khoa học (ví dụ: 'vì sao con người mất lông tiến hóa mồ hôi')")).Required(),
		schema.Property("search_depth", schema.Enum("Độ sâu tìm kiếm", "basic", "advanced")),
		schema.Property("max_results", schema.Int("Số lượng bài viết tối đa cần lấy (1-10, mặc định 5)")),
	)
}

func (t *TavilySearchTool) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var a struct {
		Query       string `json:"query"`
		SearchDepth string `json:"search_depth"`
		MaxResults  int    `json:"max_results"`
	}
	if err := json.Unmarshal(args, &a); err != nil {
		return nil, fmt.Errorf("tham số không hợp lệ: %w", err)
	}

	query := strings.TrimSpace(a.Query)
	if query == "" {
		return nil, fmt.Errorf("query không được để trống")
	}

	depth := a.SearchDepth
	if depth == "" {
		depth = "advanced"
	}
	maxRes := a.MaxResults
	if maxRes <= 0 {
		maxRes = 5
	}

	resp, err := t.client.Search(ctx, tavily.SearchRequest{
		Query:         query,
		SearchDepth:   depth,
		MaxResults:    maxRes,
		IncludeAnswer: true,
	})
	if err != nil {
		return nil, fmt.Errorf("tra cứu Tavily thất bại: %w", err)
	}

	type searchItem struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
	}

	var items []searchItem
	for _, r := range resp.Results {
		items = append(items, searchItem{
			Title:   r.Title,
			URL:     r.URL,
			Content: r.Content,
		})
	}

	resultMap := map[string]any{
		"query":        resp.Query,
		"answer":       resp.Answer,
		"results":      items,
		"total_found":  len(items),
		"instruction": "HÃY DÙNG CÁC KẾT QUẢ NÀY ĐỂ: (1) Trích dẫn cụ thể tên bài báo, tác giả/tổ chức và đường link URL vào thẻ 'NGUỒN:' ở cuối kịch bản; (2) Trích xuất các số liệu thực tế và luận điểm khoa học vào nội dung kịch bản; (3) Liệt kê các giả thuyết đối lập hoặc điểm còn đang tranh luận vào thẻ 'CẦN KIỂM CHỨNG:'.",
	}

	data, err := json.Marshal(resultMap)
	if err != nil {
		return nil, fmt.Errorf("marshal result: %w", err)
	}
	return data, nil
}
