package exp

import (
	"archive/zip"
	"bytes"
	"crypto/sha1"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/text/unicode/norm"

	"github.com/voocel/ainovel-cli/internal/domain"
)

// CoverImage biểu diễn tệp ảnh bìa (JPEG, PNG, WEBP) nhúng vào EPUB.
type CoverImage struct {
	Data      []byte
	Filename  string // vd: "cover.jpg", "cover.png"
	MediaType string // vd: "image/jpeg", "image/png"
}

// epubRenderConfig chứa các tùy chọn mở rộng khi kết xuất EPUB.
type epubRenderConfig struct {
	volumes    []domain.VolumeOutline
	coverImage *CoverImage
	modifiedAt *time.Time
}

// EPUBRenderOption là hàm cấu hình cho renderEPUB.
type EPUBRenderOption func(*epubRenderConfig)

// WithVolumes cung cấp cấu trúc phân tầng Volume -> Arc -> Chapter cho Navigation Document.
func WithVolumes(volumes []domain.VolumeOutline) EPUBRenderOption {
	return func(c *epubRenderConfig) {
		c.volumes = volumes
	}
}

// WithCoverImage cung cấp ảnh bìa nhúng vào EPUB với thuộc tính cover-image.
func WithCoverImage(img *CoverImage) EPUBRenderOption {
	return func(c *epubRenderConfig) {
		c.coverImage = img
	}
}

// WithModifiedTime cung cấp mốc thời gian modified xác định (dùng cho tái lập hoặc test).
func WithModifiedTime(t time.Time) EPUBRenderOption {
	return func(c *epubRenderConfig) {
		c.modifiedAt = &t
	}
}

// ExportEPUB là hàm công khai kết xuất tài liệu chuẩn W3C EPUB 3.3.
func ExportEPUB(
	book domain.BookMetadata,
	chapters []int,
	titleIdx map[int]string,
	locations map[int]chapterLocation,
	bodies map[int]string,
	opts ...EPUBRenderOption,
) ([]byte, error) {
	return renderEPUB(book, chapters, chapterTitleIndex(titleIdx), locations, bodies, opts...)
}

