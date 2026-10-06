package arbiter

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/voocel/ainovel-cli/internal/host/trend"
)

func TestDecideTopics_Success(t *testing.T) {
	now := time.Now()
	snap := &trend.Snapshot{
		ID:        "snap-1",
		Geo:       "VN",
		FetchedAt: now,
		Items: []trend.Item{
			{ID: "ref-gold", Title: "Giá vàng hôm nay tăng kỷ lục", Source: "google_trends", URL: "https://example.com/gold"},
			{ID: "ref-ai", Title: "Mô hình AI mới có thể tư duy như người", Source: "rss:vnexpress", URL: "https://example.com/ai"},
			{ID: "ref-xsmb", Title: "Kết quả xổ số miền Bắc hôm nay", Source: "google_trends"},
		},
	}

	modelResponse := `{
		"topics": [
			{
				"topic": "Tại sao Người Que lại tranh nhau tích trữ vỏ sò vàng?",
				"trend_ref": "ref-gold",
				"angle": "Người Que thấy tù trưởng tích vỏ sò óng ánh bèn đua nhau gom lại, ai ngờ hang đá chứa không xuể mà thịt khủng long thì không đổi được",
				"source_urls": ["https://example.com/gold"],
				"reason": "Chủ đề tài chính nóng hổi, ẩn dụ vỏ sò dễ hiểu, giải thích tâm lý đầu cơ và quy luật cung cầu"
			},
			{
				"topic": "Khi Người Que chế tạo ra hòn đá biết suy nghĩ",
				"trend_ref": "Mô hình AI mới có thể tư duy như người",
				"angle": "Người Que tìm được hòn đá biết trả lời mọi câu hỏi, nhưng nó lại bắt đầu đòi chia thịt và không chịu làm việc",
				"source_urls": ["https://example.com/ai"],
				"reason": "Chủ đề công nghệ AI rất viral trên TikTok, tương phản giữa công nghệ cao và thời đồ đá tạo tiếng cười"
			}
		],
		"reason": "Đã loại bỏ tin xổ số xsmb vì không có giá trị kiến thức; chọn 2 chủ đề tài chính và AI có tiềm năng viral cao nhất"
	}`

	model := &scriptedModel{outputs: []string{modelResponse}}
	ctx := context.Background()

	dec, err := DecideTopics(ctx, model, "prompt", snap, 5, "")
	if err != nil {
		t.Fatalf("DecideTopics failed: %v", err)
	}

	if len(dec.Topics) != 2 {
		t.Fatalf("expected 2 topics, got %d", len(dec.Topics))
	}

	// Topic 1
	if dec.Topics[0].TrendRef != "ref-gold" {
		t.Errorf("expected TrendRef 'ref-gold', got %q", dec.Topics[0].TrendRef)
	}
	if len(dec.Topics[0].SourceURLs) != 1 || dec.Topics[0].SourceURLs[0] != "https://example.com/gold" {
		t.Errorf("unexpected SourceURLs: %v", dec.Topics[0].SourceURLs)
	}

	// Topic 2: trend_ref ban đầu là title, được chuẩn hóa về ID
	if dec.Topics[1].TrendRef != "ref-ai" {
		t.Errorf("expected TrendRef normalized to 'ref-ai', got %q", dec.Topics[1].TrendRef)
	}
}

func TestDecideTopics_HallucinatedRef(t *testing.T) {
	now := time.Now()
	snap := &trend.Snapshot{
		ID:        "snap-1",
		Geo:       "VN",
		FetchedAt: now,
		Items: []trend.Item{
			{ID: "ref-gold", Title: "Giá vàng hôm nay", Source: "google_trends"},
		},
	}

	// Model tự bịa ra một trend_ref không có thật
	modelResponse := `{
		"topics": [
			{
				"topic": "Chủ đề bịa",
				"trend_ref": "ref-not-exist",
				"angle": "Ẩn dụ đồ đá",
				"reason": "Lý do"
			}
		],
		"reason": "Tổng quan"
	}`

	model := &scriptedModel{outputs: []string{modelResponse}}
	ctx := context.Background()

	_, err := DecideTopics(ctx, model, "prompt", snap, 5, "")
	if err == nil {
		t.Fatal("expected error for hallucinated trend_ref, got nil")
	}
	if !strings.Contains(err.Error(), "không tồn tại trong danh sách") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDecideTopics_EmptySnapshot(t *testing.T) {
	model := &scriptedModel{outputs: []string{}}
	ctx := context.Background()

	_, err := DecideTopics(ctx, model, "prompt", nil, 5, "")
	if err == nil {
		t.Fatal("expected error for nil snapshot, got nil")
	}
}
