package stylestat

import (
	"strings"
	"testing"
)

func chapterWith(body string) string {
	return "# 标题\n" + body
}

func TestComputeBelowMinChapters(t *testing.T) {
	in := Input{Chapters: []string{"a", "b", "c", "d"}}
	if Compute(in) != nil {
		t.Fatal("below minChapters should return nil")
	}
}

func TestComputePatterns(t *testing.T) {
	body := "他不是愤怒，而是恐惧。沉默了几息。像一盏灯。她眼中闪过慌乱，心头一紧。他觉得这是一种说不出的寒意。\n正文。\n"
	chapters := make([]string, 6)
	for i := range chapters {
		chapters[i] = chapterWith(body)
	}
	s := Compute(Input{Chapters: chapters})
	if s == nil {
		t.Fatal("expected stats")
	}
	want := map[string]int{
		"Câu phủ định sửa sai 『không phải… mà là…』":                          6,
		"Lượng từ thời gian 『X giây / tích tắc』":                             6,
		"So sánh ví von 『giống như / tựa như / như thể』":                     6,
		"Nhịp im lặng 『im lặng / không nói gì』":                              6,
		"Khuôn mẫu thần thái 『mặt tái mét / nhếch mép / mắt sáng lên』":       6,
		"Phản ứng cơ thể 『thót tim / rùng mình / toát mồ hôi』":               6,
		"Dấu vết suy nghĩ 『nghĩ rằng / nhận ra / cảm thấy』":                  6,
		"Khẩu hiệu sáo rỗng 『hóa ra là / ý nghĩa thực sự / có thể nói rằng』": 6,
	}
	matched := 0
	for _, p := range s.Patterns {
		w, ok := want[p.Name]
		if !ok {
			t.Errorf("unexpected pattern name %q", p.Name)
			continue
		}
		matched++
		if p.Total != w {
			t.Errorf("%s total: got %d want %d", p.Name, p.Total, w)
		}
		if p.PerChapter != 1.0 {
			t.Errorf("%s per_chapter: got %v want 1.0", p.Name, p.PerChapter)
		}
	}
	if matched != len(want) {
		t.Errorf("want %d pattern classes matched, got %d: %+v", len(want), matched, s.Patterns)
	}
}

func TestComputeTopPhrasesWithStopwords(t *testing.T) {
	// 「青云山巅」高频出现；「陆九渊」是角色名应被过滤
	line := "众人望向青云山巅，陆九渊负手而立。\n"
	chapters := make([]string, 10)
	for i := range chapters {
		chapters[i] = chapterWith(strings.Repeat(line, 3))
	}
	s := Compute(Input{Chapters: chapters, Stopwords: []string{"陆九渊"}})
	if s == nil {
		t.Fatal("expected stats")
	}
	var hasMountain, hasName bool
	for _, p := range s.TopPhrases {
		if strings.Contains(p.Text, "青云山") {
			hasMountain = true
		}
		if strings.Contains(p.Text, "九渊") || strings.Contains(p.Text, "陆九") {
			hasName = true
		}
	}
	if !hasMountain {
		t.Errorf("expected 青云山 phrase mined, got %+v", s.TopPhrases)
	}
	if hasName {
		t.Errorf("character name should be filtered, got %+v", s.TopPhrases)
	}
}

func TestComputeRepeatedSentences(t *testing.T) {
	motto := "此生未能远行，望你替我看看远方的山海。"
	chapters := make([]string, 6)
	for i := range chapters {
		body := "平常正文，没有什么重复。\n"
		if i%2 == 0 {
			body += motto + "\n"
		}
		chapters[i] = chapterWith(body)
	}
	s := Compute(Input{Chapters: chapters})
	if s == nil {
		t.Fatal("expected stats")
	}
	if len(s.RepeatedSentences) == 0 {
		t.Fatalf("expected repeated sentence, got none")
	}
	got := s.RepeatedSentences[0]
	if got.Chapters != 3 || got.Count != 3 {
		t.Errorf("repeated sentence: %+v", got)
	}
	if !strings.HasPrefix(got.Text, "此生未能远行") {
		t.Errorf("text: %q", got.Text)
	}
}

func TestComputeEndingAndOpening(t *testing.T) {
	short := chapterWith("一整夜没有睡。\n正文很长很长很长。\n他走了。")
	long := chapterWith("白天的事。\n正文。\n这是一个非常非常非常长的结尾句子，远远超过三十个字符的阈值长度，用来测试中位数。")
	chapters := []string{short, short, short, long, long}
	s := Compute(Input{Chapters: chapters})
	if s == nil {
		t.Fatal("expected stats")
	}
	if s.Ending.ShortRatio != 0.6 {
		t.Errorf("short_ratio: got %v want 0.6", s.Ending.ShortRatio)
	}
	if s.OpeningTimeRate != 0.6 {
		t.Errorf("opening_time_rate: got %v want 0.6", s.OpeningTimeRate)
	}
}

func TestComputeTitleFormats(t *testing.T) {
	chapters := make([]string, 5)
	for i := range chapters {
		chapters[i] = chapterWith("正文。")
	}
	// 混用 → 上报
	s := Compute(Input{Chapters: chapters, Titles: []string{"第一章 风起", "云涌", "第3章 雷动"}})
	if s.TitleFormats == nil || s.TitleFormats.WithPrefix != 2 || s.TitleFormats.WithoutPrefix != 1 {
		t.Errorf("title formats: %+v", s.TitleFormats)
	}
	// 统一 → 不上报
	s = Compute(Input{Chapters: chapters, Titles: []string{"风起", "云涌"}})
	if s.TitleFormats != nil {
		t.Errorf("uniform titles should not report: %+v", s.TitleFormats)
	}
}

