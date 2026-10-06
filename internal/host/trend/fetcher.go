package trend

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	defaultUserAgent = "Mozilla/5.0 (compatible; ainovel-cli/1.0; +https://github.com/voocel/ainovel-cli)"
	defaultTimeout   = 10 * time.Second
	maxRSSReadBytes  = 5 << 20 // 5 MB
)

// Fetcher bộ thu thập xu hướng từ các nguồn RSS.
type Fetcher struct {
	Client    *http.Client
	UserAgent string
}

// NewFetcher khởi tạo một Fetcher với client mặc định.
func NewFetcher(client *http.Client) *Fetcher {
	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}
	return &Fetcher{
		Client:    client,
		UserAgent: defaultUserAgent,
	}
}

// XML structures for Google Trends RSS
type googleTrendsRSS struct {
	XMLName xml.Name            `xml:"rss"`
	Channel googleTrendsChannel `xml:"channel"`
}

type googleTrendsChannel struct {
	Title string             `xml:"title"`
	Items []googleTrendsItem `xml:"item"`
}

type googleTrendsItem struct {
	Title         string                 `xml:"title"`
	ApproxTraffic string                 `xml:"approx_traffic"`
	Description   string                 `xml:"description"`
	Link          string                 `xml:"link"`
	PubDate       string                 `xml:"pubDate"`
	NewsItems     []googleTrendsNewsItem `xml:"news_item"`
}

type googleTrendsNewsItem struct {
	NewsItemTitle   string `xml:"news_item_title"`
	NewsItemSnippet string `xml:"news_item_snippet"`
	NewsItemURL     string `xml:"news_item_url"`
	NewsItemSource  string `xml:"news_item_source"`
}

// XML structures for standard RSS
type standardRSS struct {
	XMLName xml.Name        `xml:"rss"`
	Channel standardChannel `xml:"channel"`
}

type standardChannel struct {
	Title string         `xml:"title"`
	Items []standardItem `xml:"item"`
}

type standardItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

// StripHTML loại bỏ các thẻ HTML và giải mã ký tự HTML entities.
func StripHTML(s string) string {
	clean := htmlTagRe.ReplaceAllString(s, " ")
	clean = html.UnescapeString(clean)
	return strings.Join(strings.Fields(clean), " ")
}

func itemID(title, urlStr string) string {
	h := sha256.New()
	h.Write([]byte(title + "|" + urlStr))
	return hex.EncodeToString(h.Sum(nil))[:12]
}

// ParseGoogleTrends phân tích XML của Google Trends RSS thành danh sách Item.
func ParseGoogleTrends(data []byte, fetchedAt time.Time) ([]Item, error) {
	var feed googleTrendsRSS
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf("giải mã Google Trends XML: %w", err)
	}

	var items []Item
	for _, raw := range feed.Channel.Items {
		title := strings.TrimSpace(raw.Title)
		if title == "" {
			continue
		}
		urlStr := strings.TrimSpace(raw.Link)
		snippet := StripHTML(raw.Description)

		// Nếu có news_items con, lấy URL và snippet từ bài báo cụ thể đầu tiên
		if len(raw.NewsItems) > 0 {
			first := raw.NewsItems[0]
			if first.NewsItemURL != "" {
				urlStr = strings.TrimSpace(first.NewsItemURL)
			}
			if first.NewsItemSnippet != "" {
				snippet = StripHTML(first.NewsItemSnippet)
			}
		}

		items = append(items, Item{
			ID:        itemID(title, urlStr),
			Title:     title,
			Source:    "google_trends",
			URL:       urlStr,
			Snippet:   snippet,
			Traffic:   strings.TrimSpace(raw.ApproxTraffic),
			PubDate:   strings.TrimSpace(raw.PubDate),
			FetchedAt: fetchedAt,
		})
	}
	return items, nil
}

