// Package llmretry 是直接模型调用共用的请求层重试内核：仅重试模型适配器
// 明确标记为 retryable 的错误，遵守 Retry-After/指数退避，并经 ToolProgress
// 把进度送入既有工作台观察链。账户、鉴权、权限等终止错误会立即返回；
// retryable 错误持续重试，生命周期只由 context 控制。
package llmretry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voocel/agentcore"
)

const maxRetryDelay = 60 * time.Second

// Generator 是请求重试所需的最小模型接口。
type Generator interface {
	Generate(context.Context, []agentcore.Message, []agentcore.ToolSpec, ...agentcore.CallOption) (*agentcore.LLMResponse, error)
}

// Event 描述一次即将发生的请求重试。
type Event struct {
	Attempt int
	Delay   time.Duration
	Err     error
}

// Config 配置重试的可观测信息，不改变重试语义。
type Config struct {
	Agent   string
	OnRetry func(Event)
}

// Generate 调用 model.Generate。retryable 错误退避后持续重试，直到成功或
// context 结束；非 retryable 错误立即返回。
func Generate(ctx context.Context, model Generator, cfg Config, messages []agentcore.Message, opts ...agentcore.CallOption) (*agentcore.LLMResponse, error) {
	for retry := 1; ; retry++ {
		resp, err := model.Generate(ctx, messages, nil, opts...)
		if err == nil {
			return resp, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		if !isRetryable(err) {
			return nil, AsNonRetryable(err)
		}

		enriched := enrichDecodeError(err)
		delay := retryDelay(err, retry-1)
		event := Event{Attempt: retry, Delay: delay, Err: enriched}
		if cfg.OnRetry != nil {
			cfg.OnRetry(event)
		}
		meta, _ := json.Marshal(struct {
			DelayMS int64 `json:"retry_delay_ms"`
		}{DelayMS: delay.Milliseconds()})
		agentcore.ReportToolProgress(ctx, agentcore.ProgressPayload{
			Kind:    agentcore.ProgressRetry,
			Agent:   cfg.Agent,
			Attempt: retry,
			Message: enriched.Error(),
			Meta:    meta,
		})
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

// NonRetryableError bọc lỗi LLM không thể thử lại (như HTTP 400 exceed_context_size_error, 401, 403, 404, 422),
// triển khai agentcore.RetryableError với Retryable() == false để dừng ngay AgentLoop / subagentMaxRetries.
type NonRetryableError struct {
	Err error
}

func (e NonRetryableError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "non-retryable error"
}

func (e NonRetryableError) Unwrap() error {
	return e.Err
}

func (e NonRetryableError) Retryable() bool {
	return false
}

func (e NonRetryableError) FailoverEligible() bool {
	return false
}

// AsNonRetryable bọc err thành NonRetryableError nếu chưa bọc (đồng thời hiển thị nguyên nhân giải mã ẩn nếu có).
func AsNonRetryable(err error) error {
	if err == nil {
		return nil
	}
	err = enrichDecodeError(err)
	var nr NonRetryableError
	if errors.As(err, &nr) {
		return err
	}
	return NonRetryableError{Err: err}
}

func enrichDecodeError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		causeMsg := syntaxErr.Error()
		if strings.Contains(strings.ToLower(causeMsg), "invalid character '<'") {
			if !strings.Contains(msg, "HTML") {
				return fmt.Errorf("%w: endpoint trả về trang HTML thay vì JSON (%s) — kiểm tra Cloudflare Zero Trust Access hoặc base_url", err, causeMsg)
			}
		} else if !strings.Contains(msg, causeMsg) {
			return fmt.Errorf("%w (%s)", err, causeMsg)
		}
	}
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		causeMsg := typeErr.Error()
		if !strings.Contains(msg, causeMsg) {
			return fmt.Errorf("%w (%s)", err, causeMsg)
		}
	}
	return err
}

type statusCodeGetter interface {
	StatusCode() int
}

type httpStatusCodeGetter interface {
	HTTPStatusCode() int
}

