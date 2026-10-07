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
		"",
		"CẢNH 1 0:03-0:45",
		"LỜI: " + voiceWords(105),
		"HÌNH: Cả bộ lạc cùng đúc thêm đá.",
		"ÂM: tiếng đá lạch cạch",
		"",
		"CẢNH 2 0:45-1:45",
		"LỜI: " + voiceWords(150),
		"HÌNH: Một hòn đá đổi được ít khoai hơn.",
		"",
		"CẢNH 3 1:45-2:45",
		"LỜI: " + voiceWords(150),
		"HÌNH: Grok phát hiện mỏ đá mới ven biển.",
		"",
		"CẢNH 4 2:45-3:45",
		"LỜI: " + voiceWords(150),
		"HÌNH: Núi đá đè bẹp cán cân trao đổi.",
		"",
		"CẢNH 5 3:45-4:30",
		"LỜI: " + voiceWords(110),
		"HÌNH: Ông Ú bắt đầu đổi đá lấy công cụ thật.",
		"",
		"CHỐT 4:30-5:15",
		"LỜI: " + voiceWords(80),
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
	text := strings.Replace(validScript(), "LỜI: "+voiceWords(150), "LỜI: "+voiceWords(75)+"\n"+voiceWords(75), 1)
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
			return strings.Replace(s, "CẢNH 2 0:45-1:45", "HOOK 0:45-1:45", 1)
		}, "script_no_hook"},
		{"no closing", func(s string) string { return strings.Replace(s, "CHỐT 4:30-5:15", "CẢNH 6 4:30-5:15", 1) }, "script_no_closing"},
		{"too few words", func(s string) string { return strings.Replace(s, voiceWords(150), "ngắn", 2) }, "script_words_out_of_range"},
		{"missing visual", func(s string) string { return strings.Replace(s, "HÌNH: Ông Ú gãi đầu.", "ÂM: hết", 1) }, "script_block_missing_visual"},
		{"missing voice", func(s string) string { return strings.Replace(s, "LỜI: "+voiceWords(80), "ÂM: không lời", 1) }, "script_block_missing_voice"},
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
	// 755 từ ≈ 302 s theo tốc độ đọc nhưng mốc khai báo kéo tới 11:30 (690s) → vượt 600 s.
	text := strings.Replace(validScript(), "CHỐT 4:30-5:15", "CHỐT 4:30-11:30", 1)
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
	script := `# Vì sao tiền cứ mất giá: bí mật củ khoai đồ đá

HOOK 0:00-0:03
LỜI: Hôm qua củ khoai hai vỏ sò, hôm nay thành năm. Ai lấy mất tiền của bạn?
HÌNH: Ugg người que giơ củ khoai nướng bốc khói, hai mắt tròn xoe nhìn bảng khắc đá giá tăng vọt.

CẢNH 1 0:03-0:45
LỜI: Chào mấy bạn, tui là Ugg, dân đồ đá chính hiệu. Mấy bạn có từng đi làm quần quật cả tháng, nhận lương thấy vui vui, nhưng bước chân vào tiệm tạp hóa mua gói mì, đổ bình xăng hay trả tiền cốc trà sữa thì giật mình nhận ra: ủa sao lương mình đứng yên mà giá mọi thứ cứ tự động leo thang? Có phải người bán hàng đang cố tình bắt chẹt bạn? Hay đồng tiền trong ví bạn tự nhiên bốc hơi? Ở thời đồ đá của tụi tui, câu chuyện này cũng diễn ra y hệt, nhưng hung thủ thật sự không nằm ở người bán, mà nằm ở một thứ tinh vi hơn rất nhiều.
HÌNH: Ugg cầm giỏ đi chợ tiền sử, trước mặt là các sạp thịt mammoth, nấm rừng với bảng giá đá dựng đứng, biểu cảm dở khóc dở cười.
ÂM: Tiếng thở dài nhẹ, nhạc bộ gõ rộn ràng.

CẢNH 2 0:45-1:45
LỜI: Để hiểu rõ, quay lại hang đá của bộ lạc tụi tui một chút. Thời đó chưa có giấy bạc hay thẻ ngân hàng, tụi tui dùng vỏ sò biển làm tiền để đổi chác. Muốn có vỏ sò, bạn phải lặn lội ra tận bãi đá ngầm xa xôi, vượt sóng lớn, trèo đèo lội suối cả ngày trời mới nhặt được một vài chiếc vỏ sò óng ánh. Nhặt vỏ sò mệt đứt hơi, nên trong bộ lạc, vỏ sò cực kỳ quý giá. Ai có mười vỏ sò là coi như có một khoản tích lũy mồ hôi nước mắt. Lúc này, cả thung lũng chỉ có một bác nông dân que chuyên trồng khoai mài. Cứ một ngày công nhặt được hai vỏ sò, bạn đổi được một củ khoai nướng thơm phức ăn no bụng. Bác nông dân vui vì có vỏ sò để đi đổi rìu đá của thợ rèn, bạn vui vì có củ khoai ăn. Mọi thứ cân bằng êm đềm suốt bao mùa săn bắn.
HÌNH: Cảnh bãi biển sóng vỗ, Ugg cặm cụi mò từng vỏ sò; góc sau là bác nông dân que đang nướng khoai mài trên bếp lửa hồng, trao đổi vui vẻ.

CẢNH 3 1:45-2:45
LỜI: Nhưng chuyện quái gở bắt đầu khi một nhân vật tên là Grok xuất hiện. Ông Grok này lười đi săn nhưng lại cực kỳ ma mãnh. Một hôm, ông ta tình cờ phát hiện ra một cái hang bí mật ven biển, nơi bão đánh dạt hàng ngàn vỏ sò trôi dạt vào chất cao như núi. Thế là chẳng cần tốn một giọt mồ hôi, Grok vác về cả chục bao tải vỏ sò đầy ắp. Bỗng nhiên, Grok trở thành tỷ phú đô-la thời tiền sử chỉ sau một đêm. Có nhiều tiền quá thì làm gì? Grok chạy ngay ra sạp của bác nông dân, hét lớn: Bán cho tui hết sạch chỗ khoai này, tui trả gấp đôi, gấp ba! Bác nông dân tròn mắt sướng rơn, gom hết khoai bán cho Grok. Nhưng rồi ngày hôm sau, khi Ugg và những người dân lao động thật thà khác cầm hai vỏ sò mồ hôi nước mắt đến mua khoai, bác nông dân lắc đầu: Xin lỗi nghen, giờ khoai giá năm vỏ sò rồi!
HÌNH: Grok người que mắt híp cười toe toét, đẩy chiếc xe cút kít đá chất đầy bao tải vỏ sò ập vào sạp khoai; Ugg đứng bên cạnh cầm hai vỏ sò lẻ loi ngơ ngác.
ÂM: Tiếng chuông leng keng dồn dập, tiếng xôn xao hốt hoảng.

CẢNH 4 2:45-3:45
LỜI: Mấy bạn thấy điều gì vừa xảy ra không? Củ khoai mài trên bếp lửa đâu có to hơn, đâu có thơm ngon hơn hôm qua. Nó vẫn chỉ là một củ khoai bình thường. Cái thay đổi duy nhất là số lượng vỏ sò trong hang đá đã tăng gấp mười lần, trong khi số củ khoai vẫn y như cũ. Khi quá nhiều tiền cùng đuổi theo một lượng hàng hóa không đổi, thì từng đồng tiền buộc phải rẻ rúng đi. Đến thời hiện đại, vỏ sò được thay bằng những tờ giấy bạc in hình hoa văn đẹp đẽ và những con số nhảy múa trên màn hình ứng dụng điện thoại. Khi các cỗ máy ngân hàng trung ương in thêm tiền tràn ngập thị trường để kích thích kinh tế, thì chiếc bánh mì, ly cà phê hay căn nhà bạn mơ ước cũng y như củ khoai của Ugg. Lương của bạn tăng năm phần trăm, nhưng lượng tiền trong nền kinh tế tăng hai mươi phần trăm, thì thực chất bạn đang nghèo đi từng ngày mà không hề hay biết.
HÌNH: Cán cân đá khổng lồ: một bên đĩa cân là núi vỏ sò nặng trĩu đè bẹp xuống, bên kia đĩa cân là củ khoai bay bổng lên cao; chuyển cảnh sang que hiện đại cầm điện thoại nhìn số dư tài khoản.

CẢNH 5 3:45-4:30
LỜI: Vậy nên, lần sau khi nghe tin giá bát phở tăng năm nghìn hay tiền thuê nhà tăng thêm một triệu, đừng vội bực tức trút giận lên cô bán phở hay chú chủ nhà. Họ cũng chỉ là những người que đang cố gắng giữ cho củ khoai của mình không bị vỏ sò nhấn chìm mà thôi. Thay vì ngồi than vãn hay giữ khư khư đống vỏ sò dưới gầm giường để nhìn nó bốc hơi từng ngày, người khôn ngoan thời nay học cách đổi vỏ sò lấy những tài sản thật sự có giá trị bền vững: học thêm kỹ năng để nâng cao giá trị bản thân, hoặc đầu tư vào những thứ không thể dễ dàng in thêm được.
HÌNH: Ugg gật gù ngộ ra, bắt tay cô bán phở que; sau đó Ugg chuyển sang cầm búa đá mài giũa dụng cụ sắc bén, khuôn mặt tự tin, kiên định.

CHỐT 4:30-5:15
LỜI: Tiền không tự nhiên sinh ra và cũng không tự nhiên mất đi, nó chỉ chuyển từ túi người giữ tiền mặt sang túi người nắm giữ tài sản thật mà thôi. Bạn đang để vỏ sò của mình dưới gối hay đã đem đi đổi lấy công cụ lao động tốt hơn? Bình luận cho Ugg biết góc nhìn của bạn bên dưới nghen. Đừng quên bấm theo dõi kênh để cùng người que tụi tui giải mã những bí mật kinh tế thú vị tiếp theo.
HÌNH: Ugg người que nháy mắt cười tươi, vẫy tay chào người xem cạnh đống lửa ấm áp; góc màn hình hiện biểu tượng nút theo dõi và hộp bình luận nhấp nháy.

CAPTION: Củ khoai không đắt lên, vỏ sò mới rẻ đi. Bản chất lạm phát hiểu trong 5 phút cùng dân đồ đá.
HASHTAG: #lamphat #kinhte #taichinh #doodle #kienthuc #giaithich #xuhuong
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
	script := `# Nỗi sợ kỷ đá: vì sao bạn luôn lo âu vô cớ

