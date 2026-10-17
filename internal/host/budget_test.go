package host

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/voocel/agentcore"
	"github.com/voocel/ainovel-cli/internal/bootstrap"
)

type budgetRecorder struct {
	cost    float64
	aborts  []string
	reports []string
}

func (r *budgetRecorder) sentinel(cfg bootstrap.BudgetConfig) *BudgetSentinel {
	return NewBudgetSentinel(cfg,
		func() float64 { return r.cost },
		func(reason string) { r.aborts = append(r.aborts, reason) },
		func(level, summary string) { r.reports = append(r.reports, level+": "+summary) },
	)
}

func subagentEndEvent() agentcore.Event {
	return agentcore.Event{Type: agentcore.EventToolExecEnd, Tool: "subagent"}
}

func TestBudgetSentinelDisabled(t *testing.T) {
	r := &budgetRecorder{}
	if s := r.sentinel(bootstrap.BudgetConfig{MaxCalls: -1}); s != nil {
		t.Fatal("disabled budget should return nil sentinel")
	}
	// nil 安全
	var s *BudgetSentinel
	s.OnCost(100)
	s.HandleEvent(subagentEndEvent())
	if err := s.Refuse(); err != nil {
		t.Errorf("nil sentinel Refuse should pass: %v", err)
	}
	if s.Limit() != 0 {
		t.Error("nil sentinel Limit should be 0")
	}
}

func TestBudgetSentinelDefaultZeroDollar100Calls(t *testing.T) {
	r := &budgetRecorder{}
	s := r.sentinel(bootstrap.BudgetConfig{BookUSD: 0, MaxCalls: 0})
	if s == nil {
		t.Fatal("sentinel should be enabled by default for $0 models")
	}
	if s.MaxCalls() != 100 {
		t.Fatalf("expected default MaxCalls=100, got %d", s.MaxCalls())
	}

	// 99 calls should pass
	s.RecordUsage(99, 0)
	if err := s.Refuse(); err != nil {
		t.Fatalf("at 99 calls should pass: %v", err)
	}

	// 100th call should trigger circuit breaker with BudgetBreachedError
	s.RecordUsage(100, 0)
	err := s.Refuse()
	if err == nil {
		t.Fatal("at 100 calls Refuse should return BudgetBreachedError")
	}
	var bErr *BudgetBreachedError
	if !errors.As(err, &bErr) {
		t.Fatalf("expected *BudgetBreachedError, got %T (%v)", err, err)
	}
	if !strings.Contains(err.Error(), "100 lượt gọi LLM") {
		t.Fatalf("expected error mentioning 100 calls, got: %v", err)
	}
}

func TestBudgetSentinelWarnOnceThenBoundaryStop(t *testing.T) {
	r := &budgetRecorder{}
	s := r.sentinel(bootstrap.BudgetConfig{BookUSD: 10, WarnRatio: 0.8})

	// 未到水位：无副作用
	s.OnCost(5)
	if len(r.reports) != 0 {
		t.Fatalf("below warn ratio should be silent, got %v", r.reports)
	}

	// 越过告警水位：恰好一次 warn，重复回调不再发
	s.OnCost(8.5)
	s.OnCost(9)
	if len(r.reports) != 1 || !strings.HasPrefix(r.reports[0], "warn:") {
		t.Fatalf("expected exactly one warn, got %v", r.reports)
	}

	// 越线：进入 stopPending，发 error，但不立即停（默认等边界）
	s.OnCost(10.5)
	if len(r.reports) != 2 || !strings.HasPrefix(r.reports[1], "error:") {
		t.Fatalf("expected error report on exceeding, got %v", r.reports)
	}
	if len(r.aborts) != 0 {
		t.Fatalf("default mode should not abort before boundary, got %v", r.aborts)
	}

	// 非边界事件不触发
	s.HandleEvent(agentcore.Event{Type: agentcore.EventToolExecEnd, Tool: "novel_context"})
	if len(r.aborts) != 0 {
		t.Fatal("non-subagent boundary should not trigger stop")
	}

	// 子代理边界：恰好一次停机，重复边界不再停
	r.cost = 10.5
	if !s.HandleBoundary() {
		t.Fatal("pending budget stop should be handled at boundary")
	}
	if s.HandleBoundary() {
		t.Fatal("stopped budget should not report another handled boundary")
	}
	if len(r.aborts) != 1 {
		t.Fatalf("expected exactly one abort at boundary, got %v", r.aborts)
	}
}

