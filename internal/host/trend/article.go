package trend

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/utils"
)

const (
	defaultMaxArticleChars = 6000
	maxHTMLReadBytes       = 2 << 20 // 2 MB
)

var (
	stripTagsRe = []*regexp.Regexp{
		regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`),
		regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`),
		regexp.MustCompile(`(?is)<nav[^>]*>.*?</nav>`),
		regexp.MustCompile(`(?is)<header[^>]*>.*?</header>`),
		regexp.MustCompile(`(?is)<footer[^>]*>.*?</footer>`),
		regexp.MustCompile(`(?is)<aside[^>]*>.*?</aside>`),
		regexp.MustCompile(`(?is)<noscript[^>]*>.*?</noscript>`),
		regexp.MustCompile(`(?is)<iframe[^>]*>.*?</iframe>`),
	}
	titleTagRe = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	h1TagRe    = regexp.MustCompile(`(?is)<h1[^>]*>(.*?)</h1>`)
	pTagRe     = regexp.MustCompile(`(?is)<p[^>]*>(.*?)</p>`)
	brTagRe    = regexp.MustCompile(`(?i)<br\s*/?>`)
)

// ExtractArticle trích xuất tiêu đề và văn bản chính từ HTML của một bài viết.
func ExtractArticle(htmlStr string, maxChars int) (title string, content string) {
	if maxChars <= 0 {
		maxChars = defaultMaxArticleChars
	}

	// 1. Trích xuất tiêu đề từ <title> hoặc <h1>
	if m := h1TagRe.FindStringSubmatch(htmlStr); len(m) > 1 {
		title = StripHTML(m[1])
	} else if m := titleTagRe.FindStringSubmatch(htmlStr); len(m) > 1 {
		title = StripHTML(m[1])
	}

	// 2. Loại bỏ các khối không phải nội dung (script, style, nav...)
	cleaned := htmlStr
	for _, re := range stripTagsRe {
		cleaned = re.ReplaceAllString(cleaned, "")
	}
	cleaned = brTagRe.ReplaceAllString(cleaned, "\n")

	// 3. Tìm các đoạn <p>
	var paragraphs []string
	matches := pTagRe.FindAllStringSubmatch(cleaned, -1)
	for _, m := range matches {
		if len(m) > 1 {
			pText := strings.TrimSpace(StripHTML(m[1]))
			// Bỏ qua các đoạn quá ngắn (thường là chú thích ảnh hoặc menu rác)
			if len([]rune(pText)) >= 20 {
				paragraphs = append(paragraphs, pText)
			}
		}
	}

	// Nếu không bắt được thẻ <p> nào (ví dụ trang cấu trúc lạ), bóc toàn bộ thẻ HTML
	if len(paragraphs) == 0 {
		raw := htmlTagRe.ReplaceAllString(cleaned, "\n")
		raw = html.UnescapeString(raw)
		lines := strings.Split(raw, "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if len([]rune(l)) >= 30 {
				paragraphs = append(paragraphs, l)
			}
		}
	}

	content = strings.Join(paragraphs, "\n\n")
	if len([]rune(content)) > maxChars {
		content = utils.TruncateRunes(content, maxChars)
	}

	return strings.TrimSpace(title), strings.TrimSpace(content)
}

// FetchArticle tải nội dung một bài báo từ URL và trích xuất thành SourceArticle.
func FetchArticle(ctx context.Context, client *http.Client, userAgent, urlStr string, maxChars int) (SourceArticle, error) {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	if userAgent == "" {
		userAgent = defaultUserAgent
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return SourceArticle{}, fmt.Errorf("tạo request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return SourceArticle{}, fmt.Errorf("kết nối tới %s: %w", urlStr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SourceArticle{}, fmt.Errorf("HTTP status %d khi tải %s", resp.StatusCode, urlStr)
	}

	lr := io.LimitReader(resp.Body, maxHTMLReadBytes)
	bodyBytes, err := io.ReadAll(lr)
	if err != nil {
		return SourceArticle{}, fmt.Errorf("đọc nội dung bài báo: %w", err)
	}

	title, content := ExtractArticle(string(bodyBytes), maxChars)
	now := time.Now()

	return SourceArticle{
		URL:       urlStr,
		Title:     title,
		Content:   content,
		FetchedAt: now,
		WordCount: domain.WordCount(content),
	}, nil
}