HOOK 0:00-0:03
LỜI: Nửa đêm chuẩn bị ngủ, não bạn bỗng tua lại chuyện quê mùa năm năm trước. Ai làm trò này?
HÌNH: Que Ú nằm trên đệm rơm, mắt mở to trừng trừng nhìn lên trần hang, một cái bóng đen nhỏ hình não bộ ngồi cạnh thì thầm.

CẢNH 1 0:03-0:45
LỜI: Chào mấy bạn, tui là Que Ú. Có bao giờ bạn ngồi trong phòng máy lạnh êm ấm, công việc ổn định, đồ ăn đầy tủ lạnh, nhưng trong ngực bỗng dâng lên một cảm giác bồn chồn lo lắng khó tả? Hay vừa gửi một email cho sếp, vừa đăng một tấm ảnh lên mạng xã hội, tim bạn đã đập thình thịch như thể sắp có tai họa giáng xuống đầu? Bạn tự trách mình sao yếu đuối, sao nhạy cảm quá mức. Nhưng sự thật là: bạn không hề có lỗi. Thứ đang hành hạ bạn lúc nửa đêm không phải là sự hèn nhát, mà chính là một cỗ máy bảo vệ cổ xưa được lập trình từ thời đồ đá hàng vạn năm trước.
HÌNH: Que Ú ngồi ôm gối trong phòng ngủ hiện đại, xung quanh là những bong bóng suy nghĩ chứa hóa đơn, tin nhắn chưa trả lời và ánh mắt soi mói.
ÂM: Tiếng đồng hồ tích tắc, nhịp tim đập nhẹ.

