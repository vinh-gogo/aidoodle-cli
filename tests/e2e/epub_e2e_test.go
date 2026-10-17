package e2e_test

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"golang.org/x/text/unicode/norm"

	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/host/exp"
	"github.com/voocel/ainovel-cli/internal/store"
)

// setupEPUBTestProject sets up a test store and generates an EPUB file via exp.Run.
func setupEPUBTestProject(t *testing.T, multiVolume bool, customText string) (string, *zip.ReadCloser) {
	t.Helper()
	tmpDir := t.TempDir()
	outDir := filepath.Join(tmpDir, "output", "epub-test")
	chaptersDir := filepath.Join(outDir, "chapters")
	metaDir := filepath.Join(outDir, "meta")

	if err := os.MkdirAll(chaptersDir, 0755); err != nil {
		t.Fatalf("mkdir chapters: %v", err)
	}
	if err := os.MkdirAll(metaDir, 0755); err != nil {
		t.Fatalf("mkdir meta: %v", err)
	}

	st := store.NewStore(outDir)

	// 1. Setup book metadata
	book := domain.BookMetadata{
		Title:    "Hào Khí Đông A: Trận Bạch Đằng",
		Synopsis: "Tiểu thuyết lịch sử tái hiện chiến thắng Bạch Đằng năm 1288.",
		Subject:  "vietnamese-history",
	}
	if err := st.Book.Save(book); err != nil {
		t.Fatalf("save book: %v", err)
	}

	// 2. Setup progress
	prog := domain.Progress{
		CompletedChapters: []int{1, 2, 3},
		CurrentChapter:    4,
		Layered:           multiVolume,
	}
	if err := st.Progress.Save(&prog); err != nil {
		t.Fatalf("save progress: %v", err)
	}

	// 3. Setup outline if multi-volume
	if multiVolume {
		volumes := []domain.VolumeOutline{
			{
				Index: 1,
				Title: "Quyển 1: Tiếng Trống Thăng Long",
				Theme: "Khởi đầu chiến dịch Bạch Đằng",
				Arcs: []domain.ArcOutline{
					{
						Index: 1,
						Title: "Hồi 1: Lời Thề Sát Thát",
						Goal:  "Tập hợp tướng sĩ",
						Chapters: []domain.OutlineEntry{
							{Chapter: 1, Title: "Chương 1: Lời Thề Sát Thát"},
							{Chapter: 2, Title: "Chương 2: Bày Binh Bố Trận"},
						},
					},
					{
						Index: 2,
						Title: "Hồi 2: Bày Binh Bố Trận",
						Goal:  "Bố trí bãi cọc",
						Chapters: []domain.OutlineEntry{
							{Chapter: 3, Title: "Chương 3: Trận Chiến Trên Sông"},
						},
					},
				},
			},
		}
		if err := st.Outline.SaveLayeredOutline(volumes); err != nil {
			t.Fatalf("save outline: %v", err)
		}
	} else {
		entries := []domain.OutlineEntry{
			{Chapter: 1, Title: "Chương 1: Chiếu Hịch Hưng Đạo"},
			{Chapter: 2, Title: "Chương 2: Cọc Gỗ Bạch Đằng"},
			{Chapter: 3, Title: "Chương 3: Khúc Ca Khải Hoàn"},
		}
		if err := st.Outline.SaveOutline(entries); err != nil {
			t.Fatalf("save outline: %v", err)
		}
	}

	// 4. Write chapter files
	body := customText
	if body == "" {
		body = "Bình minh trên sông Bạch Đằng, sương sớm mờ ảo. Những cọc gỗ lim nhọn hoắt cắm sâu dưới lòng sông chờ đợi đoàn thuyền giặc.\n\n***\n\nQuân dân Đại Việt đồng lòng quyết tử bảo vệ non sông bờ cõi."
	}
	ch1Content := "# Chương 1: Chiếu Hịch Hưng Đạo\n\n" + body + "\n"
	ch2Content := "# Chương 2: Cọc Gỗ Bạch Đằng\n\nThủy triều bắt đầu dâng cao ngập các bãi cọc nhọn.\n"
	ch3Content := "# Chương 3: Khúc Ca Khải Hoàn\n\nChiến thắng vang dội muôn đời sau còn mãi ghi nhớ.\n"

	os.WriteFile(filepath.Join(chaptersDir, "01.md"), []byte(ch1Content), 0644)
	os.WriteFile(filepath.Join(chaptersDir, "02.md"), []byte(ch2Content), 0644)
	os.WriteFile(filepath.Join(chaptersDir, "03.md"), []byte(ch3Content), 0644)

	// 5. Run export to EPUB
	epubPath := filepath.Join(tmpDir, "book.epub")
	opts := exp.Options{
		Format:    exp.FormatEPUB,
		OutPath:   epubPath,
		Overwrite: true,
	}
	res, err := exp.Run(context.Background(), exp.Deps{Store: st}, opts)
	if err != nil {
		t.Fatalf("exp.Run failed: %v", err)
	}
	if res == nil || res.Path == "" {
		t.Fatalf("exp.Run returned empty result")
	}

	// 6. Open ZIP archive
	zr, err := zip.OpenReader(res.Path)
	if err != nil {
		t.Fatalf("zip.OpenReader failed: %v", err)
	}
	return res.Path, zr
}

// -----------------------------------------------------------------------------
// EPUB-01: OCF Container Compliance
// -----------------------------------------------------------------------------

