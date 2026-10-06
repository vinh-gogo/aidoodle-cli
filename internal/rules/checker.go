package rules

import (
	"strings"

	"golang.org/x/text/unicode/norm"
)

// Check 对章节正文按结构化规则进行机械检查，返回违规事实列表。
//
// 设计契约：
//   - 仅返事实，不下指令（铁律一）
//   - 不阻断任何调用方流程
//   - severity 按规则类型固定映射（参见 types.go 注释表）
//
// 参数：
//   - text：章节正文（终稿或草稿都可）
//   - s：合并后的结构化规则；IsEmpty 时直接返回 nil。
//
// Khớp chuỗi không phân biệt hoa/thường và theo chuẩn hóa Unicode NFC: tiếng Việt có
// chữ hoa đầu câu ("Có thể nói rằng") và chữ có dấu có thể ở dạng tổ hợp hoặc dựng sẵn,
// nếu so khớp byte thô thì blacklist sẽ bỏ sót.
func Check(text string, s Structured) []Violation {
	if s.IsEmpty() {
		return nil
	}
	folded := normalizeForMatch(text)

	var violations []Violation
	violations = appendForbiddenChars(violations, folded, s.ForbiddenChars)
	violations = appendForbiddenPhrases(violations, folded, s.ForbiddenPhrases)
	violations = appendFatigueWords(violations, folded, s.FatigueWords)
	return violations
}

// normalizeForMatch đưa văn bản/mục tiêu về dạng so khớp: NFC rồi hạ chữ thường.
func normalizeForMatch(s string) string {
	return strings.ToLower(norm.NFC.String(s))
}

// forbidden_chars：出现 ≥1 次即 error。
// 同一条规则只产生一条 violation，actual 是出现次数。
// text đã được chuẩn hóa bởi Check; Target giữ nguyên giá trị cấu hình ban đầu.
func appendForbiddenChars(vs []Violation, text string, list []string) []Violation {
	for _, ch := range list {
		if ch == "" {
			continue
		}
		n := strings.Count(text, normalizeForMatch(ch))
		if n == 0 {
			continue
		}
		vs = append(vs, Violation{
			Rule:     "forbidden_chars",
			Target:   ch,
			Actual:   n,
			Severity: SeverityError,
		})
	}
	return vs
}

// forbidden_phrases：出现 ≥1 次即 error；行为与 forbidden_chars 一致，仅 rule 名区分。
func appendForbiddenPhrases(vs []Violation, text string, list []string) []Violation {
	for _, ph := range list {
		if ph == "" {
			continue
		}
		n := strings.Count(text, normalizeForMatch(ph))
		if n == 0 {
			continue
		}
		vs = append(vs, Violation{
			Rule:     "forbidden_phrases",
			Target:   ph,
			Actual:   n,
			Severity: SeverityError,
		})
	}
	return vs
}

// fatigue_words：本章出现次数超过阈值才违规，warning 级。
// 不跨章累计——跨章问题后续交诊断。
func appendFatigueWords(vs []Violation, text string, m map[string]int) []Violation {
	for word, limit := range m {
		if word == "" || limit <= 0 {
			continue
		}
		n := strings.Count(text, normalizeForMatch(word))
		if n <= limit {
			continue
		}
		vs = append(vs, Violation{
			Rule:     "fatigue_words",
			Target:   word,
			Limit:    limit,
			Actual:   n,
			Severity: SeverityWarning,
		})
	}
	return vs
}