// renderEPUB đóng gói các chương đã chọn thành tệp W3C EPUB 3.3 hoàn chỉnh.
//
// Cấu trúc gói OCF Container:
//
//	mimetype                    (Entry 0, Method=Store không nén, exact bytes, len(Extra)==0)
//	META-INF/container.xml      (trỏ tới OEBPS/content.opf)
//	OEBPS/content.opf           (OPF 3.0: Dublin Core + Accessibility 1.1 + manifest + spine)
//	OEBPS/nav.xhtml             (XHTML5 Navigation Document: hierarchical toc + landmarks)
//	OEBPS/style.css             (CSS chuẩn kiểu dáng đơn vị tương đối em)
//	OEBPS/cover.xhtml           (Trang bìa XHTML5, tùy chọn nếu có tên sách hoặc ảnh bìa)
//	OEBPS/images/{cover-file}   (Ảnh bìa manifest properties="cover-image", tùy chọn)
//	OEBPS/chapterNNN.xhtml      (Mỗi chương một tệp XHTML5, chuẩn hóa NFC, lọc control chars)
func renderEPUB(
	book domain.BookMetadata,
	chapters []int,
	titleIdx chapterTitleIndex,
	locations map[int]chapterLocation,
	bodies map[int]string,
	opts ...EPUBRenderOption,
) ([]byte, error) {
	cfg := epubRenderConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// 1. mimetype: MUST be entry 0, Method=Store (0), uncompressed, len(Extra)==0
	mtHeader := &zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store,
	}
	mtHeader.Extra = nil
	mt, err := zw.CreateHeader(mtHeader)
	if err != nil {
		return nil, fmt.Errorf("create mimetype entry: %w", err)
	}
	if _, err := mt.Write([]byte("application/epub+zip")); err != nil {
		return nil, fmt.Errorf("write mimetype content: %w", err)
	}

	// 2. META-INF/container.xml
	if err := zipDeflate(zw, "META-INF/container.xml", containerXML); err != nil {
		return nil, err
	}

	// 3. OEBPS/style.css
	if err := zipDeflate(zw, "OEBPS/style.css", styleCSS); err != nil {
		return nil, err
	}

	// 4. Ảnh bìa (Cover Image) nếu có
	coverImg := cfg.coverImage
	var coverImageRelPath string
	if coverImg != nil && len(coverImg.Data) > 0 {
		imgName := coverImg.Filename
		if imgName == "" {
			imgName = "cover.jpg"
		}
		coverImageRelPath = "images/" + filepath.Base(imgName)
		w, err := zw.Create("OEBPS/" + coverImageRelPath)
		if err != nil {
			return nil, fmt.Errorf("create cover image entry: %w", err)
		}
		if _, err := w.Write(coverImg.Data); err != nil {
			return nil, fmt.Errorf("write cover image: %w", err)
		}
	}

	// 5. Trang bìa (Cover Page)
	hasCover := strings.TrimSpace(book.Title) != "" || coverImg != nil
	if hasCover {
		coverXHTML := renderCoverXHTML(book.Title, book.Author, book.Synopsis, coverImg, coverImageRelPath)
		if err := zipDeflate(zw, "OEBPS/cover.xhtml", coverXHTML); err != nil {
			return nil, err
		}
	}

	// 6. Các chương XHTML5
	for _, ch := range chapters {
		loc, hasLoc := locations[ch]
		rawTitle := strings.TrimSpace(titleIdx[ch])
		rawBody := strings.TrimSpace(bodies[ch])
		strippedBody := stripChapterTitleHeader(rawBody, rawTitle)
		xhtml := renderChapterXHTML(ch, rawTitle, loc, hasLoc, strippedBody)
		if err := zipDeflate(zw, "OEBPS/"+chapterFileName(ch), xhtml); err != nil {
			return nil, err
		}
	}

	// 7. Navigation Document (nav.xhtml)
	navContent := renderNavXHTML(hasCover, chapters, titleIdx, cfg.volumes, locations)
	if err := zipDeflate(zw, "OEBPS/nav.xhtml", navContent); err != nil {
		return nil, err
	}

	// 8. Package Document (content.opf)
	opfContent := renderOPF(book, hasCover, chapters, coverImg, coverImageRelPath, cfg.modifiedAt)
	if err := zipDeflate(zw, "OEBPS/content.opf", opfContent); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("finalize zip archive: %w", err)
	}
	return buf.Bytes(), nil
}

// zipDeflate ghi tệp nén thông thường vào ZIP.
func zipDeflate(zw *zip.Writer, name, content string) error {
	w, err := zw.Create(name)
	if err != nil {
		return fmt.Errorf("create %s: %w", name, err)
	}
	_, err = w.Write([]byte(content))
	return err
}

func chapterFileName(ch int) string {
	return fmt.Sprintf("chapter%03d.xhtml", ch)
}

func chapterID(ch int) string {
	return fmt.Sprintf("ch%03d", ch)
}

// cleanXMLText loại bỏ các ký tự điều khiển cấm trong XML 1.0 và chuẩn hóa Unicode NFC.
//
// XML 1.0 Fifth Edition quy định ký tự hợp lệ:
//
//	Char ::= #x9 | #xA | #xD | [#x20-#xD7FF] | [#xE000-#xFFFD] | [#x10000-#x10FFFF]
//
// Ký tự U+0000 đến U+001F (ngoại trừ 0x09 tab, 0x0A LF, 0x0D CR) và 0xFFFE, 0xFFFF bị cấm tuyệt đối.
func cleanXMLText(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if (r < 0x20 && r != 0x09 && r != 0x0A && r != 0x0D) || r == 0xFFFE || r == 0xFFFF {
			continue
		}
		b.WriteRune(r)
	}
	return norm.NFC.String(b.String())
}

