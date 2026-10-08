package exp

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/voocel/ainovel-cli/internal/domain"
)

func TestRenderDOCX(t *testing.T) {
	book := domain.BookMetadata{
		Title: "Quang Trung: Thiên Tài Quân Sự",
	}
	chapters := []int{1, 2}
	titleIdx := chapterTitleIndex{
		1: "Mưu Kế Nhử Giặc Vào Bẫy",
		2: "Đại Phá Quân Xiêm",
	}
	locations := map[int]chapterLocation{
		1: {VolumeIdx: 1, VolumeTitle: "Quyển 1: Tiếng Sấm Rạch Gầm", IsFirstOfVolume: true},
		2: {VolumeIdx: 1, VolumeTitle: "Quyển 1: Tiếng Sấm Rạch Gầm", IsFirstOfVolume: false},
	}
	bodies := map[int]string{
		1: "# Chương 1: Mưu Kế Nhử Giặc Vào Bẫy\n\nĐêm ngày 18 rạng sáng 19 tháng Giáp Thìn (1785), dòng sông Tiền mịt mù sương lạnh.\n\nNguyễn Huệ đứng trên mũi chiến thuyền, áo bào phần phật bay trong gió sớm.",
		2: "Trận địa mai phục đã sẵn sàng.\n\nTiếng pháo hiệu lệnh vừa dứt, hàng trăm chiến thuyền Tây Sơn từ rạch nhánh đồng loạt lao ra.",
	}

	data, err := renderDOCX(book, chapters, titleIdx, locations, bodies)
	if err != nil {
		t.Fatalf("renderDOCX failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("renderDOCX returned empty data")
	}

	// Verify that the returned bytes form a valid zip archive
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("zip.NewReader failed: %v", err)
	}

	requiredFiles := map[string]bool{
		"[Content_Types].xml":          false,
		"_rels/.rels":                  false,
		"word/_rels/document.xml.rels": false,
		"word/styles.xml":              false,
		"word/document.xml":            false,
	}

	var documentXMLContent string

	for _, f := range zr.File {
		if _, ok := requiredFiles[f.Name]; ok {
			requiredFiles[f.Name] = true
		}
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("failed to open word/document.xml: %v", err)
			}
			b, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("failed to read word/document.xml: %v", err)
			}
			documentXMLContent = string(b)
		}
	}

	for name, found := range requiredFiles {
		if !found {
			t.Errorf("missing required file in docx zip: %s", name)
		}
	}

	// Check content inside word/document.xml
	if !strings.Contains(documentXMLContent, "Quang Trung: Thiên Tài Quân Sự") {
		t.Errorf("document.xml missing book title")
	}
	if !strings.Contains(documentXMLContent, "Chương 1: Mưu Kế Nhử Giặc Vào Bẫy") {
		t.Errorf("document.xml missing chapter 1 title")
	}
	if !strings.Contains(documentXMLContent, "Nguyễn Huệ đứng trên mũi chiến thuyền") {
		t.Errorf("document.xml missing chapter 1 text")
	}
	if !strings.Contains(documentXMLContent, "Tiếng pháo hiệu lệnh vừa dứt") {
		t.Errorf("document.xml missing chapter 2 text")
	}
}
