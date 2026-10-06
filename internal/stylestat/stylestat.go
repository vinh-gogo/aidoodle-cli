// Package stylestat 对已写正文做全书级风格统计，产出纯事实。
//
// 动机：弧内评审窗口（~10 章）对全书级模式固化天然失明——句式 tic 章均几十次、
// 章末形态同构、跨章复读，单章看每处都"正常"，只有全书统计能暴露。统计归代码
// （确定性、零幻觉），裁定归 LLM（editor 按数字判维度分，writer 据此自避免）。
// Compute 供离线评测一次性全量计算；运行时使用 Tracker 按章节增量维护。
package stylestat

import (
	"regexp"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"

	"github.com/voocel/ainovel-cli/internal/utils"
)

// minChapters 少于此章数不出统计——样本太小，频率没有意义。
const minChapters = 5

// phraseWindow 动态短语挖掘只看最近 N 章：writer 需要避免的是"现在的口头禅"。
const phraseWindow = 20

// Input 统计输入。Chapters 按章号升序；Stopwords 为角色名等专有名词，
// 动态短语挖掘时跳过（出场人名天然高频，不是文风问题）。
type Input struct {
	Chapters  []string
	Titles    []string
	Stopwords []string
}

// Stats 全书风格统计结果。所有字段都是事实计数，不含任何裁定或指令。
type Stats struct {
	Chapters          int            `json:"chapters"`
	Patterns          []PatternStat  `json:"patterns,omitempty"`
	TopPhrases        []PhraseStat   `json:"top_phrases,omitempty"`
	RepeatedSentences []SentenceStat `json:"repeated_sentences,omitempty"`
	Ending            EndingStat     `json:"ending"`
	OpeningTimeRate   float64        `json:"opening_time_rate"`
	TitleFormats      *TitleStat     `json:"title_formats,omitempty"`
}

// PatternStat 固定句式模式类的全书计数（通用 AI 文风 tic）。
type PatternStat struct {
	Name       string  `json:"name"`
	Total      int     `json:"total"`
	PerChapter float64 `json:"per_chapter"`
}

// PhraseStat 最近 phraseWindow 章内挖掘出的高频短语。
type PhraseStat struct {
	Text  string `json:"text"`
	Count int    `json:"count"`
}

// SentenceStat 跨章逐字重复的长句（复读交代的直接证据）。
type SentenceStat struct {
	Text     string `json:"text"`
	Chapters int    `json:"chapters"`
	Count    int    `json:"count"`
}

// EndingStat 章末行形态分布。短结尾本身合法，全书同构才是问题。
type EndingStat struct {
	ShortRatio  float64 `json:"short_ratio"`
	MedianRunes int     `json:"median_runes"`
}

// TitleStat 章节标题「第N章」前缀混用计数（混用=机制痕迹暴露在产物里）。
type TitleStat struct {
	WithPrefix    int `json:"with_prefix"`
	WithoutPrefix int `json:"without_prefix"`
}

