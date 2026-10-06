package trend

import (
	"testing"
	"time"
)

func TestSnapshotStore(t *testing.T) {
	tempDir := t.TempDir()
	now := time.Now()

	snap := &Snapshot{
		ID:        "snap-20261006-120000",
		Geo:       "VN",
		FetchedAt: now,
		Items: []Item{
			{ID: "item1", Title: "Giá vàng", Source: "google_trends", URL: "https://example.com/gold"},
			{ID: "item2", Title: "Lãi suất", Source: "rss:vnexpress", URL: "https://example.com/interest"},
		},
	}

	if err := SaveSnapshot(tempDir, snap); err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}

	loaded, err := LoadLatestSnapshot(tempDir)
	if err != nil {
		t.Fatalf("LoadLatestSnapshot failed: %v", err)
	}

	if loaded.ID != snap.ID {
		t.Errorf("expected ID %q, got %q", snap.ID, loaded.ID)
	}
	if len(loaded.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(loaded.Items))
	}
	if loaded.Items[0].Title != "Giá vàng" {
		t.Errorf("unexpected first item: %+v", loaded.Items[0])
	}
}

func TestSourcePackStore(t *testing.T) {
	tempDir := t.TempDir()
	now := time.Now()

	sp := &SourcePack{
		Topic:    "Lạm phát đồ đá",
		TrendRef: "item1",
		Articles: []SourceArticle{
			{
				URL:       "https://example.com/gold",
				Title:     "Giá vàng biến động",
				Content:   "Nội dung bài viết về vàng...",
				FetchedAt: now,
				WordCount: 50,
			},
		},
		Summary:   "Tóm tắt dữ kiện quan trọng",
		FetchedAt: now,
	}

	if err := SaveSourcePack(tempDir, sp); err != nil {
		t.Fatalf("SaveSourcePack failed: %v", err)
	}

	loaded, err := LoadSourcePack(tempDir, "item1")
	if err != nil {
		t.Fatalf("LoadSourcePack failed: %v", err)
	}

	if loaded.Topic != sp.Topic {
		t.Errorf("expected topic %q, got %q", sp.Topic, loaded.Topic)
	}
	if len(loaded.Articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(loaded.Articles))
	}
	if loaded.Articles[0].Title != "Giá vàng biến động" {
		t.Errorf("unexpected article title: %q", loaded.Articles[0].Title)
	}
}