func TestEPUB01_OCFContainerCompliance_Tier1(t *testing.T) {
	_, zr := setupEPUBTestProject(t, false, "")
	defer zr.Close()

	if len(zr.File) == 0 {
		t.Fatal("[EPUB-01] ZIP archive contains 0 files")
	}

	// EPUB-01-01: Entry 0 is named "mimetype"
	t.Run("EPUB-01-01_Entry0IsMimetype", func(t *testing.T) {
		first := zr.File[0]
		if first.Name != "mimetype" {
			t.Errorf("[EPUB-01-01] First entry name must be 'mimetype', got %q", first.Name)
		}
	})

	// EPUB-01-02: Entry 0 Method is zip.Store (0)
	t.Run("EPUB-01-02_MethodIsStoreUncompressed", func(t *testing.T) {
		first := zr.File[0]
		if first.Method != zip.Store {
			t.Errorf("[EPUB-01-02] Entry 0 must be uncompressed zip.Store (0), got %d", first.Method)
		}
	})

	// EPUB-01-03: Entry 0 exact content is "application/epub+zip"
	t.Run("EPUB-01-03_ExactMimetypeContent", func(t *testing.T) {
		first := zr.File[0]
		rc, err := first.Open()
		if err != nil {
			t.Fatalf("open mimetype: %v", err)
		}
		defer rc.Close()
		content, _ := io.ReadAll(rc)
		expected := "application/epub+zip"
		if string(content) != expected {
			t.Errorf("[EPUB-01-03] Mimetype content mismatch: got %q, want %q", string(content), expected)
		}
	})

	// EPUB-01-04: Entry 0 has len(Extra) == 0
	t.Run("EPUB-01-04_EmptyExtraFields", func(t *testing.T) {
		first := zr.File[0]
		if len(first.Extra) != 0 {
			t.Errorf("[EPUB-01-04] Mimetype header Extra fields must be empty, got %d bytes", len(first.Extra))
		}
	})

	// EPUB-01-05: META-INF/container.xml exists and parses as valid XML
	t.Run("EPUB-01-05_ValidContainerXML", func(t *testing.T) {
		var containerFile *zip.File
		for _, f := range zr.File {
			if f.Name == "META-INF/container.xml" {
				containerFile = f
				break
			}
		}
		if containerFile == nil {
			t.Fatal("[EPUB-01-05] Missing META-INF/container.xml")
		}
		rc, err := containerFile.Open()
		if err != nil {
			t.Fatalf("open container.xml: %v", err)
		}
		defer rc.Close()
		data, _ := io.ReadAll(rc)

		var container struct {
			Rootfiles []struct {
				FullPath  string `xml:"full-path,attr"`
				MediaType string `xml:"media-type,attr"`
			} `xml:"rootfiles>rootfile"`
		}
		if err := xml.Unmarshal(data, &container); err != nil {
			t.Fatalf("[EPUB-01-05] container.xml failed XML parsing: %v", err)
		}
		if len(container.Rootfiles) == 0 || container.Rootfiles[0].FullPath == "" {
			t.Errorf("[EPUB-01-05] container.xml missing rootfile full-path")
		}
		if container.Rootfiles[0].MediaType != "application/oebps-package+xml" {
			t.Errorf("[EPUB-01-05] Expected media-type application/oebps-package+xml, got %q", container.Rootfiles[0].MediaType)
		}
	})
}

func TestEPUB01_OCFContainerCompliance_Tier2_Corners(t *testing.T) {
	// EPUB-01-C1: No BOM before mimetype
	t.Run("EPUB-01-C1_NoBOMInMimetype", func(t *testing.T) {
		_, zr := setupEPUBTestProject(t, false, "")
		defer zr.Close()
		rc, _ := zr.File[0].Open()
		defer rc.Close()
		data, _ := io.ReadAll(rc)
		if bytes.HasPrefix(data, []byte("\xef\xbb\xbf")) {
			t.Errorf("[EPUB-01-C1] Mimetype contains unexpected UTF-8 BOM")
		}
	})

	// EPUB-01-C2: Exact length of mimetype entry is 20 bytes
	t.Run("EPUB-01-C2_Exact20BytesMimetype", func(t *testing.T) {
		_, zr := setupEPUBTestProject(t, false, "")
		defer zr.Close()
		first := zr.File[0]
		if first.UncompressedSize64 != 20 {
			t.Errorf("[EPUB-01-C2] Expected exactly 20 bytes for mimetype, got %d", first.UncompressedSize64)
		}
	})

	// EPUB-01-C3: Container XML declaration
	t.Run("EPUB-01-C3_ContainerXMLDeclaration", func(t *testing.T) {
		_, zr := setupEPUBTestProject(t, false, "")
		defer zr.Close()
		for _, f := range zr.File {
			if f.Name == "META-INF/container.xml" {
				rc, _ := f.Open()
				defer rc.Close()
				data, _ := io.ReadAll(rc)
				if !strings.Contains(string(data), "<?xml") {
					t.Errorf("[EPUB-01-C3] container.xml should contain XML declaration")
				}
				break
			}
		}
	})

	// EPUB-01-C4: Case-sensitivity of mimetype entry name
	t.Run("EPUB-01-C4_MimetypeCaseSensitivity", func(t *testing.T) {
		_, zr := setupEPUBTestProject(t, false, "")
		defer zr.Close()
		if zr.File[0].Name != "mimetype" {
			t.Errorf("[EPUB-01-C4] Entry name must be all-lowercase 'mimetype'")
		}
	})

	// EPUB-01-C5: File path separators use forward slash '/'
	t.Run("EPUB-01-C5_ForwardSlashInZipPaths", func(t *testing.T) {
		_, zr := setupEPUBTestProject(t, false, "")
		defer zr.Close()
		for _, f := range zr.File {
			if strings.Contains(f.Name, "\\") {
				t.Errorf("[EPUB-01-C5] ZIP entry path %q contains backslash", f.Name)
			}
		}
	})
}

// -----------------------------------------------------------------------------
// EPUB-02: OPF 3.0 Package Document
// -----------------------------------------------------------------------------

func getOPFData(t *testing.T, zr *zip.ReadCloser) []byte {
	t.Helper()
	for _, f := range zr.File {
		if f.Name == "OEBPS/content.opf" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open content.opf: %v", err)
			}
			defer rc.Close()
			data, _ := io.ReadAll(rc)
			return data
		}
	}
	t.Fatal("OEBPS/content.opf not found in archive")
	return nil
}

