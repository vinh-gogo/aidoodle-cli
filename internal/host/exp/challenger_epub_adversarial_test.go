package exp

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"golang.org/x/text/unicode/norm"

	"github.com/voocel/ainovel-cli/internal/domain"
)

// TestAdversarial_EPUB_EdgeCases generates an EPUB with complex edge cases and asserts compliance.
func TestAdversarial_EPUB_EdgeCases(t *testing.T) {
	// 1. Setup multi-volume hierarchy: 3 volumes, 5 arcs total, 20 chapters
	// Volume 1: Arc 1 (ch 1-4), Arc 2 (ch 5-7)
	// Volume 2: Arc 3 (ch 8-11), Arc 4 (ch 12-15)
	// Volume 3: Arc 5 (ch 16-20)
	volumes := []domain.VolumeOutline{
		{
			Index: 1,
			Title: "Mùa 1: Hỗn Độn Chi Sơ & Thần Ma <Khởi Nguyên> 'Đại Kỷ'",
			Arcs: []domain.ArcOutline{
				{
					Index: 1,
					Title: "Hồi 1: Rồng Lửa Thức Giấc 🐉 & Lời Sấm Truyền",
					Chapters: []domain.OutlineEntry{
						{Chapter: 1, Title: "Chương 1: Tiếng Gầm Giữa Đêm & Bão Tố <1>"},
						{Chapter: 2, Title: "Chương 2: Kiếm Khí Trảm Ma ⚔️ \"Tử Kiếm\""},
						{Chapter: 3, Title: "Chương 3: Tâm Trí Vỡ Vụn 🧠 'Mê Hồn'"},
						{Chapter: 4, Title: "Chương 4: Hoa Đào Rơi Rụng 🌸 & Huyết Lệ"},
					},
				},
				{
					Index: 2,
					Title: "Hồi 2: Thám Tử Bóng Đêm 🕵️ <Vết Tích Bí Ẩn>",
					Chapters: []domain.OutlineEntry{
						{Chapter: 5, Title: "Chương 5: Cuộn Giấy Cổ Xưa 📜 & Lời Nguyền"},
						{Chapter: 6, Title: "Chương 6: Hỏa Ngục Trừng Phạt 🔥 \"Tro Tàn\""},
						{Chapter: 7, Title: "Chương 7: Vương Miện Hoàng Gia 👑 & Phản Bội"},
					},
				},
			},
		},
		{
			Index: 2,
			Title: "Mùa 2: Khiên Thép & Lôi Đình 🛡️ <Chiến Tuyến>",
			Arcs: []domain.ArcOutline{
				{
					Index: 3,
					Title: "Hồi 3: Tia Chớp Đêm Đen ⚡ 'Thiên Kiếp'",
					Chapters: []domain.OutlineEntry{
						{Chapter: 8, Title: "Chương 8: Vực Thẳm Vô Tận & Ác Mộng <8>"},
						{Chapter: 9, Title: "Chương 9: Tiếng Chuông Chùa Đêm & Bến Vắng"},
						{Chapter: 10, Title: "Chương 10: Trăng Tàn Dưới Nước & Khói Sương"},
						{Chapter: 11, Title: "Chương 11: Bước Chân Cô Độc \"Độc Hành\""},
					},
				},
				{
					Index: 4,
					Title: "Hồi 4: Hoàng Hôn Rực Lửa & Gió Lạnh",
					Chapters: []domain.OutlineEntry{
						{Chapter: 12, Title: "Chương 12: Đỉnh Núi Tuyết Trắng <Băng Phong>"},
						{Chapter: 13, Title: "Chương 13: Cánh Chim Lạc Lối 'Phương Nam'"},
						{Chapter: 14, Title: "Chương 14: Lửa Hận Cháy Bỏng & Tàn Tro"},
						{Chapter: 15, Title: "Chương 15: Giọt Sương Buổi Sớm 🌸 & Hy Vọng"},
					},
				},
			},
		},
		{
			Index: 3,
			Title: "Mùa 3: Định Mệnh Cuối Cùng & Khải Hoàn 👑",
			Arcs: []domain.ArcOutline{
				{
					Index: 5,
					Title: "Hồi 5: Hợp Nhất Tam Giới <Vĩnh Hằng>",
					Chapters: []domain.OutlineEntry{
						{Chapter: 16, Title: "Chương 16: Trận Chiến Sinh Tử & Huyết Chiến"},
						{Chapter: 17, Title: "Chương 17: Phá Tan Hư Không \"Trảm Thiên\""},
						{Chapter: 18, Title: "Chương 18: Thần Khí Quy Vị & Ánh Sáng"},
						{Chapter: 19, Title: "Chương 19: Lời Hứa Năm Xưa 'Tri Kỷ'"},
						{Chapter: 20, Title: "Chương 20: Đại Kết Cục: Non Sông Liền Một Dải 🐉⚔️"},
					},
				},
			},
		},
	}

	locations := buildLocations(volumes)
	var chapters []int
	titleIdx := make(map[int]string)
	bodies := make(map[int]string)

	// Comprehensive Vietnamese diacritics string
	vietnameseDiacritics := "Việt Nam đất nước ta ơi! Mênh mông biển lúa đâu trời đẹp hơn. " +
		"Á, À, Ả, Ã, Ạ, Ă, Ắ, Ằ, Ẳ, Ẵ, Ặ, Â, Ấ, Ầ, Ẩ, Ẫ, Ậ, Đ, É, È, Ẻ, Ẽ, Ẹ, Ê, Ế, Ề, Ể, Ễ, Ệ, " +
		"Í, Ì, Ỉ, Ĩ, Ị, Ó, Ò, Ỏ, Õ, Ọ, Ô, Ố, Ồ, Ổ, Ỗ, Ộ, Ơ, Ớ,Ờ, Ở, Ỡ, Ợ, Ú, Ù, Ủ, Ũ, Ụ, Ư, Ứ, Ừ, Ử, Ữ, Ự, Ý, Ỳ, Ỷ, Ỹ, Ỵ. " +
		"á à ả ã ạ ă ắ ằ ẳ ẵ ặ â ấn ầ ẩ ẫ ậ đ é è ẻ ẽ ẹ ê ế ề ể ễ ệ í ì ỉ ĩ ị ó ò ỏ õ ọ ô ố ồ ổ ỗ ộ ơ ớ ờ ở ỡ ợ ú ù ủ ũ ụ ư ứng ừ tử ữ ự ý kỳ ỷ ỹ ỵ."

	// NFD string with explicit decomposed unicode combining marks
	nfdSample := norm.NFD.String("Tiếng Việt Tổ Quốc Yêu Thương: người cầm gươm mở cõi miền Nam.")

	// XML Special characters injection payload
	xmlSpecialChars := "Nhân vật nói: <script>alert('xss attack')</script> & \"nguy hiểm\" > 'an toàn' &amp; <b>in đậm</b>."

	// Control characters 0x00-0x1F (excluding 0x09, 0x0A, 0x0D which are legal)
	controlCharsPayload := "Đoạn văn có chứa control chars: \x00\x01\x02\x03\x04\x05\x06\x07\x08\x0b\x0c\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f."

	for ch := 1; ch <= 20; ch++ {
		chapters = append(chapters, ch)
		titleIdx[ch] = fmt.Sprintf("Chương %d: Tiêu Đề Có Dấu & Ký Tự <Đặc Biệt #%d> \"Trích\"", ch, ch)

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# Chương %d\n\n", ch))
		sb.WriteString(vietnameseDiacritics + "\n\n")
		sb.WriteString("Đoạn NFD kết hợp: " + nfdSample + "\n\n")
		sb.WriteString("Đoạn ký tự đặc biệt XML: " + xmlSpecialChars + "\n\n")
		sb.WriteString(controlCharsPayload + "\n\n")
		sb.WriteString("Đoạn biểu cảm emoji: 🐉 ⚔️ 🧠 🌸 🕵️ 📜 🔥 👑 🛡️ ⚡!\n\n")
		sb.WriteString("***\n\n")
		sb.WriteString("> Trích dẫn triết lý: Người không vì mình trời tru đất diệt & kẻ vì người lưu danh thiên cổ.\n\n")
		sb.WriteString("## Hậu ký lịch sử\n\nĐây là phần hậu ký phân tích bối cảnh triều đại phong kiến xưa.")

		bodies[ch] = sb.String()
	}

	bookMeta := domain.BookMetadata{
		Title:    "Đại Việt Sử Ký & Sử Thi Anh Hùng <Toàn Thư> \"Bản Đặc Biệt\"",
		Author:   "Ngô Sĩ Liên & Lê Văn Hưu 'Sử Thần'",
		Synopsis: "Bộ sử thi ghi lại hào khí non sông đất nước qua các triều đại Đinh - Lê - Lý - Trần - Hậu Lê & Quang Trung.",
		Subject:  "Lịch sử & Huyền huyễn <Epic Fantasy>",
	}

	fixedTime := time.Date(2026, 10, 17, 8, 30, 0, 0, time.UTC)
	epubBytes, err := ExportEPUB(
		bookMeta,
		chapters,
		titleIdx,
		locations,
		bodies,
		WithVolumes(volumes),
		WithModifiedTime(fixedTime),
	)
	if err != nil {
		t.Fatalf("ExportEPUB failed with error: %v", err)
	}
	if len(epubBytes) == 0 {
		t.Fatalf("ExportEPUB returned empty byte slice")
	}

	// 2. Programmatically inspect the generated ZIP archive
	zr, err := zip.NewReader(bytes.NewReader(epubBytes), int64(len(epubBytes)))
	if err != nil {
		t.Fatalf("Generated EPUB is not a valid zip archive: %v", err)
	}

	if len(zr.File) == 0 {
		t.Fatalf("Zip archive is empty")
	}

	// Invariant: Entry 0 MUST be 'mimetype'
	entry0 := zr.File[0]
	if entry0.Name != "mimetype" {
		t.Fatalf("W3C VIOLATION: Entry 0 must be 'mimetype', got %q", entry0.Name)
	}
	if entry0.Method != zip.Store {
		t.Fatalf("W3C VIOLATION: mimetype must be uncompressed (Method=zip.Store), got method %d", entry0.Method)
	}
	if len(entry0.Extra) != 0 {
		t.Fatalf("W3C VIOLATION: mimetype len(Extra) must be 0, got %d", len(entry0.Extra))
	}

	rc0, err := entry0.Open()
	if err != nil {
		t.Fatalf("Failed to open entry 0 mimetype: %v", err)
	}
	mimetypeBytes, _ := io.ReadAll(rc0)
	_ = rc0.Close()
	if string(mimetypeBytes) != "application/epub+zip" {
		t.Fatalf("W3C VIOLATION: mimetype content must be exact 'application/epub+zip', got %q", string(mimetypeBytes))
	}

	// Read all files in archive
	zipFiles := make(map[string][]byte)
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("Failed to open entry %q: %v", f.Name, err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("Failed to read entry %q: %v", f.Name, err)
		}
		zipFiles[f.Name] = data
	}

	// Verify required OCF entries
	requiredFiles := []string{
		"META-INF/container.xml",
		"OEBPS/content.opf",
		"OEBPS/nav.xhtml",
		"OEBPS/style.css",
		"OEBPS/cover.xhtml",
	}
	for _, req := range requiredFiles {
		if _, ok := zipFiles[req]; !ok {
			t.Fatalf("Missing required EPUB file in archive: %q", req)
		}
	}

	// 3. Strict XML validation on container.xml, content.opf, nav.xhtml, cover.xhtml, and ALL 20 chapters
	validateStrictXML := func(filename string, data []byte) {
		t.Helper()
		dec := xml.NewDecoder(bytes.NewReader(data))
		dec.Strict = true
		for {
			_, err := dec.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("XML VALIDATION FAILED for %s: %v\nContent snippet:\n%s", filename, err, string(data[:min(len(data), 1000)]))
			}
		}
	}

	validateStrictXML("META-INF/container.xml", zipFiles["META-INF/container.xml"])
	validateStrictXML("OEBPS/content.opf", zipFiles["OEBPS/content.opf"])
	validateStrictXML("OEBPS/nav.xhtml", zipFiles["OEBPS/nav.xhtml"])
	validateStrictXML("OEBPS/cover.xhtml", zipFiles["OEBPS/cover.xhtml"])

	for ch := 1; ch <= 20; ch++ {
		chName := fmt.Sprintf("OEBPS/chapter%03d.xhtml", ch)
		data, ok := zipFiles[chName]
		if !ok {
			t.Fatalf("Missing chapter file %s in EPUB", chName)
		}
		validateStrictXML(chName, data)

		// 4. Verify Unicode NFC normalization on chapter text
		chContent := string(data)
		if !norm.NFC.IsNormalString(chContent) {
			t.Fatalf("UNICODE NORMALIZATION FLAW: Chapter %d is not normalized to Unicode NFC", ch)
		}

		// Verify forbidden control characters (0x00-0x1F except 0x09, 0x0A, 0x0D) are stripped
		for _, r := range chContent {
			if r < 0x20 && r != 0x09 && r != 0x0A && r != 0x0D {
				t.Fatalf("XML CONTROL CHAR LEAK: Chapter %d contains illegal control rune %#x", ch, r)
			}
		}

		// Verify raw <script> was safely escaped
		if strings.Contains(chContent, "<script>") {
			t.Fatalf("XSS INJECTION LEAK: Chapter %d contains raw <script> tag unescaped", ch)
		}
	}

	// 5. Verify presence of EPUB Accessibility 1.1 metadata in content.opf
	opfStr := string(zipFiles["OEBPS/content.opf"])
	accessibilityRequirements := []string{
		`<meta property="schema:accessMode">textual</meta>`,
		`<meta property="schema:accessModeSufficient">textual</meta>`,
		`<meta property="schema:accessibilityFeature">structuralNavigation</meta>`,
		`<meta property="schema:accessibilityHazard">none</meta>`,
		`<meta property="schema:accessibilitySummary">`,
	}
	for _, reqMeta := range accessibilityRequirements {
		if !strings.Contains(opfStr, reqMeta) {
			t.Fatalf("EPUB ACCESSIBILITY 1.1 FLAW: content.opf missing metadata element: %q", reqMeta)
		}
	}

	// Verify Dublin Core metadata in content.opf
	dublinCoreReqs := []string{
		`<dc:language>vi</dc:language>`,
		`<meta property="dcterms:modified">2026-10-17T08:30:00Z</meta>`,
		`xml:lang="vi"`,
		`version="3.0"`,
	}
	for _, reqDC := range dublinCoreReqs {
		if !strings.Contains(opfStr, reqDC) {
			t.Fatalf("OPF 3.0 METADATA FLAW: content.opf missing Dublin Core element: %q", reqDC)
		}
	}

	// 6. Verify Navigation Document (nav.xhtml) has nested hierarchy for Volume -> Arc -> Chapter
	navStr := string(zipFiles["OEBPS/nav.xhtml"])
	if !strings.Contains(navStr, `<nav epub:type="toc" id="toc">`) {
		t.Fatalf("NAV DOCUMENT FLAW: Missing <nav epub:type=\"toc\">")
	}
	if !strings.Contains(navStr, `<span>Mùa 1:`) || !strings.Contains(navStr, `<span>Hồi 1:`) {
		t.Fatalf("NAV DOCUMENT FLAW: nav.xhtml missing nested Volume or Arc titles")
	}
	if !strings.Contains(navStr, `href="chapter020.xhtml"`) {
		t.Fatalf("NAV DOCUMENT FLAW: nav.xhtml missing link to chapter 20")
	}
}