CẢNH 2 0:45-1:45
LỜI: Hãy thử tưởng tượng quay lại thảo nguyên kỷ băng hà mười nghìn năm trước. Tổ tiên của chúng ta lúc đó không có nhà lầu, không có cảnh sát hay bệnh viện. Rời khỏi miệng hang là bước vào lãnh địa của hổ răng kiếm, gấu khổng lồ và những bụi gai độc. Trong môi trường khắc nghiệt đó, có hai kiểu người que sinh sống. Kiểu thứ nhất là những người cực kỳ lạc quan, thấy bụi cây xào xạc thì mỉm cười bảo: Chắc là làn gió mát lành thổi qua thôi mà! Kết quả là chín mươi chín phần trăm những người lạc quan đó đã trở thành bữa tối ngon lành cho hổ răng kiếm trước khi kịp lập gia đình. Còn kiểu người thứ hai là những người cực kỳ cảnh giác, đa nghi và hay lo sợ. Nghe tiếng bụi cây lay động, họ lập tức dựng lông tơ, tim đập dữ dội, chân phóng hết tốc lực leo tót lên ngọn cây cao.
HÌNH: Phân chia hai nửa khung hình: bên trái người que lạc quan hái hoa bị móng vuốt hổ vồ; bên phải người que mắt tròn xoe bật nhảy tót lên cành cây trốn thoát an toàn.