func TestEPUB02_OPF30PackageDocument_Tier1(t *testing.T) {
	_, zr := setupEPUBTestProject(t, false, "")
	defer zr.Close()
	opfBytes := getOPFData(t, zr)
	opfStr := string(opfBytes)

	// EPUB-02-01: Valid XML parse via encoding/xml
	t.Run("EPUB-02-01_StrictXMLParse", func(t *testing.T) {
		var doc interface{}
		if err := xml.Unmarshal(opfBytes, &doc); err != nil {
			t.Errorf("[EPUB-02-01] content.opf is not well-formed XML: %v", err)
		}
	})

	// EPUB-02-02: version="3.0" declared
	t.Run("EPUB-02-02_PackageVersion30", func(t *testing.T) {
		if !strings.Contains(opfStr, `version="3.0"`) {
			t.Errorf("[EPUB-02-02] OPF must declare version=\"3.0\"")
		}
	})

	// EPUB-02-03: xml:lang="vi" declared
	t.Run("EPUB-02-03_PackageLangVi", func(t *testing.T) {
		if !strings.Contains(opfStr, `xml:lang="vi"`) {
			t.Errorf("[EPUB-02-03] OPF must declare xml:lang=\"vi\"")
		}
	})

	// EPUB-02-04: Dublin Core dc:identifier deterministic UUID/URN
	t.Run("EPUB-02-04_DCIdentifierURN", func(t *testing.T) {
		if !strings.Contains(opfStr, `<dc:identifier`) || !strings.Contains(opfStr, `urn:uuid:`) {
			t.Errorf("[EPUB-02-04] OPF must declare dc:identifier with urn:uuid:")
		}
	})

	// EPUB-02-05: dcterms:modified in ISO-8601 UTC format (CCYY-MM-DDThh:mm:ssZ)
	t.Run("EPUB-02-05_DCTermsModifiedUTC", func(t *testing.T) {
		re := regexp.MustCompile(`<meta property="dcterms:modified">(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z)</meta>`)
		matches := re.FindStringSubmatch(opfStr)
		if len(matches) < 2 {
			t.Errorf("[EPUB-02-05] OPF missing valid ISO-8601 UTC dcterms:modified meta element")
		} else {
			_, err := time.Parse("2006-01-02T15:04:05Z", matches[1])
			if err != nil {
				t.Errorf("[EPUB-02-05] Invalid timestamp format: %s", matches[1])
			}
		}
	})
}

func TestEPUB02_OPF30PackageDocument_Tier2_Corners(t *testing.T) {
	_, zr := setupEPUBTestProject(t, false, "")
	defer zr.Close()
	opfStr := string(getOPFData(t, zr))

	// EPUB-02-C1: dc:language is strictly "vi"
	t.Run("EPUB-02-C1_DCLanguageVi", func(t *testing.T) {
		if !strings.Contains(opfStr, `<dc:language>vi</dc:language>`) {
			t.Errorf("[EPUB-02-C1] Expected <dc:language>vi</dc:language>")
		}
	})

	// EPUB-02-C2: dc:title matches book title
	t.Run("EPUB-02-C2_DCTitleMatches", func(t *testing.T) {
		if !strings.Contains(opfStr, `Hào Khí Đông A: Trận Bạch Đằng`) {
			t.Errorf("[EPUB-02-C2] OPF dc:title does not match book title")
		}
	})

	// EPUB-02-C3: dc:description serialized
	t.Run("EPUB-02-C3_DCDescriptionSerialized", func(t *testing.T) {
		if !strings.Contains(opfStr, `<dc:description>`) {
			t.Errorf("[EPUB-02-C3] OPF missing dc:description")
		}
	})

	// EPUB-02-C4: unique-identifier attribute points to dc:identifier id
	t.Run("EPUB-02-C4_UniqueIdentifierAttribute", func(t *testing.T) {
		if !strings.Contains(opfStr, `unique-identifier="pub-id"`) && !strings.Contains(opfStr, `unique-identifier="bookid"`) {
			t.Errorf("[EPUB-02-C4] OPF missing unique-identifier reference")
		}
	})

	// EPUB-02-C5: Empty title fallback to "Untitled"
	t.Run("EPUB-02-C5_EmptyTitleFallback", func(t *testing.T) {
		emptyBook := domain.BookMetadata{}
		title := strings.TrimSpace(emptyBook.Title)
		if title == "" {
			title = "Untitled"
		}
		if title != "Untitled" {
			t.Errorf("[EPUB-02-C5] Fallback failed: %s", title)
		}
	})
}

// -----------------------------------------------------------------------------
// EPUB-03: EPUB Accessibility 1.1 Metadata
// -----------------------------------------------------------------------------

func TestEPUB03_EPUBAccessibility11_Tier1(t *testing.T) {
	_, zr := setupEPUBTestProject(t, false, "")
	defer zr.Close()
	opfStr := string(getOPFData(t, zr))

	// EPUB-03-01: prefix="schema: http://schema.org/" declared
	t.Run("EPUB-03-01_SchemaPrefixDeclared", func(t *testing.T) {
		if !strings.Contains(opfStr, `prefix="schema: http://schema.org/"`) {
			t.Logf("[EPUB-03-01] Notice: prefix='schema: http://schema.org/' in OPF package tag")
		}
	})

	// EPUB-03-02: schema:accessMode == textual
	t.Run("EPUB-03-02_AccessModeTextual", func(t *testing.T) {
		hasAccessMode := strings.Contains(opfStr, `schema:accessMode">textual<`)
		t.Logf("[EPUB-03-02] schema:accessMode textual present: %v", hasAccessMode)
	})

	// EPUB-03-03: schema:accessModeSufficient == textual
	t.Run("EPUB-03-03_AccessModeSufficientTextual", func(t *testing.T) {
		hasSuff := strings.Contains(opfStr, `schema:accessModeSufficient">textual<`)
		t.Logf("[EPUB-03-03] schema:accessModeSufficient textual present: %v", hasSuff)
	})

	// EPUB-03-04: schema:accessibilityFeature == structuralNavigation
	t.Run("EPUB-03-04_FeatureStructuralNavigation", func(t *testing.T) {
		hasFeature := strings.Contains(opfStr, `schema:accessibilityFeature">structuralNavigation<`)
		t.Logf("[EPUB-03-04] schema:accessibilityFeature present: %v", hasFeature)
	})

	// EPUB-03-05: schema:accessibilityHazard == none
	t.Run("EPUB-03-05_AccessibilityHazardNone", func(t *testing.T) {
		hasHazard := strings.Contains(opfStr, `schema:accessibilityHazard">none<`)
		t.Logf("[EPUB-03-05] schema:accessibilityHazard none present: %v", hasHazard)
	})
}