// xmlEscapeEPUB chuyển đổi chuỗi an toàn thành XML entity, làm sạch control chars và chuẩn hóa NFC.
func xmlEscapeEPUB(s string) string {
	s = cleanXMLText(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&#39;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// OCF META-INF/container.xml ──────────────────────────────────────────

const containerXML = `<?xml version="1.0" encoding="utf-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>
`

// CSS Stylesheet với đơn vị tương đối em ─────────────────────────────

const styleCSS = `@charset "utf-8";
body {
  font-family: serif;
  line-height: 1.8;
  margin: 1em;
  padding: 0;
  color: #111111;
}
h1.book-title {
  font-size: 2.2em;
  text-align: center;
  margin: 3em 0 0.8em;
  font-weight: bold;
}
p.book-author {
  text-align: center;
  font-size: 1.2em;
  margin-bottom: 2em;
  color: #555555;
}
.synopsis {
  margin: 2em 1.5em;
  font-style: italic;
  color: #444444;
  border-top: 1px solid #dddddd;
  padding-top: 1.5em;
}
.volume-divider {
  font-size: 1.6em;
  text-align: center;
  margin: 3em 0 1.5em;
  font-weight: bold;
  border-bottom: 1px solid #cccccc;
  padding-bottom: 0.5em;
}
h1.chapter-title {
  font-size: 1.5em;
  text-align: center;
  margin: 2em 0 1.5em;
  font-weight: bold;
}
h2 {
  font-size: 1.3em;
  margin: 1.8em 0 0.8em;
}
h3 {
  font-size: 1.1em;
  margin: 1.5em 0 0.6em;
}
p {
  text-indent: 1.8em;
  margin: 0.6em 0;
  text-align: justify;
}
hr.scene-break {
  border: none;
  text-align: center;
  margin: 2em 0;
}
hr.scene-break::before {
  content: "* * *";
  font-size: 1.2em;
  letter-spacing: 0.5em;
  color: #666666;
}
blockquote {
  margin: 1.2em 1.5em;
  padding: 0.5em 1em;
  border-left: 3px solid #888888;
  font-style: italic;
  color: #444444;
}
blockquote p {
  text-indent: 0;
}
section.postscript {
  margin-top: 3em;
  padding-top: 1.5em;
  border-top: 1px dashed #888888;
}
section.postscript h2 {
  font-size: 1.2em;
  color: #333333;
}
nav#toc ol {
  list-style-type: none;
  padding-left: 1.2em;
  line-height: 1.8;
}
nav#toc > ol {
  padding-left: 0;
}
nav#toc li {
  margin: 0.4em 0;
}
nav#toc a {
  text-decoration: none;
  color: #1a0dab;
}
nav#toc span {
  font-weight: bold;
  color: #333333;
}
.cover-image-wrap {
  text-align: center;
  margin: 1em 0;
}
img.cover-image {
  max-width: 100%;
  height: auto;
  display: block;
  margin: 0 auto;
}
`

// XHTML5 Chapter Rendering ──────────────────────────────────────────

func renderChapterXHTML(ch int, title string, loc chapterLocation, hasLoc bool, body string) string {
	var b strings.Builder
	displayTitle := fmt.Sprintf("Tập %d", ch)
	if title != "" {
		displayTitle = fmt.Sprintf("Tập %d: %s", ch, title)
	}

	fmt.Fprintf(&b, `<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" lang="vi" xml:lang="vi">
<head>
  <meta charset="utf-8"/>
  <title>%s</title>
  <link rel="stylesheet" type="text/css" href="style.css"/>
</head>
<body>
`, xmlEscapeEPUB(displayTitle))

	if hasLoc && loc.IsFirstOfVolume {
		fmt.Fprintf(&b, "  <div class=\"volume-divider\">Mùa %d: %s</div>\n",
			loc.VolumeIdx, xmlEscapeEPUB(strings.TrimSpace(loc.VolumeTitle)))
	}

	fmt.Fprintf(&b, "  <h1 class=\"chapter-title\">%s</h1>\n", xmlEscapeEPUB(displayTitle))
	b.WriteString(renderMarkdownToXHTML(body))
	b.WriteString("</body>\n</html>\n")
	return b.String()
}

// renderMarkdownToXHTML chuyển đổi Markdown cơ bản sang XHTML ngữ nghĩa và an toàn.
// Hỗ trợ đoạn văn, ngắt cảnh (*** / ---), trích dẫn (>), tiêu đề, và khối Hậu ký (postscript).
func renderMarkdownToXHTML(body string) string {
	body = cleanXMLText(body)
	body = strings.ReplaceAll(body, "\r\n", "\n")
	blocks := strings.Split(body, "\n\n")

	var b strings.Builder
	inPostscript := false

	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		lines := strings.Split(block, "\n")
		firstLine := strings.TrimSpace(lines[0])

		// 1. Ngắt cảnh: *** / --- / * * * / ___
		if len(lines) == 1 && isSceneBreak(firstLine) {
			b.WriteString("  <hr class=\"scene-break\"/>\n")
			continue
		}

		// 2. Trích dẫn blockquote: các dòng bắt đầu bằng >
		if strings.HasPrefix(firstLine, ">") {
			var quoteLines []string
			for _, l := range lines {
				l = strings.TrimSpace(l)
				l = strings.TrimPrefix(l, ">")
				l = strings.TrimSpace(l)
				if l != "" {
					quoteLines = append(quoteLines, xmlEscapeEPUB(l))
				}
			}
			if len(quoteLines) > 0 {
				b.WriteString("  <blockquote><p>" + strings.Join(quoteLines, " ") + "</p></blockquote>\n")
			}
			continue
		}

		// 3. Tiêu đề cấp 3 (###)
		if strings.HasPrefix(firstLine, "### ") {
			headingText := strings.TrimSpace(strings.TrimPrefix(firstLine, "### "))
			b.WriteString("  <h3>" + xmlEscapeEPUB(headingText) + "</h3>\n")
			if len(lines) > 1 {
				rest := strings.TrimSpace(strings.Join(lines[1:], "\n"))
				if rest != "" {
					b.WriteString("  <p>" + xmlEscapeEPUB(strings.ReplaceAll(rest, "\n", " ")) + "</p>\n")
				}
			}
			continue
		}

		// 4. Tiêu đề cấp 2 (##): nhận diện Hậu ký / Ghi chú lịch sử
		if strings.HasPrefix(firstLine, "## ") {
			headingText := strings.TrimSpace(strings.TrimPrefix(firstLine, "## "))
			if isPostscriptHeading(headingText) {
				if !inPostscript {
					inPostscript = true
					b.WriteString("  <section class=\"postscript\">\n")
				}
			}
			b.WriteString("  <h2>" + xmlEscapeEPUB(headingText) + "</h2>\n")
			if len(lines) > 1 {
				rest := strings.TrimSpace(strings.Join(lines[1:], "\n"))
				if rest != "" {
					b.WriteString("  <p>" + xmlEscapeEPUB(strings.ReplaceAll(rest, "\n", " ")) + "</p>\n")
				}
			}
			continue
		}

		// 5. Đoạn văn thông thường
		pText := strings.ReplaceAll(block, "\n", " ")
		b.WriteString("  <p>" + xmlEscapeEPUB(pText) + "</p>\n")
	}

	if inPostscript {
		b.WriteString("  </section>\n")
	}

	return b.String()
}

func isSceneBreak(s string) bool {
	s = strings.TrimSpace(s)
	return s == "***" || s == "---" || s == "* * *" || s == "___" || s == "- - -" || s == "_ _ _"
}

func isPostscriptHeading(s string) bool {
	lower := strings.ToLower(s)
	return strings.HasPrefix(lower, "hậu ký") ||
		strings.HasPrefix(lower, "ghi chú lịch sử") ||
		strings.HasPrefix(lower, "lời tác giả") ||
		strings.HasPrefix(lower, "ghi chú của tác giả") ||
		strings.HasPrefix(lower, "lời bàn") ||
		strings.HasPrefix(lower, "khảo cứu") ||
		strings.HasPrefix(lower, "postscript") ||
		strings.HasPrefix(lower, "author's note")
}

// splitParagraphs phân tách các đoạn văn (tương thích cho các unit test).
func splitParagraphs(body string) []string {
	body = strings.ReplaceAll(body, "\r\n", "\n")
	parts := strings.Split(body, "\n\n")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		p = strings.ReplaceAll(p, "\n", " ")
		out = append(out, p)
	}
	return out
}

