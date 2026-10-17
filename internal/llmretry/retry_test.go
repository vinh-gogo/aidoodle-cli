package llmretry

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/voocel/agentcore"
)

func TestRetryDelayCapsWithoutOverflow(t *testing.T) {
	if got := retryDelay(nil, 10_000); got != maxRetryDelay {
		t.Fatalf("retryDelay = %s, want %s", got, maxRetryDelay)
	}
	if got := retryDelay(nil, 3); got != 8*time.Second {
		t.Fatalf("retryDelay = %s, want 8s", got)
	}
}

type mockGenerator struct {
	calls atomic.Int32
	fn    func(attempt int) (*agentcore.LLMResponse, error)
}

func (m *mockGenerator) Generate(ctx context.Context, msgs []agentcore.Message, tools []agentcore.ToolSpec, opts ...agentcore.CallOption) (*agentcore.LLMResponse, error) {
	call := int(m.calls.Add(1))
	if m.fn != nil {
		return m.fn(call)
	}
	return nil, nil
}

type statusErr struct {
	msg  string
	code int
}

func (e statusErr) Error() string   { return e.msg }
func (e statusErr) StatusCode() int { return e.code }

type testRetryableErr struct {
	msg string
}

func (e testRetryableErr) Error() string   { return e.msg }
func (e testRetryableErr) Retryable() bool { return true }

func TestGenerate_NonRetryable4xxErrors(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{
			name: "exceed_context_size_error",
			err:  errors.New(`{"error": {"message": "exceed_context_size_error: prompt too long", "type": "invalid_request_error"}}`),
		},
		{
			name: "http 400 bad request",
			err:  statusErr{msg: "400 Bad Request: prompt contains invalid tokens", code: 400},
		},
		{
			name: "http 401 unauthorized",
			err:  statusErr{msg: "401 Unauthorized: invalid api key", code: 401},
		},
		{
			name: "http 403 forbidden",
			err:  statusErr{msg: "403 Forbidden: access denied", code: 403},
		},
		{
			name: "http 422 unprocessable",
			err:  statusErr{msg: "422 Unprocessable Entity", code: 422},
		},
		{
			name: "explicit non retryable wrapper",
			err:  AsNonRetryable(errors.New("some custom non retryable error")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockGenerator{
				fn: func(attempt int) (*agentcore.LLMResponse, error) {
					return nil, tc.err
				},
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			_, err := Generate(ctx, mock, Config{Agent: "test"}, nil)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if mock.calls.Load() != 1 {
				t.Fatalf("expected exactly 1 call (no retries), got %d", mock.calls.Load())
			}
		})
	}
}

func TestGenerate_RetryableRecovers(t *testing.T) {
	mock := &mockGenerator{
		fn: func(attempt int) (*agentcore.LLMResponse, error) {
			if attempt == 1 {
				return nil, testRetryableErr{msg: "temporary network glitch"}
			}
			return &agentcore.LLMResponse{
				Message: agentcore.Message{
					Role: agentcore.RoleAssistant,
					Content: []agentcore.ContentBlock{
						agentcore.TextBlock("Thành công"),
					},
				},
			}, nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := Generate(ctx, mock, Config{Agent: "test"}, nil)
	if err != nil {
		t.Fatalf("expected recovery, got err: %v", err)
	}
	if resp == nil || len(resp.Message.Content) == 0 {
		t.Fatalf("unexpected resp: %v", resp)
	}
	if mock.calls.Load() != 2 {
		t.Fatalf("expected 2 calls, got %d", mock.calls.Load())
	}
}