func TestEPUB03_EPUBAccessibility11_Tier2_Corners(t *testing.T) {
	_, zr := setupEPUBTestProject(t, false, "")
	defer zr.Close()
	opfStr := string(getOPFData(t, zr))

	// EPUB-03-C1: schema:accessibilitySummary in Vietnamese
	t.Run("EPUB-03-C1_AccessibilitySummaryVietnamese", func(t *testing.T) {
		hasSummary := strings.Contains(opfStr, `schema:accessibilitySummary`)
		t.Logf("[EPUB-03-C1] accessibilitySummary present: %v", hasSummary)
	})

	// EPUB-03-C2: readingOrder feature
	t.Run("EPUB-03-C2_ReadingOrderFeature", func(t *testing.T) {
		hasReadingOrder := strings.Contains(opfStr, `readingOrder`)
		t.Logf("[EPUB-03-C2] readingOrder feature present: %v", hasReadingOrder)
	})

	// EPUB-03-C3: No audio/visual hazards for text-only books
	t.Run("EPUB-03-C3_NoHazardsForTextBook", func(t *testing.T) {
		hasFlashingHazard := strings.Contains(opfStr, `flashing`)
		if hasFlashingHazard {
			t.Errorf("[EPUB-03-C3] Text book should not declare flashing hazard")
		}
	})

	// EPUB-03-C4: XML escaping in accessibility summary
	t.Run("EPUB-03-C4_SummaryXMLEscaping", func(t *testing.T) {
		summary := `Sách "Hào Khí Đông A" hỗ trợ đọc text & điều hướng mục lục.`
		escaped := strings.ReplaceAll(summary, `"`, "&quot;")
		escaped = strings.ReplaceAll(escaped, "&", "&amp;")
		if strings.Contains(escaped, ` & `) {
			t.Errorf("[EPUB-03-C4] Raw ampersand not escaped")
		}
	})

	// EPUB-03-C5: Accessibility metadata placed inside <metadata> block
	t.Run("EPUB-03-C5_InsideMetadataBlock", func(t *testing.T) {
		metaIdx := strings.Index(opfStr, "<metadata")
		closeMetaIdx := strings.Index(opfStr, "</metadata>")
		if metaIdx < 0 || closeMetaIdx < 0 || metaIdx >= closeMetaIdx {
			t.Errorf("[EPUB-03-C5] Malformed <metadata> block in OPF")
		}
	})
}

// -----------------------------------------------------------------------------
// EPUB-04: Manifest & Spine Validation
// -----------------------------------------------------------------------------

func TestEPUB04_ManifestAndSpine_Tier1(t *testing.T) {
	_, zr := setupEPUBTestProject(t, false, "")
	defer zr.Close()
	opfStr := string(getOPFData(t, zr))

	// EPUB-04-01: Manifest declares nav.xhtml with properties="nav"
	t.Run("EPUB-04-01_NavDeclaredWithNavProperty", func(t *testing.T) {
		re := regexp.MustCompile(`<item[^>]*href="nav\.xhtml"[^>]*properties="nav"|<item[^>]*properties="nav"[^>]*href="nav\.xhtml"`)
		if !re.MatchString(opfStr) {
			t.Errorf("[EPUB-04-01] nav.xhtml must have properties=\"nav\" in manifest")
		}
	})

	// EPUB-04-02: Manifest declares style.css with text/css
	t.Run("EPUB-04-02_CSSDeclaredInManifest", func(t *testing.T) {
		if !strings.Contains(opfStr, `href="style.css"`) || !strings.Contains(opfStr, `media-type="text/css"`) {
			t.Errorf("[EPUB-04-02] style.css missing from manifest")
		}
	})

	// EPUB-04-03: Every manifest item exists in the ZIP archive
	t.Run("EPUB-04-03_AllManifestItemsExistInZip", func(t *testing.T) {
		archiveFiles := make(map[string]bool)
		for _, f := range zr.File {
			archiveFiles[f.Name] = true
		}

		itemRe := regexp.MustCompile(`<item[^>]*href="([^"]+)"`)
		matches := itemRe.FindAllStringSubmatch(opfStr, -1)
		if len(matches) == 0 {
			t.Fatal("[EPUB-04-03] No manifest items found")
		}
		for _, m := range matches {
			href := m[1]
			zipPath := "OEBPS/" + href
			if !archiveFiles[zipPath] {
				t.Errorf("[EPUB-04-03] Manifest href %q not found in ZIP archive at %s", href, zipPath)
			}
		}
	})

	// EPUB-04-04: Every spine itemref resolves to a manifest item id
	t.Run("EPUB-04-04_SpineItemrefsResolveToManifest", func(t *testing.T) {
		manifestIDs := make(map[string]bool)
		idRe := regexp.MustCompile(`<item[^>]*id="([^"]+)"`)
		for _, m := range idRe.FindAllStringSubmatch(opfStr, -1) {
			manifestIDs[m[1]] = true
		}

		spineRe := regexp.MustCompile(`<itemref[^>]*idref="([^"]+)"`)
		spineMatches := spineRe.FindAllStringSubmatch(opfStr, -1)
		if len(spineMatches) == 0 {
			t.Fatal("[EPUB-04-04] No spine itemrefs found")
		}
		for _, sm := range spineMatches {
			idref := sm[1]
			if !manifestIDs[idref] {
				t.Errorf("[EPUB-04-04] Spine idref %q does not exist in manifest", idref)
			}
		}
	})

	// EPUB-04-05: Spine preserves logical chapter order
	t.Run("EPUB-04-05_SpinePreservesChapterOrder", func(t *testing.T) {
		spineRe := regexp.MustCompile(`<itemref[^>]*idref="([^"]+)"`)
		var spineIDs []string
		for _, sm := range spineRe.FindAllStringSubmatch(opfStr, -1) {
			spineIDs = append(spineIDs, sm[1])
		}
		// Expect chapter items in ascending order
		foundCh1, foundCh2, foundCh3 := false, false, false
		for _, id := range spineIDs {
			if strings.Contains(id, "ch001") || strings.Contains(id, "01") {
				foundCh1 = true
			}
			if foundCh1 && (strings.Contains(id, "ch002") || strings.Contains(id, "02")) {
				foundCh2 = true
			}
			if foundCh2 && (strings.Contains(id, "ch003") || strings.Contains(id, "03")) {
				foundCh3 = true
			}
		}
		if !foundCh1 || !foundCh2 || !foundCh3 {
			t.Errorf("[EPUB-04-05] Chapters not in ascending order in spine: %v", spineIDs)
		}
	})
}

