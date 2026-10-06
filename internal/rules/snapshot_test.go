package rules

import (
	"strings"
	"testing"

	"golang.org/x/text/unicode/norm"
)

func TestBuildSnapshot_FieldOverridePrecedence(t *testing.T) {
	// 低→高：defaults 设 修仙，project 覆盖为 都市；高优先级胜出。
	snap := BuildSnapshot([]Candidate{
		{Source: "system_defaults", Structured: Structured{Genre: "修仙"}},
		{Source: "project:a.md", Structured: Structured{Genre: "都市"}},
	})
	if snap.Structured.Genre != "都市" {
		t.Fatalf("期望 project 覆盖 defaults，得到 %q", snap.Structured.Genre)
	}
	if snap.Status != StatusReady {
		t.Fatalf("期望 ready，得到 %s", snap.Status)
	}
	if snap.Version != SnapshotVersion {
		t.Fatalf("version 应为 %d，得到 %d", SnapshotVersion, snap.Version)
	}
}

func TestBuildSnapshot_EmptyAndZeroAreAbsent(t *testing.T) {
	// 归一化器吐占位：genre:""、空串元素——都必须当缺失，不覆盖低优先级真值。
	snap := BuildSnapshot([]Candidate{
		{Source: "system_defaults", Structured: Structured{
			Genre: "修仙",
		}},
		{Source: "startup_prompt", Structured: Structured{
			Genre:            "",                 // 占位空串 → 不覆盖
			ForbiddenPhrases: []string{"", "  "}, // 全空 → 丢弃
		}},
	})
	if snap.Structured.Genre != "修仙" {
		t.Fatalf("空 genre 不应覆盖，期望 修仙，得到 %q", snap.Structured.Genre)
	}
	if len(snap.Structured.ForbiddenPhrases) != 0 {
		t.Fatalf("全空 forbidden_phrases 应被丢弃，得到 %v", snap.Structured.ForbiddenPhrases)
	}
}

func TestBuildSnapshot_PreferencesPrecedenceOrder(t *testing.T) {
	snap := BuildSnapshot([]Candidate{
		{Source: "global:g.md", Preferences: "全局偏好"},
		{Source: "project:p.md", Preferences: "项目偏好"},
	})
	gi := strings.Index(snap.Preferences, "全局偏好")
	pi := strings.Index(snap.Preferences, "项目偏好")
	if gi < 0 || pi < 0 || gi > pi {
		t.Fatalf("preferences 应按优先级低→高拼接（项目在后），得到:\n%s", snap.Preferences)
	}
	if !strings.Contains(snap.Preferences, "## [global:g.md]") {
		t.Fatalf("preferences 应带来源标题，得到:\n%s", snap.Preferences)
	}
}

func TestBuildSnapshot_FatigueWordsMergeByWord(t *testing.T) {
	snap := BuildSnapshot([]Candidate{
		{Source: "system_defaults", Structured: Structured{FatigueWords: map[string]int{"竟然": 1, "仿佛": 2}}},
		{Source: "project:p.md", Structured: Structured{FatigueWords: map[string]int{"仿佛": 5}}},
	})
	if snap.Structured.FatigueWords["竟然"] != 1 {
		t.Fatalf("竟然 应保留 defaults 阈值 1，得到 %d", snap.Structured.FatigueWords["竟然"])
	}
	if snap.Structured.FatigueWords["仿佛"] != 5 {
		t.Fatalf("仿佛 应被 project 覆盖为 5，得到 %d", snap.Structured.FatigueWords["仿佛"])
	}
}

func TestBuildSnapshot_DegradedPropagates(t *testing.T) {
	snap := BuildSnapshot([]Candidate{
		{Source: "system_defaults", Structured: Structured{FatigueWords: map[string]int{"竟然": 1}}},
		{Source: "project:bad.md", Preferences: "原文降级", Degraded: true},
	})
	if snap.Status != StatusDegraded {
		t.Fatalf("任一来源降级则 status=degraded，得到 %s", snap.Status)
	}
	// 降级来源仍以 raw preferences 进入，不阻断；其它来源 structured 照常。
	if len(snap.Structured.FatigueWords) == 0 {
		t.Fatalf("降级不应影响其它来源的 structured")
	}
	if !strings.Contains(snap.Preferences, "原文降级") {
		t.Fatalf("降级来源应作为 raw preferences 保留")
	}
}

func TestSystemDefaults_VietnameseBaseline(t *testing.T) {
	d := SystemDefaults().Structured
	if len(d.ForbiddenPhrases) != 6 {
		t.Fatalf("mặc định cần 6 cụm cấm, có %d", len(d.ForbiddenPhrases))
	}
	if len(d.FatigueWords) != 16 {
		t.Fatalf("mặc định cần 16 từ mệt mỏi, có %d", len(d.FatigueWords))
	}
	var all []string
	all = append(all, d.ForbiddenPhrases...)
	for w, limit := range d.FatigueWords {
		all = append(all, w)
		if limit <= 0 {
			t.Errorf("ngưỡng của %q phải > 0, có %d", w, limit)
		}
	}
	for _, s := range all {
		if s != strings.ToLower(s) {
			t.Errorf("%q phải viết thường (Check so khớp không phân biệt hoa/thường)", s)
		}
		if hanRe.MatchString(s) {
			t.Errorf("%q còn chữ Hán: đường cơ sở phải là tiếng Việt", s)
		}
	}
}

// Đường cơ sở phải thực sự bắt được văn sáo tiếng Việt, kể cả chữ hoa đầu câu và
// chữ có dấu ở dạng tổ hợp (NFD).
func TestSystemDefaults_CatchesVietnameseText(t *testing.T) {
	s := SystemDefaults().Structured

	vs := Check("Có thể nói rằng người tiền sử đã đúng.", s)
	if findViolation(vs, "forbidden_phrases", "có thể nói rằng") == nil {
		t.Errorf("chữ hoa đầu câu phải bị bắt: %+v", vs)
	}

	nfd := norm.NFD.String("hãy cùng khám phá nhé")
	if nfd == "hãy cùng khám phá nhé" {
		t.Fatal("test cần chuỗi NFD khác chuỗi NFC")
	}
	if findViolation(Check(nfd, s), "forbidden_phrases", "hãy cùng khám phá") == nil {
		t.Error("chuỗi NFD phải khớp sau khi chuẩn hóa")
	}

	over := Check("Thực sự thì thực sự rất vui, thực sự đấy.", s)
	v := findViolation(over, "fatigue_words", "thực sự")
	if v == nil || v.Actual != 3 || v.Limit != 2 {
		t.Errorf("thực sự x3 vượt ngưỡng 2: %+v", over)
	}

	if vs := Check("Ông Gậy vẽ một vòng tròn lên vách đá rồi cười.", s); len(vs) != 0 {
		t.Errorf("câu sạch không được bị bắt: %+v", vs)
	}
}
