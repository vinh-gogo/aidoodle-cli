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

	type searchItem struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
	}

	// Nếu client chưa có API key hoặc client nil, phản hồi mềm để LLM dùng tri thức có sẵn thay vì báo lỗi đứt luồng.
	if t.client == nil || strings.TrimSpace(t.client.APIKey) == "" {
		resultMap := map[string]any{
			"query":       query,
			"answer":      "Dịch vụ Tavily chưa được cấu hình API key (hoặc không kích hoạt). Hãy vận dụng trực tiếp nguồn tri thức khoa học, tâm lý học và học thuật sâu rộng sẵn có của bạn để giải thích cơ chế, trích dẫn tác giả nghiên cứu kinh điển và hoàn thành kịch bản mà KHÔNG CẦN gọi lại tavily_search.",
			"results":     []searchItem{},
			"total_found": 0,
			"warning":     "Chưa cấu hình TAVILY_API_KEY. Tiếp tục sáng tác dựa trên tri thức sẵn có.",
			"instruction": "Không gọi lại tavily_search. Hãy tiến hành viết nội dung ngay dựa trên cơ sở kiến thức học thuật sẵn có.",
		}
		data, _ := json.Marshal(resultMap)
		return data, nil
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
		// Thay vì báo lỗi làm đứt gãy luồng ReAct (khiến LLM thử lại liên tục dẫn tới bế tắc),
		// trả về kết quả dự phòng hướng dẫn LLM tiếp tục sáng tác dựa trên tri thức sẵn có.
		fallbackMap := map[string]any{
			"query":       query,
			"answer":      fmt.Sprintf("Dịch vụ tra cứu Tavily hiện không khả dụng (%v). Bạn hãy sử dụng nguồn tri thức khoa học, tâm lý học và học thuật sâu rộng sẵn có của bạn để giải thích các cơ chế, trích dẫn tác giả nghiên cứu kinh điển và hoàn thành kịch bản mà KHÔNG CẦN gọi lại tavily_search.", err),
			"results":     []searchItem{},
			"total_found": 0,
			"warning":     fmt.Sprintf("Tra cứu Tavily không thành công: %v. Hãy tiếp tục sáng tác dựa trên tri thức sẵn có.", err),
			"instruction": "Không gọi lại tavily_search. Hãy tiến hành viết nội dung ngay dựa trên cơ sở kiến thức học thuật sẵn có.",
		}
		data, _ := json.Marshal(fallbackMap)
		return data, nil
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
