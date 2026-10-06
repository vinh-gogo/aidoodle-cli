package trend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const sampleGoogleTrendsXML = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:ht="https://trends.google.com/trends/trendingsearches/daily">
  <channel>
    <title>Daily Search Trends</title>
    <item>
      <title>Giá vàng hôm nay</title>
      <ht:approx_traffic>50K+</ht:approx_traffic>
      <description>Cập nhật giá vàng miếng và vàng nhẫn 9999</description>
      <link>https://trends.google.com/trending?geo=VN</link>
      <pubDate>Mon, 06 Oct 2026 04:00:00 +0000</pubDate>
      <ht:news_item>
        <ht:news_item_title>Giá vàng trong nước biến động mạnh</ht:news_item_title>
        <ht:news_item_snippet>Giá vàng miếng SJC sáng nay tăng nhẹ theo đà thế giới...</ht:news_item_snippet>
        <ht:news_item_url>https://example.com/gia-vang-bien-dong</ht:news_item_url>
        <ht:news_item_source>VnExpress</ht:news_item_source>
      </ht:news_item>
    </item>
    <item>
      <title>iPhone 18 Pro</title>
      <ht:approx_traffic>20K+</ht:approx_traffic>
      <link>https://trends.google.com/trending?geo=VN</link>
      <pubDate>Mon, 06 Oct 2026 03:00:00 +0000</pubDate>
    </item>
  </channel>
</rss>`

const sampleVnExpressRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>VnExpress - Tin mới nhất</title>
    <item>
      <title>Ngân hàng trung ương hạ lãi suất điều hành</title>
      <link>https://vnexpress.net/ngan-hang-trung-uong-ha-lai-suat.html</link>
      <description><![CDATA[<a href="..."><img src="..."/></a>Ngân hàng Nhà nước quyết định giảm thêm 0,5% lãi suất điều hành từ ngày mai.]]></description>
      <pubDate>Mon, 06 Oct 2026 08:30:00 +0700</pubDate>
    </item>
  </channel>
</rss>`

func TestParseGoogleTrends(t *testing.T) {
	now := time.Now()
	items, err := ParseGoogleTrends([]byte(sampleGoogleTrendsXML), now)
	if err != nil {
		t.Fatalf("ParseGoogleTrends failed: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	item1 := items[0]
	if item1.Title != "Giá vàng hôm nay" {
		t.Errorf("expected title 'Giá vàng hôm nay', got %q", item1.Title)
	}
	if item1.Traffic != "50K+" {
		t.Errorf("expected traffic '50K+', got %q", item1.Traffic)
	}
	if item1.URL != "https://example.com/gia-vang-bien-dong" {
		t.Errorf("expected URL from news_item, got %q", item1.URL)
	}
	if item1.Snippet == "" {
		t.Errorf("expected non-empty snippet")
	}

	item2 := items[1]
	if item2.Title != "iPhone 18 Pro" {
		t.Errorf("expected title 'iPhone 18 Pro', got %q", item2.Title)
	}
	if item2.Traffic != "20K+" {
		t.Errorf("expected traffic '20K+', got %q", item2.Traffic)
	}
}

func TestParseStandardRSS(t *testing.T) {
	now := time.Now()
	items, err := ParseStandardRSS([]byte(sampleVnExpressRSS), "rss:vnexpress", now)
	if err != nil {
		t.Fatalf("ParseStandardRSS failed: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	item := items[0]
	if item.Title != "Ngân hàng trung ương hạ lãi suất điều hành" {
		t.Errorf("expected title, got %q", item.Title)
	}
	if item.URL != "https://vnexpress.net/ngan-hang-trung-uong-ha-lai-suat.html" {
		t.Errorf("expected URL, got %q", item.URL)
	}
	if item.Source != "rss:vnexpress" {
		t.Errorf("expected source 'rss:vnexpress', got %q", item.Source)
	}
	// Snippet phải được strip HTML (loại bỏ thẻ <a>, <img>)
	if item.Snippet != "Ngân hàng Nhà nước quyết định giảm thêm 0,5% lãi suất điều hành từ ngày mai." {
		t.Errorf("expected clean snippet, got %q", item.Snippet)
	}
}

func TestFetcher_FetchAll(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.Write([]byte(sampleGoogleTrendsXML))
	}))
	defer ts.Close()

	fetcher := NewFetcher(ts.Client())
	ctx := context.Background()

	// Truyền trực tiếp URL của test server
	snap, err := fetcher.FetchAll(ctx, []string{ts.URL}, "VN", 10)
	if err != nil {
		t.Fatalf("FetchAll failed: %v", err)
	}

	if len(snap.Items) != 2 {
		t.Fatalf("expected 2 items in snapshot, got %d", len(snap.Items))
	}
	if snap.Items[0].Title != "Giá vàng hôm nay" {
		t.Errorf("unexpected first item: %+v", snap.Items[0])
	}
}
