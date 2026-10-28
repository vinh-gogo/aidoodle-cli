package exp

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/text/unicode/norm"

	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/store"
)

// epubInspector kiểm tra và giải mã cấu trúc tệp .epub
type epubInspector struct {
	data     []byte
	zr       *zip.Reader
	files    map[string][]byte
	manifest map[string]string // id -> href
	spine    []string          // idrefs
}

func inspectEPUB(t *testing.T, data []byte) *epubInspector {
	t.Helper()
	if len(data) == 0 {
		t.Fatal("empty epub data")
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("data is not a valid zip archive: %v", err)
	}

	files := make(map[string][]byte, len(zr.File))
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("failed to open zip entry %q: %v", f.Name, err)
		}
		content, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("failed to read zip entry %q: %v", f.Name, err)
		}
		files[f.Name] = content
	}

	insp := &epubInspector{
		data:     data,
		zr:       zr,
		files:    files,
		manifest: make(map[string]string),
	}

	// Phân tích OPF để lấy manifest và spine
	if opfData, ok := files["OEBPS/content.opf"]; ok {
		insp.parseOPF(t, opfData)
	}

	return insp
}

func (insp *epubInspector) parseOPF(t *testing.T, opfData []byte) {
	t.Helper()
	dec := xml.NewDecoder(bytes.NewReader(opfData))
	inManifest := false
	inSpine := false

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("XML parse error in content.opf: %v", err)
		}

		switch se := tok.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "manifest":
				inManifest = true
			case "spine":
				inSpine = true
			case "item":
				if inManifest {
					var id, href string
					for _, attr := range se.Attr {
						if attr.Name.Local == "id" {
							id = attr.Value
						} else if attr.Name.Local == "href" {
							href = attr.Value
						}
					}
					if id != "" && href != "" {
						insp.manifest[id] = href
					}
				}
			case "itemref":
				if inSpine {
					for _, attr := range se.Attr {
						if attr.Name.Local == "idref" {
							insp.spine = append(insp.spine, attr.Value)
						}
					}
				}
			}
		case xml.EndElement:
			switch se.Name.Local {
			case "manifest":
				inManifest = false
			case "spine":
				inSpine = false
			}
		}
	}
}

// assertWellFormedXML kiểm tra tính hợp lệ cú pháp XML nghiêm ngặt
func assertWellFormedXML(t *testing.T, name string, data []byte) {
	t.Helper()
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		_, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("strict XML validation failed for %q: %v\ncontent snippet:\n%s",
				name, err, string(data[:min(len(data), 500)]))
		}
	}
}