// patternDefs Các mẫu câu sáo AI thường gặp (hỗ trợ cả tiếng Việt và tiếng Trung).
// Dùng để đối chiếu đường cơ sở nội bộ của tác phẩm, không làm phân tích cú pháp nghiêm ngặt.
var patternDefs = []struct {
	name string
	re   *regexp.Regexp
}{
	{"Câu phủ định sửa sai 『không phải… mà là…』", regexp.MustCompile(`(?i)(?:不是[^。！？\n]{1,24}?[，、]?(?:而)?是|không\s+phải[^.!?\n]{1,50}?[,;]?(?:mà)?\s*là)`)},
	{"Lượng từ thời gian 『X giây / tích tắc』", regexp.MustCompile(`(?i)(?:[一两二三四五六七八九十几数半][息瞬]|\b\d+\s*(?:giây|phút|tích\s+tắc|chớp\s+mắt)\b)`)},
	{"So sánh ví von 『giống như / tựa như / như thể』", regexp.MustCompile(`(?i)(?:像一|仿佛|如同|宛如|giống\s+như|tựa\s+như|chẳng\s+khác\s+nào|như\s+thể)`)},
	{"Nhịp im lặng 『im lặng / không nói gì』", regexp.MustCompile(`(?i)(?:沉默了|没有说话|没有回头|im\s+lặng|không\s+nói\s+gì|không\s+quay\s+đầu)`)},
	{"Khuôn mẫu thần thái 『mặt tái mét / nhếch mép / mắt sáng lên』", regexp.MustCompile(`(?i)(?:眼[中底]闪过|目光一凝|瞳孔一缩|眼眶微红|嘴角[微轻一]?[勾扬翘]|咬了咬唇|不可置信|mắt\s+sáng\s+lên|nhếch\s+mép|cắn\s+môi|không\s+thể\s+tin\s+được|mặt\s+tái\s+mét)`)},
	{"Phản ứng cơ thể 『thót tim / rùng mình / toát mồ hôi』", regexp.MustCompile(`(?i)(?:心头一[紧沉颤]|身子一[颤震僵]|倒吸(?:了)?一口凉气|thót\s+tim|rùng\s+mình|hít\s+sâu\s+một\s+hơi|toát\s+mồ\s+hôi)`)},
	{"Dấu vết suy nghĩ 『nghĩ rằng / nhận ra / cảm thấy』", regexp.MustCompile(`(?i)(?:心想|意识到|感到|觉得|nghĩ\s+rằng|nhận\s+ra|cảm\s+thấy|cho\s+rằng)`)},
	{"Khẩu hiệu sáo rỗng 『hóa ra là / ý nghĩa thực sự / có thể nói rằng』", regexp.MustCompile(`(?i)(?:一种说不出的|说不清[的道]|的意义在于|真正的[^。！？\n]{1,10}是|ý\s+nghĩa\s+thực\s+sự|bài\s+học\s+rút\s+ra|hóa\s+ra\s+là|nói\s+cách\s+khác|có\s+thể\s+nói\s+rằng)`)},
}

var (
	sentenceSplit = regexp.MustCompile(`[.!?。！？\n]+`)
	openingTimeRe = regexp.MustCompile(`(?i)(?:夜|清晨|黎明|天亮|醒来|晨光|一整夜|hôm\s+qua|hôm\s+nay|sáng\s+nay|ngày\s+xưa|thời\s+đồ\s+đá|vừa\s+qua|mới\s+đây|ban\s+đêm)`)
	titlePrefixRe = regexp.MustCompile(`^(?:#{0,2}\s*第[零〇一二三四五六七八九十百千万\d]+章|#{0,2}\s*[Cc]hương\s+\d+|#{0,2}\s*[Tt]ập\s+\d+)`)
)

// shortEndingRunes 末行不超过此字数计为"短结尾"。
const shortEndingRunes = 30

// Compute 计算全书风格统计；章数不足时返回 nil。
func Compute(in Input) *Stats {
	n := len(in.Chapters)
	if n < minChapters {
		return nil
	}
	all := strings.Join(in.Chapters, "\n")

	s := &Stats{Chapters: n}
	for _, def := range patternDefs {
		total := len(def.re.FindAllStringIndex(all, -1))
		if total == 0 {
			continue
		}
		s.Patterns = append(s.Patterns, PatternStat{
			Name:       def.name,
			Total:      total,
			PerChapter: round1(float64(total) / float64(n)),
		})
	}
	s.TopPhrases = minePhrases(recentWindow(in.Chapters), in.Stopwords)
	s.RepeatedSentences = repeatedSentences(in.Chapters)
	s.Ending = endingShape(in.Chapters)
	s.OpeningTimeRate = openingTimeRate(in.Chapters)
	s.TitleFormats = titleFormats(in.Titles)
	return s
}

func recentWindow(chapters []string) []string {
	if len(chapters) <= phraseWindow {
		return chapters
	}
	return chapters[len(chapters)-phraseWindow:]
}

