package exp

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"

	"github.com/voocel/ainovel-cli/internal/domain"
)

const docxContentTypes = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
  <Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
</Types>`

const docxRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

const docxDocumentRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`

const docxStyles = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:docDefaults>
    <w:rPrDefault>
      <w:rPr>
        <w:rFonts w:ascii="Times New Roman" w:eastAsia="Times New Roman" w:hAnsi="Times New Roman" w:cs="Times New Roman"/>
        <w:sz w:val="24"/>
        <w:szCs w:val="24"/>
        <w:lang w:val="vi-VN"/>
      </w:rPr>
    </w:rPrDefault>
    <w:pPrDefault>
      <w:pPr>
        <w:spacing w:after="160" w:line="276" w:lineRule="auto"/>
      </w:pPr>
    </w:pPrDefault>
  </w:docDefaults>

  <w:style w:type="paragraph" w:default="1" w:styleId="Normal">
    <w:name w:val="Normal"/>
    <w:qFormat/>
  </w:style>

  <w:style w:type="paragraph" w:styleId="Title">
    <w:name w:val="Title"/>
    <w:basedOn w:val="Normal"/>
    <w:next w:val="Normal"/>
    <w:qFormat/>
    <w:pPr>
      <w:spacing w:before="240" w:after="360"/>
      <w:jc w:val="center"/>
    </w:pPr>
    <w:rPr>
      <w:b/>
      <w:sz w:val="56"/>
      <w:szCs w:val="56"/>
      <w:color w:val="1B365D"/>
    </w:rPr>
  </w:style>

  <w:style w:type="paragraph" w:styleId="Heading1">
    <w:name w:val="heading 1"/>
    <w:basedOn w:val="Normal"/>
    <w:next w:val="Normal"/>
    <w:qFormat/>
    <w:pPr>
      <w:spacing w:before="360" w:after="180"/>
      <w:jc w:val="center"/>
    </w:pPr>
    <w:rPr>
      <w:b/>
      <w:sz w:val="36"/>
      <w:szCs w:val="36"/>
      <w:color w:val="2E5B88"/>
    </w:rPr>
  </w:style>

  <w:style w:type="paragraph" w:styleId="Heading2">
    <w:name w:val="heading 2"/>
    <w:basedOn w:val="Normal"/>
    <w:next w:val="Normal"/>
    <w:qFormat/>
    <w:pPr>
      <w:spacing w:before="300" w:after="140"/>
    </w:pPr>
    <w:rPr>
      <w:b/>
      <w:sz w:val="32"/>
      <w:szCs w:val="32"/>
      <w:color w:val="2E5B88"/>
    </w:rPr>
  </w:style>
