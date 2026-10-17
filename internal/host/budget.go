package host

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/voocel/agentcore"
	"github.com/voocel/ainovel-cli/internal/bootstrap"
)

// 预算状态机：单调递进，每次迁移恰好触发一次副作用，不回退。
// 上调预算 = 用户重新授权 = 改配置后重启/新 Host 实例，不在本实例内回退状态。
const (
	budgetNormal      int32 = iota // 未到告警水位
	budgetWarned                   // 已发告警，未越线
	budgetStopPending              // 已越线，等子代理边界停机
	budgetStopped                  // 已执行停机
)

// BudgetSentinel 监视累计成本，执行用户的预算政策（config budget 块）。
//
// 合宪定位（architecture.md §8.3/§10）：不评估模型行为——越线停机等同于用户在
// 那一刻手动 Abort，Host 只是代为执行一条预先签署的指令。它影响控制流，因此
// 不是观察者，定位为与 flow.Dispatcher 平级的 Host 政策组件；Route/工具层不感知。
//
// 停机时机：默认在子代理边界（Host 同步调用 HandleBoundary），不浪费 in-flight 章节；
// hardStop=true 时越线立即停。边界处理先于 flow.Dispatcher 派发下一步，Route/工具层不感知预算。
type BudgetSentinel struct {
	limit     float64
	warnRatio float64
	hardStop  bool

	maxCalls    int
	maxTokens   int64
	maxDuration time.Duration
	startTime   time.Time

	calls  atomic.Int64
	tokens atomic.Int64

	reasonMu      sync.RWMutex
	pendingReason string

	costNow func() float64              // 当前累计成本（usage.Totals 包装；可注入测试桩）
	abort   func(reason string)         // Host 停机包装（带原因事件）
	report  func(level, summary string) // 告警出口（emitEvent + notify，由 Host 注入）

	state atomic.Int32

	// 计费盲区检测：注册表无价且 provider 不自报 cost 的模型每笔记账增量为 $0，
	// 预算静默失效。按"连续多笔零增量"判定而非 total==0——后者抓不住长跑中途
	// /model 切到无价模型的场景（total 停在历史值非零但不再增长）。
	// 免费模型同样命中，提示"预算不会触发"对其同样成立。
	lastTotal   atomic.Uint64 // math.Float64bits(上次回调的累计成本)
	zeroStreak  atomic.Int32
	blindWarned atomic.Bool
}

// blindZeroStreak 连续零增量记账多少笔后告警。正常计价模型每笔增量必 > 0
// （cost 是 float 累计不取整），取 5 仅为避免极端毛刺，不是可调策略阈值。
const blindZeroStreak = 5

// BudgetBreachedError đại diện cho lỗi vi phạm ngân sách hoặc kích hoạt cầu chì ngắt mạch.
type BudgetBreachedError struct {
	Reason string
}

func (e *BudgetBreachedError) Error() string {
	return e.Reason
}

// NewBudgetSentinel 创建预算哨兵；政策未启用时返回 nil（所有方法 nil 安全）。
func NewBudgetSentinel(cfg bootstrap.BudgetConfig, costNow func() float64, abort func(reason string), report func(level, summary string)) *BudgetSentinel {
	if cfg.MaxCalls < 0 || !cfg.Enabled() {
		return nil
	}
	var maxDuration time.Duration
	if cfg.MaxDuration != "" {
		if d, err := time.ParseDuration(cfg.MaxDuration); err == nil {
			maxDuration = d
		}
	}
	maxCalls := cfg.MaxCalls
	// Khi chạy model $0 (BookUSD <= 0) và chưa cấu hình MaxCalls, đặt mặc định 100 calls/session limit
	if cfg.BookUSD <= 0 && maxCalls <= 0 {
		maxCalls = 100
	}
	return &BudgetSentinel{
		limit:       cfg.BookUSD,
		warnRatio:   cfg.WarnRatio,
		hardStop:    cfg.HardStop,
		maxCalls:    maxCalls,
		maxTokens:   cfg.MaxTokens,
		maxDuration: maxDuration,
		startTime:   time.Now(),
		costNow:     costNow,
		abort:       abort,
		report:      report,
	}
}

// setPendingReason lưu lý do dừng máy chờ thực thi ở biên subagent.
func (s *BudgetSentinel) setPendingReason(reason string) {
	s.reasonMu.Lock()
	s.pendingReason = reason
	s.reasonMu.Unlock()
}