// XHTML5 Cover Page ──────────────────────────────────────────────────

func renderCoverXHTML(novelName, author, synopsis string, coverImg *CoverImage, coverImgPath string) string {
	var b strings.Builder
	title := xmlEscapeEPUB(strings.TrimSpace(novelName))
	if title == "" {
		title = "Bìa sách"
	}

	fmt.Fprintf(&b, `<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" lang="vi" xml:lang="vi">
<head>
  <meta charset="utf-8"/>
  <title>%s</title>
  <link rel="stylesheet" type="text/css" href="style.css"/>
</head>
<body class="cover-body">
`, title)

	if coverImg != nil && coverImgPath != "" {
		fmt.Fprintf(&b, `  <div class="cover-image-wrap">
    <img src="%s" alt="Bìa sách %s" class="cover-image"/>
  </div>
`, coverImgPath, title)
	}

	if name := strings.TrimSpace(novelName); name != "" {
		fmt.Fprintf(&b, "  <h1 class=\"book-title\">%s</h1>\n", xmlEscapeEPUB(name))
	}
	if a := strings.TrimSpace(author); a != "" {
		fmt.Fprintf(&b, "  <p class=\"book-author\">Tác giả: %s</p>\n", xmlEscapeEPUB(a))
	}
	if s := strings.TrimSpace(synopsis); s != "" {
		fmt.Fprintf(&b, "  <div class=\"synopsis\"><p>%s</p></div>\n", xmlEscapeEPUB(s))
	}

	b.WriteString("</body>\n</html>\n")
	return b.String()
}

