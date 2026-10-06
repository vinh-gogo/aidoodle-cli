package exp

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"

	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/tools"
)

// slugify chuyển đổi chuỗi tiếng Việt thành slug URL an toàn cho tên tệp:
// loại bỏ dấu thanh, chuyển đ/Đ thành d, chuyển chữ hoa thành chữ thường, thay khoảng trắng/ký tự đặc biệt bằng '-'.
func slugify(s string) string {
	s = norm.NFD.String(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		switch r {
		case 'đ', 'Đ':
			b.WriteRune('d')
		default:
			if !unicode.Is(unicode.Mn, r) {
				b.WriteRune(r)
			}
		}
	}
	raw := strings.ToLower(b.String())
	var out strings.Builder
	lastDash := false
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			out.WriteRune(r)
			lastDash = false
		} else if !lastDash && out.Len() > 0 {
			out.WriteRune('-')
			lastDash = true
		}
	}
	res := strings.Trim(out.String(), "-")
	if res == "" {
		return "video"
	}
	return res
}

// renderVideoPackage xuất gói kịch bản video TikTok vào outDir:
//   - scripts/NN-slug.md: kịch bản hoàn chỉnh
//   - voiceover/NN-slug.txt: chỉ nội dung LỜI: để đưa vào TTS / thu âm
//   - shotlist.csv: danh sách phân cảnh (Tập, Khối, LỜI, HÌNH, CHỮ, ÂM)
//   - publish.csv: dữ liệu đăng tải (Tập, Tiêu đề, Thời lượng, Số từ, CAPTION, HASHTAG, NGUỒN, CẦN KIỂM CHỨNG)
func renderVideoPackage(outDir string, chapters []int, titleIdx map[int]string, bodies map[int]string, overwrite bool) (int, error) {
	scriptsDir := filepath.Join(outDir, "scripts")
	voiceoverDir := filepath.Join(outDir, "voiceover")
	shotlistPath := filepath.Join(outDir, "shotlist.csv")
	publishPath := filepath.Join(outDir, "publish.csv")

	if !overwrite {
		for _, p := range []string{scriptsDir, voiceoverDir, shotlistPath, publishPath} {
			if _, err := os.Stat(p); err == nil {
				return 0, fmt.Errorf("thư mục hoặc tệp xuất video đã tồn tại: %s (thêm --overwrite để ghi đè)", p)
			}
		}
	}

	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		return 0, fmt.Errorf("tạo thư mục scripts thất bại: %w", err)
	}
	if err := os.MkdirAll(voiceoverDir, 0o755); err != nil {
		return 0, fmt.Errorf("tạo thư mục voiceover thất bại: %w", err)
	}

	var totalBytes int

	// 1. Xuất scripts/NN-slug.md và voiceover/NN-slug.txt
	type parsedInfo struct {
		ch      int
		title   string
		slug    string
		doc     tools.ScriptDoc
		words   int
		seconds int
	}

	var parsedList []parsedInfo

	for _, ch := range chapters {
		body := bodies[ch]
		doc := tools.ParseScript(body)

		title := titleIdx[ch]
		if title == "" && doc.Title != "" {
			title = doc.Title
		}
		if title == "" {
			title = fmt.Sprintf("Tập %d", ch)
		}

		words := 0
		for _, v := range doc.Voice {
			words += domain.WordCount(v)
		}
		declared := 0
		for _, b := range doc.Blocks {
			declared = max(declared, b.EndSec)
		}
		seconds := max(declared, domain.SpeechSeconds(words, 0))

		slug := slugify(title)
		baseName := fmt.Sprintf("%02d-%s", ch, slug)

		scriptPath := filepath.Join(scriptsDir, baseName+".md")
		scriptBytes := []byte(body)
		if err := atomicWrite(scriptPath, scriptBytes); err != nil {
			return 0, fmt.Errorf("ghi tệp kịch bản %s thất bại: %w", scriptPath, err)
		}
		totalBytes += len(scriptBytes)

		voiceText := strings.Join(doc.Voice, "\n\n")
		voicePath := filepath.Join(voiceoverDir, baseName+".txt")
		voiceBytes := []byte(voiceText)
		if err := atomicWrite(voicePath, voiceBytes); err != nil {
			return 0, fmt.Errorf("ghi tệp lời đọc %s thất bại: %w", voicePath, err)
		}
		totalBytes += len(voiceBytes)

		parsedList = append(parsedList, parsedInfo{
			ch:      ch,
			title:   title,
			slug:    slug,
			doc:     doc,
			words:   words,
			seconds: seconds,
		})
	}

	// 2. Xuất shotlist.csv (kèm UTF-8 BOM để Excel hiển thị đúng tiếng Việt)
	var shotlistBuf bytes.Buffer
	shotlistBuf.WriteString("\xef\xbb\xbf") // UTF-8 BOM
	sw := csv.NewWriter(&shotlistBuf)
	if err := sw.Write([]string{"Tập", "Khối/Cảnh", "Mốc thời gian", "LỜI", "HÌNH", "CHỮ", "ÂM"}); err != nil {
		return 0, err
	}

	for _, info := range parsedList {
		for _, b := range info.doc.Blocks {
			voice := strings.Join(b.Tags["LỜI"], " ")
			visual := strings.Join(b.Tags["HÌNH"], " ")
			textOnScreen := strings.Join(b.Tags["CHỮ"], " ")
			audio := strings.Join(b.Tags["ÂM"], " ")
			timeStr := ""
			if b.EndSec > 0 {
				timeStr = fmt.Sprintf("kết thúc %02d:%02d", b.EndSec/60, b.EndSec%60)
			}
			if err := sw.Write([]string{
				fmt.Sprintf("Tập %02d", info.ch),
				b.Name,
				timeStr,
				voice,
				visual,
				textOnScreen,
				audio,
			}); err != nil {
				return 0, err
			}
		}
	}
	sw.Flush()
	if err := sw.Error(); err != nil {
		return 0, fmt.Errorf("xuất shotlist.csv thất bại: %w", err)
	}

	shotlistData := shotlistBuf.Bytes()
	if err := atomicWrite(shotlistPath, shotlistData); err != nil {
		return 0, fmt.Errorf("ghi shotlist.csv thất bại: %w", err)
	}
	totalBytes += len(shotlistData)

	// 3. Xuất publish.csv (kèm UTF-8 BOM)
	var pubBuf bytes.Buffer
	pubBuf.WriteString("\xef\xbb\xbf") // UTF-8 BOM
	pw := csv.NewWriter(&pubBuf)
	if err := pw.Write([]string{
		"Tập", "Tiêu đề", "Thời lượng (giây)", "Số từ LỜI",
		"CAPTION", "HASHTAG", "NGUỒN", "CẦN KIỂM CHỨNG",
	}); err != nil {
		return 0, err
	}

	for _, info := range parsedList {
		caption := strings.Join(info.doc.Footer["CAPTION"], " ")
		hashtag := strings.Join(info.doc.Footer["HASHTAG"], " ")
		source := strings.Join(info.doc.Footer["NGUỒN"], " | ")
		unverified := strings.Join(info.doc.Footer["CẦN KIỂM CHỨNG"], " | ")

		if err := pw.Write([]string{
			fmt.Sprintf("Tập %02d", info.ch),
			info.title,
			strconv.Itoa(info.seconds),
			strconv.Itoa(info.words),
			caption,
			hashtag,
			source,
			unverified,
		}); err != nil {
			return 0, err
		}
	}
	pw.Flush()
	if err := pw.Error(); err != nil {
		return 0, fmt.Errorf("xuất publish.csv thất bại: %w", err)
	}

	pubData := pubBuf.Bytes()
	if err := atomicWrite(publishPath, pubData); err != nil {
		return 0, fmt.Errorf("ghi publish.csv thất bại: %w", err)
	}
	totalBytes += len(pubData)

	return totalBytes, nil
}
