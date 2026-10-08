package agents

import (
	"context"
	"testing"

	"github.com/voocel/agentcore"
	"github.com/voocel/agentcore/llm"
)

type mockCapabilityModel struct {
	caps     llm.Capabilities
	info     llm.ModelInfo
	provName string
}

func (m *mockCapabilityModel) Generate(_ context.Context, _ []agentcore.Message, _ []agentcore.ToolSpec, _ ...agentcore.CallOption) (*agentcore.LLMResponse, error) {
	return nil, nil
}

func (m *mockCapabilityModel) GenerateStream(_ context.Context, _ []agentcore.Message, _ []agentcore.ToolSpec, _ ...agentcore.CallOption) (<-chan agentcore.StreamEvent, error) {
	return nil, nil
}

func (m *mockCapabilityModel) SupportsTools() bool {
	return true
}

func (m *mockCapabilityModel) Capabilities() llm.Capabilities {
	return m.caps
}

func (m *mockCapabilityModel) Info() llm.ModelInfo {
	return m.info
}

func (m *mockCapabilityModel) ProviderName() string {
	return m.provName
}

func TestModelSupportsThinking_OpenAIModels(t *testing.T) {
	tests := []struct {
		name      string
		provider  string
		modelName string
		want      bool
	}{
		{
			name:      "gemma-4-26b on openai",
			provider:  "openai",
			modelName: "gemma-4-26b",
			want:      false,
		},
		{
			name:      "qwen-27b on openai",
			provider:  "openai",
			modelName: "qwen-27b",
			want:      false,
		},
		{
			name:      "gpt-4o on openai",
			provider:  "openai",
			modelName: "gpt-4o",
			want:      false,
		},
		{
			name:      "gpt-5 on openai",
			provider:  "openai",
			modelName: "gpt-5",
			want:      true,
		},
		{
			name:      "bedrock claude",
			provider:  "bedrock",
			modelName: "anthropic.claude-3-5-sonnet",
			want:      true,
		},
		{
			name:      "bedrock non-claude",
			provider:  "bedrock",
			modelName: "meta.llama3",
			want:      false,
		},
		{
			name:      "openrouter deepseek r1",
			provider:  "openrouter",
			modelName: "deepseek/deepseek-r1",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockCapabilityModel{
				caps: llm.Capabilities{
					Provider: tt.provider,
					Model:    tt.modelName,
					Thinking: llm.ThinkingCapabilities{
						Supported: llm.SupportPartial,
						Disable:   llm.SupportPartial,
						Efforts: []agentcore.ThinkingLevel{
							agentcore.ThinkingLow,
							agentcore.ThinkingMedium,
							agentcore.ThinkingHigh,
						},
					},
				},
				info: llm.ModelInfo{
					Provider: tt.provider,
					Name:     tt.modelName,
				},
				provName: tt.provider,
			}

			got := ModelSupportsThinking(m)
			if got != tt.want {
				t.Errorf("ModelSupportsThinking(%s/%s) = %v, want %v", tt.provider, tt.modelName, got, tt.want)
			}
		})
	}
}

func TestResolveThinkingForModel_NonReasoningOpenAI(t *testing.T) {
	m := &mockCapabilityModel{
		caps: llm.Capabilities{
			Provider: "openai",
			Model:    "gemma-4-26b",
			Thinking: llm.ThinkingCapabilities{
				Supported: llm.SupportPartial,
				Disable:   llm.SupportPartial,
				Efforts: []agentcore.ThinkingLevel{
					agentcore.ThinkingLow,
					agentcore.ThinkingMedium,
					agentcore.ThinkingHigh,
				},
			},
		},
		info: llm.ModelInfo{
			Provider: "openai",
			Name:     "gemma-4-26b",
		},
		provName: "colab-llama",
	}

	// Khi model không hỗ trợ thinking, dù truyền "off", "high" hay bất kỳ giá trị nào,
	// đều phải trả về ThinkingAuto ("") để không gửi thinking param đến litellm
	level, ok := ResolveThinkingForModel(m, agentcore.ThinkingOff)
	if level != agentcore.ThinkingAuto {
		t.Fatalf("ResolveThinkingForModel with 'off' = %q, want ThinkingAuto ('')", level)
	}
	if !ok {
		t.Fatalf("ResolveThinkingForModel with 'off' ok = false, want true")
	}

	level, ok = ResolveThinkingForModel(m, agentcore.ThinkingHigh)
	if level != agentcore.ThinkingAuto {
		t.Fatalf("ResolveThinkingForModel with 'high' = %q, want ThinkingAuto ('')", level)
	}
	if ok {
		t.Fatalf("ResolveThinkingForModel with 'high' ok = true, want false (model cannot satisfy high)")
	}

	avail := AvailableThinkingForModel(m)
	if len(avail) != 1 || avail[0] != agentcore.ThinkingAuto {
		t.Fatalf("AvailableThinkingForModel = %v, want [%q]", avail, agentcore.ThinkingAuto)
	}
}
