package rules

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/voocel/ainovel-cli/internal/domain"
)

// ValidationError mô tả lỗi thẩm định cơ học đầu ra.
type ValidationError struct {
	Rule     string   `json:"rule"`
	Target   string   `json:"target,omitempty"`
	Limit    any      `json:"limit,omitempty"`
	Actual   any      `json:"actual"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
}

func (ve ValidationError) ToViolation() Violation {
	return Violation{
		Rule:     ve.Rule,
		Target:   ve.Target,
		Limit:    ve.Limit,
		Actual:   ve.Actual,
		Severity: ve.Severity,
	}
}

// Regex CJK mở rộng: bắt toàn bộ chữ Hán (Han), Hiragana, Katakana, Hangul và Bopomofo
var cjkRegex = regexp.MustCompile(`[\p{Han}\p{Hiragana}\p{Katakana}\p{Hangul}\p{Bopomofo}]+`)

// Regex phát hiện rò rỉ thẻ kịch bản video vào tiểu thuyết văn xuôi
var scriptTagLeakRegex = regexp.MustCompile(`(?im)^\s*(?:HOOK(?:\s+\d+)?|CẢNH\s+\d+|CHỐT(?:\s+\d+)?|LỜI\s*:|HÌNH\s*:|CAPTION\s*:|HASHTAG\s*:|NGUỒN\s*:)`)

// Danh sách từ convert Hán-Việt thô áp dụng cho toàn bộ các thể loại
var generalConvertBanned = []string{
	"chẳng qua là",
	"nói không chừng",
	"liền là",
	"nhịn không được",
	"hít một hơi khí lạnh",
	"dở khóc dở cười",
	"sắc mặt khó coi",
	"khóe miệng co giật",
	"trong lòng lộp bộp",
	"tê cả da đầu",
	"tiểu tử này",
	"lão gia hỏa",
}

// Danh sách sáo ngữ và từ cấm đặc thù theo từng thể loại
var genreBannedPhrases = map[string][]string{
	"suspense": {
		"thám tử đại tài",
		"chân tướng chỉ có một",
		"vở kịch đã hạ màn",
		"hung thủ không ai khác chính là",
		"bầu không khí quỷ dị đến cực điểm",
	},
	"romance": {
		"tổng tài bá đạo",
		"nữ nhân, ngươi đang đùa với lửa",
		"ánh mắt tà mị",
		"bảo bối ngoan",
		"tiểu yêu tinh",
	},
	"fantasy": {
		"khí thế ngập trời",
		"kinh hãi vạn phần",
		"đồng tử co rút lại",
		"hừ lạnh một tiếng",
		"chết không có chỗ chôn",
	},
	"vietnamese-history": {
		"bó tay",
		"bá đạo",
		"ảo ma",
		"chém gió",
		"đỉnh nóc kịch trần",
		"ô kê",
		"vãi",
		"toang",
	},
	"doodle-explainer": {
		"hãy cùng khám phá",
		"chào mừng các bạn đã quay trở lại",
		"trong thế giới ngày nay",
		"đừng quên bấm like và đăng ký kênh",
	},
	"behavioral-psychology": {
		"hãy cùng khám phá",
		"chào mừng các bạn đã quay trở lại",
		"trong thế giới ngày nay",
		"đừng quên bấm like và đăng ký kênh",
	},
}

// IsVideoMode xác định một mode có phải là video explainer không.
func IsVideoMode(mode string) bool {
	m := strings.ToLower(strings.TrimSpace(mode))
	return m == "doodle-explainer" || m == "behavioral-psychology" || m == "psychology" || m == "video"
}

// ValidateChapterOutput thực hiện 6 bộ kiểm tra cơ học xác định cho bản thảo chương/kịch bản.
// Tuân thủ Thiết Quân Luật 1: Chỉ trả về SỰ THẬT khách quan có thể chứng minh được bằng thuật toán.
func ValidateChapterOutput(mode string, title string, content string) ([]ValidationError, error) {
	var errs []ValidationError

	// VAL-01: Chặn ký tự CJK (Chinese, Japanese, Korean)
	errs = append(errs, validateCJK(content)...)

	// VAL-02: Kiểm tra giới hạn số từ theo từng thể loại
	errs = append(errs, validateWordCountBounds(mode, content)...)

	// VAL-03: Phát hiện vòng lặp lặp câu / lặp đoạn
	errs = append(errs, validateRepetitionLoops(content)...)

	// VAL-04: Phát hiện chương cụt / kết thúc giữa câu
	errs = append(errs, validateTruncatedEnding(content)...)

	// VAL-05: Danh sách từ cấm convert & sáo ngữ theo thể loại
	errs = append(errs, validateBannedPhrases(mode, content)...)

	// VAL-06: Ngăn chặn rò rỉ thẻ kịch bản video vào văn xuôi tiểu thuyết
	errs = append(errs, validateScriptTagLeak(mode, content)...)

	return errs, nil
}

// ValidateMechanicalRules trả về danh sách Violation phù hợp với pipeline commit_chapter.
func ValidateMechanicalRules(mode string, title string, content string) []Violation {
	errs, _ := ValidateChapterOutput(mode, title, content)
	violations := make([]Violation, 0, len(errs))
	for _, e := range errs {
		violations = append(violations, e.ToViolation())
	}
	return violations
}

// VAL-01: Quét ký tự CJK
func validateCJK(text string) []ValidationError {
	matches := cjkRegex.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}

	totalRunes := 0
	seen := make(map[string]struct{})
	var examples []string
	for _, m := range matches {
		totalRunes += utf8.RuneCountInString(m)
		if _, ok := seen[m]; !ok {
			seen[m] = struct{}{}
			if len(examples) < 3 {
				examples = append(examples, m)
			}
		}
	}

	target := strings.Join(examples, ", ")
	return []ValidationError{
		{
			Rule:     "cjk_residue",
			Target:   target,
			Actual:   totalRunes,
			Severity: SeverityError,
			Message:  fmt.Sprintf("Phát hiện %d ký tự CJK (tiếng Trung/Nhật/Hàn) trong nội dung tiếng Việt: %s", totalRunes, target),
		},
	}
}

// VAL-02: Giới hạn số từ theo từng thể loại
func validateWordCountBounds(mode string, text string) []ValidationError {
	if IsVideoMode(mode) {
		// Video modes: kiểm tra số từ của các dòng LỜI: (700 - 1.500 từ)
		spokenWords := 0
		lines := strings.Split(text, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "LỜI:") || strings.HasPrefix(trimmed, "LỜI :") {
				after := strings.TrimPrefix(trimmed, "LỜI:")
				after = strings.TrimPrefix(after, "LỜI :")
				spokenWords += domain.WordCount(strings.TrimSpace(after))
			}
		}

		// Nếu không tìm thấy thẻ LỜI: nào (có thể do lỗi định dạng), đếm toàn văn
		if spokenWords == 0 {
			spokenWords = domain.WordCount(text)
		}

		var errs []ValidationError
		if spokenWords < 700 || spokenWords > 1500 {
			errs = append(errs, ValidationError{
				Rule:     "script_word_count_out_of_range",
				Target:   "spoken_words",
				Limit:    "700-1500",
				Actual:   spokenWords,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("Số từ lời đọc kịch bản video (%d từ) nằm ngoài khoảng chuẩn 700 - 1.500 từ", spokenWords),
			})
		}

		// Ước tính thời lượng nói (~2.5 từ/giây): yêu cầu 300 - 600 giây (5 - 10 phút)
		estSeconds := int(float64(spokenWords) / 2.5)
		if estSeconds < 300 || estSeconds > 600 {
			errs = append(errs, ValidationError{
				Rule:     "script_duration_out_of_range",
				Target:   "duration_seconds",
				Limit:    "300-600s",
				Actual:   estSeconds,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("Thời lượng ước tính (%ds / %.1f phút) nằm ngoài khoảng chuẩn 300 - 600s", estSeconds, float64(estSeconds)/60.0),
			})
		}
		return errs
	}

	// 5 Novel modes: 2.000 - 4.500 từ
	words := domain.WordCount(text)
	if words < 2000 {
		sev := SeverityWarning
		if words < 1200 {
			sev = SeverityError
		}
		return []ValidationError{
			{
				Rule:     "chapter_too_short",
				Target:   "word_count",
				Limit:    "2000-4500",
				Actual:   words,
				Severity: sev,
				Message:  fmt.Sprintf("Chương truyện quá ngắn: %d từ (tiêu chuẩn 2.000 - 4.500 từ)", words),
			},
		}
	}

	if words > 4500 {
		return []ValidationError{
			{
				Rule:     "chapter_too_long",
				Target:   "word_count",
				Limit:    "2000-4500",
				Actual:   words,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("Chương truyện quá dài: %d từ (tiêu chuẩn 2.000 - 4.500 từ)", words),
			},
		}
	}

	return nil
}

// VAL-03: Phát hiện lặp câu liên tiếp, lặp câu tần suất cao, và vòng lặp đoạn văn
func validateRepetitionLoops(text string) []ValidationError {
	var errs []ValidationError

	// 1. Tách và chuẩn hóa câu
	sentences := splitSentences(text)
	sentenceCounts := make(map[string]int)
	sentenceOriginal := make(map[string]string)

	for i, s := range sentences {
		clean := normalizeSentence(s)
		runeLen := utf8.RuneCountInString(clean)
		if runeLen < 15 {
			continue
		}

		sentenceCounts[clean]++
		if _, ok := sentenceOriginal[clean]; !ok {
			sentenceOriginal[clean] = strings.TrimSpace(s)
		}

		// Lặp câu liên tiếp (S_i == S_{i+1})
		if i > 0 {
			prevClean := normalizeSentence(sentences[i-1])
			if clean == prevClean {
				sample := truncateString(sentenceOriginal[clean], 50)
				errs = append(errs, ValidationError{
					Rule:     "repetition_consecutive_sentence",
					Target:   sample,
					Actual:   2,
					Severity: SeverityError,
					Message:  fmt.Sprintf("Phát hiện câu lặp lại liên tiếp 2 lần: %q", sample),
				})
			}
		}
	}

	// Lặp câu tần suất cao (>= 3 lần trong chương)
	for clean, count := range sentenceCounts {
		if count >= 3 {
			sample := truncateString(sentenceOriginal[clean], 50)
			errs = append(errs, ValidationError{
				Rule:     "repetition_frequent_sentence",
				Target:   sample,
				Actual:   count,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("Phát hiện câu lặp lại bất thường %d lần trong cùng chương: %q", count, sample),
			})
		}
	}

	// 2. Phát hiện lặp đoạn văn (Paragraph loops)
	lines := strings.Split(text, "\n")
	var paragraphs []string
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if utf8.RuneCountInString(t) >= 30 {
			paragraphs = append(paragraphs, t)
		}
	}

	for i := 0; i < len(paragraphs); i++ {
		pClean := normalizeSentence(paragraphs[i])

		// Đoạn văn lặp liên tiếp (P_i == P_{i+1})
		if i+1 < len(paragraphs) {
			pNextClean := normalizeSentence(paragraphs[i+1])
			if pClean == pNextClean {
				sample := truncateString(paragraphs[i], 60)
				errs = append(errs, ValidationError{
					Rule:     "repetition_paragraph_loop",
					Target:   sample,
					Actual:   "consecutive",
					Severity: SeverityError,
					Message:  fmt.Sprintf("Phát hiện đoạn văn lặp lại liên tiếp: %q", sample),
				})
			}
		}

		// Đoạn văn lặp chu kỳ A-B-A-B (P_i == P_{i+2} và P_{i+1} == P_{i+3})
		if i+3 < len(paragraphs) {
			p2Clean := normalizeSentence(paragraphs[i+2])
			p1Clean := normalizeSentence(paragraphs[i+1])
			p3Clean := normalizeSentence(paragraphs[i+3])
			if pClean == p2Clean && p1Clean == p3Clean && pClean != p1Clean {
				sample := truncateString(paragraphs[i], 40) + " / " + truncateString(paragraphs[i+1], 40)
				errs = append(errs, ValidationError{
					Rule:     "repetition_paragraph_loop",
					Target:   sample,
					Actual:   "abab_cycle",
					Severity: SeverityError,
					Message:  fmt.Sprintf("Phát hiện chuỗi đoạn văn lặp chu kỳ A-B-A-B: %q", sample),
				})
			}
		}
	}

	return errs
}

// VAL-04: Phát hiện chương cụt / kết thúc giữa câu
func validateTruncatedEnding(text string) []ValidationError {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return []ValidationError{
			{
				Rule:     "truncated_ending",
				Target:   "",
				Severity: SeverityError,
				Message:  "Nội dung chương trống hoàn toàn",
			},
		}
	}

	// Lấy dòng phi rỗng cuối cùng
	lines := strings.Split(trimmed, "\n")
	var lastLine string
	for i := len(lines) - 1; i >= 0; i-- {
		l := strings.TrimSpace(lines[i])
		if l != "" {
			lastLine = l
			break
		}
	}

	// Bỏ qua các dòng chú thích Markdown hoặc metadata ở chân kịch bản video
	if strings.HasPrefix(lastLine, "HASHTAG:") || strings.HasPrefix(lastLine, "CAPTION:") || strings.HasPrefix(lastLine, "NGUỒN:") {
		return nil
	}

	// Cắt bỏ các ký tự bọc ngoài cuối câu (khoảng trắng, dấu ngoặc, markdown bold/italic)
	cleanEnd := strings.TrimRightFunc(lastLine, func(r rune) bool {
		return unicode.IsSpace(r) || r == '*' || r == '_' || r == ')' || r == ']' || r == '}' || r == '`'
	})

	if cleanEnd == "" {
		return nil
	}

	// Kiểm tra ký tự cuối cùng
	lastRune, _ := utf8.DecodeLastRuneInString(cleanEnd)

	// Các dấu kết thúc hợp lệ
	validTerminals := map[rune]bool{
		'.':  true,
		'!':  true,
		'?':  true,
		'…':  true,
		'"':  true,
		'”':  true,
		'\'': true,
		'’':  true,
		'>':  true, // trích dẫn kết thúc
	}

	var errs []ValidationError

	if !validTerminals[lastRune] {
		// Kiểm tra nếu 3 ký tự cuối là "..."
		if !strings.HasSuffix(cleanEnd, "...") {
			sample := truncateString(lastLine, 40)
			errs = append(errs, ValidationError{
				Rule:     "truncated_ending",
				Target:   string(lastRune),
				Severity: SeverityError,
				Message:  fmt.Sprintf("Chương kết thúc giữa chừng không có dấu chấm câu hợp lệ (kết thúc bằng %q): %q", lastRune, sample),
			})
		}
	} else if lastRune == '"' || lastRune == '”' || lastRune == '\'' || lastRune == '’' {
		// Nếu kết thúc bằng ngoặc kép, ký tự ngay trước ngoặc kép phải là dấu kết câu hợp lệ
		withoutQuote := strings.TrimRight(cleanEnd, "\"'”’")
		if withoutQuote != "" {
			innerRune, _ := utf8.DecodeLastRuneInString(withoutQuote)
			innerTerminals := map[rune]bool{'.': true, '!': true, '?': true, '…': true}
			if !innerTerminals[innerRune] && !strings.HasSuffix(withoutQuote, "...") {
				sample := truncateString(lastLine, 40)
				errs = append(errs, ValidationError{
					Rule:     "truncated_ending",
					Target:   string(innerRune),
					Severity: SeverityError,
					Message:  fmt.Sprintf("Câu thoại cuối chương kết thúc bằng ngoặc kép nhưng thiếu dấu ngắt câu bên trong: %q", sample),
				})
			}
		}
	}

	// Kiểm tra số lượng dấu ngoặc kép lẻ (chưa đóng)
	asciiQuoteCount := strings.Count(text, "\"")
	if asciiQuoteCount%2 != 0 {
		errs = append(errs, ValidationError{
			Rule:     "unclosed_dialogue_quote",
			Target:   "\"",
			Actual:   asciiQuoteCount,
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("Chương có số lượng dấu ngoặc kép đối thoại lẻ (%d dấu), nghi ngờ chưa đóng ngoặc kép", asciiQuoteCount),
		})
	}

	return errs
}

