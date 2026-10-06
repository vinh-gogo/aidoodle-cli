package trend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const sampleArticleHTML = `<!DOCTYPE html>
<html>
<head>
  <title>Ngân hàng trung ương hạ lãi suất điều hành - VnExpress</title>
  <script>console.log("tracking");</script>
  <style>.header { color: red; }</style>
</head>
<body>
  <header><nav>Menu rác</nav></header>
  <h1>Ngân hàng trung ương hạ lãi suất điều hành</h1>
  <p>Ngân hàng Nhà nước quyết định giảm thêm 0,5% lãi suất điều hành từ ngày mai nhằm hỗ trợ thanh khoản.</p>
  <p>Đây là lần giảm lãi suất thứ tư liên tiếp trong năm nhằm kích cầu nền kinh tế và hỗ trợ doanh nghiệp tiếp cận vốn giá rẻ.</p>
  <p>ngắn</p>
  <footer>Bản quyền 2026</footer>
</body>
</html>`

func TestExtractArticle(t *testing.T) {
	title, content := ExtractArticle(sampleArticleHTML, 1000)

	if !strings.Contains(title, "Ngân hàng trung ương hạ lãi suất điều hành") {
		t.Errorf("unexpected title: %q", title)
	}

	if strings.Contains(content, "console.log") {
		t.Errorf("content should not contain script: %q", content)
	}
	if strings.Contains(content, "Menu rác") {
		t.Errorf("content should not contain nav: %q", content)
	}
	if strings.Contains(content, "Bản quyền") {
		t.Errorf("content should not contain footer: %q", content)
	}

	if !strings.Contains(content, "Ngân hàng Nhà nước quyết định giảm thêm 0,5%") {
		t.Errorf("content missing first paragraph: %q", content)
	}
	if !strings.Contains(content, "Đây là lần giảm lãi suất thứ tư") {
		t.Errorf("content missing second paragraph: %q", content)
	}
}

func TestFetchArticle(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(sampleArticleHTML))
	}))
	defer ts.Close()

	ctx := context.Background()
	article, err := FetchArticle(ctx, ts.Client(), "", ts.URL, 1000)
	if err != nil {
		t.Fatalf("FetchArticle failed: %v", err)
	}

	if !strings.Contains(article.Title, "Ngân hàng trung ương") {
		t.Errorf("unexpected article title: %q", article.Title)
	}
	if article.WordCount <= 0 {
		t.Errorf("expected positive word count, got %d", article.WordCount)
	}
	if article.URL != ts.URL {
		t.Errorf("expected URL %q, got %q", ts.URL, article.URL)
	}
}
