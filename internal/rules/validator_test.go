package rules

import (
	"strings"
	"testing"
)

func TestValidateChapterOutput_VAL01_CJKRejection(t *testing.T) {
	// 1. Chữ Hán
	hanText := "Hắn ngước nhìn lên bầu trời rồi thở dài: 太阳 mọc rồi."
	errs, _ := ValidateChapterOutput("novel-manga", "Chương 1", hanText)
	if !hasRule(errs, "cjk_residue") {
		t.Fatalf("expected cjk_residue for Han characters, got: %+v", errs)
	}

	// 2. Chữ Nhật (Hiragana / Katakana)
	kanaText := "Cô bé cúi đầu chào: こんにちは, rồi cầm chiếc ô màu đỏ đi mất."
	errs, _ = ValidateChapterOutput("novel-manga", "Chương 1", kanaText)
	if !hasRule(errs, "cjk_residue") {
		t.Fatalf("expected cjk_residue for Japanese Kana, got: %+v", errs)
	}

	// 3. Chữ Hàn (Hangul)
	hangulText := "Người đàn ông gật đầu và nói 안녕하세요 với vị khách lạ."
	errs, _ = ValidateChapterOutput("novel-manga", "Chương 1", hangulText)
	if !hasRule(errs, "cjk_residue") {
		t.Fatalf("expected cjk_residue for Korean Hangul, got: %+v", errs)
	}

	// 4. Văn bản thuần tiếng Việt không dính CJK
	cleanText := generateCleanNovelWords(2100)
	errs, _ = ValidateChapterOutput("novel-manga", "Chương 1", cleanText)
	if hasRule(errs, "cjk_residue") {
		t.Fatalf("clean Vietnamese text should not trigger cjk_residue, got: %+v", errs)
	}
}

func TestValidateChapterOutput_VAL02_WordCountBounds(t *testing.T) {
	// Novel mode quá ngắn (< 1200 từ -> Error)
	shortText := generateCleanNovelWords(500)
	errs, _ := ValidateChapterOutput("novel-manga", "Chương 1", shortText)
	err := findRule(errs, "chapter_too_short")
	if err == nil || err.Severity != SeverityError {
		t.Fatalf("expected chapter_too_short with SeverityError for 500 words, got: %+v", errs)
	}

	// Novel mode ngắn vừa (1500 từ -> Warning)
	mediumText := generateCleanNovelWords(1500)
	errs, _ = ValidateChapterOutput("novel-manga", "Chương 1", mediumText)
	err = findRule(errs, "chapter_too_short")
	if err == nil || err.Severity != SeverityWarning {
		t.Fatalf("expected chapter_too_short with SeverityWarning for 1500 words, got: %+v", errs)
	}

	// Novel mode chuẩn (2500 từ -> Không lỗi)
	normalText := generateCleanNovelWords(2500)
	errs, _ = ValidateChapterOutput("novel-manga", "Chương 1", normalText)
	if hasRule(errs, "chapter_too_short") || hasRule(errs, "chapter_too_long") {
		t.Fatalf("normal 2500 words should be within bounds, got: %+v", errs)
	}

	// Novel mode quá dài (5000 từ -> Warning)
	longText := generateCleanNovelWords(5000)
	errs, _ = ValidateChapterOutput("novel-manga", "Chương 1", longText)
	if !hasRule(errs, "chapter_too_long") {
		t.Fatalf("expected chapter_too_long for 5000 words, got: %+v", errs)
	}

	// Video mode: 400 từ lời đọc -> Warning
	shortVideo := "# Tiêu đề\n\nHOOK 0:00-0:03\nLỜI: " + generateCleanWords(400) + "\nHÌNH: Người que đứng nói chuyện.\n"
	errs, _ = ValidateChapterOutput("doodle-explainer", "Tập 1", shortVideo)
	if !hasRule(errs, "script_word_count_out_of_range") {
		t.Fatalf("expected script_word_count_out_of_range for 400 spoken words, got: %+v", errs)
	}

	// Video mode: 1000 từ lời đọc -> Hợp lệ
	validVideo := "# Tiêu đề\n\nHOOK 0:00-0:03\nLỜI: " + generateCleanWords(1000) + "\nHÌNH: Người que đứng nói chuyện.\n"
	errs, _ = ValidateChapterOutput("doodle-explainer", "Tập 1", validVideo)
	if hasRule(errs, "script_word_count_out_of_range") || hasRule(errs, "script_duration_out_of_range") {
		t.Fatalf("valid 1000 words video should pass bounds, got: %+v", errs)
	}
}

