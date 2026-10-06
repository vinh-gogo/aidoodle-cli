package trend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/voocel/ainovel-cli/internal/bootstrap"
	"github.com/voocel/ainovel-cli/internal/store"
)

// RunIntake thu thập xu hướng từ các nguồn cấu hình và lưu vào meta/trends/.
func RunIntake(ctx context.Context, cfg bootstrap.Config, outputDir string, fetcher *Fetcher) (*Snapshot, error) {
	if fetcher == nil {
		fetcher = NewFetcher(nil)
	}

	sources := cfg.Trends.Sources
	if len(sources) == 0 {
		sources = []string{"google_trends", "rss:vnexpress"}
	}
	geo := cfg.Trends.Geo
	if geo == "" {
		geo = "VN"
	}
	maxItems := cfg.Trends.MaxItems
	if maxItems <= 0 {
		maxItems = 30
	}

	snap, err := fetcher.FetchAll(ctx, sources, geo, maxItems)
	if err != nil {
		return nil, fmt.Errorf("thu thập xu hướng thất bại: %w", err)
	}

	if err := SaveSnapshot(outputDir, snap); err != nil {
		return nil, fmt.Errorf("lưu snapshot xu hướng thất bại: %w", err)
	}

	return snap, nil
}

// FetchAndPrepareSources tải bài báo từ các URL nguồn của một chủ đề và lưu thành SourcePack.
func FetchAndPrepareSources(ctx context.Context, client *http.Client, userAgent, outputDir string, dec TopicDecision, maxChars int) (*SourcePack, error) {
	if maxChars <= 0 {
		maxChars = defaultMaxArticleChars
	}

	var articles []SourceArticle
	for _, u := range dec.SourceURLs {
		if u == "" {
			continue
		}
		art, err := FetchArticle(ctx, client, userAgent, u, maxChars)
		if err != nil {
			// Bỏ qua lỗi từng bài báo, thu thập tối đa có thể
			continue
		}
		articles = append(articles, art)
	}

	sp := &SourcePack{
		Topic:     dec.Topic,
		TrendRef:  dec.TrendRef,
		Articles:  articles,
		Summary:   fmt.Sprintf("Góc nhìn: %s | Lý do: %s", dec.Angle, dec.Reason),
		FetchedAt: time.Now(),
	}

	if err := SaveSourcePack(outputDir, sp); err != nil {
		return nil, fmt.Errorf("lưu source pack: %w", err)
	}

	return sp, nil
}

// RecordTopicsDecision lưu quyết định chọn chủ đề của Arbiter vào meta/decisions.jsonl để phục vụ audit/replay.
func RecordTopicsDecision(st *store.Store, modelName string, snap *Snapshot, topics []TopicDecision, reason string) error {
	if st == nil {
		return nil
	}

	factsRaw, _ := json.Marshal(map[string]any{
		"snapshot_id": snap.ID,
		"geo":         snap.Geo,
		"item_count":  len(snap.Items),
	})
	decRaw, _ := json.Marshal(topics)

	_, err := st.Decisions.Append(store.DecisionRecord{
		Kind:     "trend_topics",
		Decider:  "arbiter",
		Model:    modelName,
		Facts:    factsRaw,
		Decision: decRaw,
		Reason:   reason,
	})
	return err
}
