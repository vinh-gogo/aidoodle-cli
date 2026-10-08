package web

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/store"
)

// ProjectItem lưu trữ thông tin tóm tắt của một dự án trong thư mục output.
type ProjectItem struct {
	DirName      string    `json:"dir_name"`
	RelPath      string    `json:"rel_path"`
	FullPath     string    `json:"full_path"`
	Title        string    `json:"title"`
	Prompt       string    `json:"prompt"`
	Style        string    `json:"style"`
	StyleBadge   string    `json:"style_badge"`
	ProgressText string    `json:"progress_text"`
	Phase        string    `json:"phase"`
	TotalCh      int       `json:"total_ch"`
	CompletedCh  int       `json:"completed_ch"`
	ModTime      time.Time `json:"mod_time"`
	IsCurrent    bool      `json:"is_current"`
}

// ChapterItem lưu trữ thông tin về một chương trong dự án.
type ChapterItem struct {
	Chapter   int    `json:"chapter"`
	Title     string `json:"title"`
	CoreEvent string `json:"core_event"`
	WordCount int    `json:"word_count"`
	Completed bool   `json:"completed"`
	Path      string `json:"path"`
}

var timestampSuffixRe = regexp.MustCompile(`^(.+)-\d{8}-\d{4}(?:-\d+)?$`)

// NextOutputDir tạo đường dẫn thư mục mới theo kiểu output/novel-YYYYMMDD-HHMM.
func NextOutputDir(currentDir string, now time.Time) string {
	if currentDir == "" {
		currentDir = filepath.Join("output", "novel")
	}
	parent := filepath.Dir(currentDir)
	base := filepath.Base(currentDir)

	prefix := base
	if m := timestampSuffixRe.FindStringSubmatch(base); len(m) > 1 {
		prefix = m[1]
	}

	ts := now.Format("20060102-1504")
	candidate := filepath.Join(parent, fmt.Sprintf("%s-%s", prefix, ts))
	if _, err := os.Stat(candidate); os.IsNotExist(err) {
		return candidate
	}

	for i := 1; i <= 99; i++ {
		alt := filepath.Join(parent, fmt.Sprintf("%s-%s-%02d", prefix, ts, i))
		if _, err := os.Stat(alt); os.IsNotExist(err) {
			return alt
		}
	}
	return filepath.Join(parent, fmt.Sprintf("%s-%s-%s", prefix, ts, now.Format("05")))
}

// DetectStyleBadge tạo nhãn hiển thị phong cách trực quan cho từng mode.
func DetectStyleBadge(style string) string {
	switch strings.ToLower(style) {
	case "behavioral-psychology", "psychology":
		return "[🧠 Tâm lý học hành vi]"
	case "doodle-explainer":
		return "[🦴 Video TikTok]"
	case "vietnamese-history":
		return "[⚔️ Lịch sử Việt Nam]"
	case "novel-manga", "novel", "manga", "default", "":
		return "[📖 Tiểu thuyết / Manga]"
	default:
		return fmt.Sprintf("[🏷️ %s]", style)
	}
}

// DetectProjectStyle trích xuất style của một dự án từ đĩa.
func DetectProjectStyle(dir string) string {
	// 1. Thử đọc meta/run.json
	runJSONPath := filepath.Join(dir, "meta", "run.json")
	if data, err := os.ReadFile(runJSONPath); err == nil {
		var meta struct {
			Style string `json:"style"`
		}
		if err := json.Unmarshal(data, &meta); err == nil && strings.TrimSpace(meta.Style) != "" {
			return strings.TrimSpace(meta.Style)
		}
	}

	// 2. Thử đọc meta/user_rules.json
	userRulesPath := filepath.Join(dir, "meta", "user_rules.json")
	if data, err := os.ReadFile(userRulesPath); err == nil {
		var r struct {
			Structured struct {
				Genre string `json:"genre"`
			} `json:"structured"`
		}
		if err := json.Unmarshal(data, &r); err == nil && strings.TrimSpace(r.Structured.Genre) != "" {
			return strings.TrimSpace(r.Structured.Genre)
		}
	}

	return ""
}