CẢNH 3 1:45-2:45
LỜI: Dù tiếng xào xạc đó chín mươi chín lần chỉ là một cơn gió thổi, nhưng chỉ cần đúng một lần có thú dữ thật, thì người hay lo âu đã giữ được mạng sống quý giá. Và đoán xem, bạn là hậu duệ của ai? Đúng rồi, bạn chính là con cháu ruột thịt của những người que hay lo âu nhất thời tiền sử! Não bộ của bạn thừa hưởng một trung tâm báo động mang tên hạch hạnh nhân. Nhiệm vụ tối thượng của nó suốt hàng trăm thế hệ qua là quét tìm hiểm họa để bạn không bị chết đói hay bị ăn thịt. Vấn đề trớ trêu nằm ở chỗ: nền văn minh của loài người phát triển quá nhanh chỉ trong vài trăm năm gần đây, nhưng bộ não sinh học của chúng ta thì vẫn giữ nguyên cấu trúc đồ đá từ hàng triệu năm trước. Cỗ máy sinh tồn đó chưa kịp cập nhật phiên bản thích ứng với thế giới văn phòng và điện thoại thông minh ngày nay.
HÌNH: Que Ú cầm tấm bản đồ sinh học chỉ vào chấm đỏ trong não; chuyển cảnh sang que tiền sử vác giáo mác đứng giữa ngã tư đường phố xe cộ tấp nập.
ÂM: Tiếng còi báo động khẩn cấp rồi chuyển thành nhạc hài hước.