// getPendingReason lấy lý do dừng máy chờ thực thi.
func (s *BudgetSentinel) getPendingReason() string {
	s.reasonMu.RLock()
	defer s.reasonMu.RUnlock()
	return s.pendingReason
}

// checkTripwires kiểm tra các chốt ngắt mạch (circuit-breakers) cho model $0 hoặc khi vượt quá ngưỡng token/call/thời gian.
func (s *BudgetSentinel) checkTripwires() {
	if s == nil {
		return
	}
	isZeroCost := s.limit <= 0 || s.zeroStreak.Load() >= blindZeroStreak
	if !isZeroCost && s.maxCalls <= 0 && s.maxTokens <= 0 && s.maxDuration <= 0 {
		return
	}

	var tripReason string
	if s.maxCalls > 0 && s.calls.Load() >= int64(s.maxCalls) {
		tripReason = fmt.Sprintf("Cầu chì tự ngắt: Đã đạt giới hạn %d lượt gọi LLM cho mô hình tự host ($0 cost)", s.maxCalls)
	} else if s.maxTokens > 0 && s.tokens.Load() >= s.maxTokens {
		tripReason = fmt.Sprintf("Cầu chì tự ngắt: Đã đạt giới hạn %d tokens tích lũy cho mô hình tự host ($0 cost)", s.maxTokens)
	} else if s.maxDuration > 0 && time.Since(s.startTime) >= s.maxDuration {
		tripReason = fmt.Sprintf("Cầu chì tự ngắt: Đã đạt giới hạn thời gian chạy %s cho mô hình tự host ($0 cost)", s.maxDuration)
	}

	if tripReason != "" {
		if s.state.CompareAndSwap(budgetNormal, budgetStopPending) || s.state.CompareAndSwap(budgetWarned, budgetStopPending) {
			s.setPendingReason(tripReason)
			if s.hardStop {
				s.report("error", tripReason+", lập tức dừng máy")
				s.stopWithReason(tripReason)
				return
			}
			s.report("error", tripReason+", sẽ dừng máy sau khi nhiệm vụ subagent hiện tại kết thúc")
		}
	}
}

// OnUsage cập nhật số lượt gọi và token từ UsageTracker, đồng thời kiểm tra circuit breaker.
func (s *BudgetSentinel) OnUsage(calls int64, tokens int64) {
	if s == nil {
		return
	}
	if calls > 0 {
		s.calls.Store(calls)
	}
	if tokens > 0 {
		s.tokens.Store(tokens)
	}
	s.checkTripwires()
}

// RecordUsage ghi nhận số lượt gọi và tokens phục vụ kiểm thử hoặc ghi nhận trực tiếp.
func (s *BudgetSentinel) RecordUsage(calls int64, tokens int64) {
	s.OnUsage(calls, tokens)
}

// OnCost 由 UsageTracker 每次记账后携带最新累计成本调用（锁外）。
// 一次回调可能连跨两级（normal→warned→stopPending），两次副作用各触发一次。
func (s *BudgetSentinel) OnCost(total float64) {
	if s == nil {
		return
	}
	s.calls.Add(1)

	if prev := s.lastTotal.Swap(math.Float64bits(total)); total == math.Float64frombits(prev) {
		if s.zeroStreak.Add(1) >= blindZeroStreak && s.blindWarned.CompareAndSwap(false, true) {
			s.report("warn", fmt.Sprintf("Điểm mù ngân sách: Liên tục ghi nhận nhưng chi phí tích lũy dừng ở $%.2f không tăng thêm (mô hình hiện tại không có giá trong bảng đăng ký và provider không tự báo cost, hoặc là mô hình miễn phí) — giới hạn ngân sách sẽ không được kích hoạt", total))
		}
	} else {
		s.zeroStreak.Store(0)
	}

	if s.limit > 0 {
		if total >= s.limit*s.warnRatio && s.state.CompareAndSwap(budgetNormal, budgetWarned) {
			s.report("warn", fmt.Sprintf("Cảnh báo ngân sách: Đã chi tiêu $%.2f, đạt %.0f%% của mức ngân sách $%.2f", total, s.warnRatio*100, s.limit))
		}
		if total >= s.limit && s.state.CompareAndSwap(budgetWarned, budgetStopPending) {
			reason := fmt.Sprintf("Hết ngân sách: Đã chi tiêu $%.2f, vượt ngân sách $%.2f", total, s.limit)
			s.setPendingReason(reason)
			if s.hardStop {
				s.report("error", reason+", lập tức dừng máy")
				s.stopWithReason(reason)
				return
			}
			s.report("error", reason+", sẽ dừng máy sau khi nhiệm vụ subagent hiện tại kết thúc")
		}
	}

	s.checkTripwires()
}