// ParseStandardRSS phân tích XML của RSS thông thường (VnExpress, Tuổi Trẻ...) thành danh sách Item.
func ParseStandardRSS(data []byte, sourceName string, fetchedAt time.Time) ([]Item, error) {
	var feed standardRSS
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf("giải mã RSS XML (%s): %w", sourceName, err)
	}

	var items []Item
	for _, raw := range feed.Channel.Items {
		title := strings.TrimSpace(raw.Title)
		if title == "" {
			continue
		}
		urlStr := strings.TrimSpace(raw.Link)
		snippet := StripHTML(raw.Description)

		items = append(items, Item{
			ID:        itemID(title, urlStr),
			Title:     title,
			Source:    sourceName,
			URL:       urlStr,
			Snippet:   snippet,
			PubDate:   strings.TrimSpace(raw.PubDate),
			FetchedAt: fetchedAt,
		})
	}
	return items, nil
}

// fetchURL thực hiện GET request tới URL với giới hạn kích thước đọc.
func (f *Fetcher) fetchURL(ctx context.Context, urlStr string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", f.UserAgent)
	req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml, */*")

	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status %d khi tải %s", resp.StatusCode, urlStr)
	}

	lr := io.LimitReader(resp.Body, maxRSSReadBytes)
	return io.ReadAll(lr)
}

// ResolveSourceURL chuyển đổi tên nguồn dạng rút gọn thành URL thực tế.
func ResolveSourceURL(source, geo string) (url string, isGoogleTrends bool) {
	source = strings.TrimSpace(source)
	lower := strings.ToLower(source)
	switch {
	case lower == "google_trends" || lower == "google":
		g := strings.ToUpper(strings.TrimSpace(geo))
		if g == "" {
			g = "VN"
		}
		return fmt.Sprintf("https://trends.google.com/trending/rss?geo=%s", g), true
	case lower == "rss:vnexpress" || lower == "vnexpress":
		return "https://vnexpress.net/rss/tin-moi-nhat.rss", false
	case lower == "rss:tuoitre" || lower == "tuoitre":
		return "https://tuoitre.vn/rss/tin-moi-nhat.rss", false
	case lower == "rss:thanhnien" || lower == "thanhnien":
		return "https://thanhnien.vn/rss/home.rss", false
	case strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://"):
		return source, false
	default:
		return source, false
	}
}

// FetchSource tải và phân tích một nguồn xu hướng cụ thể.
func (f *Fetcher) FetchSource(ctx context.Context, source, geo string) ([]Item, error) {
	urlStr, isGoogleTrends := ResolveSourceURL(source, geo)
	data, err := f.fetchURL(ctx, urlStr)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if isGoogleTrends {
		return ParseGoogleTrends(data, now)
	}
	return ParseStandardRSS(data, source, now)
}

// FetchAll thu thập xu hướng từ tất cả các nguồn cấu hình, gộp và khử trùng lặp theo tiêu đề.
func (f *Fetcher) FetchAll(ctx context.Context, sources []string, geo string, maxItems int) (*Snapshot, error) {
	if len(sources) == 0 {
		sources = []string{"google_trends", "rss:vnexpress"}
	}
	if maxItems <= 0 {
		maxItems = 30
	}

	now := time.Now()
	seenTitles := make(map[string]struct{})
	var allItems []Item

	for _, src := range sources {
		items, err := f.FetchSource(ctx, src, geo)
		if err != nil {
			// Một nguồn lỗi không được làm đổ cả mẻ thu thập
			continue
		}
		for _, it := range items {
			key := strings.ToLower(strings.TrimSpace(it.Title))
			if _, exists := seenTitles[key]; exists {
				continue
			}
			seenTitles[key] = struct{}{}
			allItems = append(allItems, it)
			if len(allItems) >= maxItems {
				break
			}
		}
		if len(allItems) >= maxItems {
			break
		}
	}

	snap := &Snapshot{
		ID:        fmt.Sprintf("snap-%s", now.Format("20060102-150405")),
		Geo:       geo,
		FetchedAt: now,
		Items:     allItems,
	}
	return snap, nil
}
