package arbiter

import (
	"context"
	"fmt"
	"strings"

	"github.com/voocel/agentcore"
	"github.com/voocel/agentcore/schema"
	"github.com/voocel/ainovel-cli/internal/host/trend"
	"github.com/voocel/ainovel-cli/internal/llmcontract"
)

// TopicsDecision gói phán quyết lựa chọn chủ đề từ danh sách xu hướng.
type TopicsDecision struct {
	Topics []trend.TopicDecision `json:"topics"`
	Reason string                `json:"reason"`
}

// Validate kiểm tra tính hợp lệ cơ học của TopicsDecision.
func (d *TopicsDecision) Validate() error {
	if len(d.Topics) == 0 {
		return fmt.Errorf("danh sách chủ đề được chọn không được để trống")
	}
	if strings.TrimSpace(d.Reason) == "" {
		return fmt.Errorf("reason không được để trống")
	}
	for i, t := range d.Topics {
		if strings.TrimSpace(t.Topic) == "" {
			return fmt.Errorf("chủ đề thứ %d: topic không được để trống", i+1)
		}
		if strings.TrimSpace(t.TrendRef) == "" {
			return fmt.Errorf("chủ đề thứ %d: trend_ref không được để trống", i+1)
		}
		if strings.TrimSpace(t.Angle) == "" {
			return fmt.Errorf("chủ đề thứ %d: angle không được để trống", i+1)
		}
		if strings.TrimSpace(t.Reason) == "" {
			return fmt.Errorf("chủ đề thứ %d: reason không được để trống", i+1)
		}
	}
	return nil
}

var topicsContract = llmcontract.Contract{
	Name:        "arbiter_trend_topics",
	Description: "Phán quyết lựa chọn chủ đề kịch bản video TikTok doodle explainer từ các xu hướng",
	Schema: schema.Object(
		schema.Property("topics", schema.Array("Danh sách các chủ đề được lựa chọn", schema.Object(
			schema.Property("topic", schema.String("Tên chủ đề kịch bản đề xuất")).Required(),
			schema.Property("trend_ref", schema.String("ID hoặc tiêu đề của xu hướng nguồn đã cho")).Required(),
			schema.Property("angle", schema.String("Góc nhìn giải thích bằng ẩn dụ người que thời đồ đá")).Required(),
			schema.Property("source_urls", schema.Array("Các URL bài báo nguồn để khai thác dữ kiện", schema.String("URL bài báo"))),
			schema.Property("reason", schema.String("Lý do lựa chọn")).Required(),
		))).Required(),
		schema.Property("reason", schema.String("Lý do tổng quan lựa chọn bộ chủ đề này")).Required(),
	),
}

type topicsPayload struct {
	Count       int             `json:"count"`
	Preferences string          `json:"preferences,omitempty"`
	Items       []trendItemView `json:"items"`
}

type trendItemView struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Source  string `json:"source"`
	Snippet string `json:"snippet,omitempty"`
	Traffic string `json:"traffic,omitempty"`
	URL     string `json:"url,omitempty"`
}

// DecideTopics lựa chọn các chủ đề nóng phù hợp nhất từ snapshot xu hướng để làm kịch bản doodle explainer.
func DecideTopics(ctx context.Context, model agentcore.ChatModel, systemPrompt string, snapshot *trend.Snapshot, count int, preferences string) (TopicsDecision, error) {
	if snapshot == nil || len(snapshot.Items) == 0 {
		return TopicsDecision{}, fmt.Errorf("snapshot xu hướng rỗng")
	}
	if count <= 0 {
		count = 8
	}

	views := make([]trendItemView, 0, len(snapshot.Items))
	itemMap := make(map[string]trend.Item)
	for _, it := range snapshot.Items {
		views = append(views, trendItemView{
			ID:      it.ID,
			Title:   it.Title,
			Source:  it.Source,
			Snippet: it.Snippet,
			Traffic: it.Traffic,
			URL:     it.URL,
		})
		itemMap[it.ID] = it
		itemMap[strings.ToLower(strings.TrimSpace(it.Title))] = it
	}

	payload, err := marshalPayload(topicsPayload{
		Count:       count,
		Preferences: preferences,
		Items:       views,
	})
	if err != nil {
		return TopicsDecision{}, err
	}

	dec, err := decide(ctx, model, topicsContract, systemPrompt, payload, (*TopicsDecision).Validate)
	if err != nil {
		return TopicsDecision{}, err
	}

	// Hậu kiểm cơ học: mỗi trend_ref phải khớp với một item có thật trong snapshot
	for i := range dec.Topics {
		ref := dec.Topics[i].TrendRef
		matched, exists := itemMap[ref]
		if !exists {
			matched, exists = itemMap[strings.ToLower(strings.TrimSpace(ref))]
		}
		if !exists {
			return TopicsDecision{}, fmt.Errorf("chủ đề %q chứa trend_ref %q không tồn tại trong danh sách xu hướng đầu vào", dec.Topics[i].Topic, ref)
		}
		// Chuẩn hóa trend_ref về ID và gắn URL nếu thiếu
		dec.Topics[i].TrendRef = matched.ID
		if len(dec.Topics[i].SourceURLs) == 0 && matched.URL != "" {
			dec.Topics[i].SourceURLs = []string{matched.URL}
		}
	}

	// Giới hạn số lượng chủ đề tối đa
	if len(dec.Topics) > count {
		dec.Topics = dec.Topics[:count]
	}

	return dec, nil
}
