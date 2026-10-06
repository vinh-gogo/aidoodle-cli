package tools

import (
	"slices"
	"strings"
	"testing"

	"github.com/voocel/ainovel-cli/internal/rules"
)

// voiceWords sinh một câu lời đọc có đúng n từ.
func voiceWords(n int) string {
	return strings.TrimSpace(strings.Repeat("từ ", n))
}

func validScript() string {
	return strings.Join([]string{
		"# Vì sao tiền cứ mất giá?",
		"",
		"HOOK 0:00-0:03",
		"LỜI: " + voiceWords(10),
		"HÌNH: Ông Ú cầm hòn đá, mặt hoảng.",
		"CHỮ: Lạm phát là gì?",
		"",
		"CẢNH 1 0:03-0:25",
		"LỜI: " + voiceWords(60),
		"HÌNH: Cả bộ lạc cùng đúc thêm đá.",
		"ÂM: tiếng đá lạch cạch",
		"",
		"CẢNH 2 0:25-0:55",
		"LỜI: " + voiceWords(60),
		"HÌNH: Một hòn đá đổi được ít khoai hơn.",
		"",
		"CHỐT 0:55-1:10",
		"LỜI: " + voiceWords(40),
		"HÌNH: Ông Ú gãi đầu.",
		"",
		"CAPTION: Tiền mất giá, nói đơn giản thế này.",
		"HASHTAG: #lamphat #taichinh #doodle",
		"NGUỒN: không có (kiến thức nền)",
		"CẦN KIỂM CHỨNG: không có",
	}, "\n")
}

func lintRuleNames(text string) []string {
	var names []string
	for _, v := range lintScript(text) {
		names = append(names, v.Rule)
	}
	return names
}

func TestLintScript_ValidScriptHasNoWarnings(t *testing.T) {
	if got := lintRuleNames(validScript()); len(got) != 0 {
		t.Fatalf("kịch bản hợp lệ không được có cảnh báo, got %v", got)
	}
}

func TestLintScript_CRLFAndLowercaseTags(t *testing.T) {
	text := strings.ReplaceAll(validScript(), "\n", "\r\n")
	text = strings.ReplaceAll(text, "LỜI:", "lời:")
	if got := lintRuleNames(text); len(got) != 0 {
		t.Fatalf("CRLF/thẻ chữ thường vẫn phải hợp lệ, got %v", got)
	}
}

func TestLintScript_MultilineVoiceIsCounted(t *testing.T) {
	text := strings.Replace(validScript(), "LỜI: "+voiceWords(60), "LỜI: "+voiceWords(30)+"\n"+voiceWords(30), 1)
	if got := lintRuleNames(text); len(got) != 0 {
		t.Fatalf("LỜI nhiều dòng phải được cộng đủ từ, got %v", got)
	}
}

func TestLintScript_OnlyVoiceCountsTowardWords(t *testing.T) {
	// HÌNH/CHỮ/CAPTION dài không được tính vào số từ.
	text := strings.Replace(validScript(), "HÌNH: Ông Ú gãi đầu.", "HÌNH: "+voiceWords(500), 1)
	if got := lintRuleNames(text); len(got) != 0 {
		t.Fatalf("thẻ không phải LỜI không được tính từ, got %v", got)
	}
}

func TestLintScript_Findings(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(string) string
		want   string
	}{
		{"no title", func(s string) string { return strings.TrimPrefix(s, "# Vì sao tiền cứ mất giá?\n\n") }, "script_no_title"},
		{"no hook", func(s string) string { return strings.Replace(s, "HOOK 0:00-0:03", "CẢNH 0 0:00-0:03", 1) }, "script_no_hook"},
		{"hook not first", func(s string) string {
			s = strings.Replace(s, "HOOK 0:00-0:03", "CẢNH 0 0:00-0:03", 1)
			return strings.Replace(s, "CẢNH 2 0:25-0:55", "HOOK 0:25-0:55", 1)
		}, "script_no_hook"},
		{"no closing", func(s string) string { return strings.Replace(s, "CHỐT 0:55-1:10", "CẢNH 3 0:55-1:10", 1) }, "script_no_closing"},
		{"too few words", func(s string) string { return strings.Replace(s, voiceWords(60), "ngắn", 2) }, "script_words_out_of_range"},
		{"missing visual", func(s string) string { return strings.Replace(s, "HÌNH: Ông Ú gãi đầu.", "CHỮ: hết", 1) }, "script_block_missing_visual"},
		{"missing voice", func(s string) string { return strings.Replace(s, "LỜI: "+voiceWords(40), "CHỮ: không lời", 1) }, "script_block_missing_voice"},
		{"missing caption", func(s string) string {
			return strings.Replace(s, "CAPTION: Tiền mất giá, nói đơn giản thế này.\n", "", 1)
		}, "script_missing_caption"},
		{"missing hashtag", func(s string) string {
			return strings.Replace(s, "HASHTAG: #lamphat #taichinh #doodle", "HASHTAG: không có thẻ", 1)
		}, "script_missing_hashtag"},
		{"missing source", func(s string) string { return strings.Replace(s, "NGUỒN: không có (kiến thức nền)\n", "", 1) }, "script_missing_source"},
		{"missing unverified", func(s string) string { return strings.Replace(s, "CẦN KIỂM CHỨNG: không có", "", 1) }, "script_missing_unverified"},
	}
	for _, tc := range cases {
		got := lintRuleNames(tc.mutate(validScript()))
		if !slices.Contains(got, tc.want) {
			t.Errorf("%s: want %s in %v", tc.name, tc.want, got)
		}
	}
}

