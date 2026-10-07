package tui

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNextOutputDir(t *testing.T) {
	fixedTime := time.Date(2026, 10, 7, 18, 25, 0, 0, time.Local)

	t.Run("default output/novel", func(t *testing.T) {
		cur := filepath.Join("output", "novel")
		got := nextOutputDir(cur, fixedTime)
		want := filepath.Join("output", "novel-20261007-1825")
		if got != want {
			t.Fatalf("nextOutputDir(%q) = %q, want %q", cur, got, want)
		}
	})

	t.Run("empty string defaults to novel prefix", func(t *testing.T) {
		got := nextOutputDir("", fixedTime)
		want := filepath.Join("output", "novel-20261007-1825")
		if got != want {
			t.Fatalf("nextOutputDir(\"\") = %q, want %q", got, want)
		}
	})

	t.Run("consecutive /new strips previous timestamp", func(t *testing.T) {
		cur := filepath.Join("output", "novel-20261007-1825")
		laterTime := time.Date(2026, 10, 7, 18, 30, 0, 0, time.Local)
		got := nextOutputDir(cur, laterTime)
		want := filepath.Join("output", "novel-20261007-1830")
		if got != want {
			t.Fatalf("nextOutputDir(%q) = %q, want %q", cur, got, want)
		}
	})

	t.Run("collision in same minute appends counter", func(t *testing.T) {
		tmpDir := t.TempDir()
		cur := filepath.Join(tmpDir, "novel")

		first := nextOutputDir(cur, fixedTime)
		wantFirst := filepath.Join(tmpDir, "novel-20261007-1825")
		if first != wantFirst {
			t.Fatalf("first candidate = %q, want %q", first, wantFirst)
		}

		// Create first dir to simulate existing directory
		if err := os.MkdirAll(first, 0o755); err != nil {
			t.Fatal(err)
		}

		// Second call with same minute
		second := nextOutputDir(cur, fixedTime)
		wantSecond := filepath.Join(tmpDir, "novel-20261007-1825-01")
		if second != wantSecond {
			t.Fatalf("second candidate = %q, want %q", second, wantSecond)
		}
	})
}
