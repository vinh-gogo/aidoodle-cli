package tools

import (
	"slices"
	"testing"

	"golang.org/x/text/unicode/norm"

	"github.com/voocel/ainovel-cli/internal/domain"
)

func TestParsePremiseSections(t *testing.T) {
	premise := `# Tiền đề cốt truyện

## Thể loại và sắc thái
Giải thích hài hước, giọng dí dỏm.

## Định vị thể loại
Doodle explainer cho người trẻ 18-30 tuổi.

## Xung đột cốt lõi
Người tiền sử đối mặt với một khái niệm hiện đại khó hiểu.

## Chuyển hướng giữa kỳ
Chuyển từ giải thích sang tranh luận.
`

	sections := parsePremiseSections(premise)
	for _, heading := range []string{headingGenreTone, headingPositioning, headingCoreConflict, headingMidTurn} {
		if sections[heading] == "" {
			t.Fatalf("expected section %q, got %+v", heading, sections)
		}
	}
}

// Model hay viết kèm chú thích trong ngoặc, dấu hai chấm, hoặc đổi hoa/thường.
func TestParsePremiseSections_ToleratesDecoratedHeadings(t *testing.T) {
	premise := `## Định vị thể loại (độc giả mục tiêu, điểm bán cốt lõi)
Người trẻ.

## móc câu khác biệt:
Hook ba giây.

## HƯỚNG KẾT THÚC
Khép lại bằng một cú twist.
`
	sections := parsePremiseSections(premise)
	for _, heading := range []string{headingPositioning, headingHook, headingEnding} {
		if sections[heading] == "" {
			t.Fatalf("expected section %q, got %+v", heading, sections)
		}
	}
}

// Tiêu đề tiếng Việt ở dạng NFD (dấu tổ hợp) phải khớp như dạng dựng sẵn.
func TestParsePremiseSections_NFDHeading(t *testing.T) {
	heading := norm.NFD.String("Xung đột cốt lõi")
	if heading == "Xung đột cốt lõi" {
		t.Fatal("test cần chuỗi NFD khác chuỗi NFC")
	}
	premise := "## " + heading + "\nNội dung.\n"
	if parsePremiseSections(premise)[headingCoreConflict] == "" {
		t.Fatal("expected core conflict section")
	}
}

// Prompt liệt kê tiêu đề dạng "Tên: mô tả"; model có thể chép cả cụm vào dòng tiêu đề.
func TestParsePremiseSections_HeadingWithInlineDescription(t *testing.T) {
	premise := `## Móc câu khác biệt: Điểm độc đáo đáng theo dõi nhất của series này
Hook ba giây đầu.

## Cam kết cốt lõi: Người xem nhận được gì sau mỗi video
Hiểu một khái niệm trong 60 giây.
`
	sections := parsePremiseSections(premise)
	for _, heading := range []string{headingHook, headingPromise} {
		if sections[heading] == "" {
			t.Fatalf("expected section %q, got %+v", heading, sections)
		}
	}
}

func TestPremiseStructure(t *testing.T) {
	premise := `## Thể loại và sắc thái
Giải thích hài hước.

## Định vị thể loại
Người trẻ.

## Xung đột cốt lõi
Xung đột.

## Mục tiêu của nhân vật chính
Mục tiêu.

## Hướng kết thúc
Kết.

## Vùng cấm kỵ khi viết
Cấm.

## Điểm bán khác biệt
Điểm bán.

## Móc câu khác biệt
Móc câu.

## Cam kết cốt lõi
Cam kết.

## Động cơ câu chuyện
Động cơ.

## Chuyển hướng giữa kỳ
Chuyển hướng.
`

	structure := premiseStructure(premise, domain.PlanningTierMid)
	if ready, _ := structure["template_ready"].(bool); !ready {
		t.Fatalf("expected template_ready, got %+v", structure)
	}
	missing, _ := structure["missing"].([]string)
	if len(missing) != 0 {
		t.Fatalf("expected no missing headings, got %+v", missing)
	}
}

// Danh sách `missing` được gửi cho model, nên phải là tiếng Việt, không phải tiếng Trung.
func TestPremiseStructure_MissingHeadingsAreVietnamese(t *testing.T) {
	structure := premiseStructure("## Xung đột cốt lõi\nXung đột.\n", domain.PlanningTierShort)
	if ready, _ := structure["template_ready"].(bool); ready {
		t.Fatalf("premise thiếu tiêu đề không được ready: %+v", structure)
	}
	missing, _ := structure["missing"].([]string)
	if !slices.Contains(missing, headingShortFit) || slices.Contains(missing, headingCoreConflict) {
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

func TestPremiseStructureShortAcceptsLegacyHeadingAlias(t *testing.T) {
	premise := `## 题材和基调
单卷高压营救。

## 题材定位
短篇高密度冒险。

## 核心冲突
主角必须在一夜内救出人质。

## 主角目标
救出人质并活着离开。

## 结局方向
完成任务但付出代价。

## 写作禁区
不扩展成长期连载。

## 差异化卖点
时限压力与连续反转。

## 差异化钩子
每次选择都缩短救援时间。

## 核心兑现承诺
紧迫感、抉择与反转。

## 本作为什么适合短篇/单卷收束
核心矛盾和人物弧线都能在单次任务中完成。
`

	structure := premiseStructure(premise, domain.PlanningTierShort)
	if ready, _ := structure["template_ready"].(bool); !ready {
		t.Fatalf("expected short template_ready via legacy aliases, got %+v", structure)
	}
}

func TestPremiseStructureShortAcceptsVietnameseAlternateLabel(t *testing.T) {
	premise := `## Thể loại và sắc thái
a
## Định vị thể loại
a
## Xung đột cốt lõi
a
## Mục tiêu của nhân vật chính
a
## Hướng kết thúc
a
## Vùng cấm kỵ khi viết
a
## Điểm bán khác biệt
a
## Móc câu khác biệt
a
## Cam kết cốt lõi
a
## Lý do phù hợp với truyện ngắn / khép lại trong một quyển
a
`
	structure := premiseStructure(premise, domain.PlanningTierShort)
	if ready, _ := structure["template_ready"].(bool); !ready {
		t.Fatalf("expected ready, got %+v", structure)
	}
}
