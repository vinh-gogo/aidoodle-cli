package rules

import (
	"strings"
	"testing"
)

func TestLint_CleanVietnameseText(t *testing.T) {
	text := "# Vì sao hết tiền cuối tháng?\n" +
		"Ông Gậy ngồi trước hang, đếm sáu viên đá cuội rồi thở dài.\n" +
		"Con mồi chạy mất, tin xấu đến nhanh hơn cả gió."
	if vs := Lint(text); len(vs) != 0 {
		t.Errorf("tiếng Việt sạch không được bị cảnh báo: %+v", vs)
	}
}

// Hồi quy: regex Latin cũ [A-Za-z]{2,} khớp các từ tiếng Việt thường ("con", "tin"...),
// khiến mọi kịch bản đều bị cảnh báo giả.
func TestLint_VietnameseWordsWithPlainASCIIAreNotFlagged(t *testing.T) {
	text := "# Tiêu đề\ncon tin ban an on nhu the nao"
	if vs := Lint(text); len(vs) != 0 {
		t.Errorf("từ ASCII thuần trong tiếng Việt không phải sự cố: %+v", vs)
	}
}

func TestLint_MarkdownResidue(t *testing.T) {
	text := "# Tiêu đề\nĐây là **trọng điểm** của bài.\n## Tiểu mục\nNội dung."
	vs := Lint(text)
	bold := findViolation(vs, "markdown_residue", "**")
	if bold == nil || bold.Actual != 2 {
		t.Errorf("expected ** residue x2: %+v", vs)
	}
	heading := findViolation(vs, "markdown_residue", "#")
	if heading == nil || heading.Actual != 1 {
		t.Errorf("expected 1 heading beyond first line: %+v", vs)
	}
}

func TestLint_HanResidue(t *testing.T) {
	text := "# Tiêu đề\nÔng Gậy nhìn thấy 一个 pattern, rồi nói 他发现了 điều lạ, lại 一个 nữa."
	vs := Lint(text)
	var v *Violation
	for i := range vs {
		if vs[i].Rule == "han_residue" {
			v = &vs[i]
			break
		}
	}
	if v == nil {
		t.Fatalf("expected han_residue violation: %+v", vs)
	}
	// 一个(2) + 他发现了(4) + 一个(2) = 8 chữ Hán
	if v.Actual != 8 {
		t.Errorf("tổng chữ Hán: got %v want 8", v.Actual)
	}
	target := v.Target
	if strings.Count(target, "一个") != 1 || !strings.Contains(target, "他发现了") {
		t.Errorf("ví dụ phải không trùng lặp: %q", target)
	}
	if v.Severity != SeverityError {
		t.Errorf("severity: %v", v.Severity)
	}
	// Từ Latin như "pattern" không còn là sự cố.
	if findViolation(vs, "non_cjk_fragments", "pattern") != nil {
		t.Errorf("non_cjk_fragments đã bị loại bỏ: %+v", vs)
	}
}

func TestLint_HanResidueExamplesCapped(t *testing.T) {
	text := "# T\na 一 b 二 c 三 d 四 e 五"
	vs := Lint(text)
	for _, v := range vs {
		if v.Rule != "han_residue" {
			continue
		}
		if got := len(strings.Split(v.Target, ", ")); got != 3 {
			t.Errorf("tối đa 3 ví dụ, có %d: %q", got, v.Target)
		}
		if v.Actual != 5 {
			t.Errorf("tổng chữ Hán: got %v want 5", v.Actual)
		}
		return
	}
	t.Fatal("expected han_residue violation")
}
