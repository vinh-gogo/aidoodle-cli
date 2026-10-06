package domain

import "testing"

func TestWordCount(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		{"rỗng", "", 0},
		{"chỉ khoảng trắng", "  \n\t ", 0},
		{"tiếng Việt đếm theo âm tiết", "Xin chào các bạn", 4},
		{"dấu câu dính vào từ không tách", "Ê, nghe này!", 3},
		{"nhiều dòng và khoảng trắng thừa", "một  hai\nba\t\tbốn", 4},
		{"chữ có dấu tổ hợp vẫn là một token", "Việt Nam", 2},
		{"chữ Hán tính từng ký tự", "他发现了", 4},
		{"trộn Hán và Latin", "他 nhìn thấy pattern", 4},
	}
	for _, c := range cases {
		if got := WordCount(c.in); got != c.want {
			t.Errorf("%s: WordCount(%q) = %d, want %d", c.name, c.in, got, c.want)
		}
	}
}

// Số từ phải xấp xỉ số âm tiết, không phải số rune (lệch ~4–5 lần với tiếng Việt).
func TestWordCount_NotRuneCount(t *testing.T) {
	text := "Người tiền sử cũng biết đếm bằng que củi"
	if got := WordCount(text); got != 9 {
		t.Fatalf("got %d, want 9 (len runes = %d)", got, len([]rune(text)))
	}
}

func TestSpeechSeconds(t *testing.T) {
	cases := []struct {
		words int
		wps   float64
		want  int
	}{
		{0, 0, 0},
		{-5, 0, 0},
		{150, 0, 60},  // mặc định 2,5 từ/giây
		{450, 0, 180}, // đầu trên của video 60–180s
		{151, 0, 61},  // làm tròn lên
		{100, 4, 25},
	}
	for _, c := range cases {
		if got := SpeechSeconds(c.words, c.wps); got != c.want {
			t.Errorf("SpeechSeconds(%d, %v) = %d, want %d", c.words, c.wps, got, c.want)
		}
	}
}