// HandleEvent 在子代理边界执行待定的停机。订阅必须先于 Dispatcher。
// 不跳过 IsError——出错返回同样是边界，停机不应因子代理失败而推迟。
func (s *BudgetSentinel) HandleEvent(ev agentcore.Event) {
	if s == nil {
		return
	}
	if ev.Type != agentcore.EventToolExecEnd || ev.Tool != "subagent" {
		return
	}
	s.HandleBoundary()
}

func (s *BudgetSentinel) HandleBoundary() bool {
	if s == nil || s.state.Load() != budgetStopPending {
		return false
	}
	reason := s.getPendingReason()
	if reason == "" {
		cost := s.costNow()
		reason = fmt.Sprintf("Dừng máy do ngân sách: Đã chi tiêu $%.2f, vượt ngân sách $%.2f; hãy tăng budget.book_usd để có thể khôi phục tiếp tục chạy", cost, s.limit)
	}
	s.stopWithReason(reason)
	return true
}

func (s *BudgetSentinel) stopWithReason(reason string) {
	if s.state.CompareAndSwap(budgetStopPending, budgetStopped) {
		s.abort(reason)
	}
}

func (s *BudgetSentinel) stop(total float64) {
	reason := fmt.Sprintf("Dừng máy do ngân sách: Đã chi tiêu $%.2f, vượt ngân sách $%.2f; hãy tăng budget.book_usd để có thể khôi phục tiếp tục chạy", total, s.limit)
	s.stopWithReason(reason)
}

// Refuse 启动前置检查：预算已超返回拒绝错误（Start/Resume/Continue 恢复路径调用）。
// 用户上调预算 = 重新授权，新配置下 Refuse 自然放行。
func (s *BudgetSentinel) Refuse() error {
	if s == nil {
		return nil
	}
	if s.limit > 0 {
		if cost := s.costNow(); cost >= s.limit {
			return &BudgetBreachedError{Reason: fmt.Sprintf("Sách này đã chi tiêu $%.2f, đạt giới hạn ngân sách $%.2f; vui lòng tăng cấu hình budget.book_usd rồi thử lại", cost, s.limit)}
		}
	}
	if s.maxCalls > 0 && s.calls.Load() >= int64(s.maxCalls) {
		return &BudgetBreachedError{Reason: fmt.Sprintf("Sách này đã đạt giới hạn %d lượt gọi LLM cho mô hình tự host ($0 cost); vui lòng tăng cấu hình budget.max_calls rồi thử lại", s.maxCalls)}
	}
	if s.maxTokens > 0 && s.tokens.Load() >= s.maxTokens {
		return &BudgetBreachedError{Reason: fmt.Sprintf("Sách này đã đạt giới hạn %d tokens tích lũy cho mô hình tự host ($0 cost); vui lòng tăng cấu hình budget.max_tokens rồi thử lại", s.maxTokens)}
	}
	if s.maxDuration > 0 && time.Since(s.startTime) >= s.maxDuration {
		return &BudgetBreachedError{Reason: fmt.Sprintf("Sách này đã chạy quá thời hạn %s cho mô hình tự host ($0 cost); vui lòng tăng cấu hình budget.max_duration rồi thử lại", s.maxDuration)}
	}
	return nil
}

// Limit 返回预算上限（UI 展示用）；未启用返回 0。
func (s *BudgetSentinel) Limit() float64 {
	if s == nil {
		return 0
	}
	return s.limit
}

func (s *BudgetSentinel) MaxCalls() int {
	if s == nil {
		return 0
	}
	return s.maxCalls
}

func (s *BudgetSentinel) MaxTokens() int64 {
	if s == nil {
		return 0
	}
	return s.maxTokens
}

func (s *BudgetSentinel) MaxDuration() time.Duration {
	if s == nil {
		return 0
	}
	return s.maxDuration
}

func (s *BudgetSentinel) Calls() int64 {
	if s == nil {
		return 0
	}
	return s.calls.Load()
}

func (s *BudgetSentinel) Tokens() int64 {
	if s == nil {
		return 0
	}
	return s.tokens.Load()
}