func TestComputePatterns_Vietnamese(t *testing.T) {
	body := "Hắn nhận ra rằng điều này không phải tiền mặt, mà là vỏ sò. Sau 5 giây, hắn im lặng chẳng khác nào pho tượng. Mắt sáng lên, thót tim khi thấy một con khủng long. Hắn nghĩ rằng hóa ra là có thể nói rằng mọi chuyện đã an bài.\nNội dung chính văn ở đây.\n"
	chapters := make([]string, 6)
	for i := range chapters {
		chapters[i] = chapterWith(body)
	}
	s := Compute(Input{Chapters: chapters})
	if s == nil {
		t.Fatal("expected stats")
	}
	want := map[string]int{
		"Câu phủ định sửa sai 『không phải… mà là…』":                          6,
		"Lượng từ thời gian 『X giây / tích tắc』":                             6,
		"So sánh ví von 『giống như / tựa như / như thể』":                     6,
		"Nhịp im lặng 『im lặng / không nói gì』":                              6,
		"Khuôn mẫu thần thái 『mặt tái mét / nhếch mép / mắt sáng lên』":       6,
		"Phản ứng cơ thể 『thót tim / rùng mình / toát mồ hôi』":               6,
		"Dấu vết suy nghĩ 『nghĩ rằng / nhận ra / cảm thấy』":                  6,
		"Khẩu hiệu sáo rỗng 『hóa ra là / ý nghĩa thực sự / có thể nói rằng』": 6,
	}
	matched := 0
	for _, p := range s.Patterns {
		w, ok := want[p.Name]
		if !ok {
			t.Errorf("unexpected pattern name %q", p.Name)
			continue
		}
		matched++
		if p.Total < w {
			t.Errorf("%s total: got %d want at least %d", p.Name, p.Total, w)
		}
	}
	if matched != len(want) {
		t.Errorf("want %d pattern classes matched, got %d: %+v", len(want), matched, s.Patterns)
	}
}

func TestComputeTopPhrases_Vietnamese(t *testing.T) {
	// "vỏ sò biển" tần suất cao; "Que Bự" là tên nhân vật cần bị lọc bỏ
	line := "Người Que dùng vỏ sò biển để đổi lấy thức ăn, Que Bự đứng nhìn từ xa.\n"
	chapters := make([]string, 10)
	for i := range chapters {
		chapters[i] = chapterWith(strings.Repeat(line, 3))
	}
	s := Compute(Input{Chapters: chapters, Stopwords: []string{"Que Bự"}})
	if s == nil {
		t.Fatal("expected stats")
	}
	var hasShell, hasStopword bool
	for _, p := range s.TopPhrases {
		if strings.Contains(strings.ToLower(p.Text), "vỏ sò") {
			hasShell = true
		}
		if strings.Contains(strings.ToLower(p.Text), "que bự") {
			hasStopword = true
		}
	}
	if !hasShell {
		t.Errorf("expected 'vỏ sò' phrase mined, got %+v", s.TopPhrases)
	}
	if hasStopword {
		t.Errorf("character stopword should be filtered, got %+v", s.TopPhrases)
	}
}

func TestComputeRepeatedSentences_Vietnamese(t *testing.T) {
	motto := "Hóa ra người que thời đồ đá cũng biết dùng vỏ sò để tích lũy tài sản."
	chapters := make([]string, 6)
	for i := range chapters {
		body := ""
		if i%2 == 0 {
			body = motto + "\n"
		}
		chapters[i] = chapterWith(body)
	}
	s := Compute(Input{Chapters: chapters})
	if s == nil {
		t.Fatal("expected stats")
	}
	if len(s.RepeatedSentences) == 0 {
		t.Fatalf("expected repeated sentence, got none")
	}
	got := s.RepeatedSentences[0]
	if got.Chapters != 3 || got.Count != 3 {
		t.Errorf("repeated sentence: %+v", got)
	}
	if !strings.HasPrefix(got.Text, "Hóa ra người que") {
		t.Errorf("text: %q", got.Text)
	}
}

func TestComputeTitleFormats_Vietnamese(t *testing.T) {
	chapters := make([]string, 5)
	for i := range chapters {
		chapters[i] = chapterWith("Nội dung kịch bản.")
	}
	// Trộn Tập 1, Chương 2 và không tiền tố -> báo cáo
	s := Compute(Input{Chapters: chapters, Titles: []string{"Tập 1 Lạm phát đồ đá", "Thẻ tín dụng", "Tập 3 Bong bóng tài sản"}})
	if s.TitleFormats == nil || s.TitleFormats.WithPrefix != 2 || s.TitleFormats.WithoutPrefix != 1 {
		t.Errorf("title formats: %+v", s.TitleFormats)
	}
	// Đồng nhất Tập N -> không báo cáo
	s = Compute(Input{Chapters: chapters, Titles: []string{"Tập 1 Lạm phát", "Tập 2 Lãi suất"}})
	if s.TitleFormats != nil {
		t.Errorf("uniform titles should not report: %+v", s.TitleFormats)
	}
}