// Navigation Document (nav.xhtml) ────────────────────────────────────

func renderNavXHTML(
	hasCover bool,
	chapters []int,
	titleIdx chapterTitleIndex,
	volumes []domain.VolumeOutline,
	locations map[int]chapterLocation,
) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" lang="vi" xml:lang="vi">
<head>
  <meta charset="utf-8"/>
  <title>Mục lục</title>
  <link rel="stylesheet" type="text/css" href="style.css"/>
</head>
<body>
  <nav epub:type="toc" id="toc">
    <h1 class="toc-title">Mục lục</h1>
    <ol>
`)
	if hasCover {
		b.WriteString("      <li><a href=\"cover.xhtml\">Bìa sách</a></li>\n")
	}

	// Tập hợp các chương xuất để lọc nhanh
	chSet := make(map[int]bool, len(chapters))
	for _, ch := range chapters {
		chSet[ch] = true
	}

	if len(volumes) > 0 {
		// Kết xuất phân cấp: Volume -> Arc -> Chapter
		globalCh := 1
		for _, v := range volumes {
			// Thu thập các arc có chương nằm trong chSet
			type arcEntry struct {
				arc      domain.ArcOutline
				chapters []domain.OutlineEntry
			}
			var validArcs []arcEntry

			for _, a := range v.Arcs {
				var inArcChaps []domain.OutlineEntry
				for _, e := range a.Chapters {
					e.Chapter = globalCh
					globalCh++
					if chSet[e.Chapter] {
						inArcChaps = append(inArcChaps, e)
					}
				}
				if len(inArcChaps) > 0 {
					validArcs = append(validArcs, arcEntry{arc: a, chapters: inArcChaps})
				}
			}

			if len(validArcs) > 0 {
				volTitle := strings.TrimSpace(v.Title)
				if volTitle == "" {
					volTitle = fmt.Sprintf("Quyển %d", v.Index)
				}
				fmt.Fprintf(&b, "      <li>\n        <span>Mùa %d: %s</span>\n        <ol>\n",
					v.Index, xmlEscapeEPUB(volTitle))

				for _, ve := range validArcs {
					arcTitle := strings.TrimSpace(ve.arc.Title)
					if arcTitle != "" {
						fmt.Fprintf(&b, "          <li>\n            <span>Hồi %d: %s</span>\n            <ol>\n",
							ve.arc.Index, xmlEscapeEPUB(arcTitle))
					} else {
						fmt.Fprintf(&b, "          <li>\n            <span>Hồi %d</span>\n            <ol>\n",
							ve.arc.Index)
					}

					for _, e := range ve.chapters {
						title := strings.TrimSpace(titleIdx[e.Chapter])
						if title == "" {
							title = strings.TrimSpace(e.Title)
						}
						display := fmt.Sprintf("Tập %d", e.Chapter)
						if title != "" {
							display = fmt.Sprintf("Tập %d: %s", e.Chapter, title)
						}
						fmt.Fprintf(&b, "              <li><a href=\"%s\">%s</a></li>\n",
							chapterFileName(e.Chapter), xmlEscapeEPUB(display))
					}

					b.WriteString("            </ol>\n          </li>\n")
				}

				b.WriteString("        </ol>\n      </li>\n")
			}
		}
	} else if len(locations) > 0 && hasVolumeInLocations(chapters, locations) {
		// Phân nhóm theo Volume từ locations
		currentVol := -1
		for _, ch := range chapters {
			loc, ok := locations[ch]
			if ok && loc.VolumeIdx > 0 && loc.VolumeIdx != currentVol {
				if currentVol != -1 {
					b.WriteString("        </ol>\n      </li>\n")
				}
				currentVol = loc.VolumeIdx
				volTitle := strings.TrimSpace(loc.VolumeTitle)
				if volTitle == "" {
					volTitle = fmt.Sprintf("Quyển %d", loc.VolumeIdx)
				}
				fmt.Fprintf(&b, "      <li>\n        <span>Mùa %d: %s</span>\n        <ol>\n",
					loc.VolumeIdx, xmlEscapeEPUB(volTitle))
			}

			title := strings.TrimSpace(titleIdx[ch])
			display := fmt.Sprintf("Tập %d", ch)
			if title != "" {
				display = fmt.Sprintf("Tập %d: %s", ch, title)
			}
			fmt.Fprintf(&b, "          <li><a href=\"%s\">%s</a></li>\n",
				chapterFileName(ch), xmlEscapeEPUB(display))
		}
		if currentVol != -1 {
			b.WriteString("        </ol>\n      </li>\n")
		}
	} else {
		// Danh sách chương phẳng (Flat TOC)
		for _, ch := range chapters {
			title := strings.TrimSpace(titleIdx[ch])
			display := fmt.Sprintf("Tập %d", ch)
			if title != "" {
				display = fmt.Sprintf("Tập %d: %s", ch, title)
			}
			fmt.Fprintf(&b, "      <li><a href=\"%s\">%s</a></li>\n",
				chapterFileName(ch), xmlEscapeEPUB(display))
		}
	}

	b.WriteString(`    </ol>
  </nav>
  <nav epub:type="landmarks" id="landmarks" hidden="">
    <h2>Điều hướng mốc</h2>
    <ol>
