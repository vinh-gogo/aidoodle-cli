package host

import (
	"errors"
	"testing"

	"github.com/voocel/ainovel-cli/assets"
	"github.com/voocel/ainovel-cli/internal/bootstrap"
	"github.com/voocel/ainovel-cli/internal/domain"
	storepkg "github.com/voocel/ainovel-cli/internal/store"
)

func newTestHostConfig(dir string) bootstrap.Config {
	return bootstrap.Config{
		OutputDir: dir,
		Provider:  "test",
		ModelName: "dummy",
		Providers: map[string]bootstrap.ProviderConfig{
			"test": {
				Type:    "openai",
				BaseURL: "http://127.0.0.1:9",
				APIKey:  "dummy",
			},
		},
	}
}

func TestHostAdvanceMode_DefaultsToAuto(t *testing.T) {
	dir := t.TempDir()
	cfg := newTestHostConfig(dir)
	bundle := assets.Load("default", assets.DefaultLoadOptions(dir))

	h, err := New(cfg, bundle)
	if err != nil {
		t.Fatalf("New host: %v", err)
	}
	defer h.Close()

	st := storepkg.NewStore(dir)
	rm, err := st.RunMeta.Load()
	if err != nil || rm == nil {
		t.Fatalf("Load RunMeta: %v", err)
	}
	if rm.AdvanceMode != domain.ChapterAdvanceAuto {
		t.Fatalf("mặc định không có trends phải là auto, got %s", rm.AdvanceMode)
	}
}

func TestHostAdvanceMode_TrendsDefaultsToReview(t *testing.T) {
	dir := t.TempDir()
	cfg := newTestHostConfig(dir)
	cfg.Trends.Enabled = true
	bundle := assets.Load("default", assets.DefaultLoadOptions(dir))

	h, err := New(cfg, bundle)
	if err != nil {
		t.Fatalf("New host: %v", err)
	}
	defer h.Close()

	st := storepkg.NewStore(dir)
	rm, err := st.RunMeta.Load()
	if err != nil || rm == nil {
		t.Fatalf("Load RunMeta: %v", err)
	}
	if rm.AdvanceMode != domain.ChapterAdvanceReview {
		t.Fatalf("dự án mới khi bật trends phải mặc định là review, got %s", rm.AdvanceMode)
	}
}

func TestHostAdvanceMode_ExplicitConfigOverrides(t *testing.T) {
	dir := t.TempDir()
	cfg := newTestHostConfig(dir)
	cfg.AdvanceMode = "review"
	bundle := assets.Load("default", assets.DefaultLoadOptions(dir))

	h, err := New(cfg, bundle)
	if err != nil {
		t.Fatalf("New host: %v", err)
	}
	defer h.Close()

	st := storepkg.NewStore(dir)
	rm, err := st.RunMeta.Load()
	if err != nil || rm == nil {
		t.Fatalf("Load RunMeta: %v", err)
	}
	if rm.AdvanceMode != domain.ChapterAdvanceReview {
		t.Fatalf("cấu hình rõ ràng advance_mode=review phải được áp dụng, got %s", rm.AdvanceMode)
	}
}

func TestHostAdvanceMode_InvalidModeRejected(t *testing.T) {
	dir := t.TempDir()
	cfg := newTestHostConfig(dir)
	cfg.AdvanceMode = "invalid_mode"
	bundle := assets.Load("default", assets.DefaultLoadOptions(dir))

	_, err := New(cfg, bundle)
	if err == nil {
		t.Fatal("chế độ advance mode không hợp lệ phải trả về lỗi")
	}
	var unsupported *domain.UnsupportedAdvanceModeError
	if !errors.As(err, &unsupported) {
		t.Fatalf("lỗi phải là UnsupportedAdvanceModeError, got %v", err)
	}
}