// minePhrases khai phá cụm từ lặp tần suất cao trong cửa sổ gần nhất.
// Hỗ trợ cả cụm từ tiếng Việt (2-4 từ) và n-gram chữ Hán (3-6 ký tự).
// Lọc: bỏ dấu câu, bỏ từ đệm ở đầu/cuối, loại bỏ tên riêng/stopwords; khử trùng: cụm con của cụm đã chọn bị bỏ.
func minePhrases(chapters []string, stopwords []string) []PhraseStat {
	text := strings.Join(chapters, "\n")
	threshold := max(8, len(chapters)/2)
	counts := make(map[string]int)

	// 1. Cụm từ tiếng Việt / Latin: tách câu rồi khai phá 2-4 từ
	for _, sentence := range sentenceSplit.Split(text, -1) {
		sentence = trimWrappedQuotes(sentence)
		words := extractWords(sentence)
		for size := 2; size <= 4; size++ {
			for i := 0; i+size <= len(words); i++ {
				gram := words[i : i+size]
				if !validWordGram(gram) {
					continue
				}
				phrase := strings.Join(gram, " ")
				counts[phrase]++
			}
		}
	}

	// 2. Chữ Hán n-gram (3-6 ký tự) để tương thích ngược với dữ liệu tiếng Trung
	runes := []rune(text)
	for size := 3; size <= 6; size++ {
		for i := 0; i+size <= len(runes); i++ {
			gram := runes[i : i+size]
			if !validHanGram(gram) {
				continue
			}
			counts[string(gram)]++
		}
	}

	stopGrams := stopwordBigrams(stopwords)
	type cand struct {
		text  string
		count int
	}
	var cands []cand
	for g, c := range counts {
		if c < threshold || hitStopword(g, stopGrams) {
			continue
		}
		cands = append(cands, cand{g, c})
	}
	sort.Slice(cands, func(i, j int) bool {
		if cands[i].count != cands[j].count {
			return cands[i].count > cands[j].count
		}
		// Đồng tần suất ưu tiên cụm dài hơn, sau đó sắp xếp theo từ điển
		if len([]rune(cands[i].text)) != len([]rune(cands[j].text)) {
			return len([]rune(cands[i].text)) > len([]rune(cands[j].text))
		}
		return cands[i].text < cands[j].text
	})

	var out []PhraseStat
	for _, c := range cands {
		if len(out) >= 8 {
			break
		}
		dup := false
		lowerC := strings.ToLower(c.text)
		for _, picked := range out {
			lowerP := strings.ToLower(picked.Text)
			if strings.Contains(lowerP, lowerC) || strings.Contains(lowerC, lowerP) {
				dup = true
				break
			}
		}
		if !dup {
			out = append(out, PhraseStat{Text: c.text, Count: c.count})
		}
	}
	return out
}

func extractWords(s string) []string {
	var words []string
	var cur []rune
	for _, r := range norm.NFC.String(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur = append(cur, unicode.ToLower(r))
		} else if len(cur) > 0 {
			words = append(words, string(cur))
			cur = cur[:0]
		}
	}
	if len(cur) > 0 {
		words = append(words, string(cur))
	}
	return words
}

var vnEdgeStops = map[string]struct{}{
	"và": {}, "của": {}, "là": {}, "thì": {}, "mà": {}, "ở": {}, "trong": {},
	"được": {}, "bị": {}, "cho": {}, "với": {}, "các": {}, "những": {}, "cái": {},
	"một": {}, "này": {}, "đó": {}, "ra": {}, "vào": {}, "lại": {}, "đã": {},
	"đang": {}, "sẽ": {}, "có": {}, "không": {}, "thế": {}, "nhưng": {}, "rồi": {},
	"khi": {}, "từ": {}, "tui": {}, "tôi": {}, "ai": {}, "đến": {}, "về": {},
	"đây": {}, "đấy": {}, "nào": {}, "gì": {}, "sao": {}, "như": {}, "thêm": {},
}

func validWordGram(gram []string) bool {
	if len(gram) == 0 {
		return false
	}
	allHan := true
	for _, w := range gram {
		for _, r := range w {
			if r < 0x4E00 || r > 0x9FFF {
				allHan = false
				break
			}
		}
	}
	if allHan {
		return false
	}
	if _, bad := vnEdgeStops[gram[0]]; bad {
		return false
	}
	if _, bad := vnEdgeStops[gram[len(gram)-1]]; bad {
		return false
	}
	return true
}

// gramEdgeStop 首尾为这些虚词/代词的 n-gram 不是文风短语，跳过。
const gramEdgeStop = "的了着是在和与就也都还又把被他她它我你这那"

func validHanGram(gram []rune) bool {
	for _, r := range gram {
		if r < 0x4E00 || r > 0x9FFF { // 仅纯汉字片段
			return false
		}
	}
	if strings.ContainsRune(gramEdgeStop, gram[0]) || strings.ContainsRune(gramEdgeStop, gram[len(gram)-1]) {
		return false
	}
	return true
}