type statusGetter interface {
	Status() int
}

// IsNonRetryable xác định một lỗi có phải là lỗi 4xx / tràn ngữ cảnh / xác thực / phản hồi HTML-sai định dạng không thể thử lại hay không.
func IsNonRetryable(err error) bool {
	if err == nil {
		return false
	}
	var retryable agentcore.RetryableError
	if errors.As(err, &retryable) && !retryable.Retryable() {
		return true
	}
	if agentcore.IsContextOverflow(err) {
		return true
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		return true
	}
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		causeLower := strings.ToLower(syntaxErr.Error())
		if strings.Contains(causeLower, "invalid character '<'") ||
			(syntaxErr.Offset <= 2 && strings.Contains(causeLower, "looking for beginning of value")) {
			return true
		}
	}

	var sc statusCodeGetter
	if errors.As(err, &sc) {
		code := sc.StatusCode()
		if code >= 400 && code < 500 && code != 429 {
			return true
		}
	}
	var hsc httpStatusCodeGetter
	if errors.As(err, &hsc) {
		code := hsc.HTTPStatusCode()
		if code >= 400 && code < 500 && code != 429 {
			return true
		}
	}
	var sg statusGetter
	if errors.As(err, &sg) {
		code := sg.Status()
		if code >= 400 && code < 500 && code != 429 {
			return true
		}
	}

	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "429") || strings.Contains(msg, "rate limit") {
		return false
	}

	if strings.Contains(msg, "exceed_context_size_error") ||
		strings.Contains(msg, "context_length_exceeded") ||
		strings.Contains(msg, "maximum context length") ||
		strings.Contains(msg, "context window") ||
		strings.Contains(msg, "context_overflow") {
		return true
	}

	if strings.Contains(msg, "400 bad request") ||
		strings.Contains(msg, "status: 400") ||
		strings.Contains(msg, "status code: 400") ||
		strings.Contains(msg, "code: 400") ||
		strings.Contains(msg, "http 400") ||
		strings.Contains(msg, "invalid_request_error") {
		return true
	}

	if strings.Contains(msg, "401 unauthorized") ||
		strings.Contains(msg, "status: 401") ||
		strings.Contains(msg, "status code: 401") ||
		strings.Contains(msg, "code: 401") ||
		strings.Contains(msg, "http 401") ||
		strings.Contains(msg, "invalid api key") ||
		strings.Contains(msg, "authentication_error") {
		return true
	}

	if strings.Contains(msg, "403 forbidden") ||
		strings.Contains(msg, "status: 403") ||
		strings.Contains(msg, "status code: 403") ||
		strings.Contains(msg, "code: 403") ||
		strings.Contains(msg, "http 403") ||
		strings.Contains(msg, "permission_denied") {
		return true
	}

	if strings.Contains(msg, "404 not found") ||
		strings.Contains(msg, "status: 404") ||
		strings.Contains(msg, "status code: 404") ||
		strings.Contains(msg, "code: 404") ||
		strings.Contains(msg, "http 404") {
		return true
	}

	if strings.Contains(msg, "422 unprocessable") ||
		strings.Contains(msg, "status: 422") ||
		strings.Contains(msg, "status code: 422") ||
		strings.Contains(msg, "code: 422") ||
		strings.Contains(msg, "http 422") {
		return true
	}

	return false
}

func isRetryable(err error) bool {
	if IsNonRetryable(err) {
		return false
	}
	var retryable agentcore.RetryableError
	return errors.As(err, &retryable) && retryable.Retryable()
}

func retryDelay(err error, attempt int) time.Duration {
	var hinter agentcore.RetryHinter
	if errors.As(err, &hinter) {
		if delay := hinter.RetryAfter(); delay > 0 {
			if delay > maxRetryDelay {
				return maxRetryDelay
			}
			return delay
		}
	}
	delay := time.Second
	for i := 0; i < attempt && delay < maxRetryDelay; i++ {
		delay *= 2
	}
	if delay > maxRetryDelay {
		return maxRetryDelay
	}
	return delay
}