func TestBudgetSentinelJumpStraightPastLimit(t *testing.T) {
	r := &budgetRecorder{}
	s := r.sentinel(bootstrap.BudgetConfig{BookUSD: 10, WarnRatio: 0.8})

	// 一次回调直接跨过告警与上限：warn 与 error 各恰好一次
	s.OnCost(12)
	if len(r.reports) != 2 {
		t.Fatalf("expected warn+error in single jump, got %v", r.reports)
	}
}

func TestBudgetSentinelHardStop(t *testing.T) {
	r := &budgetRecorder{}
	s := r.sentinel(bootstrap.BudgetConfig{BookUSD: 10, WarnRatio: 0.8, HardStop: true})

	s.OnCost(11)
	if len(r.aborts) != 1 {
		t.Fatalf("hard_stop should abort immediately, got %v", r.aborts)
	}
	// 后续边界不再重复停
	r.cost = 11
	s.HandleEvent(subagentEndEvent())
	if len(r.aborts) != 1 {
		t.Fatalf("stopped state should not abort again, got %v", r.aborts)
	}
}

func TestBudgetSentinelRefuse(t *testing.T) {
	r := &budgetRecorder{cost: 9.99}
	s := r.sentinel(bootstrap.BudgetConfig{BookUSD: 10, WarnRatio: 0.8})

	if err := s.Refuse(); err != nil {
		t.Errorf("below limit should pass: %v", err)
	}
	r.cost = 10 // 恰好等于上限 → 拒绝
	if err := s.Refuse(); err == nil {
		t.Error("at limit should refuse")
	} else if !strings.Contains(err.Error(), "book_usd") {
		t.Errorf("refuse error should mention how to recover, got %v", err)
	}
}

func TestBudgetSentinelZeroCostBlindWarning(t *testing.T) {
	r := &budgetRecorder{}
	s := r.sentinel(bootstrap.BudgetConfig{BookUSD: 10, WarnRatio: 0.8})

	// 连续零成本记账：到 blindZeroStreak 笔时恰好一次盲区告警，之后静默
	for range blindZeroStreak + 3 {
		s.OnCost(0)
	}
	if len(r.reports) != 1 || (!strings.Contains(r.reports[0], "Điểm mù ngân sách") && !strings.Contains(r.reports[0], "预算盲区")) {
		t.Fatalf("expected exactly one blind warning, got %v", r.reports)
	}
	if len(r.aborts) != 0 {
		t.Fatal("blind warning must not abort")
	}

	// 正常计价模型不应误报：每笔记账总额递增
	r2 := &budgetRecorder{}
	s2 := r2.sentinel(bootstrap.BudgetConfig{BookUSD: 10, WarnRatio: 0.8})
	for i := range blindZeroStreak + 3 {
		s2.OnCost(0.1 * float64(i+1))
	}
	for _, rep := range r2.reports {
		if strings.Contains(rep, "Điểm mù") || strings.Contains(rep, "盲区") {
			t.Fatalf("priced model should not trigger blind warning: %v", r2.reports)
		}
	}
}