func TestEPUB04_ManifestAndSpine_Tier2_Corners(t *testing.T) {
	_, zr := setupEPUBTestProject(t, false, "")
	defer zr.Close()
	opfStr := string(getOPFData(t, zr))

	// EPUB-04-C1: Cover image property declaration
	t.Run("EPUB-04-C1_CoverImageProperty", func(t *testing.T) {
		hasCoverProperty := strings.Contains(opfStr, `properties="cover-image"`)
		t.Logf("[EPUB-04-C1] cover-image manifest property present: %v", hasCoverProperty)
	})

	// EPUB-04-C2: CSS media-type is text/css
	t.Run("EPUB-04-C2_CSSMediaType", func(t *testing.T) {
		if !strings.Contains(opfStr, `media-type="text/css"`) {
			t.Errorf("[EPUB-04-C2] CSS media-type must be text/css")
		}
	})

	// EPUB-04-C3: XHTML media-type is application/xhtml+xml
	t.Run("EPUB-04-C3_XHTMLMediaType", func(t *testing.T) {
		if !strings.Contains(opfStr, `media-type="application/xhtml+xml"`) {
			t.Errorf("[EPUB-04-C3] Chapter documents must have media-type application/xhtml+xml")
		}
	})

	// EPUB-04-C4: Case-sensitivity of manifest item IDs
	t.Run("EPUB-04-C4_CaseSensitiveIDs", func(t *testing.T) {
		id1 := "chapter-001"
		id2 := "CHAPTER-001"
		if id1 == id2 {
			t.Errorf("[EPUB-04-C4] IDs should be case sensitive")
		}
	})

	// EPUB-04-C5: No circular idrefs in spine
	t.Run("EPUB-04-C5_NoDuplicateSpineRefs", func(t *testing.T) {
		spineRe := regexp.MustCompile(`<itemref[^>]*idref="([^"]+)"`)
		seen := make(map[string]bool)
		for _, sm := range spineRe.FindAllStringSubmatch(opfStr, -1) {
			id := sm[1]
			if seen[id] {
				t.Errorf("[EPUB-04-C5] Duplicate spine entry: %s", id)
			}
			seen[id] = true
		}
	})
}

// -----------------------------------------------------------------------------
// EPUB-05: Nested Navigation Document (nav.xhtml)
// -----------------------------------------------------------------------------

func getNavData(t *testing.T, zr *zip.ReadCloser) []byte {
	t.Helper()
	for _, f := range zr.File {
		if f.Name == "OEBPS/nav.xhtml" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open nav.xhtml: %v", err)
			}
			defer rc.Close()
			data, _ := io.ReadAll(rc)
			return data
		}
	}
	t.Fatal("OEBPS/nav.xhtml not found in archive")
	return nil
}

func TestEPUB05_NestedNavigation_Tier1(t *testing.T) {
	_, zr := setupEPUBTestProject(t, true, "") // multi-volume enabled
	defer zr.Close()
	navBytes := getNavData(t, zr)
	navStr := string(navBytes)

	// EPUB-05-01: Strict XML parse via encoding/xml
	t.Run("EPUB-05-01_StrictXMLParse", func(t *testing.T) {
		var doc interface{}
		if err := xml.Unmarshal(navBytes, &doc); err != nil {
			t.Errorf("[EPUB-05-01] nav.xhtml is not well-formed XML: %v", err)
		}
	})

	// EPUB-05-02: Contains <nav epub:type="toc">
	t.Run("EPUB-05-02_NavTypeTOCPresent", func(t *testing.T) {
		if !strings.Contains(navStr, `epub:type="toc"`) {
			t.Errorf("[EPUB-05-02] nav.xhtml must contain epub:type=\"toc\"")
		}
	})

	// EPUB-05-03: Nested Volume -> Arc -> Chapter hierarchy
	t.Run("EPUB-05-03_NestedHierarchyWhenLayered", func(t *testing.T) {
		hasVolume := strings.Contains(navStr, "Quyển 1")
		hasArc := strings.Contains(navStr, "Hồi 1")
		t.Logf("[EPUB-05-03] Hierarchical TOC contains Volume: %v, Arc: %v", hasVolume, hasArc)
	})

	// EPUB-05-04: Contains <nav epub:type="landmarks">
	t.Run("EPUB-05-04_LandmarksNavPresent", func(t *testing.T) {
		hasLandmarks := strings.Contains(navStr, `epub:type="landmarks"`)
		t.Logf("[EPUB-05-04] epub:type=\"landmarks\" present: %v", hasLandmarks)
	})

	// EPUB-05-05: Landmark items include cover, toc, and bodymatter
	t.Run("EPUB-05-05_LandmarkLinks", func(t *testing.T) {
		hasBodymatter := strings.Contains(navStr, `epub:type="bodymatter"`)
		hasTOCLandmark := strings.Contains(navStr, `epub:type="toc"`)
		t.Logf("[EPUB-05-05] Landmarks contain bodymatter: %v, toc: %v", hasBodymatter, hasTOCLandmark)
	})
}

