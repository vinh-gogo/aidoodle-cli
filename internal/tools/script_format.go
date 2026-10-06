package tools

import (
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/unicode/norm"

	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/rules"
)

// Validator định dạng kịch bản doodle explainer (đặc tả: docs/script-format.md, mục 2–3).
//
// Tuân thủ Iron Law 1 và 4: chỉ trả SỰ THẬT mà máy đóng/kiểm chứng được (có/không thẻ, số từ
// lời đọc, mốc thời gian) dưới dạng violation mức warning, không chặn commit. Chất lượng hook,
// độ hài hước, độ đúng sự thật là phán đoán ngữ nghĩa của Editor, KHÔNG viết thành luật cơ học ở đây.
const (
	scriptMinWords   = 150
	scriptMaxWords   = 450
	scriptMinSeconds = 60
	scriptMaxSeconds = 180
)

var (
	// Dòng mở khối: "HOOK 0:00-0:03", "CẢNH 2 0:20-0:45", "CHỐT 0:50-1:05".
	// Mốc thời gian là tùy chọn khi phân tích (thiếu mốc không phải lỗi định dạng khối).
	scriptBlockRe = regexp.MustCompile(`(?i)^(HOOK|CẢNH(?:\s+\d+)?|CHỐT)(?:\s+(\d{1,2}):(\d{2})\s*[-–—]\s*(\d{1,2}):(\d{2}))?\s*:?$`)
	// Dòng thẻ: "LỜI: ...", "HÌNH: ...", "CHỮ: ...", "ÂM: ...", "CAPTION: ..." ...
	scriptTagRe     = regexp.MustCompile(`(?i)^(LỜI|HÌNH|CHỮ|ÂM|CAPTION|HASHTAG|NGUỒN|CẦN KIỂM CHỨNG)\s*:\s*(.*)$`)
	scriptHashtagRe = regexp.MustCompile(`#[\p{L}\p{N}_]+`)
)

const (
	tagVoice      = "LỜI"
	tagVisual     = "HÌNH"
	tagCaption    = "CAPTION"
	tagHashtag    = "HASHTAG"
	tagSource     = "NGUỒN"
	tagUnverified = "CẦN KIỂM CHỨNG"
)

type scriptBlock struct {
	name   string // "HOOK", "CẢNH 2", "CHỐT"
	kind   string // "HOOK" / "CẢNH" / "CHỐT"
	endSec int    // 0 nếu không khai báo mốc thời gian
	tags   map[string][]string
}

type scriptDoc struct {
	hasTitle bool
	blocks   []*scriptBlock
	footer   map[string][]string
	voice    []string // toàn bộ nội dung các thẻ LỜI (kể cả LỜI nằm ngoài khối)
}

// parseScript phân tích chính văn thành khối/thẻ. Dòng không phải tiêu đề/khối/thẻ được coi là phần
// nối tiếp của thẻ ngay trước đó cho đến dòng trống; ngoài ra bị bỏ qua.
func parseScript(text string) scriptDoc {
	doc := scriptDoc{footer: map[string][]string{}}
	var cur *scriptBlock
	var curTag string // thẻ đang nhận phần nối tiếp ("" nếu không có)
	seenContent := false

	appendTag := func(tag, value string) {
		value = strings.TrimSpace(value)
		switch tag {
		case tagVoice, tagVisual, "CHỮ", "ÂM":
			if tag == tagVoice {
				doc.voice = append(doc.voice, value)
			}
			if cur != nil {
				cur.tags[tag] = append(cur.tags[tag], value)
			}
		default:
			doc.footer[tag] = append(doc.footer[tag], value)
		}
	}
	extendTag := func(tag, value string) {
		value = strings.TrimSpace(value)
		switch tag {
		case tagVoice, tagVisual, "CHỮ", "ÂM":
			if tag == tagVoice && len(doc.voice) > 0 {
				doc.voice[len(doc.voice)-1] += " " + value
			}
			if cur != nil {
				if v := cur.tags[tag]; len(v) > 0 {
					v[len(v)-1] += " " + value
				}
			}
		default:
			if v := doc.footer[tag]; len(v) > 0 {
				v[len(v)-1] += " " + value
			}
		}
	}

	for _, raw := range strings.Split(norm.NFC.String(text), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			curTag = ""
			continue
		}
		first := !seenContent
		seenContent = true

		if first && strings.HasPrefix(line, "# ") {
			doc.hasTitle = strings.TrimSpace(line[2:]) != ""
			continue
		}
		if m := scriptBlockRe.FindStringSubmatch(line); m != nil {
			kind := strings.ToUpper(strings.Fields(m[1])[0])
			b := &scriptBlock{name: strings.ToUpper(strings.Join(strings.Fields(m[1]), " ")), kind: kind, tags: map[string][]string{}}
			if m[4] != "" {
				b.endSec = scriptClock(m[4], m[5])
			}
			doc.blocks = append(doc.blocks, b)
			cur = b
			curTag = ""
			continue
		}
		if m := scriptTagRe.FindStringSubmatch(line); m != nil {
			tag := strings.ToUpper(strings.Join(strings.Fields(m[1]), " "))
			appendTag(tag, m[2])
			curTag = tag
			continue
		}
		if curTag != "" {
			extendTag(curTag, line)
		}
	}
	return doc
}

