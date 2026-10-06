package rules

import (
	"regexp"
	"strings"
)

// Lint 内置产品底线检查：扫描正文中的机制残留，与用户规则无关，commit 时始终执行。
// 与 Check 同契约——仅返事实（铁律一），不阻断流程，由评审/用户裁定。
//
// Hiện có hai loại (đều xuất phát từ lỗi thật khi chạy dài):
//   - markdown_residue：正文残留 ** 加粗、首行之外的 # 标题行（导出 txt 会裸露符号）
//   - han_residue：chữ Hán lọt vào lời đọc tiếng Việt (model Qwen/GLM hay trộn tiếng Trung).
//     Thay thế non_cjk_fragments cũ: regex Latin cũ khớp cả từ tiếng Việt thường như
//     "con", "tin", nên mọi kịch bản đều bị cảnh báo giả, còn chữ Hán thì không bị bắt.
func Lint(text string) []Violation {
	var vs []Violation
	vs = appendMarkdownResidue(vs, text)
	vs = appendHanResidue(vs, text)
	return vs
}

func appendMarkdownResidue(vs []Violation, text string) []Violation {
	if n := strings.Count(text, "**"); n > 0 {
		vs = append(vs, Violation{
			Rule:     "markdown_residue",
			Target:   "**",
			Actual:   n,
			Severity: SeverityWarning,
		})
	}
	headings := 0
	seenContent := false
	for line := range strings.SplitSeq(text, "\n") {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		// 第一个非空行的 # 标题是章文件的合法格式（不按行号写死，容忍前导空行）
		first := !seenContent
		seenContent = true
		if !first && strings.HasPrefix(t, "#") {
			headings++
		}
	}
	if headings > 0 {
		vs = append(vs, Violation{
			Rule:     "markdown_residue",
			Target:   "#",
			Actual:   headings,
			Severity: SeverityWarning,
		})
	}
	return vs
}

// hanRe khớp một chuỗi liên tiếp các chữ Hán.
var hanRe = regexp.MustCompile(`\p{Han}+`)

// appendHanResidue báo tổng số chữ Hán và vài đoạn ví dụ không trùng lặp.
// Lời đọc bắt buộc 100% tiếng Việt nên mọi chữ Hán đều là sự cố ngôn ngữ; mức error là
// một sự thật để Editor/người dùng phán quyết, không chặn commit.
func appendHanResidue(vs []Violation, text string) []Violation {
	runs := hanRe.FindAllString(text, -1)
	if len(runs) == 0 {
		return vs
	}
	total := 0
	seen := make(map[string]struct{})
	var examples []string
	for _, run := range runs {
		total += len([]rune(run))
		if _, ok := seen[run]; ok {
			continue
		}
		seen[run] = struct{}{}
		if len(examples) < 3 {
			examples = append(examples, truncateRunes(run, 12))
		}
	}
	return append(vs, Violation{
		Rule:     "han_residue",
		Target:   strings.Join(examples, ", "),
		Actual:   total,
		Severity: SeverityError,
	})
}

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