func TestValidateChapterOutput_VAL03_RepetitionLoops(t *testing.T) {
	// 1. Lặp câu liên tiếp
	consecutiveLoop := "Hắn ngước nhìn lên bầu trời đêm tĩnh mịch một mình. Hắn ngước nhìn lên bầu trời đêm tĩnh mịch một mình. Gió thổi qua khe cửa sổ rít từng hồi lạnh buốt."
	errs, _ := ValidateChapterOutput("novel-manga", "Chương 1", consecutiveLoop)
	if !hasRule(errs, "repetition_consecutive_sentence") {
		t.Fatalf("expected repetition_consecutive_sentence, got: %+v", errs)
	}

	// 2. Lặp câu tần suất cao (>= 3 lần trong chương)
	frequentLoop := "Đêm đó trời đổ cơn mưa rất lớn không dứt ở ngoại ô thành phố.\n" +
		"Hắn bước đi trong màn đêm tĩnh mịch vô tận.\n" +
		"Đêm đó trời đổ cơn mưa rất lớn không dứt ở ngoại ô thành phố.\n" +
		"Hắn quay lại quán trọ cũ kỹ bên đường.\n" +
		"Đêm đó trời đổ cơn mưa rất lớn không dứt ở ngoại ô thành phố.\n"
	errs, _ = ValidateChapterOutput("novel-manga", "Chương 1", frequentLoop)
	if !hasRule(errs, "repetition_frequent_sentence") {
		t.Fatalf("expected repetition_frequent_sentence, got: %+v", errs)
	}

	// 3. Lặp đoạn văn liên tiếp
	para := "Bóng tối bao trùm lấy toàn bộ căn nhà gỗ hoang phế nằm sâu trong thung lũng sương mù dày đặc."
	paraLoop := para + "\n\n" + para + "\n\nNgày hôm sau trời bắt đầu có nắng nhẹ trở lại."
	errs, _ = ValidateChapterOutput("novel-manga", "Chương 1", paraLoop)
	if !hasRule(errs, "repetition_paragraph_loop") {
		t.Fatalf("expected repetition_paragraph_loop, got: %+v", errs)
	}

	// 4. Lặp đoạn văn chu kỳ A-B-A-B
	pA := "Tiếng bước chân dồn dập vang lên từ phía cầu thang gỗ cũ kỹ kêu cót két."
	pB := "Gió rít qua từng khe hở của vách đá làm ngọn nến le lói sắp sửa tàn lụi."
	cycleLoop := pA + "\n\n" + pB + "\n\n" + pA + "\n\n" + pB + "\n\nMọi thứ đột ngột dừng lại."
	errs, _ = ValidateChapterOutput("novel-manga", "Chương 1", cycleLoop)
	if !hasRule(errs, "repetition_paragraph_loop") {
		t.Fatalf("expected abab repetition_paragraph_loop, got: %+v", errs)
	}
}

