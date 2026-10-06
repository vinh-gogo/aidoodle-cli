package trend

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func sourcePackID(key string) string {
	h := sha256.New()
	h.Write([]byte(strings.TrimSpace(key)))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func atomicWriteJSON(path string, v any) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	// Trên Windows, os.Rename có thể lỗi nếu file đích đang mở;
	// nhưng với file mới/ghi đè đơn lẻ thông thường thì Rename là chuẩn POSIX/Go.
	_ = os.Remove(path) // Xóa file cũ trước trên Windows nếu cần
	return os.Rename(tmpPath, path)
}

// SaveSnapshot lưu snapshot xu hướng vào cả file phiên bản lẫn latest.json.
func SaveSnapshot(outputDir string, snap *Snapshot) error {
	if snap == nil {
		return fmt.Errorf("snapshot rỗng")
	}
	baseDir := filepath.Join(outputDir, "meta", "trends")
	if snap.ID == "" {
		snap.ID = fmt.Sprintf("snap-%s", snap.FetchedAt.Format("20060102-150405"))
	}

	versionPath := filepath.Join(baseDir, snap.ID+".json")
	if err := atomicWriteJSON(versionPath, snap); err != nil {
		return fmt.Errorf("lưu snapshot %s: %w", snap.ID, err)
	}

	latestPath := filepath.Join(baseDir, "latest.json")
	if err := atomicWriteJSON(latestPath, snap); err != nil {
		return fmt.Errorf("lưu latest.json: %w", err)
	}

	return nil
}

// LoadLatestSnapshot tải snapshot xu hướng mới nhất từ thư mục meta/trends/.
func LoadLatestSnapshot(outputDir string) (*Snapshot, error) {
	path := filepath.Join(outputDir, "meta", "trends", "latest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("giải mã snapshot: %w", err)
	}
	return &snap, nil
}

// SaveSourcePack lưu gói tài liệu nguồn cho một chủ đề kịch bản.
func SaveSourcePack(outputDir string, sp *SourcePack) error {
	if sp == nil {
		return fmt.Errorf("source pack rỗng")
	}
	id := sourcePackID(sp.Topic)
	if sp.TrendRef != "" {
		id = sourcePackID(sp.TrendRef)
	}
	path := filepath.Join(outputDir, "meta", "trends", "sources", id+".json")
	return atomicWriteJSON(path, sp)
}

// LoadSourcePack tải gói tài liệu nguồn theo chủ đề hoặc mã trendRef.
func LoadSourcePack(outputDir string, topicOrTrendRef string) (*SourcePack, error) {
	id := sourcePackID(topicOrTrendRef)
	path := filepath.Join(outputDir, "meta", "trends", "sources", id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var sp SourcePack
	if err := json.Unmarshal(data, &sp); err != nil {
		return nil, fmt.Errorf("giải mã source pack: %w", err)
	}
	return &sp, nil
}
