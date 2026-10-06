package domain

import (
	"fmt"
	"math"
	"unicode"
)

// ReviewInterval 全局审阅间隔（每 N 章触发一次）。
const ReviewInterval = 5

// ShouldReview 根据已完成章节数判断是否需要全局审阅（短篇/中篇模式）。
func ShouldReview(completedCount int) (bool, string) {
	if completedCount > 0 && completedCount%ReviewInterval == 0 {
		return true, fmt.Sprintf("已完成 %d 章，触发全局审阅", completedCount)
	}
	return false, ""
}

// ShouldArcReview 长篇模式下判断是否需要弧级/卷级评审。
func ShouldArcReview(isArcEnd, isVolumeEnd bool, volume, arc int) (bool, string) {
	if isVolumeEnd {
		return true, fmt.Sprintf("第 %d 卷第 %d 弧结束（卷结束），触发弧级+卷级评审", volume, arc)
	}
	if isArcEnd {
		return true, fmt.Sprintf("第 %d 卷第 %d 弧结束，触发弧级评审", volume, arc)
	}
	return false, ""
}

// WordCount đếm số "từ" của văn bản.
//
// Tiếng Việt viết tách từng âm tiết bằng khoảng trắng, nên mỗi token ngăn cách bởi
// khoảng trắng tính là 1 từ (đúng với cách đo lời đọc: số âm tiết). Đếm theo rune như
// trước sẽ cộng cả dấu cách và dấu câu, làm số liệu lệch khoảng 4–5 lần.
// Chữ Hán vẫn tính mỗi ký tự 1 từ để nội dung cũ bằng tiếng Trung không bị đếm thiếu.
func WordCount(content string) int {
	count := 0
	inToken := false
	for _, r := range content {
		switch {
		case unicode.IsSpace(r):
			inToken = false
		case unicode.Is(unicode.Han, r):
			count++
			inToken = false
		default:
			if !inToken {
				count++
				inToken = true
			}
		}
	}
	return count
}

// DefaultWordsPerSecond là tốc độ đọc lời dẫn mặc định (từ/giây). 2,5 từ/giây khớp với
// khoảng 150–450 từ cho video 60–180 giây của phong cách doodle explainer, đã gồm cả
// khoảng ngắt cho hình vẽ. Con số này là ước lượng, cần hiệu chỉnh bằng bản đọc thật.
const DefaultWordsPerSecond = 2.5

// SpeechSeconds ước lượng thời lượng đọc (giây, làm tròn lên) của số từ cho trước.
// wordsPerSecond <= 0 dùng DefaultWordsPerSecond.
func SpeechSeconds(words int, wordsPerSecond float64) int {
	if words <= 0 {
		return 0
	}
	if wordsPerSecond <= 0 {
		wordsPerSecond = DefaultWordsPerSecond
	}
	return int(math.Ceil(float64(words) / wordsPerSecond))
}