func TestValidateChapterOutput_VAL04_TruncatedEnding(t *testing.T) {
	// Kết thúc hợp lệ với dấu chấm
	validEnding := "Hắn quay lưng bước đi trong bóng hoàng hôn đang dần tắt."
	errs, _ := ValidateChapterOutput("novel-manga", "Chương 1", validEnding)
	if hasRule(errs, "truncated_ending") {
		t.Fatalf("valid terminal dot should not trigger truncated_ending, got: %+v", errs)
	}

	// Kết thúc bị cắt giữa chừng (kết thúc bằng chữ cái)
	truncatedWord := "Bỗng nhiên có một bóng đen vụt qua khe cửa sổ phòng khách làm hắn kinh hãi kêu"
	errs, _ = ValidateChapterOutput("novel-manga", "Chương 1", truncatedWord)
	if !hasRule(errs, "truncated_ending") {
		t.Fatalf("expected truncated_ending for sentence ending with letter, got: %+v", errs)
	}

	// Kết thúc bằng dấu phẩy
	commaEnding := "Hắn quay lại nhìn nàng lần cuối trước khi rời đi,"
	errs, _ = ValidateChapterOutput("novel-manga", "Chương 1", commaEnding)
	if !hasRule(errs, "truncated_ending") {
		t.Fatalf("expected truncated_ending for comma ending, got: %+v", errs)
	}

	// Ngoặc kép thoại hợp lệ
	validDialogue := "Hắn khẽ mỉm cười và đáp: \"Ta nhất định sẽ trở về.\""
	errs, _ = ValidateChapterOutput("novel-manga", "Chương 1", validDialogue)
	if hasRule(errs, "truncated_ending") || hasRule(errs, "unclosed_dialogue_quote") {
		t.Fatalf("valid dialogue ending should not trigger error, got: %+v", errs)
	}

	// Ngoặc kép thoại thiếu dấu kết câu bên trong
	badDialogue := "Hắn khẽ mỉm cười và đáp: \"Ta nhất định sẽ trở về\""
	errs, _ = ValidateChapterOutput("novel-manga", "Chương 1", badDialogue)
	if !hasRule(errs, "truncated_ending") {
		t.Fatalf("dialogue without inner terminal punctuation should trigger truncated_ending, got: %+v", errs)
	}

	// Dấu ngoặc kép lẻ
	oddQuotes := "Hắn nói: \"Ta sẽ đi. Nhưng không ai trả lời."
	errs, _ = ValidateChapterOutput("novel-manga", "Chương 1", oddQuotes)
	if !hasRule(errs, "unclosed_dialogue_quote") {
		t.Fatalf("expected unclosed_dialogue_quote for odd quote count, got: %+v", errs)
	}
}

func TestValidateChapterOutput_VAL05_BannedPhrases(t *testing.T) {
	// 1. Từ convert chung
	convertText := "Hắn nhịn không được bèn hít một hơi khí lạnh khi nhìn thấy bảo vật."
	errs, _ := ValidateChapterOutput("novel-manga", "Chương 1", convertText)
	if !hasRule(errs, "forbidden_phrases") {
		t.Fatalf("expected forbidden_phrases for convert terms, got: %+v", errs)
	}

	// 2. Sáo ngữ suspense
	suspenseText := "Vị thám tử mỉm cười tự tin: Chân tướng chỉ có một mà thôi!"
	errs, _ = ValidateChapterOutput("suspense", "Chương 1", suspenseText)
	if !hasRule(errs, "forbidden_phrases") {
		t.Fatalf("expected forbidden_phrases for suspense cliché, got: %+v", errs)
	}

	// 3. Sáo ngữ romance
	romanceText := "Vị tổng tài bá đạo bước vào phòng và nói: Nữ nhân, ngươi đang đùa với lửa."
	errs, _ = ValidateChapterOutput("romance", "Chương 1", romanceText)
	if !hasRule(errs, "forbidden_phrases") {
		t.Fatalf("expected forbidden_phrases for romance cliché, got: %+v", errs)
	}

	// 4. Sáo ngữ fantasy
	fantasyText := "Khí thế ngập trời của hắn làm mọi người xung quanh kinh hãi vạn phần."
	errs, _ = ValidateChapterOutput("fantasy", "Chương 1", fantasyText)
	if !hasRule(errs, "forbidden_phrases") {
		t.Fatalf("expected forbidden_phrases for fantasy cliché, got: %+v", errs)
	}

	// 5. Tiếng lóng anachronisms trong lịch sử
	historyText := "Quân giặc nghe tin đại bại liền bảo nhau chuyến này toang rồi, đúng là ảo ma thật."
	errs, _ = ValidateChapterOutput("vietnamese-history", "Chương 1", historyText)
	if !hasRule(errs, "forbidden_phrases") {
		t.Fatalf("expected forbidden_phrases for history anachronisms, got: %+v", errs)
	}

	// 6. Sáo ngữ video
	videoText := "Chào mừng các bạn đã quay trở lại với kênh, đừng quên bấm like và đăng ký kênh nhé."
	errs, _ = ValidateChapterOutput("doodle-explainer", "Tập 1", videoText)
	if !hasRule(errs, "forbidden_phrases") {
		t.Fatalf("expected forbidden_phrases for video cliché, got: %+v", errs)
	}
}