// ScanProjects quét toàn bộ các thư mục dự án trong output/ và trích xuất siêu dữ liệu.
func ScanProjects(baseDir, currentDir string) []ProjectItem {
	if baseDir == "" {
		baseDir = "output"
		if currentDir != "" {
			parent := filepath.Dir(currentDir)
			if parent != "" && parent != "." {
				baseDir = parent
			}
		}
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil && baseDir != "output" {
		baseDir = "output"
		entries, err = os.ReadDir(baseDir)
	}
	if err != nil {
		return nil
	}

	currentClean := ""
	if currentDir != "" {
		currentClean = filepath.Clean(currentDir)
	}

	var items []ProjectItem
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := entry.Name()
		projectPath := filepath.Join(baseDir, dirName)

		metaDir := filepath.Join(projectPath, "meta")
		if !strings.HasPrefix(dirName, "novel") {
			if fi, err := os.Stat(metaDir); err != nil || !fi.IsDir() {
				continue
			}
		}

		info, err := entry.Info()
		modTime := time.Now()
		if err == nil {
			modTime = info.ModTime()
		}

		item := ProjectItem{
			DirName:  dirName,
			RelPath:  filepath.Join("output", dirName),
			FullPath: projectPath,
			ModTime:  modTime,
		}

		if currentClean != "" && filepath.Clean(projectPath) == currentClean {
			item.IsCurrent = true
		}

		// 1. Đọc meta/run.json
		runJSONPath := filepath.Join(projectPath, "meta", "run.json")
		var runMeta struct {
			StartedAt   string `json:"started_at"`
			Style       string `json:"style"`
			StartPrompt string `json:"start_prompt"`
		}
		if data, err := os.ReadFile(runJSONPath); err == nil {
			if err := json.Unmarshal(data, &runMeta); err == nil {
				item.Style = runMeta.Style
				item.Prompt = runMeta.StartPrompt
				if t, err := time.Parse(time.RFC3339, runMeta.StartedAt); err == nil {
					item.ModTime = t
				}
			}
		}
		if item.Style == "" {
			item.Style = DetectProjectStyle(projectPath)
		}
		item.StyleBadge = DetectStyleBadge(item.Style)

		// 2. Tìm Title
		bookJSONPath := filepath.Join(projectPath, "meta", "book.json")
		if data, err := os.ReadFile(bookJSONPath); err == nil {
			var b struct {
				Title string `json:"title"`
			}
			if err := json.Unmarshal(data, &b); err == nil && strings.TrimSpace(b.Title) != "" {
				item.Title = strings.TrimSpace(b.Title)
			}
		}
		if item.Title == "" {
			if data, err := os.ReadFile(filepath.Join(projectPath, "book.json")); err == nil {
				var b struct {
					Title string `json:"title"`
				}
				if err := json.Unmarshal(data, &b); err == nil && strings.TrimSpace(b.Title) != "" {
					item.Title = strings.TrimSpace(b.Title)
				}
			}
		}
		if item.Title == "" && item.Prompt != "" {
			p := strings.ReplaceAll(item.Prompt, "\n", " ")
			p = strings.TrimSpace(p)
			if len([]rune(p)) > 40 {
				p = string([]rune(p)[:40]) + "..."
			}
			item.Title = p
		}
		if item.Title == "" {
			item.Title = dirName
		}

		// 3. Tiến độ từ meta/progress.json
		progJSONPath := filepath.Join(projectPath, "meta", "progress.json")
		if data, err := os.ReadFile(progJSONPath); err == nil {
			var p struct {
				Phase             string `json:"phase"`
				CurrentChapter    int    `json:"current_chapter"`
				CompletedChapters []int  `json:"completed_chapters"`
			}
			if err := json.Unmarshal(data, &p); err == nil {
				item.Phase = p.Phase
				item.CompletedCh = len(p.CompletedChapters)
			}
		}

		// 4. Tổng số chương từ meta/outline.json
		outlineJSONPath := filepath.Join(projectPath, "meta", "outline.json")
		if data, err := os.ReadFile(outlineJSONPath); err == nil {
			var outline []struct {
				Chapter int `json:"chapter"`
			}
			if err := json.Unmarshal(data, &outline); err == nil {
				item.TotalCh = len(outline)
			}
		}

		if item.Phase == string(domain.PhaseComplete) || (item.TotalCh > 0 && item.CompletedCh >= item.TotalCh) {
			item.ProgressText = fmt.Sprintf("✓ Đã xong (%d/%d tập)", item.CompletedCh, item.TotalCh)
		} else if item.TotalCh > 0 {
			item.ProgressText = fmt.Sprintf("⏳ Đang viết (%d/%d tập)", item.CompletedCh, item.TotalCh)
		} else if item.CompletedCh > 0 {
			item.ProgressText = fmt.Sprintf("⏳ Đã viết %d chương", item.CompletedCh)
		} else {
			item.ProgressText = "🌱 Khởi tạo"
		}

		items = append(items, item)
	}

	// Sắp xếp thời gian mới nhất lên đầu
	sort.Slice(items, func(i, j int) bool {
		return items[i].ModTime.After(items[j].ModTime)
	})

	return items
}