func scriptClock(min, sec string) int {
	m, _ := strconv.Atoi(min)
	s, _ := strconv.Atoi(sec)
	return m*60 + s
}

func nonEmpty(values []string) bool {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return true
		}
	}
	return false
}

// lintScript kiểm tra định dạng kịch bản và trả các violation mức warning (không chặn commit).
func lintScript(text string) []rules.Violation {
	doc := parseScript(text)
	var vs []rules.Violation
	warn := func(rule, target string, limit, actual any) {
		vs = append(vs, rules.Violation{Rule: rule, Target: target, Limit: limit, Actual: actual, Severity: rules.SeverityWarning})
	}

	if !doc.hasTitle {
		warn("script_no_title", "dòng đầu phải là '# Tiêu đề video'", nil, 0)
	}

	hookIdx, closeIdx := -1, -1
	for i, b := range doc.blocks {
		switch {
		case b.kind == "HOOK" && hookIdx < 0:
			hookIdx = i
		case b.kind == "CHỐT":
			closeIdx = i
		}
	}
	switch {
	case hookIdx < 0:
		warn("script_no_hook", "thiếu khối HOOK 0:00-0:03", nil, 0)
	case hookIdx != 0:
		warn("script_no_hook", "khối HOOK phải là khối đầu tiên", nil, hookIdx+1)
	}
	if closeIdx < 0 {
		warn("script_no_closing", "thiếu khối CHỐT", nil, 0)
	}

	words := 0
	for _, v := range doc.voice {
		words += domain.WordCount(v)
	}
	if words < scriptMinWords || words > scriptMaxWords {
		warn("script_words_out_of_range", "số từ LỜI", strconv.Itoa(scriptMinWords)+"-"+strconv.Itoa(scriptMaxWords), words)
	}

	declared := 0
	for _, b := range doc.blocks {
		declared = max(declared, b.endSec)
	}
	seconds := max(declared, domain.SpeechSeconds(words, 0))
	if seconds < scriptMinSeconds || seconds > scriptMaxSeconds {
		warn("script_duration_out_of_range", "thời lượng ước tính (giây)", strconv.Itoa(scriptMinSeconds)+"-"+strconv.Itoa(scriptMaxSeconds), seconds)
	}

	for _, b := range doc.blocks {
		if !nonEmpty(b.tags[tagVisual]) {
			warn("script_block_missing_visual", b.name, nil, 0)
		}
		if !nonEmpty(b.tags[tagVoice]) {
			warn("script_block_missing_voice", b.name, nil, 0)
		}
	}

	if !nonEmpty(doc.footer[tagCaption]) {
		warn("script_missing_caption", tagCaption, nil, 0)
	}
	if !scriptHashtagRe.MatchString(strings.Join(doc.footer[tagHashtag], " ")) {
		warn("script_missing_hashtag", tagHashtag, nil, 0)
	}
	if !nonEmpty(doc.footer[tagSource]) {
		warn("script_missing_source", tagSource, nil, 0)
	}
	if !nonEmpty(doc.footer[tagUnverified]) {
		warn("script_missing_unverified", tagUnverified, nil, 0)
	}
	return vs
}