func TestLintScript_DurationFromDeclaredTimes(t *testing.T) {
	// 170 từ ≈ 68 s theo tốc độ đọc nhưng mốc khai báo kéo tới 3:30 → vượt 180 s.
	text := strings.Replace(validScript(), "CHỐT 0:55-1:10", "CHỐT 0:55-3:30", 1)
	if got := lintRuleNames(text); !slices.Contains(got, "script_duration_out_of_range") {
		t.Fatalf("want script_duration_out_of_range in %v", got)
	}
}

func TestLintScript_ViolationsAreWarnings(t *testing.T) {
	for _, v := range lintScript("không phải kịch bản") {
		if v.Severity != "warning" {
			t.Fatalf("violation %s phải ở mức warning, got %s", v.Rule, v.Severity)
		}
	}
}

func TestLintScript_NoMarkdownResidueOnValidScript(t *testing.T) {
	// Định dạng chuẩn không được kích hoạt lint markdown_residue (dòng HASHTAG chứa '#').
	for _, v := range rules.Lint(validScript()) {
		if v.Rule == "markdown_residue" || v.Rule == "han_residue" {
			t.Fatalf("kịch bản chuẩn không được có %s", v.Rule)
		}
	}
}

func TestLintScript_WriterPromptExample(t *testing.T) {
	script := `# Lạm phát: vì sao củ khoai đắt lên mỗi ngày

HOOK 0:00-0:03
LỜI: Hôm qua củ khoai nướng giá hai vỏ sò. Hôm nay giá ba. Ai làm đây?
HÌNH: Ugg, người que mặc da thú, giơ củ khoai bốc khói, mắt tròn hoảng hốt, sau lưng là bảng giá khắc trên đá.
CHỮ: LẠM PHÁT LÀ GÌ?

CẢNH 1 0:03-0:25
LỜI: Chào mấy bạn, tui là Ugg, dân đồ đá. Ở bộ lạc tui, vỏ sò chính là tiền. Muốn có vỏ sò thì phải ra biển nhặt. Mà nhặt được thì mệt lắm. Nên vỏ sò mới quý. Một củ khoai đổi hai vỏ sò, ai cũng vui.
HÌNH: Cảnh bãi biển, Ugg còng lưng nhặt từng vỏ sò bỏ vào giỏ; mặt trời mỉm cười.
ÂM: Tiếng sóng, nhạc bongo nhẹ.

CẢNH 2 0:25-0:50
LỜI: Rồi một hôm, ông Grok tìm ra bãi biển đầy vỏ sò. Cả bộ lạc chạy ra nhặt. Ai cũng giàu lên trông thấy! Nhưng khoai thì vẫn chỉ có bấy nhiêu. Người nào cũng cầm đầy vỏ sò, và ai cũng muốn mua khoai. Thế là bà bán khoai nghĩ: mấy người trả nhiều thế, mình tăng giá thôi.
HÌNH: Cả bộ lạc người que ôm giỏ vỏ sò đầy ắp xếp hàng trước một sạp khoai bé xíu; bà bán khoai gõ bảng giá, đổi số hai thành số ba.
CHỮ: TIỀN NHIỀU HƠN, HÀNG KHÔNG ĐỔI

CẢNH 3 0:50-1:10
LỜI: Đó, gọi là lạm phát. Vỏ sò thì nhiều hơn, nên mỗi vỏ sò mua được ít khoai hơn. Vỏ sò không hỏng, nhưng giá trị của nó thì mòn dần. Ngày nay người ta không nhặt vỏ sò, mà in tiền. Nghe quen không?
HÌNH: Một con mammoth bụng phệ ngồi trên đống vỏ sò khổng lồ; bên cạnh, củ khoai nhỏ xíu, nhìn rất tội.

CHỐT 1:10-1:20
LỜI: Nên lần sau thấy giá tăng, đừng trách bà bán khoai. Hãy hỏi: ai vừa tìm ra bãi biển mới?
HÌNH: Ugg gãi đầu nhìn ra xa, bóng ông Grok vác cả bao vỏ sò đi qua.
CHỮ: THEO DÕI ĐỂ HIỂU TIẾP

CAPTION: Củ khoai không đắt lên, vỏ sò mới rẻ đi. Lạm phát giải thích bằng đồ đá.
HASHTAG: #lamphat #kinhte #doodle #giaithich #hoccungtiktok
NGUỒN: không có (kiến thức nền)
CẦN KIỂM CHỨNG: không có`

	for _, v := range rules.Lint(script) {
		t.Errorf("rules.Lint reported unexpected violation: %+v", v)
	}
	if vs := lintScript(script); len(vs) != 0 {
		t.Errorf("lintScript reported unexpected violations: %+v", vs)
	}
}

