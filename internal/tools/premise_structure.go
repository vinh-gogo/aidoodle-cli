package tools

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"

	"github.com/voocel/ainovel-cli/internal/domain"
)

// Tiêu đề chuẩn của premise = series bible của kênh doodle explainer (tiếng Việt).
// Danh sách này PHẢI đồng bộ với docs/script-format.md (mục 4) và prompt Architect
// (assets/prompts/architect-short.md, architect-long.md). Key chuẩn được đưa nguyên vào
// novel_context (premise_sections / premise_structure) nên model nhìn thấy đúng tiếng Việt.
const (
	headingChannel        = "Kênh và khán giả"
	headingVoiceCast      = "Giọng kể và nhân vật dẫn chuyện"
	headingCoreQuestion   = "Câu hỏi cốt lõi của series"
	headingFormula        = "Công thức video"
	headingHookFormula    = "Công thức hook"
	headingDoodleLaw      = "Luật vũ trụ doodle"
	headingSourcing       = "Chuẩn nguồn và kiểm chứng"
	headingTaboo          = "Vùng cấm kỵ khi viết"
	headingDifferentiator = "Điểm khác biệt của kênh"
	headingPromise        = "Cam kết với người xem"
	headingSeasonPlan     = "Kế hoạch mùa"
	headingBatches        = "Các đợt chủ đề"
	headingGagCallback    = "Running gag và callback"
	headingGrowth         = "Hướng phát triển series"
)

// premiseHeadingSpecs liệt kê, cho mỗi tiêu đề chuẩn, các tên khác được chấp nhận
// (model hay viết biến thể gần giống). Đây là nơi duy nhất cần sửa khi đổi bộ tiêu đề
// premise (phải đồng bộ với prompt Architect).
var premiseHeadingSpecs = []struct {
	canonical string
	aliases   []string
}{
	{headingChannel, []string{"Kênh và người xem", "Định vị kênh"}},
	{headingVoiceCast, []string{"Giọng kể", "Nhân vật dẫn chuyện", "Giọng điệu và nhân vật dẫn chuyện"}},
	{headingCoreQuestion, []string{"Câu hỏi cốt lõi"}},
	{headingFormula, []string{"Công thức tập", "Cấu trúc video"}},
	{headingHookFormula, []string{"Công thức mở đầu", "Hook"}},
	{headingDoodleLaw, []string{"Luật vũ trụ", "Luật thế giới doodle"}},
	{headingSourcing, []string{"Chuẩn nguồn", "Nguồn và kiểm chứng"}},
	{headingTaboo, []string{"Vùng cấm", "Vùng cấm kỵ"}},
	{headingDifferentiator, []string{"Điểm khác biệt", "Điểm khác biệt của series"}},
	{headingPromise, []string{"Cam kết với khán giả", "Cam kết"}},
	{headingSeasonPlan, []string{"Kế hoạch mùa video", "Kế hoạch series"}},
	{headingBatches, []string{"Đợt chủ đề", "Các đợt chủ đề video"}},
	{headingGagCallback, []string{"Running gag", "Gag và callback"}},
	{headingGrowth, []string{"Hướng phát triển", "Hướng đi của series"}},
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
// "Công thức hook (3 giây đầu)" hoặc chép cả "Cam kết với người xem: Người xem nhận được gì"
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
		headingChannel,
		headingVoiceCast,
		headingCoreQuestion,
		headingFormula,
		headingHookFormula,
		headingDoodleLaw,
		headingSourcing,
		headingTaboo,
		headingDifferentiator,
		headingPromise,
	}

	switch tier {
	case domain.PlanningTierLong:
		return append(common,
			headingBatches,
			headingGagCallback,
			headingGrowth,
		)
	case domain.PlanningTierMid:
		return append(common,
			headingBatches,
			headingGagCallback,
		)
	case domain.PlanningTierShort:
		return append(common,
			headingSeasonPlan,
		)
	default:
		return common
	}
}
