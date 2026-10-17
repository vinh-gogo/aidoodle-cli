package tui

import (
	"testing"

	"github.com/voocel/ainovel-cli/internal/host/exp"
)

func TestParseExportArgs_EPUBFlag(t *testing.T) {
	opts, err := parseExportArgs([]string{"--epub", "--overwrite"})
	if err != nil {
		t.Fatalf("unexpected error parsing --epub: %v", err)
	}
	if opts.Format != exp.FormatEPUB {
		t.Errorf("expected FormatEPUB, got %q", opts.Format)
	}
	if !opts.Overwrite {
		t.Errorf("expected Overwrite to be true")
	}

	opts2, err := parseExportArgs([]string{"from=1", "to=5", "format=epub"})
	if err != nil {
		t.Fatalf("unexpected error parsing format=epub: %v", err)
	}
	if opts2.Format != exp.FormatEPUB {
		t.Errorf("expected FormatEPUB, got %q", opts2.Format)
	}
	if opts2.From != 1 || opts2.To != 5 {
		t.Errorf("expected from=1 to=5, got from=%d to=%d", opts2.From, opts2.To)
	}

	opts3, err := parseExportArgs([]string{"--txt"})
	if err != nil {
		t.Fatalf("unexpected error parsing --txt: %v", err)
	}
	if opts3.Format != exp.FormatTXT {
		t.Errorf("expected FormatTXT, got %q", opts3.Format)
	}
}