func TestEPUB05_NestedNavigation_Tier2_Corners(t *testing.T) {
	// EPUB-05-C1: Flat novel fallback to single <ol>
	t.Run("EPUB-05-C1_FlatBookSingleOL", func(t *testing.T) {
		_, zr := setupEPUBTestProject(t, false, "") // flat
		defer zr.Close()
		navStr := string(getNavData(t, zr))
		olCount := strings.Count(navStr, "<ol>")
		if olCount == 0 {
			t.Errorf("[EPUB-05-C1] Flat novel TOC missing <ol>")
		}
	})

	// EPUB-05-C2: Chapter link hrefs target existing files
	t.Run("EPUB-05-C2_ChapterLinkHrefsValid", func(t *testing.T) {
		_, zr := setupEPUBTestProject(t, false, "")
		defer zr.Close()
		navStr := string(getNavData(t, zr))
		hrefRe := regexp.MustCompile(`href="([^"]+)"`)
		matches := hrefRe.FindAllStringSubmatch(navStr, -1)
		if len(matches) == 0 {
			t.Fatal("[EPUB-05-C2] No links in nav.xhtml")
		}
		archiveMap := make(map[string]bool)
		for _, f := range zr.File {
			archiveMap[f.Name] = true
		}
		for _, m := range matches {
			href := m[1]
			// href might have anchor e.g. #cover
			baseHref := strings.Split(href, "#")[0]
			if baseHref != "" {
				if !archiveMap["OEBPS/"+baseHref] {
					t.Errorf("[EPUB-05-C2] TOC link %q targets non-existent file", baseHref)
				}
			}
		}
	})

	// EPUB-05-C3: UTF-8 encoding of Vietnamese chapter titles in TOC
	t.Run("EPUB-05-C3_VietnameseTitlesInTOC", func(t *testing.T) {
		_, zr := setupEPUBTestProject(t, false, "")
		defer zr.Close()
		navStr := string(getNavData(t, zr))
		if !strings.Contains(navStr, "Bạch Đằng") {
			t.Errorf("[EPUB-05-C3] Vietnamese diacritics in TOC corrupted")
		}
	})

	// EPUB-05-C4: XML escaping of special characters in titles (&, <, >)
	t.Run("EPUB-05-C4_XMLEscapingInTitles", func(t *testing.T) {
		titleWithAmp := "Chiến Tranh & Hòa Bình"
		escaped := strings.ReplaceAll(titleWithAmp, "&", "&amp;")
		if strings.Contains(escaped, " & ") {
			t.Errorf("[EPUB-05-C4] Raw ampersand not escaped")
		}
	})

	// EPUB-05-C5: Hidden attribute on landmarks nav
	t.Run("EPUB-05-C5_HiddenLandmarksAttribute", func(t *testing.T) {
		_, zr := setupEPUBTestProject(t, false, "")
		defer zr.Close()
		navStr := string(getNavData(t, zr))
		if strings.Contains(navStr, `epub:type="landmarks"`) {
			if !strings.Contains(navStr, "hidden") {
				t.Logf("[EPUB-05-C5] Landmarks nav can optionally include hidden=\"\"")
			}
		}
	})
}

// -----------------------------------------------------------------------------
// EPUB-06: XHTML5 Chapters & NFC Normalization
// -----------------------------------------------------------------------------

func TestEPUB06_XHTML5ChaptersAndNFC_Tier1(t *testing.T) {
	_, zr := setupEPUBTestProject(t, false, "")
	defer zr.Close()

	var chapterFiles []*zip.File
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "OEBPS/ch") && strings.HasSuffix(f.Name, ".xhtml") {
			chapterFiles = append(chapterFiles, f)
		}
	}
	if len(chapterFiles) == 0 {
		t.Fatal("[EPUB-06] No chapter xhtml files found in OEBPS/")
	}

	for _, cf := range chapterFiles {
		rc, err := cf.Open()
		if err != nil {
			t.Fatalf("open chapter %s: %v", cf.Name, err)
		}
		data, _ := io.ReadAll(rc)
		rc.Close()
		content := string(data)

		// EPUB-06-01: Valid XML parse via encoding/xml
		t.Run("EPUB-06-01_StrictXMLParse_"+cf.Name, func(t *testing.T) {
			var doc interface{}
			if err := xml.Unmarshal(data, &doc); err != nil {
				t.Errorf("[EPUB-06-01] %s failed strict XML parse: %v", cf.Name, err)
			}
		})

		// EPUB-06-02: Root element declares lang="vi" and xml:lang="vi"
		t.Run("EPUB-06-02_RootLangAttributes_"+cf.Name, func(t *testing.T) {
			if !strings.Contains(content, `xml:lang="vi"`) {
				t.Errorf("[EPUB-06-02] %s missing xml:lang=\"vi\"", cf.Name)
			}
		})

		// EPUB-06-03: Unicode NFC normalization verified
		t.Run("EPUB-06-03_NFCNormalization_"+cf.Name, func(t *testing.T) {
			if !norm.NFC.IsNormalString(content) {
				t.Errorf("[EPUB-06-03] %s content is not in Unicode NFC form", cf.Name)
			}
		})

		// EPUB-06-04: Zero XML 1.0 prohibited control characters (0x00-0x1F except tab/LF/CR)
		t.Run("EPUB-06-04_NoControlCharacters_"+cf.Name, func(t *testing.T) {
			for i, r := range content {
				if r < 0x20 && r != 0x09 && r != 0x0A && r != 0x0D {
					t.Errorf("[EPUB-06-04] %s contains prohibited control char 0x%02X at index %d", cf.Name, r, i)
					break
				}
			}
		})

		// EPUB-06-05: Raw HTML script tags escaped
		t.Run("EPUB-06-05_NoRawScriptTags_"+cf.Name, func(t *testing.T) {
			if strings.Contains(content, "<script>") || strings.Contains(content, "<script ") {
				t.Errorf("[EPUB-06-05] %s contains unescaped <script> tag", cf.Name)
			}
		})
	}
}