func TestValidateChapterOutput_VAL06_LeakedScriptTagGuard(t *testing.T) {
	// Rò rỉ thẻ kịch bản vào novel-manga -> Bị chặn
	novelWithScriptTags := "Trời vừa hửng sáng trên ngọn đồi.\n\nHOOK 0:00-0:03\nLỜI: Hắn bước ra khỏi lều trại.\nHÌNH: Người que cầm kiếm.\n\nHắn bắt đầu cuộc hành trình."
	errs, _ := ValidateChapterOutput("novel-manga", "Chương 1", novelWithScriptTags)
	if !hasRule(errs, "leaked_script_tag") {
		t.Fatalf("expected leaked_script_tag in novel prose mode, got: %+v", errs)
	}

	// Video mode dùng thẻ kịch bản -> Hợp lệ, không bị chặn
	videoScript := "# Tiêu đề video\n\nHOOK 0:00-0:03\nLỜI: Bạn có bao giờ tự hỏi vì sao không?\nHÌNH: Người que gãi đầu bối rối.\n\nCHỐT 4:30-5:00\nLỜI: Cảm ơn bạn đã xem.\nHÌNH: Vẫy tay chào."
	errs, _ = ValidateChapterOutput("doodle-explainer", "Tập 1", videoScript)
	if hasRule(errs, "leaked_script_tag") {
		t.Fatalf("video mode should NOT trigger leaked_script_tag, got: %+v", errs)
	}
}

func hasRule(errs []ValidationError, rule string) bool {
	for _, e := range errs {
		if e.Rule == rule {
			return true
		}
	}
	return false
}

func findRule(errs []ValidationError, rule string) *ValidationError {
	for _, e := range errs {
		if e.Rule == rule {
			return &e
		}
	}
	return nil
}

func generateCleanWords(count int) string {
	words := []string{"hành", "vi", "con", "người", "hình", "thành", "từ", "nhận", "thức", "và", "môi", "trường", "xung", "quanh"}
	var b strings.Builder
	for i := 0; i < count; i++ {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(words[i%len(words)])
	}
	b.WriteString(".")
	return b.String()
}

func generateCleanNovelWords(count int) string {
	baseWords := []string{
		"Trời", "đêm", "mênh", "mông", "gió", "thổi", "qua", "từng", "hàng", "cây",
		"trên", "đỉnh", "núi", "cao", "vời", "vợi", "xa", "xa", "có", "ánh",
		"lửa", "chập", "chờn", "trong", "sương", "lạnh", "bước", "chân", "người", "lữ",
		"thứ", "dừng", "lại", "bên", "con", "suối", "nhỏ", "nước", "chảy", "róc",
	}
	var b strings.Builder
	for i := 0; i < count; i++ {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(baseWords[i%len(baseWords)])
		if (i+1)%25 == 0 {
			b.WriteString(".")
		}
	}
	if !strings.HasSuffix(b.String(), ".") {
		b.WriteString(".")
	}
	return b.String()
}