// ListProjectChapters đọc danh sách các chương của dự án tại dir.
func ListProjectChapters(dir string) ([]ChapterItem, error) {
	st := store.NewStore(dir)

	outline, _ := st.Outline.LoadOutline()
	progress, _ := st.Progress.Load()

	completedMap := make(map[int]bool)
	wordCounts := make(map[int]int)
	if progress != nil {
		for _, ch := range progress.CompletedChapters {
			completedMap[ch] = true
		}
		for ch, wc := range progress.ChapterWordCounts {
			wordCounts[ch] = wc
		}
	}

	titleMap := make(map[int]string)
	coreEventMap := make(map[int]string)
	for _, entry := range outline {
		titleMap[entry.Chapter] = entry.Title
		coreEventMap[entry.Chapter] = entry.CoreEvent
	}

	// Đọc thêm summary title nếu có
	for ch := range completedMap {
		if summaryTitle, err := st.Summaries.LoadSummaryTitle(ch); err == nil && strings.TrimSpace(summaryTitle) != "" {
			titleMap[ch] = summaryTitle
		}
	}

	// Thu thập các chương
	var allChapters []int
	seen := make(map[int]bool)

	for _, entry := range outline {
		if !seen[entry.Chapter] {
			seen[entry.Chapter] = true
			allChapters = append(allChapters, entry.Chapter)
		}
	}
	for ch := range completedMap {
		if !seen[ch] {
			seen[ch] = true
			allChapters = append(allChapters, ch)
		}
	}

	// Quét các file trong chapters/ nếu chưa có trong outline
	chaptersDir := filepath.Join(dir, "chapters")
	if files, err := os.ReadDir(chaptersDir); err == nil {
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".md") {
				var chNum int
				if n, _ := fmt.Sscanf(f.Name(), "%02d.md", &chNum); n == 1 && chNum > 0 {
					if !seen[chNum] {
						seen[chNum] = true
						allChapters = append(allChapters, chNum)
					}
				}
			}
		}
	}

	sort.Ints(allChapters)

	var items []ChapterItem
	for _, ch := range allChapters {
		title := titleMap[ch]
		if title == "" {
			title = fmt.Sprintf("Chương %02d", ch)
		}
		item := ChapterItem{
			Chapter:   ch,
			Title:     title,
			CoreEvent: coreEventMap[ch],
			WordCount: wordCounts[ch],
			Completed: completedMap[ch],
			Path:      filepath.Join(chaptersDir, fmt.Sprintf("%02d.md", ch)),
		}
		items = append(items, item)
	}

	return items, nil
}

// ReadProjectChapter đọc nội dung file markdown của chương ch trong dự án dir.
func ReadProjectChapter(dir string, ch int) (string, error) {
	st := store.NewStore(dir)

	// Thử đọc từ st.Drafts
	text, err := st.Drafts.LoadChapterText(ch)
	if err == nil && strings.TrimSpace(text) != "" {
		return text, nil
	}

	// Thử đọc file chapters/%02d.md trực tiếp
	chPath := filepath.Join(dir, "chapters", fmt.Sprintf("%02d.md", ch))
	if data, err := os.ReadFile(chPath); err == nil && len(data) > 0 {
		return string(data), nil
	}

	// Thử tìm trong các thư mục video scripts
	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		if entry.IsDir() && strings.HasSuffix(entry.Name(), "-video") {
			scriptsDir := filepath.Join(dir, entry.Name(), "scripts")
			sFiles, _ := os.ReadDir(scriptsDir)
			for _, sf := range sFiles {
				prefix := fmt.Sprintf("%02d-", ch)
				if strings.HasPrefix(sf.Name(), prefix) && strings.HasSuffix(sf.Name(), ".md") {
					if data, err := os.ReadFile(filepath.Join(scriptsDir, sf.Name())); err == nil {
						return string(data), nil
					}
				}
			}
		}
	}

	// Thử đọc bản nháp draft
	draft, draftErr := st.Drafts.LoadDraft(ch)
	if draftErr == nil && strings.TrimSpace(draft) != "" {
		return draft, nil
	}

	return "", fmt.Errorf("chưa có nội dung cho chương %d", ch)
}