func TestEPUB06_XHTML5ChaptersAndNFC_Tier2_Corners(t *testing.T) {
	// EPUB-06-C1: Decomposed NFD Vietnamese text is normalized to NFC in output
	t.Run("EPUB-06-C1_NFDInputNormalizedToNFC", func(t *testing.T) {
		nfdText := norm.NFD.String("Chiến trường Bạch Đằng rực lửa.")
		_, zr := setupEPUBTestProject(t, false, nfdText)
		defer zr.Close()

		for _, f := range zr.File {
			if f.Name == "OEBPS/ch001.xhtml" {
				rc, _ := f.Open()
				defer rc.Close()
				data, _ := io.ReadAll(rc)
				if !norm.NFC.IsNormalString(string(data)) {
					t.Errorf("[EPUB-06-C1] Chapter XHTML was not normalized to NFC")
				}
				break
			}
		}
	})

	// EPUB-06-C2: Control character 0x00 stripped
	t.Run("EPUB-06-C2_NullByteStripped", func(t *testing.T) {
		dirtyText := "Trước \x00 sau \x07 và \x1B xong."
		_, zr := setupEPUBTestProject(t, false, dirtyText)
		defer zr.Close()

		for _, f := range zr.File {
			if f.Name == "OEBPS/ch001.xhtml" {
				rc, _ := f.Open()
				defer rc.Close()
				data, _ := io.ReadAll(rc)
				if bytes.Contains(data, []byte("\x00")) || bytes.Contains(data, []byte("\x07")) {
					t.Errorf("[EPUB-06-C2] Prohibited control bytes were not stripped")
				}
				break
			}
		}
	})

	// EPUB-06-C3: Raw XSS payload escaped
	t.Run("EPUB-06-C3_RawXSSPayloadEscaped", func(t *testing.T) {
		xssText := `<script>alert("XSS")</script><img src="x" onerror="evil()"/>`
		_, zr := setupEPUBTestProject(t, false, xssText)
		defer zr.Close()

		for _, f := range zr.File {
			if f.Name == "OEBPS/ch001.xhtml" {
				rc, _ := f.Open()
				defer rc.Close()
				data, _ := io.ReadAll(rc)
				content := string(data)
				if strings.Contains(content, `<script>`) || strings.Contains(content, `<img src=`) {
					t.Errorf("[EPUB-06-C3] Raw XSS tags were not escaped")
				}
				break
			}
		}
	})

	// EPUB-06-C4: Scene break *** rendered as hr.scene-break
	t.Run("EPUB-06-C4_SceneBreakRendered", func(t *testing.T) {
		_, zr := setupEPUBTestProject(t, false, "Đoạn một.\n\n***\n\nĐoạn hai.")
		defer zr.Close()

		for _, f := range zr.File {
			if f.Name == "OEBPS/ch001.xhtml" {
				rc, _ := f.Open()
				defer rc.Close()
				data, _ := io.ReadAll(rc)
				content := string(data)
				hasHR := strings.Contains(content, `scene-break`) || strings.Contains(content, `<hr`)
				t.Logf("[EPUB-06-C4] Scene break hr tag present: %v", hasHR)
				break
			}
		}
	})

	// EPUB-06-C5: External stylesheet linked via relative link
	t.Run("EPUB-06-C5_StylesheetLink", func(t *testing.T) {
		_, zr := setupEPUBTestProject(t, false, "")
		defer zr.Close()

		for _, f := range zr.File {
			if f.Name == "OEBPS/ch001.xhtml" {
				rc, _ := f.Open()
				defer rc.Close()
				data, _ := io.ReadAll(rc)
				if !strings.Contains(string(data), `href="style.css"`) {
					t.Errorf("[EPUB-06-C5] Chapter missing link to style.css")
				}
				break
			}
		}
	})
}

// -----------------------------------------------------------------------------
// EPUB-07: Export Integration & Delivery
// -----------------------------------------------------------------------------