func TestLintScript_ChapterTemplateExample(t *testing.T) {
	script := `# Lạm phát: vì sao rìu đá của bạn bốc hơi

HOOK 0:00-0:03
LỜI: Rìu đá của bạn đang bốc hơi. Ai lấy mất?
HÌNH: Que Ú ôm mười cái rìu đá, từng cái nhẹ bẫng bay lên thành khói. Mặt Que Ú tái mét.
CHỮ: RÌU ĐÁ BỐC HƠI?

CẢNH 1 0:03-0:22
LỜI: Bộ lạc Que có đúng một con mammoth. Hôm qua, một con mammoth đổi được mười cái rìu. Rồi thủ lĩnh Que Bự nghĩ ra chuyện hay ho: phát vỏ sò cho cả bộ lạc để tiện mua bán. Ai cũng khoái, vì tự dưng giàu hẳn lên.
HÌNH: Que Bự rải vỏ sò từ trên đỉnh đá xuống như mưa. Cả bộ lạc nhảy cẫng, ôm đầy vỏ sò.
ÂM: tiếng vỏ sò lách cách, nhạc vui.

CẢNH 2 0:22-0:47
LỜI: Nhưng mammoth thì vẫn chỉ có một con. Người bán mammoth nhìn đống vỏ sò mới, nhún vai: giờ một con mammoth giá hai mươi vỏ sò. Hôm qua còn mười. Mammoth đâu có to ra, chỉ là vỏ sò dễ kiếm hơn nên nó bớt quý đi. Cái đó gọi là lạm phát.
HÌNH: Cân hai đĩa: bên trái một con mammoth, bên phải chồng vỏ sò cao dần lên. Chữ giá nhảy từ 10 thành 20.
CHỮ: Tiền nhiều lên = tiền bớt quý

CẢNH 3 0:47-1:02
LỜI: Còn Que Ú thì sao? Tuần này ông vẫn lãnh mười vỏ sò công săn bắn. Mười vỏ hôm qua mua được một con mammoth, mười vỏ hôm nay chỉ mua nổi nửa cái đùi.
HÌNH: Que Ú cầm mười vỏ sò, nhìn nửa cái đùi mammoth trên sạp, nước mắt chảy thành hai vệt.

CHỐT 1:02-1:16
LỜI: Vậy ai lấy mất rìu của bạn? Chính là cái tay vỏ sò phát ra quá nhiều. Lần sau thấy giá tăng, hãy hỏi: ai đang rải thêm vỏ sò?
HÌNH: Que Bự nấp sau tảng đá, tay vẫn đang rải vỏ sò. Que Ú chỉ thẳng vào mặt.
CHỮ: Giá tăng? Hỏi: ai in thêm tiền?

CAPTION: Lạm phát giải thích bằng vỏ sò và mammoth, 60 giây là hiểu.
HASHTAG: #lamphat #kinhte #giaithichdongian #doodle
NGUỒN: không có (kiến thức nền)
CẦN KIỂM CHỨNG: không có`

	for _, v := range rules.Lint(script) {
		t.Errorf("rules.Lint reported unexpected violation: %+v", v)
	}
	if vs := lintScript(script); len(vs) != 0 {
		t.Errorf("lintScript reported unexpected violations: %+v", vs)
	}
}
