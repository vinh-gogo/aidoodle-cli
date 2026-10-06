package tools

import (
	"slices"
	"testing"

	"golang.org/x/text/unicode/norm"

	"github.com/voocel/ainovel-cli/internal/domain"
)

func TestParsePremiseSections(t *testing.T) {
	premise := `# Series bible

## Kênh và khán giả
Người trẻ 18-30 tuổi thích tin nhanh.

## Giọng kể và nhân vật dẫn chuyện
Hài hước, dí dỏm; host là Ông Ú.

## Câu hỏi cốt lõi của series
Người đá đối mặt với một khái niệm hiện đại khó hiểu thì sao?

## Các đợt chủ đề
Đợt 1: tiền bạc.
`

	sections := parsePremiseSections(premise)
	for _, heading := range []string{headingChannel, headingVoiceCast, headingCoreQuestion, headingBatches} {
		if sections[heading] == "" {
			t.Fatalf("expected section %q, got %+v", heading, sections)
		}
	}
}

// Model hay viết kèm chú thích trong ngoặc, dấu hai chấm, hoặc đổi hoa/thường.
func TestParsePremiseSections_ToleratesDecoratedHeadings(t *testing.T) {
	premise := `## Kênh và khán giả (ai xem, xem khi nào)
Người trẻ.

## công thức hook:
Hook ba giây.

## CAM KẾT VỚI NGƯỜI XEM
Hiểu một khái niệm trong 60 giây.
`
	sections := parsePremiseSections(premise)
	for _, heading := range []string{headingChannel, headingHookFormula, headingPromise} {
		if sections[heading] == "" {
			t.Fatalf("expected section %q, got %+v", heading, sections)
		}
	}
}

// Tiêu đề tiếng Việt ở dạng NFD (dấu tổ hợp) phải khớp như dạng dựng sẵn.
func TestParsePremiseSections_NFDHeading(t *testing.T) {
	heading := norm.NFD.String(headingCoreQuestion)
	if heading == headingCoreQuestion {
		t.Fatal("test cần chuỗi NFD khác chuỗi NFC")
	}
	premise := "## " + heading + "\nNội dung.\n"
	if parsePremiseSections(premise)[headingCoreQuestion] == "" {
		t.Fatal("expected core question section")
	}
}

// Prompt liệt kê tiêu đề dạng "Tên: mô tả"; model có thể chép cả cụm vào dòng tiêu đề.
func TestParsePremiseSections_HeadingWithInlineDescription(t *testing.T) {
	premise := `## Công thức hook: Cách mở đầu 3 giây đầu của mọi video
Hook ba giây đầu.

## Cam kết với người xem: Người xem nhận được gì sau mỗi video
Hiểu một khái niệm trong 60 giây.
`
	sections := parsePremiseSections(premise)
	for _, heading := range []string{headingHookFormula, headingPromise} {
		if sections[heading] == "" {
			t.Fatalf("expected section %q, got %+v", heading, sections)
		}
	}
}

func commonPremiseFixture() string {
	return `# Series bible

## Kênh và khán giả
a
## Giọng kể và nhân vật dẫn chuyện
a
## Câu hỏi cốt lõi của series
a
## Công thức video
a
## Công thức hook
a
## Luật vũ trụ doodle
a
## Chuẩn nguồn và kiểm chứng
a
## Vùng cấm kỵ khi viết
a
## Điểm khác biệt của kênh
a
## Cam kết với người xem
a
`
}

func TestPremiseStructure(t *testing.T) {
	cases := []struct {
		tier  domain.PlanningTier
		extra string
	}{
		{domain.PlanningTierShort, "## Kế hoạch mùa\na\n"},
		{domain.PlanningTierMid, "## Các đợt chủ đề\na\n## Running gag và callback\na\n"},
		{domain.PlanningTierLong, "## Các đợt chủ đề\na\n## Running gag và callback\na\n## Hướng phát triển series\na\n"},
	}
	for _, tc := range cases {
		structure := premiseStructure(commonPremiseFixture()+tc.extra, tc.tier)
		if ready, _ := structure["template_ready"].(bool); !ready {
			t.Fatalf("tier %v: expected template_ready, got %+v", tc.tier, structure)
		}
		if missing, _ := structure["missing"].([]string); len(missing) != 0 {
			t.Fatalf("tier %v: expected no missing headings, got %+v", tc.tier, missing)
		}
	}
}

// Danh sách `missing` được gửi cho model, nên phải là tiếng Việt, không phải tiếng Trung.
func TestPremiseStructure_MissingHeadingsAreVietnamese(t *testing.T) {
	structure := premiseStructure("## Công thức video\nCông thức.\n", domain.PlanningTierShort)
	if ready, _ := structure["template_ready"].(bool); ready {
		t.Fatalf("premise thiếu tiêu đề không được ready: %+v", structure)
	}
	missing, _ := structure["missing"].([]string)
	if !slices.Contains(missing, headingSeasonPlan) || slices.Contains(missing, headingFormula) {
		t.Fatalf("missing = %+v", missing)
	}
	for _, h := range missing {
		for _, r := range h {
			if r >= 0x4e00 && r <= 0x9fff {
				t.Fatalf("tiêu đề thiếu còn chữ Hán: %q", h)
			}
		}
	}
}

func TestPremiseStructure_AcceptsShortAlias(t *testing.T) {
	premise := commonPremiseFixture() + "## Kế hoạch mùa video\na\n"
	structure := premiseStructure(premise, domain.PlanningTierShort)
	if ready, _ := structure["template_ready"].(bool); !ready {
		t.Fatalf("expected ready via alias, got %+v", structure)
	}
}
