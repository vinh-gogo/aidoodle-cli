package trend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/voocel/ainovel-cli/internal/bootstrap"
	"github.com/voocel/ainovel-cli/internal/store"
)

func TestRunIntake(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.Write([]byte(sampleGoogleTrendsXML))
	}))
	defer ts.Close()

	tempDir := t.TempDir()
	cfg := bootstrap.Config{
		Trends: bootstrap.TrendsConfig{
			Enabled:  true,
			Sources:  []string{ts.URL},
			Geo:      "VN",
			MaxItems: 10,
		},
	}

	fetcher := NewFetcher(ts.Client())
	ctx := context.Background()

	snap, err := RunIntake(ctx, cfg, tempDir, fetcher)
	if err != nil {
		t.Fatalf("RunIntake failed: %v", err)
	}

	if len(snap.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(snap.Items))
	}

	// Xác minh snapshot đã được lưu đĩa
	loaded, err := LoadLatestSnapshot(tempDir)
	if err != nil {
		t.Fatalf("LoadLatestSnapshot failed: %v", err)
	}
	if len(loaded.Items) != 2 {
		t.Errorf("expected 2 items in saved snapshot, got %d", len(loaded.Items))
	}
}

func TestFetchAndPrepareSources(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(sampleArticleHTML))
	}))
	defer ts.Close()

	tempDir := t.TempDir()
	dec := TopicDecision{
		Topic:      "Lãi suất hạ",
		TrendRef:   "trend-1",
		Angle:      "Ẩn dụ kho thịt",
		SourceURLs: []string{ts.URL},
		Reason:     "Chủ đề hot",
	}

	ctx := context.Background()
	sp, err := FetchAndPrepareSources(ctx, ts.Client(), "", tempDir, dec, 1000)
	if err != nil {
		t.Fatalf("FetchAndPrepareSources failed: %v", err)
	}

	if len(sp.Articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(sp.Articles))
	}
	if sp.Topic != "Lãi suất hạ" {
		t.Errorf("expected topic, got %q", sp.Topic)
	}

	// Xác minh source pack đã được lưu đĩa
	loaded, err := LoadSourcePack(tempDir, "trend-1")
	if err != nil {
		t.Fatalf("LoadSourcePack failed: %v", err)
	}
	if loaded.Topic != "Lãi suất hạ" {
		t.Errorf("loaded topic mismatch: %q", loaded.Topic)
	}
}

func TestRecordTopicsDecision(t *testing.T) {
	tempDir := t.TempDir()
	st := store.NewStore(tempDir)
	if err := st.Init(); err != nil {
		t.Fatalf("store Init failed: %v", err)
	}

	now := time.Now()
	snap := &Snapshot{
		ID:        "snap-1",
		Geo:       "VN",
		FetchedAt: now,
		Items:     []Item{{ID: "it1", Title: "Xu hướng 1"}},
	}
	topics := []TopicDecision{
		{Topic: "Chủ đề 1", TrendRef: "it1", Angle: "Góc nhìn", Reason: "Lý do"},
	}

	if err := RecordTopicsDecision(st, "model-test", snap, topics, "Lý do chung"); err != nil {
		t.Fatalf("RecordTopicsDecision failed: %v", err)
	}

	history, err := st.Decisions.Recent(10)
	if err != nil {
		t.Fatalf("Recent failed: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("expected 1 decision record, got %d", len(history))
	}
	if history[0].Kind != "trend_topics" {
		t.Errorf("expected kind 'trend_topics', got %q", history[0].Kind)
	}
}