// VAL-05: Danh sách từ cấm convert & sáo ngữ theo thể loại
func validateBannedPhrases(mode string, text string) []ValidationError {
	var errs []ValidationError
	textLower := strings.ToLower(text)

	checkList := func(phrases []string, genreLabel string) {
		for _, phrase := range phrases {
			pLower := strings.ToLower(phrase)
			if n := strings.Count(textLower, pLower); n > 0 {
				errs = append(errs, ValidationError{
					Rule:     "forbidden_phrases",
					Target:   phrase,
					Actual:   n,
					Severity: SeverityError,
					Message:  fmt.Sprintf("Phát hiện từ ngữ/sáo ngữ bị cấm trong %s: %q (xuất hiện %d lần)", genreLabel, phrase, n),
				})
			}
		}
	}

	// 1. Kiểm tra từ convert Hán-Việt chung
	checkList(generalConvertBanned, "thể loại chung")

	// 2. Kiểm tra từ cấm đặc thù theo thể loại
	modeNorm := strings.ToLower(strings.TrimSpace(mode))
	if phrases, ok := genreBannedPhrases[modeNorm]; ok {
		checkList(phrases, "thể loại "+modeNorm)
	}

	return errs
}

// VAL-06: Ngăn chặn rò rỉ thẻ kịch bản video vào văn xuôi tiểu thuyết
func validateScriptTagLeak(mode string, text string) []ValidationError {
	// Chỉ kiểm tra cho các thể loại tiểu thuyết văn xuôi (không áp dụng cho video modes)
	if IsVideoMode(mode) {
		return nil
	}

	matches := scriptTagLeakRegex.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]struct{})
	var examples []string
	for _, m := range matches {
		t := strings.TrimSpace(m)
		if _, ok := seen[t]; !ok {
			seen[t] = struct{}{}
			if len(examples) < 3 {
				examples = append(examples, t)
			}
		}
	}

	target := strings.Join(examples, ", ")
	return []ValidationError{
		{
			Rule:     "leaked_script_tag",
			Target:   target,
			Actual:   len(matches),
			Severity: SeverityError,
			Message:  fmt.Sprintf("Phát hiện %d thẻ kịch bản video rò rỉ vào chương tiểu thuyết: %s", len(matches), target),
		},
	}
}

// Các hàm bổ trợ
func splitSentences(text string) []string {
	var sentences []string
	var current strings.Builder
	for _, r := range text {
		current.WriteRune(r)
		if r == '.' || r == '!' || r == '?' || r == '…' || r == '\n' {
			s := strings.TrimSpace(current.String())
			if s != "" {
				sentences = append(sentences, s)
			}
			current.Reset()
		}
	}
	if current.Len() > 0 {
		s := strings.TrimSpace(current.String())
		if s != "" {
			sentences = append(sentences, s)
		}
	}
	return sentences
}

func normalizeSentence(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	space := false
	for _, r := range s {
		if unicode.IsSpace(r) || r == '\n' || r == '\r' || r == '\t' {
			if !space {
				b.WriteRune(' ')
				space = true
			}
		} else if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			continue
		} else {
			b.WriteRune(r)
			space = false
		}
	}
	return strings.TrimSpace(b.String())
}

func truncateString(s string, maxRunes int) string {
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	return string(r[:maxRunes]) + "…"
}