// 1. Programmatic ZIP Inspector: Entry 0 is uncompressed mimetype, empty Extra
func TestEPUBValidator_Entry0_Mimetype(t *testing.T) {
	data, err := renderEPUB(
		domain.BookMetadata{Title: "Tiêu đề kiểm thử", Synopsis: "Giới thiệu tóm tắt"},
		[]int{1},
		chapterTitleIndex{1: "Chương một"},
		nil,
		map[int]string{1: "Nội dung chương một."},
	)
	if err != nil {
		t.Fatalf("renderEPUB: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("not a valid zip: %v", err)
	}

	if len(zr.File) == 0 {
		t.Fatal("archive has no entries")
	}

	first := zr.File[0]
	// Entry 0 MUST be named "mimetype"
	if first.Name != "mimetype" {
		t.Errorf("entry 0 must be 'mimetype', got %q", first.Name)
	}
	// Method MUST be zip.Store (0, uncompressed)
	if first.Method != zip.Store {
		t.Errorf("entry 0 must be uncompressed (zip.Store == 0), got method %d", first.Method)
	}
	// Extra fields MUST be empty
	if len(first.Extra) != 0 {
		t.Errorf("entry 0 Extra field must be empty (len == 0), got len %d", len(first.Extra))
	}

	rc, err := first.Open()
	if err != nil {
		t.Fatalf("open mimetype: %v", err)
	}
	content, err := io.ReadAll(rc)
	_ = rc.Close()
	if err != nil {
		t.Fatalf("read mimetype: %v", err)
	}
	if string(content) != "application/epub+zip" {
		t.Errorf("mimetype content must be 'application/epub+zip', got %q", string(content))
	}
}

// 2. Strict XML Validation of container.xml, content.opf, nav.xhtml, cover.xhtml, and chapters
func TestEPUBValidator_AllXMLFilesStrictlyWellFormed(t *testing.T) {
	book := domain.BookMetadata{
		Title:    "Đại Việt Sử Ký",
		Synopsis: "Sử thi hào hùng thời Lý - Trần & Lê Sơ.",
		Author:   "Lê Văn Hưu",
		Subject:  "Lịch sử Việt Nam",
	}

	chapters := []int{1, 2, 3}
	titles := chapterTitleIndex{
		1: "Hịch Tướng Sĩ",
		2: "Bạch Đằng Giang",
		3: "Bình Ngô Đại Cáo",
	}
	bodies := map[int]string{
		1: "# Hịch Tướng Sĩ\n\nTa thường nghe Kỷ Tín đem mình chết thay cho Cao Đế...\n\n***\n\nQuân ta trăm trận trăm thắng.",
		2: "> Tráng sĩ vung gươm bóng nguyệt mờ.\n\nSóng Bạch Đằng cuồn cuộn chảy về đông.\n\n## Hậu ký và Ghi chú lịch sử của tác giả\n\nTrận đánh năm 1288 là đỉnh cao quân sự thời Trần.",
		3: "Từng nghe:\nViệc nhân nghĩa cốt ở yên dân\nQuân điếu phạt trước lo trừ bạo.",
	}

	data, err := renderEPUB(book, chapters, titles, nil, bodies)
	if err != nil {
		t.Fatalf("renderEPUB: %v", err)
	}

	insp := inspectEPUB(t, data)

	// Verify all XML/XHTML files in the archive
	xmlFilesChecked := 0
	for name, content := range insp.files {
		if strings.HasSuffix(name, ".xml") || strings.HasSuffix(name, ".opf") || strings.HasSuffix(name, ".xhtml") {
			assertWellFormedXML(t, name, content)
			xmlFilesChecked++
		}
	}

	if xmlFilesChecked < 5 {
		t.Errorf("expected at least 5 XML documents checked, got %d", xmlFilesChecked)
	}
}

// 3. OPF 3.0 & EPUB Accessibility 1.1 Assertions
func TestEPUBValidator_OPF3_And_Accessibility11(t *testing.T) {
	fixedTime := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	book := domain.BookMetadata{
		Title:    "Mặt Trời Đêm",
		Synopsis: "Tiểu thuyết trinh thám hình sự.",
		Author:   "Trần Văn A",
		Subject:  "Trinh thám & Kinh dị",
	}

	data, err := renderEPUB(
		book,
		[]int{1, 2},
		chapterTitleIndex{1: "Vụ án mạng", 2: "Manh mối"},
		nil,
		map[int]string{1: "Nội dung chương 1.", 2: "Nội dung chương 2."},
		WithModifiedTime(fixedTime),
	)
	if err != nil {
		t.Fatalf("renderEPUB: %v", err)
	}

	insp := inspectEPUB(t, data)
	opf := string(insp.files["OEBPS/content.opf"])

	// OPF Package element assertions
	if !strings.Contains(opf, `version="3.0"`) {
		t.Errorf("OPF must declare version=\"3.0\"")
	}
	if !strings.Contains(opf, `unique-identifier="pub-id"`) {
		t.Errorf("OPF must declare unique-identifier=\"pub-id\"")
	}
	if !strings.Contains(opf, `xml:lang="vi"`) {
		t.Errorf("OPF must declare xml:lang=\"vi\"")
	}
	if !strings.Contains(opf, `prefix="schema: http://schema.org/"`) {
		t.Errorf("OPF must declare prefix for schema: http://schema.org/")
	}

	// Dublin Core Metadata assertions
	for _, expected := range []string{
		`<dc:identifier id="pub-id">urn:uuid:`,
		`<dc:title>Mặt Trời Đêm</dc:title>`,
		`<dc:language>vi</dc:language>`,
		`<dc:creator id="creator">Trần Văn A</dc:creator>`,
		`<dc:description>Tiểu thuyết trinh thám hình sự.</dc:description>`,
		`<dc:subject>Trinh thám &amp; Kinh dị</dc:subject>`,
		`<dc:date>2026-10-17</dc:date>`,
		`<meta property="dcterms:modified">2026-10-17T12:00:00Z</meta>`,
	} {
		if !strings.Contains(opf, expected) {
			t.Errorf("OPF missing expected Dublin Core element %q", expected)
		}
	}

	// EPUB Accessibility 1.1 Assertions
	accessibilityElements := []string{
		`<meta property="schema:accessMode">textual</meta>`,
		`<meta property="schema:accessModeSufficient">textual</meta>`,
		`<meta property="schema:accessibilityFeature">structuralNavigation</meta>`,
		`<meta property="schema:accessibilityHazard">none</meta>`,
		`<meta property="schema:accessibilitySummary">`,
	}
	for _, elem := range accessibilityElements {
		if !strings.Contains(opf, elem) {
			t.Errorf("OPF missing Accessibility 1.1 element %q", elem)
		}
	}

	// Manifest and Spine cross-validation: every manifest href exists in ZIP
	for id, href := range insp.manifest {
		zipPath := "OEBPS/" + href
		if _, ok := insp.files[zipPath]; !ok {
			t.Errorf("manifest item %q references %q, but %s not found in zip archive", id, href, zipPath)
		}
	}

	// Every spine idref resolves to a manifest item
	for _, idref := range insp.spine {
		if _, ok := insp.manifest[idref]; !ok {
			t.Errorf("spine itemref %q does not exist in manifest", idref)
		}
	}
}

// 4. Unicode NFC Normalization, Control Character Stripping & XML Escaping
func TestEPUBValidator_UnicodeNFC_And_ControlCharacterStripping(t *testing.T) {
	// e + combining acute accent (\u0301) is NFD; normalized NFC is é (\u00e9)
	nfdTitle := "Tie\u0302\u0301ng Vie\u0323\u0302t" // Tiếng Việt in NFD
	dirtyBody := "Khởi đầu câu chuyện.\x00\x07\x0b\x0c\x1b\x1fVăn bản sạch.\n\n" +
		"Đoạn 2 chứa ký tự đặc biệt: <script>alert('xss')</script> & \"ngoặc kép\" 'nháy đơn' < >.\n\n" +
		"***\n\n" +
		"Hào khí Đông A — ngàn năm văn hiến."

	data, err := renderEPUB(
		domain.BookMetadata{Title: nfdTitle, Synopsis: "Tóm tắt chứa dấu tổ hợp rời: e\u0301"},
		[]int{1},
		chapterTitleIndex{1: "Chương diacritics: o\u0309 u\u0303"},
		nil,
		map[int]string{1: dirtyBody},
	)
	if err != nil {
		t.Fatalf("renderEPUB: %v", err)
	}

	insp := inspectEPUB(t, data)

	// 1. Assert all XHTML files are strict valid XML
	for name, content := range insp.files {
		if strings.HasSuffix(name, ".xhtml") {
			assertWellFormedXML(t, name, content)

			// 2. Assert Unicode NFC normalization
			if !norm.NFC.IsNormal(content) {
				t.Errorf("file %q is not normalized to Unicode NFC", name)
			}

			// 3. Assert absence of forbidden XML control characters (0x00 - 0x1F except 0x09, 0x0A, 0x0D)
			for idx, b := range content {
				if b < 0x20 && b != 0x09 && b != 0x0A && b != 0x0D {
					t.Errorf("file %q contains forbidden XML control byte 0x%02x at byte index %d", name, b, idx)
					break
				}
			}
		}
	}

	ch1 := string(insp.files["OEBPS/chapter001.xhtml"])
	// Assert raw HTML script tag is safely escaped and NOT executable
	if strings.Contains(ch1, "<script>") {
		t.Errorf("chapter XHTML contains unescaped <script> tag: dangerous raw HTML leakage!")
	}
	if !strings.Contains(ch1, "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;") {
		t.Errorf("expected escaped script tag in body, got: %s", ch1)
	}

	// Assert scene break was rendered
	if !strings.Contains(ch1, `<hr class="scene-break"/>`) {
		t.Errorf("missing scene break hr tag in chapter XHTML")
	}

	// Assert root html element has lang="vi" xml:lang="vi"
	if !strings.Contains(ch1, `lang="vi" xml:lang="vi"`) {
		t.Errorf("chapter root html tag missing lang=\"vi\" xml:lang=\"vi\"")
	}
}

// 5. Hierarchical Navigation Document (Volume -> Arc -> Chapter) & Landmarks
func TestEPUBValidator_HierarchicalNav_And_Landmarks(t *testing.T) {
	volumes := []domain.VolumeOutline{
		{
			Index: 1,
			Title: "Khởi Đầu Gian Nan",
			Arcs: []domain.ArcOutline{
				{
					Index: 1,
					Title: "Rời Làng",
					Chapters: []domain.OutlineEntry{
						{Title: "Chia tay người thân"},
						{Title: "Đêm đầu trên rừng"},
					},
				},
				{
					Index: 2,
					Title: "Gặp Gỡ Minh Chủ",
					Chapters: []domain.OutlineEntry{
						{Title: "Lạc vào quân doanh"},
					},
				},
			},
		},
		{
			Index: 2,
			Title: "Khởi Nghĩa Toàn Thắng",
			Arcs: []domain.ArcOutline{
				{
					Index: 1,
					Title: "Trận Đại Thắng",
					Chapters: []domain.OutlineEntry{
						{Title: "Quyết chiến Lam Sơn"},
					},
				},
			},
		},
	}

	book := domain.BookMetadata{
		Title:    "Lam Sơn Thực Lục",
		Synopsis: "Sử thi khởi nghĩa Lam Sơn.",
		Author:   "Nguyễn Trãi",
	}

	chapters := []int{1, 2, 3, 4}
	titleIdx := chapterTitleIndex{
		1: "Chia tay người thân",
		2: "Đêm đầu trên rừng",
		3: "Lạc vào quân doanh",
		4: "Quyết chiến Lam Sơn",
	}
	bodies := map[int]string{
		1: "Nội dung chương 1",
		2: "Nội dung chương 2",
		3: "Nội dung chương 3",
		4: "Nội dung chương 4",
	}

	data, err := renderEPUB(
		book, chapters, titleIdx, nil, bodies,
		WithVolumes(volumes),
	)
	if err != nil {
		t.Fatalf("renderEPUB: %v", err)
	}

	insp := inspectEPUB(t, data)
	nav := string(insp.files["OEBPS/nav.xhtml"])

	assertWellFormedXML(t, "OEBPS/nav.xhtml", insp.files["OEBPS/nav.xhtml"])

	// Check Manifest declares nav.xhtml with properties="nav"
	opf := string(insp.files["OEBPS/content.opf"])
	if !strings.Contains(opf, `href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"`) {
		t.Errorf("manifest must declare nav with properties=\"nav\"")
	}

	// Check Hierarchical TOC elements
	for _, want := range []string{
		`<nav epub:type="toc" id="toc">`,
		`<span>Mùa 1: Khởi Đầu Gian Nan</span>`,
		`<span>Hồi 1: Rời Làng</span>`,
		`href="chapter001.xhtml">Tập 1: Chia tay người thân</a>`,
		`href="chapter002.xhtml">Tập 2: Đêm đầu trên rừng</a>`,
		`<span>Hồi 2: Gặp Gỡ Minh Chủ</span>`,
		`href="chapter003.xhtml">Tập 3: Lạc vào quân doanh</a>`,
		`<span>Mùa 2: Khởi Nghĩa Toàn Thắng</span>`,
		`<span>Hồi 1: Trận Đại Thắng</span>`,
		`href="chapter004.xhtml">Tập 4: Quyết chiến Lam Sơn</a>`,
	} {
		if !strings.Contains(nav, want) {
			t.Errorf("nav.xhtml missing hierarchical TOC element %q", want)
		}
	}

	// Check Landmarks block
	for _, want := range []string{
		`<nav epub:type="landmarks" id="landmarks" hidden="">`,
		`<li><a epub:type="cover" href="cover.xhtml">Bìa sách</a></li>`,
		`<li><a epub:type="toc" href="nav.xhtml">Mục lục</a></li>`,
		`<li><a epub:type="bodymatter" href="chapter001.xhtml">Bắt đầu đọc</a></li>`,
	} {
		if !strings.Contains(nav, want) {
			t.Errorf("nav.xhtml missing Landmark element %q", want)
		}
	}
}

// 6. Cover Image Embedding and Properties Assertion
func TestEPUBValidator_CoverImage(t *testing.T) {
	fakeJPEG := []byte("\xff\xd8\xff\xe0\x00\x10JFIF\x00\x01\x01\x00\x00\x01\x00\x01\x00\x00\xff\xdb\x00C\x00\xff\xd9")
	cover := &CoverImage{
		Data:      fakeJPEG,
		Filename:  "cover.jpg",
		MediaType: "image/jpeg",
	}

	book := domain.BookMetadata{
		Title:    "Sách Có Bìa Minh Họa",
		Synopsis: "Sách thử nghiệm nhúng ảnh bìa chuẩn W3C EPUB 3.3.",
	}

	data, err := renderEPUB(
		book,
		[]int{1},
		chapterTitleIndex{1: "Chương duy nhất"},
		nil,
		map[int]string{1: "Nội dung chương."},
		WithCoverImage(cover),
	)
	if err != nil {
		t.Fatalf("renderEPUB: %v", err)
	}

	insp := inspectEPUB(t, data)

	// 1. Assert image exists in ZIP
	if _, ok := insp.files["OEBPS/images/cover.jpg"]; !ok {
		t.Fatalf("OEBPS/images/cover.jpg missing from ZIP")
	}

	// 2. Assert manifest has properties="cover-image"
	opf := string(insp.files["OEBPS/content.opf"])
	if !strings.Contains(opf, `id="cover-image" href="images/cover.jpg" media-type="image/jpeg" properties="cover-image"`) {
		t.Errorf("OPF manifest missing cover-image declaration with properties=\"cover-image\"")
	}

	// 3. Assert cover.xhtml displays the cover image
	coverXHTML := string(insp.files["OEBPS/cover.xhtml"])
	if !strings.Contains(coverXHTML, `<img src="images/cover.jpg"`) {
		t.Errorf("cover.xhtml missing <img> tag pointing to images/cover.jpg")
	}
}

// 7. Fallback Chapter Lookup: chapters -> drafts -> video scripts
func TestEPUBValidator_ChapterFallbackPriority(t *testing.T) {
	dir := t.TempDir()
	st := store.NewStore(dir)
	if err := st.Init(); err != nil {
		t.Fatal(err)
	}
	if err := st.Progress.Init(3); err != nil {
		t.Fatal(err)
	}
	if err := st.Progress.UpdatePhase(domain.PhaseWriting); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 3; i++ {
		_ = st.Progress.StartChapter(i)
		_ = st.Progress.MarkChapterComplete(i, 1000, "cliff", "main")
	}
	if err := st.Book.Save(domain.BookMetadata{Title: "Fallback Test", Synopsis: "Testing fallback priorities"}); err != nil {
		t.Fatal(err)
	}

	// Chapter 1: Standard final chapter in chapters/01.md
	if err := st.Drafts.SaveFinalChapter(1, "Nội dung từ chapters/01.md"); err != nil {
		t.Fatal(err)
	}

	// Chapter 2: chapters/02.md is missing, but drafts/02.draft.md exists
	// Also create a deceptive video script in *-video/scripts/02-video.md
	// The exporter MUST prioritize drafts/02.draft.md over video scripts!
	if err := st.Drafts.SaveDraft(2, "Nội dung từ drafts/02.draft.md"); err != nil {
		t.Fatal(err)
	}
	videoDir := filepath.Join(dir, "my-video", "scripts")
	if err := os.MkdirAll(videoDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(videoDir, "02-doodle.md"), []byte("Kịch bản video cũ không được ưu tiên!"), 0644); err != nil {
		t.Fatal(err)
	}

	// Chapter 3: chapters/03.md and drafts/03.draft.md are missing, but *-video/scripts/03-video.md exists
	if err := os.WriteFile(filepath.Join(videoDir, "03-doodle.md"), []byte("LỜI: Nội dung video fallback hợp lệ khi không có draft."), 0644); err != nil {
		t.Fatal(err)
	}

	outEPUB := filepath.Join(dir, "fallback_test.epub")
	res, err := Run(context.Background(), Deps{Store: st}, Options{
		Format:    FormatEPUB,
		OutPath:   outEPUB,
		Overwrite: true,
	})
	if err != nil {
		t.Fatalf("Run export with fallback failed: %v", err)
	}
	if res.Chapters != 3 {
		t.Errorf("expected 3 chapters exported, got %d", res.Chapters)
	}

	epubBytes, err := os.ReadFile(outEPUB)
	if err != nil {
		t.Fatalf("read exported epub: %v", err)
	}
	insp := inspectEPUB(t, epubBytes)

	// Chapter 1 must have content from chapters/01.md
	ch1 := string(insp.files["OEBPS/chapter001.xhtml"])
	if !strings.Contains(ch1, "Nội dung từ chapters/01.md") {
		t.Errorf("ch1 expected content from chapters/01.md, got: %s", ch1)
	}

	// Chapter 2 MUST have content from drafts/02.draft.md, NOT the video script
	ch2 := string(insp.files["OEBPS/chapter002.xhtml"])
	if !strings.Contains(ch2, "Nội dung từ drafts/02.draft.md") {
		t.Errorf("ch2 expected content from drafts/02.draft.md, got: %s", ch2)
	}
	if strings.Contains(ch2, "Kịch bản video cũ không được ưu tiên!") {
		t.Errorf("ch2 incorrectly used video script over draft!")
	}

	// Chapter 3 must have fallback content from video script
	ch3 := string(insp.files["OEBPS/chapter003.xhtml"])
	if !strings.Contains(ch3, "Nội dung video fallback hợp lệ") {
		t.Errorf("ch3 expected video fallback content, got: %s", ch3)
	}
}

// 8. Long Book (30+ chapters) Stress & Memory Test
func TestEPUBValidator_LargeBookExport(t *testing.T) {
	numChapters := 35
	chapters := make([]int, numChapters)
	titleIdx := make(chapterTitleIndex, numChapters)
	bodies := make(map[int]string, numChapters)

	for i := 1; i <= numChapters; i++ {
		chapters[i-1] = i
		titleIdx[i] = fmt.Sprintf("Hồi thứ %d: Thử thách phong vân", i)
		bodies[i] = fmt.Sprintf(
			"Đây là nội dung chương %d của tác phẩm trường thiên đồ sộ.\n\n"+
				"Gió thét mây trôi, anh hùng xuất thế tại hồi %d.\n\n***\n\n"+
				"> Trích dẫn danh ngôn chương %d.\n\nKết thúc chương %d.",
			i, i, i, i,
		)
	}

	start := time.Now()
	data, err := renderEPUB(
		domain.BookMetadata{
			Title:    "Bách Niên Phong Vân",
			Synopsis: "Tác phẩm trường thiên lịch sử quy mô lớn.",
			Author:   "Sử gia",
			Subject:  "Kỳ ảo lịch sử",
		},
		chapters,
		titleIdx,
		nil,
		bodies,
	)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("renderEPUB large book: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("empty epub output for large book")
	}

	// Export should be fast (well under 2 seconds)
	if duration > 3*time.Second {
		t.Errorf("large book export too slow: took %v", duration)
	}

	insp := inspectEPUB(t, data)
	if len(insp.spine) != numChapters+2 { // cover + nav + numChapters
		t.Errorf("expected spine length %d, got %d", numChapters+2, len(insp.spine))
	}

	// Verify all chapters well-formed
	for i := 1; i <= numChapters; i++ {
		chFile := fmt.Sprintf("OEBPS/chapter%03d.xhtml", i)
		content, ok := insp.files[chFile]
		if !ok {
			t.Fatalf("missing %s", chFile)
		}
		assertWellFormedXML(t, chFile, content)
	}
}

// 9. Web UI Export Integration: RFC 5987 / RFC 6266 Headers and EPUB download
func TestEPUBValidator_WebExportEndpoint(t *testing.T) {
	dir := t.TempDir()
	st := store.NewStore(dir)
	if err := st.Init(); err != nil {
		t.Fatal(err)
	}
	if err := st.Progress.Init(1); err != nil {
		t.Fatal(err)
	}
	if err := st.Progress.UpdatePhase(domain.PhaseWriting); err != nil {
		t.Fatal(err)
	}
	_ = st.Progress.StartChapter(1)
	_ = st.Progress.MarkChapterComplete(1, 500, "cliff", "main")
	vietnameseTitle := "Ngọc Sáng Trời Nam (Bản Chuẩn)"
	if err := st.Book.Save(domain.BookMetadata{Title: vietnameseTitle, Synopsis: "Sách kiểm tra xuất bản web"}); err != nil {
		t.Fatal(err)
	}
	if err := st.Drafts.SaveFinalChapter(1, "Chương 1 mở đầu."); err != nil {
		t.Fatal(err)
	}

	// Simulate Web UI Export handler logic
	opts := Options{
		Format:    FormatEPUB,
		OutPath:   filepath.Join(dir, sanitizeFileName(vietnameseTitle)+".epub"),
		Overwrite: true,
	}
	res, err := Run(context.Background(), Deps{Store: st}, opts)
	if err != nil {
		t.Fatalf("Run EPUB export: %v", err)
	}

	fileName := filepath.Base(res.Path)
	data, err := os.ReadFile(res.Path)
	if err != nil {
		t.Fatalf("read exported file: %v", err)
	}

	// Verify RFC 5987 / RFC 6266 headers as constructed in Web UI
	rec := httptest.NewRecorder()
	rec.Header().Set("Content-Type", "application/epub+zip")
	asciiFallback := sanitizeFileName(fileName)
	encodedFileName := url.PathEscape(fileName)
	rec.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q; filename*=UTF-8''%s", asciiFallback, encodedFileName))
	rec.WriteHeader(http.StatusOK)
	_, _ = rec.Write(data)

	resp := rec.Result()
	if resp.Header.Get("Content-Type") != "application/epub+zip" {
		t.Errorf("expected Content-Type application/epub+zip, got %q", resp.Header.Get("Content-Type"))
	}
	disp := resp.Header.Get("Content-Disposition")
	if !strings.Contains(disp, "attachment;") || !strings.Contains(disp, "filename*=") || !strings.Contains(disp, "UTF-8''") {
		t.Errorf("Content-Disposition header does not comply with RFC 5987 / 6266: %q", disp)
	}

	// Assert the served bytes form a valid EPUB
	insp := inspectEPUB(t, data)
	if insp.files["mimetype"] == nil {
		t.Errorf("served EPUB is missing mimetype entry")
	}
}