func TestBudgetSentinelBlindWarningAfterModelSwitch(t *testing.T) {
	// 长跑中途 /model 切到无价模型：total 停在历史值非零但不再增长，同样要告警
	r := &budgetRecorder{}
	s := r.sentinel(bootstrap.BudgetConfig{BookUSD: 100, WarnRatio: 0.8})

	for i := range 5 {
		s.OnCost(1.0 * float64(i+1)) // 计价阶段：总额递增到 $5
	}
	for range blindZeroStreak {
		s.OnCost(5.0) // 切到无价模型：总额钉死
	}
	if len(r.reports) != 1 || (!strings.Contains(r.reports[0], "Điểm mù") && !strings.Contains(r.reports[0], "盲区")) {
		t.Fatalf("expected blind warning after switch to unpriced model, got %v", r.reports)
	}
}

func TestBudgetSentinelCircuitBreakerMaxCalls(t *testing.T) {
	r := &budgetRecorder{}
	s := r.sentinel(bootstrap.BudgetConfig{BookUSD: 0, MaxCalls: 5, HardStop: false})
	if s == nil {
		t.Fatal("sentinel should be enabled for MaxCalls > 0")
	}

	for i := 0; i < 4; i++ {
		s.OnCost(0)
	}
	if len(r.aborts) != 0 {
		t.Fatalf("should not abort before MaxCalls, got %v", r.aborts)
	}

	// Lần gọi thứ 5 đạt giới hạn 5 calls
	s.OnCost(0)
	if len(r.reports) == 0 || !strings.Contains(r.reports[len(r.reports)-1], "Cầu chì tự ngắt") {
		t.Fatalf("expected circuit breaker report, got %v", r.reports)
	}
	if len(r.aborts) != 0 {
		t.Fatal("should wait for boundary when HardStop is false")
	}

	// Tại ranh giới subagent
	if !s.HandleBoundary() {
		t.Fatal("HandleBoundary should return true when stop is pending")
	}
	if len(r.aborts) != 1 || !strings.Contains(r.aborts[0], "5 lượt gọi LLM") {
		t.Fatalf("expected abort with calls limit, got %v", r.aborts)
	}
}

func TestBudgetSentinelCircuitBreakerMaxTokensHardStop(t *testing.T) {
	r := &budgetRecorder{}
	s := r.sentinel(bootstrap.BudgetConfig{BookUSD: 0, MaxTokens: 1000, HardStop: true})
	if s == nil {
		t.Fatal("sentinel should be enabled for MaxTokens > 0")
	}

	s.RecordUsage(1, 500)
	if len(r.aborts) != 0 {
		t.Fatal("should not abort below token limit")
	}

	// Vượt ngưỡng token với HardStop = true -> dừng ngay lập tức
	s.RecordUsage(2, 1200)
	if len(r.aborts) != 1 || !strings.Contains(r.aborts[0], "1000 tokens") {
		t.Fatalf("expected immediate abort on tokens limit, got %v", r.aborts)
	}
}

func TestBudgetSentinelCircuitBreakerMaxDuration(t *testing.T) {
	r := &budgetRecorder{}
	s := r.sentinel(bootstrap.BudgetConfig{BookUSD: 0, MaxDuration: "20ms", HardStop: true})
	if s == nil {
		t.Fatal("sentinel should be enabled for MaxDuration")
	}

	time.Sleep(30 * time.Millisecond)
	s.OnCost(0)
	if len(r.aborts) != 1 || !strings.Contains(r.aborts[0], "thời gian chạy") {
		t.Fatalf("expected duration abort, got %v", r.aborts)
	}
}

func TestBudgetSentinelRefuseCircuitBreakers(t *testing.T) {
	r := &budgetRecorder{}
	s := r.sentinel(bootstrap.BudgetConfig{BookUSD: 0, MaxCalls: 3})

	if err := s.Refuse(); err != nil {
		t.Fatalf("initial Refuse should pass: %v", err)
	}
	s.OnCost(0)
	s.OnCost(0)
	s.OnCost(0)

	if err := s.Refuse(); err == nil {
		t.Fatal("Refuse should fail once MaxCalls reached")
	} else if !strings.Contains(err.Error(), "max_calls") {
		t.Fatalf("expected refuse error to mention max_calls, got %v", err)
	}
}
