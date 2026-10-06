package tools

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"

	"github.com/voocel/ainovel-cli/internal/domain"
)

// Tiêu đề chuẩn của premise (tiếng Việt, khớp đúng tên tiêu đề mà prompt Architect yêu cầu).
// Key chuẩn này được đưa nguyên vào novel_context (premise_sections / premise_structure)
// nên model nhìn thấy đúng tiếng Việt thay vì tiêu đề tiếng Trung như bản gốc.
const (
	headingGenreTone      = "Thể loại và sắc thái"
	headingPositioning    = "Định vị thể loại"
	headingCoreConflict   = "Xung đột cốt lõi"
	headingProtagonist    = "Mục tiêu của nhân vật chính"
	headingEnding         = "Hướng kết thúc"
	headingTaboo          = "Vùng cấm kỵ khi viết"
	headingSellingPoint   = "Điểm bán khác biệt"
	headingHook           = "Móc câu khác biệt"
	headingPromise        = "Cam kết cốt lõi"
	headingStoryEngine    = "Động cơ câu chuyện"
	headingRelationArc    = "Tuyến chính quan hệ/trưởng thành"
	headingProgression    = "Lộ trình nâng cấp"
	headingMidTurn        = "Chuyển hướng giữa kỳ"
	headingEndingQuestion = "Mệnh đề hồi kết"
	headingShortFit       = "Khả năng thích ứng truyện ngắn"
)

// premiseHeadingSpecs liệt kê, cho mỗi tiêu đề chuẩn, các tên được chấp nhận. Tiêu đề tiếng
// Trung của bản gốc được giữ làm alias để dữ liệu cũ vẫn phân tích được.
// Đây là nơi duy nhất cần sửa khi đổi bộ tiêu đề premise (phải đồng bộ với prompt Architect).
var premiseHeadingSpecs = []struct {
	canonical string
	aliases   []string
}{
	{headingGenreTone, []string{"题材和基调"}},
	{headingPositioning, []string{"题材定位"}},
	{headingCoreConflict, []string{"核心冲突"}},
	{headingProtagonist, []string{"主角目标"}},
	{headingEnding, []string{"终局方向", "结局方向", "Hướng kết cục"}},
	{headingTaboo, []string{"写作禁区", "Vùng cấm khi viết"}},
	{headingSellingPoint, []string{"差异化卖点"}},
	{headingHook, []string{"差异化钩子"}},
	{headingPromise, []string{"核心兑现承诺", "Cam kết thực hiện cốt lõi"}},
	{headingStoryEngine, []string{"故事引擎"}},
	{headingRelationArc, []string{"关系/成长主线", "Tuyến chính quan hệ và trưởng thành"}},
	{headingProgression, []string{"升级路径"}},
	{headingMidTurn, []string{"中段转折", "中期转向", "Chuyển hướng giữa truyện", "Bước ngoặt giữa kỳ"}},
	{headingEndingQuestion, []string{"终局命题"}},
	{headingShortFit, []string{
		"短篇适配性", "本作为什么适合短篇/单卷收束",
		"Lý do phù hợp với truyện ngắn / khép lại trong một quyển",
		"Khả năng thích ứng truyện ngắn",
	}},
}

// premiseHeadingAliases ánh xạ tên tiêu đề đã chuẩn hóa -> tiêu đề chuẩn.
var premiseHeadingAliases = buildPremiseHeadingAliases()

func buildPremiseHeadingAliases() map[string]string {
	m := make(map[string]string)
	for _, spec := range premiseHeadingSpecs {
		m[normalizeHeading(spec.canonical)] = spec.canonical
		for _, a := range spec.aliases {
			m[normalizeHeading(a)] = spec.canonical
		}
	}
	return m
}

// normalizeHeading đưa tiêu đề về dạng so khớp: NFC, hạ chữ thường, bỏ phần mô tả sau
// dấu hai chấm / trong ngoặc ở cuối và dấu câu thừa. Model hay viết
// "Định vị thể loại (độc giả mục tiêu…)" hoặc chép cả "Móc câu khác biệt: Điểm độc đáo…"
// thay vì đúng tên tiêu đề (chính prompt cũng liệt kê theo dạng "Tên: mô tả").
func normalizeHeading(s string) string {
	s = norm.NFC.String(strings.TrimSpace(s))
	if i := strings.IndexAny(s, "(（:："); i > 0 {
		s = s[:i]
	}
	s = strings.TrimRightFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	})
	return strings.ToLower(strings.TrimSpace(s))
}

func parsePremiseSections(premise string) map[string]string {
	lines := strings.Split(premise, "\n")
	sections := make(map[string]string)
	var current string
	var body []string

	flush := func() {
		if current == "" {
			return
		}
		text := strings.TrimSpace(strings.Join(body, "\n"))
		if text != "" {
			sections[current] = text
		}
		body = body[:0]
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if heading, ok := canonicalPremiseHeading(trimmed); ok {
			flush()
			current = heading
			continue
		}
		if current != "" {
			body = append(body, line)
		}
	}
	flush()
	return sections
}

func canonicalPremiseHeading(line string) (string, bool) {
	if !strings.HasPrefix(line, "#") {
		return "", false
	}
	title := strings.TrimSpace(strings.TrimLeft(line, "#"))
	if title == "" {
		return "", false
	}
	canonical, ok := premiseHeadingAliases[normalizeHeading(title)]
	return canonical, ok
}

func premiseStructure(premise string, tier domain.PlanningTier) map[string]any {
	sections := parsePremiseSections(premise)
	required := requiredPremiseHeadings(tier)
	found := make([]string, 0, len(required))
	var missing []string
	for _, heading := range required {
		if _, ok := sections[heading]; ok {
			found = append(found, heading)
			continue
		}
		missing = append(missing, heading)
	}

	structure := map[string]any{
		"template_ready": len(missing) == 0,
		"found":          found,
		"missing":        missing,
	}
	if len(sections) > 0 {
		structure["section_count"] = len(sections)
	}
	return structure
}

func requiredPremiseHeadings(tier domain.PlanningTier) []string {
	common := []string{
		headingGenreTone,
		headingPositioning,
		headingCoreConflict,
		headingProtagonist,
		headingEnding,
		headingTaboo,
		headingSellingPoint,
		headingHook,
		headingPromise,
	}

	switch tier {
	case domain.PlanningTierLong:
		return append(common,
			headingStoryEngine,
			headingRelationArc,
			headingProgression,
			headingMidTurn,
			headingEndingQuestion,
		)
	case domain.PlanningTierMid:
		return append(common,
			headingStoryEngine,
			headingMidTurn,
		)
	case domain.PlanningTierShort:
		return append(common,
			headingShortFit,
		)
	default:
		return common
	}
}