CẢNH 4 2:45-3:45
LỜI: Khi bạn nhận được một tin nhắn thông báo họp gấp từ sếp lúc năm giờ chiều, hạch hạnh nhân trong đầu bạn không hiểu khái niệm công việc văn phòng hay đánh giá hiệu suất là gì. Trong mắt nó, ánh mắt nghiêm nghị của sếp hay tiếng chuông điện thoại réo rắt cũng nguy hiểm y hệt như tiếng gầm của một con hổ răng kiếm đang rình rập sau tảng đá. Ngay lập tức, cơ thể bạn tự động kích hoạt cơ chế chiến hay biến: adrenaline được bơm ồ ạt vào máu, tim đập thình thịch, cơ bắp căng cứng và dạ dày thắt lại. Nhưng ngặt nỗi, bạn không thể cầm giáo mác lao vào đâm chiếc máy tính, cũng không thể bỏ chạy thục mạng ra khỏi phòng làm việc. Toàn bộ năng lượng sinh tồn khổng lồ đó không được giải phóng ra ngoài, đành phải quay ngược vào trong cắn xé tâm can bạn, tạo thành những cơn hoảng loạn và lo âu kéo dài.
HÌNH: Sếp que hiện hình bóng con hổ răng kiếm to lớn; Que Ú ngồi trước laptop mồ hôi vã như tắm, tim ngoài lồng ngực nhảy tưng tưng như quả bóng.

CẢNH 5 3:45-4:30
LỜI: Vậy làm sao để sống yên ổn với người tiền sử gác cổng mẫn cán bên trong bạn? Đừng cố gắng đàn áp hay chửi bới nỗi sợ của mình, vì càng chống cự thì hạch hạnh nhân càng tin rằng bạn đang gặp nguy hiểm thật sự. Thay vào đó, mỗi khi thấy tim đập nhanh hay bồn chồn lo lắng, hãy mỉm cười hít một hơi thật sâu và tự nhủ: Cảm ơn bạn gác cổng nhé, tui biết bạn đang muốn bảo vệ tui khỏi hổ răng kiếm, nhưng chỗ này an toàn rồi. Sau đó, hãy đứng dậy đi lại vài vòng hoặc vận động nhẹ nhàng để đốt cháy lượng adrenaline dư thừa, giúp cơ thể trở về trạng thái bình yên.
HÌNH: Que Ú ngồi xếp bằng hít thở sâu, tay vỗ vai an ủi một chú người que nhỏ xíu đội mũ thợ săn đứng bên cạnh; cả hai cùng thở ra làn khói nhẹ nhõm.

CHỐT 4:30-5:15
LỜI: Lo âu không phải là khiếm khuyết tâm lý, mà là bằng chứng cho thấy tổ tiên bạn đã chiến đấu ngoan cường thế nào để dòng giống sinh tồn đến tận hôm nay. Hãy học cách làm hòa với người gác cổng đồ đá bên trong để sống nhẹ nhõm hơn mỗi ngày. Bạn thường hay lo lắng nhất về chuyện gì khi đêm xuống? Hãy chia sẻ bên dưới để cùng Que Ú giải tỏa nghen. Đừng quên bấm theo dõi kênh để lắng nghe thêm nhiều câu chuyện thấu hiểu tâm lý thú vị.
HÌNH: Que Ú mỉm cười thanh thản ngả lưng ngủ ngon giấc dưới bầu trời đầy sao lung linh; dòng chữ bình luận hiện lên mời gọi tương tác cùng biểu tượng theo dõi kênh.

CAPTION: Nỗi sợ kỷ đá: vì sao bạn luôn lo âu vô cớ và bí quyết làm hòa với người gác cổng bên trong.
HASHTAG: #tamly #loau #kienthuc #doodle #suckhoetinhthan #chualanh #xuhuong
NGUỒN: không có (kiến thức nền)
CẦN KIỂM CHỨNG: không có`

	for _, v := range rules.Lint(script) {
		t.Errorf("rules.Lint reported unexpected violation: %+v", v)
	}
	if vs := lintScript(script); len(vs) != 0 {
		t.Errorf("lintScript reported unexpected violations: %+v", vs)
	}
}
