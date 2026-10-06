package trend

import "time"

// Item đại diện cho một mục xu hướng thô thu thập được từ RSS hoặc Google Trends.
type Item struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Source    string    `json:"source"` // "google_trends", "rss:vnexpress", v.v.
	URL       string    `json:"url"`
	Snippet   string    `json:"snippet,omitempty"`
	Traffic   string    `json:"traffic,omitempty"` // Lượng tìm kiếm ước tính (Google Trends, ví dụ "10K+")
	PubDate   string    `json:"pub_date,omitempty"`
	FetchedAt time.Time `json:"fetched_at"`
}

// Snapshot ảnh chụp tập hợp các xu hướng tại một thời điểm nhất định.
type Snapshot struct {
	ID        string    `json:"id"`
	Geo       string    `json:"geo"`
	FetchedAt time.Time `json:"fetched_at"`
	Items     []Item    `json:"items"`
}

// SourceArticle nội dung văn bản trích xuất từ bài báo/nguồn tin để làm tư liệu dữ kiện.
type SourceArticle struct {
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Content   string    `json:"content"` // Văn bản thuần túy đã làm sạch (tối đa N ký tự)
	FetchedAt time.Time `json:"fetched_at"`
	WordCount int       `json:"word_count"`
}

// SourcePack gói tài liệu nguồn hoàn chỉnh cho một chủ đề kịch bản, dùng để nạp vào novel_context.
type SourcePack struct {
	Topic     string          `json:"topic"`
	TrendRef  string          `json:"trend_ref"`
	Articles  []SourceArticle `json:"articles"`
	Summary   string          `json:"summary,omitempty"`
	FetchedAt time.Time       `json:"fetched_at"`
}

// TopicDecision phán quyết lựa chọn chủ đề của Arbiter từ snapshot xu hướng.
type TopicDecision struct {
	Topic      string   `json:"topic"`       // Tên chủ đề kịch bản đề xuất
	TrendRef   string   `json:"trend_ref"`   // Mã hoặc tiêu đề xu hướng gốc được chọn
	Angle      string   `json:"angle"`       // Góc nhìn giải thích bằng ẩn dụ người que thời đồ đá
	SourceURLs []string `json:"source_urls"` // Các URL bài báo nguồn để khai thác dữ kiện
	Reason     string   `json:"reason"`      // Lý do lựa chọn và tiềm năng viral/giải thích
}