// stopwordBigrams 把专有名词拆成 2 字片段：人名常以部分形式入文
// （"九渊负手"含"九渊"），按整名匹配会漏网。宁可过滤偏严——短语事实少一条
// 无碍，人名混进口头禅清单才是噪声。
func stopwordBigrams(stopwords []string) []string {
	var grams []string
	for _, w := range stopwords {
		w = strings.ToLower(strings.TrimSpace(w))
		if w == "" {
			continue
		}
		grams = append(grams, w)
		runes := []rune(w)
		if len(runes) >= 2 {
			for i := 0; i+2 <= len(runes); i++ {
				grams = append(grams, string(runes[i:i+2]))
			}
		}
	}
	return grams
}

func hitStopword(gram string, stopGrams []string) bool {
	lower := strings.ToLower(gram)
	for _, g := range stopGrams {
		if strings.Contains(lower, g) {
			return true
		}
	}
	return false
}

// repeatedSentences 找跨 ≥3 章逐字重复的 ≥12 字句子，按次数取 top 5。
func repeatedSentences(chapters []string) []SentenceStat {
	type rec struct {
		count    int
		chapters map[int]struct{}
	}
	seen := make(map[string]*rec)
	for ci, text := range chapters {
		for sent, count := range chapterSentenceCounts(text) {
			r := seen[sent]
			if r == nil {
				r = &rec{chapters: make(map[int]struct{})}
				seen[sent] = r
			}
			r.count += count
			r.chapters[ci] = struct{}{}
		}
	}

	var out []SentenceStat
	for sent, r := range seen {
		if len(r.chapters) < 3 {
			continue
		}
		out = append(out, SentenceStat{Text: utils.TruncateRunes(sent, 40), Chapters: len(r.chapters), Count: r.count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Text < out[j].Text
	})
	if len(out) > 5 {
		out = out[:5]
	}
	return out
}

// trimWrappedQuotes 剥掉包裹引号：同一句台词带/不带前引号不应算成两条。
func trimWrappedQuotes(sentence string) string {
	return strings.Trim(strings.TrimSpace(sentence), `"“”‘’「」『』`)
}

func endingShape(chapters []string) EndingStat {
	var lengths []int
	short := 0
	for _, text := range chapters {
		line := lastNonEmptyLine(text)
		if line == "" {
			continue
		}
		n := len([]rune(line))
		lengths = append(lengths, n)
		if n <= shortEndingRunes {
			short++
		}
	}
	if len(lengths) == 0 {
		return EndingStat{}
	}
	sort.Ints(lengths)
	return EndingStat{
		ShortRatio:  round2(float64(short) / float64(len(lengths))),
		MedianRunes: lengths[len(lengths)/2],
	}
}

func openingTimeRate(chapters []string) float64 {
	hit := 0
	for _, text := range chapters {
		if openingTimeRe.MatchString(firstParagraph(text)) {
			hit++
		}
	}
	return round2(float64(hit) / float64(len(chapters)))
}

func titleFormats(titles []string) *TitleStat {
	if len(titles) == 0 {
		return nil
	}
	t := &TitleStat{}
	for _, title := range titles {
		if strings.TrimSpace(title) == "" {
			continue
		}
		if titlePrefixRe.MatchString(title) {
			t.WithPrefix++
		} else {
			t.WithoutPrefix++
		}
	}
	// 只有混用才值得上报；统一格式不是事实意义上的问题
	if t.WithPrefix == 0 || t.WithoutPrefix == 0 {
		return nil
	}
	return t
}

func lastNonEmptyLine(text string) string {
	lines := strings.Split(text, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if line := strings.TrimSpace(lines[i]); line != "" {
			return line
		}
	}
	return ""
}

// firstParagraph 取第一个非空且非 Markdown 标题的行（章文件首行常是 # 标题）。
func firstParagraph(text string) string {
	for line := range strings.SplitSeq(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line
	}
	return ""
}

func round1(f float64) float64 { return float64(int(f*10+0.5)) / 10 }
func round2(f float64) float64 { return float64(int(f*100+0.5)) / 100 }