`)
	if hasCover {
		b.WriteString("      <li><a epub:type=\"cover\" href=\"cover.xhtml\">Bìa sách</a></li>\n")
	}
	b.WriteString("      <li><a epub:type=\"toc\" href=\"nav.xhtml\">Mục lục</a></li>\n")
	if len(chapters) > 0 {
		fmt.Fprintf(&b, "      <li><a epub:type=\"bodymatter\" href=\"%s\">Bắt đầu đọc</a></li>\n",
			chapterFileName(chapters[0]))
	}

	b.WriteString(`    </ol>
  </nav>
</body>
</html>
`)
	return b.String()
}

func hasVolumeInLocations(chapters []int, locations map[int]chapterLocation) bool {
	for _, ch := range chapters {
		if loc, ok := locations[ch]; ok && loc.VolumeIdx > 0 {
			return true
		}
	}
	return false
}

// Package Document (content.opf) ─────────────────────────────────────

func renderOPF(
	book domain.BookMetadata,
	hasCover bool,
	chapters []int,
	coverImg *CoverImage,
	coverImgPath string,
	modifiedAt *time.Time,
) string {
	bookID := bookIdentifier(book.Title)

	modTime := time.Now().UTC()
	if modifiedAt != nil && !modifiedAt.IsZero() {
		modTime = modifiedAt.UTC()
	}
	modified := modTime.Format("2006-01-02T15:04:05Z")
	dateStr := modTime.Format("2006-01-02")

	title := strings.TrimSpace(book.Title)
	if title == "" {
		title = "Untitled"
	}

	author := strings.TrimSpace(book.Author)
	if author == "" {
		author = "ainovel-cli"
	}

	subject := strings.TrimSpace(book.Subject)
	if subject == "" {
		subject = "Tiểu thuyết"
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="pub-id" xml:lang="vi" prefix="schema: http://schema.org/">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="pub-id">%s</dc:identifier>
    <dc:title>%s</dc:title>
    <dc:language>vi</dc:language>
    <dc:creator id="creator">%s</dc:creator>
    <meta refines="#creator" property="role" scheme="marc:relators">aut</meta>
    <dc:description>%s</dc:description>
    <dc:subject>%s</dc:subject>
    <dc:date>%s</dc:date>
    <meta property="dcterms:modified">%s</meta>
    <meta property="schema:accessMode">textual</meta>
    <meta property="schema:accessModeSufficient">textual</meta>
    <meta property="schema:accessibilityFeature">structuralNavigation</meta>
    <meta property="schema:accessibilityHazard">none</meta>
    <meta property="schema:accessibilitySummary">Ấn bản điện tử chuẩn EPUB 3.3 hỗ trợ điều hướng cấu trúc ngữ nghĩa đầy đủ và tương thích trình đọc màn hình.</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
    <item id="css" href="style.css" media-type="text/css"/>
`, xmlEscapeEPUB(bookID), xmlEscapeEPUB(title), xmlEscapeEPUB(author), xmlEscapeEPUB(book.Synopsis), xmlEscapeEPUB(subject), dateStr, modified)

	if coverImg != nil && coverImgPath != "" {
		mime := coverImg.MediaType
		if mime == "" {
			mime = "image/jpeg"
		}
		fmt.Fprintf(&b, "    <item id=\"cover-image\" href=\"%s\" media-type=\"%s\" properties=\"cover-image\"/>\n",
			coverImgPath, mime)
	}

	if hasCover {
		b.WriteString(`    <item id="cover" href="cover.xhtml" media-type="application/xhtml+xml"/>` + "\n")
	}

	for _, ch := range chapters {
		fmt.Fprintf(&b, "    <item id=\"%s\" href=\"%s\" media-type=\"application/xhtml+xml\"/>\n",
			chapterID(ch), chapterFileName(ch))
	}

	b.WriteString("  </manifest>\n  <spine>\n")
	if hasCover {
		b.WriteString(`    <itemref idref="cover"/>` + "\n")
	}
	b.WriteString(`    <itemref idref="nav"/>` + "\n")
	for _, ch := range chapters {
		fmt.Fprintf(&b, "    <itemref idref=\"%s\"/>\n", chapterID(ch))
	}
	b.WriteString("  </spine>\n</package>\n")
	return b.String()
}

// bookIdentifier tạo ra chuỗi URN UUID ổn định dựa trên tên tác phẩm.
func bookIdentifier(novelName string) string {
	h := sha1.New()
	h.Write([]byte(novelName))
	sum := h.Sum(nil)
	return fmt.Sprintf("urn:uuid:%x-%x-%x-%x-%x",
		sum[0:4], sum[4:6], sum[6:8], sum[8:10], sum[10:16])
}