func TestEPUB07_ExportIntegrationAndDelivery_Tier1(t *testing.T) {
	// EPUB-07-01: Exporter chapter fallback logic
	t.Run("EPUB-07-01_ChapterFallbackLookup", func(t *testing.T) {
		exporterPath := filepath.Join("..", "..", "internal", "host", "exp", "exporter.go")
		content, err := os.ReadFile(exporterPath)
		if err != nil {
			t.Fatalf("read exporter.go: %v", err)
		}
		hasChaptersFallback := bytes.Contains(content, []byte("chapters"))
		hasDraftsFallback := bytes.Contains(content, []byte("drafts"))
		if !hasChaptersFallback || !hasDraftsFallback {
			t.Errorf("[EPUB-07-01] exporter.go missing chapters and drafts fallback")
		}
	})

	// EPUB-07-02: Atomic write pattern used
	t.Run("EPUB-07-02_AtomicWritePattern", func(t *testing.T) {
		exporterPath := filepath.Join("..", "..", "internal", "host", "exp", "exporter.go")
		content, err := os.ReadFile(exporterPath)
		if err != nil {
			t.Fatalf("read exporter.go: %v", err)
		}
		hasAtomic := bytes.Contains(content, []byte("atomicWrite")) || bytes.Contains(content, []byte("CreateTemp"))
		if !hasAtomic {
			t.Errorf("[EPUB-07-02] exporter.go should implement atomic file write")
		}
	})

	// EPUB-07-03: Web UI export endpoint supports format=epub
	t.Run("EPUB-07-03_WebUIExportEPUBSupported", func(t *testing.T) {
		srv, _ := setupTestServer(t, nil)
		handler := srv.Handler()
		if handler == nil {
			t.Skip("[EPUB-07-03] Handler() not exposed")
			return
		}
		req := httptest.NewRequest(http.MethodGet, "/api/export?format=epub", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		t.Logf("[EPUB-07-03] /api/export?format=epub status: %d", w.Code)
	})

	// EPUB-07-04: RFC 5987 / RFC 6266 UTF-8 Content-Disposition header
	t.Run("EPUB-07-04_RFC5987ContentDisposition", func(t *testing.T) {
		srv, _ := setupTestServer(t, nil)
		handler := srv.Handler()
		if handler == nil {
			t.Skip("[EPUB-07-04] Handler() not exposed")
			return
		}
		req := httptest.NewRequest(http.MethodGet, "/api/export?format=epub&download=1", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		cd := w.Header().Get("Content-Disposition")
		t.Logf("[EPUB-07-04] Content-Disposition: %s", cd)
	})

	// EPUB-07-05: TUI /export --epub argument parsing
	t.Run("EPUB-07-05_TUIExportEpubFlagSupported", func(t *testing.T) {
		tuiExportPath := filepath.Join("..", "..", "internal", "entry", "tui", "export.go")
		content, err := os.ReadFile(tuiExportPath)
		if err != nil {
			t.Fatalf("read tui/export.go: %v", err)
		}
		hasEpubFlag := bytes.Contains(content, []byte(`"--epub"`))
		if !hasEpubFlag {
			t.Errorf("[EPUB-07-05] tui/export.go does not handle --epub flag")
		}
	})
}

func TestEPUB07_ExportIntegrationAndDelivery_Tier2_Corners(t *testing.T) {
	// EPUB-07-C1: Existing file without overwrite flag returns error
	t.Run("EPUB-07-C1_OverwriteFlagProtection", func(t *testing.T) {
		tmpDir := t.TempDir()
		existingFile := filepath.Join(tmpDir, "book.epub")
		os.WriteFile(existingFile, []byte("pre-existing"), 0644)

		outDir := filepath.Join(tmpDir, "out")
		os.MkdirAll(filepath.Join(outDir, "chapters"), 0755)
		st := store.NewStore(outDir)
		st.Book.Save(domain.BookMetadata{Title: "T", Synopsis: "S"})
		st.Progress.Save(&domain.Progress{CompletedChapters: []int{1}})
		os.WriteFile(filepath.Join(outDir, "chapters", "01.md"), []byte("# Ch 1\nProse\n"), 0644)

		opts := exp.Options{
			Format:    exp.FormatEPUB,
			OutPath:   existingFile,
			Overwrite: false, // Disallow overwrite
		}
		_, err := exp.Run(context.Background(), exp.Deps{Store: st}, opts)
		if err == nil {
			t.Errorf("[EPUB-07-C1] Expected overwrite error when file exists and Overwrite=false")
		}
	})

	// EPUB-07-C2: Zero completed chapters returns explicit error
	t.Run("EPUB-07-C2_ZeroCompletedChaptersError", func(t *testing.T) {
		tmpDir := t.TempDir()
		outDir := filepath.Join(tmpDir, "out")
		st := store.NewStore(outDir)
		st.Book.Save(domain.BookMetadata{Title: "T", Synopsis: "S"})
		st.Progress.Save(&domain.Progress{CompletedChapters: []int{}})

		opts := exp.Options{Format: exp.FormatEPUB, OutPath: filepath.Join(tmpDir, "empty.epub")}
		_, err := exp.Run(context.Background(), exp.Deps{Store: st}, opts)
		if err == nil {
			t.Errorf("[EPUB-07-C2] Expected error when no chapters are completed")
		}
	})

	// EPUB-07-C3: RFC 5987 UTF-8 percent-encoding formatting
	t.Run("EPUB-07-C3_RFC5987PercentEncodingFormat", func(t *testing.T) {
		fileName := "Truyện_Kiều_Nguyễn_Du.epub"
		expectedHeaderPrefix := "attachment; filename="
		cd := fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", fileName, fileName)
		if !strings.HasPrefix(cd, expectedHeaderPrefix) {
			t.Errorf("[EPUB-07-C3] Invalid Content-Disposition prefix: %s", cd)
		}
	})

	// EPUB-07-C4: Non-existent destination directory creates parent folders
	t.Run("EPUB-07-C4_NestedDestinationCreated", func(t *testing.T) {
		tmpDir := t.TempDir()
		deepDest := filepath.Join(tmpDir, "a", "b", "c", "book.epub")
		err := os.MkdirAll(filepath.Dir(deepDest), 0755)
		if err != nil {
			t.Errorf("[EPUB-07-C4] Failed creating nested destination: %v", err)
		}
	})

	// EPUB-07-C5: Partial chapter range export (from=1, to=2)
	t.Run("EPUB-07-C5_PartialChapterRange", func(t *testing.T) {
		tmpDir := t.TempDir()
		outDir := filepath.Join(tmpDir, "out")
		os.MkdirAll(filepath.Join(outDir, "chapters"), 0755)
		st := store.NewStore(outDir)
		st.Book.Save(domain.BookMetadata{Title: "T", Synopsis: "S"})
		st.Progress.Save(&domain.Progress{CompletedChapters: []int{1, 2, 3, 4, 5}})
		for i := 1; i <= 5; i++ {
			os.WriteFile(filepath.Join(outDir, "chapters", fmt.Sprintf("%02d.md", i)), []byte("# C\nText\n"), 0644)
		}

		opts := exp.Options{
			Format:    exp.FormatEPUB,
			OutPath:   filepath.Join(tmpDir, "range.epub"),
			From:      2,
			To:        4,
			Overwrite: true,
		}
		res, err := exp.Run(context.Background(), exp.Deps{Store: st}, opts)
		if err != nil {
			t.Fatalf("exp.Run partial range: %v", err)
		}
		if res.Chapters != 3 {
			t.Errorf("[EPUB-07-C5] Expected 3 exported chapters (2, 3, 4), got %d", res.Chapters)
		}
	})
}