</w:styles>`

// renderDOCX đóng gói tập hợp các chương thành file Word (.docx) chuẩn Office Open XML.
func renderDOCX(
	book domain.BookMetadata,
	chapters []int,
	titleIdx chapterTitleIndex,
	locations map[int]chapterLocation,
	bodies map[int]string,
) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	if err := zipDeflate(zw, "[Content_Types].xml", docxContentTypes); err != nil {
		return nil, err
	}
	if err := zipDeflate(zw, "_rels/.rels", docxRels); err != nil {
		return nil, err
	}
	if err := zipDeflate(zw, "word/_rels/document.xml.rels", docxDocumentRels); err != nil {
		return nil, err
	}
	if err := zipDeflate(zw, "word/styles.xml", docxStyles); err != nil {
		return nil, err
	}

	docXML := buildDocumentXML(book.Title, chapters, titleIdx, locations, bodies)
	if err := zipDeflate(zw, "word/document.xml", docXML); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("đóng gói zip docx: %w", err)
	}

	return buf.Bytes(), nil
}

func buildDocumentXML(
	bookTitle string,
	chapters []int,
	titleIdx chapterTitleIndex,
	locations map[int]chapterLocation,
	bodies map[int]string,
) string {
	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`)
	b.WriteString(`<w:body>`)

	// 1. Tiêu đề tác phẩm
	if strings.TrimSpace(bookTitle) != "" {
		b.WriteString(`<w:p>`)
		b.WriteString(`<w:pPr><w:pStyle w:val="Title"/><w:jc w:val="center"/></w:pPr>`)
		b.WriteString(`<w:r>`)
		b.WriteString(`<w:rPr><w:b/><w:sz w:val="56"/><w:szCs w:val="56"/><w:color w:val="1B365D"/></w:rPr>`)
		b.WriteString(fmt.Sprintf(`<w:t xml:space="preserve">%s</w:t>`, xmlEscape(bookTitle)))
		b.WriteString(`</w:r>`)
		b.WriteString(`</w:p>`)
	}

	lastVol := -1

	for idx, ch := range chapters {
		// Phân cách quyển / volume nếu có
		if loc, ok := locations[ch]; ok && loc.VolumeIdx != lastVol {
			lastVol = loc.VolumeIdx
			volTitle := loc.VolumeTitle
			if volTitle == "" {
				volTitle = fmt.Sprintf("Quyển %d", loc.VolumeIdx)
			}
			b.WriteString(`<w:p>`)
			b.WriteString(`<w:pPr><w:pStyle w:val="Heading1"/><w:jc w:val="center"/></w:pPr>`)
			b.WriteString(`<w:r>`)
			b.WriteString(`<w:rPr><w:b/><w:sz w:val="36"/><w:szCs w:val="36"/><w:color w:val="2E5B88"/></w:rPr>`)
			b.WriteString(fmt.Sprintf(`<w:t xml:space="preserve">%s</w:t>`, xmlEscape(volTitle)))
			b.WriteString(`</w:r>`)
			b.WriteString(`</w:p>`)
		}

		// Tiêu đề chương
		rawTitle := titleIdx[ch]
		chTitle := fmt.Sprintf("Chương %d", ch)
		if strings.TrimSpace(rawTitle) != "" {
			chTitle = fmt.Sprintf("Chương %d: %s", ch, rawTitle)
		}

		b.WriteString(`<w:p>`)
		b.WriteString(`<w:pPr><w:pStyle w:val="Heading2"/></w:pPr>`)
		b.WriteString(`<w:r>`)
		b.WriteString(`<w:rPr><w:b/><w:sz w:val="32"/><w:szCs w:val="32"/><w:color w:val="2E5B88"/></w:rPr>`)
		b.WriteString(fmt.Sprintf(`<w:t xml:space="preserve">%s</w:t>`, xmlEscape(chTitle)))
		b.WriteString(`</w:r>`)
		b.WriteString(`</w:p>`)

		// Nội dung thân chương
		body := bodies[ch]
		body = stripChapterTitleHeader(body, rawTitle)
		lines := strings.Split(body, "\n")

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Xử lý các tiêu đề con markdown (#, ##, ###)
			isHeading := false
			if strings.HasPrefix(line, "#") {
				isHeading = true
				line = strings.TrimLeft(line, "# ")
			}

			b.WriteString(`<w:p>`)
			if isHeading {
				b.WriteString(`<w:pPr><w:spacing w:before="200" w:after="100"/></w:pPr>`)
				b.WriteString(`<w:r>`)
				b.WriteString(`<w:rPr><w:b/><w:sz w:val="26"/><w:szCs w:val="26"/></w:rPr>`)
				b.WriteString(fmt.Sprintf(`<w:t xml:space="preserve">%s</w:t>`, xmlEscape(line)))
				b.WriteString(`</w:r>`)
			} else {
				// Đoạn văn thông thường có thụt đầu dòng (firstLine indent)
				b.WriteString(`<w:pPr><w:ind w:firstLine="360"/><w:spacing w:after="140" w:line="276" w:lineRule="auto"/></w:pPr>`)
				b.WriteString(`<w:r>`)
				b.WriteString(`<w:rPr><w:sz w:val="24"/><w:szCs w:val="24"/></w:rPr>`)
				b.WriteString(fmt.Sprintf(`<w:t xml:space="preserve">%s</w:t>`, xmlEscape(line)))
				b.WriteString(`</w:r>`)
			}
			b.WriteString(`</w:p>`)
		}

		// Ngắt trang giữa các chương (trừ chương cuối cùng)
		if idx < len(chapters)-1 {
			b.WriteString(`<w:p><w:r><w:br w:type="page"/></w:r></w:p>`)
		}
	}

	// Kích thước trang A4 và lề tiêu chuẩn (1 inch = 1440 twips)
	b.WriteString(`<w:sectPr>`)
	b.WriteString(`<w:pgSz w:w="11906" w:h="16838"/>`)
	b.WriteString(`<w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440" w:header="720" w:footer="720"/>`)
	b.WriteString(`</w:sectPr>`)

	b.WriteString(`</w:body>`)
	b.WriteString(`</w:document>`)

	return b.String()
}

// xmlEscape mã hóa các ký tự đặc biệt trong XML và loại bỏ ký tự điều khiển không hợp lệ.
func xmlEscape(s string) string {
	var b strings.Builder
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
			b.WriteString("&apos;")
		default:
			// Loại bỏ ký tự điều khiển ASCII < 0x20 ngoại trừ \t, \n, \r
			if r < 0x20 && r != '\t' && r != '\n' && r != '\r' {
				continue
			}
			b.WriteRune(r)
		}
	}
	return b.String()
}
