package airuntime

import (
	"errors"
	"strings"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Provider    string         `json:"provider"`
	Model       string         `json:"model"`
	Messages    []ChatMessage  `json:"messages"`
	Temperature *float64       `json:"temperature,omitempty"`
	MaxTokens   *int           `json:"max_tokens,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type ChatResponse struct {
	Text       string         `json:"text"`
	Message    ChatMessage    `json:"message"`
	TokenUsage TokenUsage     `json:"token_usage"`
	Raw        map[string]any `json:"raw,omitempty"`
}

type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

func (r ChatRequest) Validate() error {
	if strings.TrimSpace(r.Provider) == "" {
		return errors.New("provider is required")
	}

	if strings.TrimSpace(r.Model) == "" {
		return errors.New("model is required")
	}

	if len(r.Messages) == 0 {
		return errors.New("messages is required")
	}

	for index, message := range r.Messages {
		if strings.TrimSpace(message.Role) == "" {
			return errors.New("message role is required")
		}

		if strings.TrimSpace(message.Content) == "" {
			return errors.New("message content is required")
		}

		switch message.Role {
		case "system", "user", "assistant", "tool":
		default:
			return errors.New("unsupported message role at index " + string(rune(index)))
		}
	}

	if r.Temperature != nil {
		if *r.Temperature < 0 || *r.Temperature > 2 {
			return errors.New("temperature must be between 0 and 2")
		}
	}

	if r.MaxTokens != nil {
		if *r.MaxTokens <= 0 {
			return errors.New("max_tokens must be greater than 0")
		}
	}

	return nil
}

func (r ChatResponse) ResponseText() string {
	if strings.TrimSpace(r.Text) != "" {
		return r.Text
	}

	return r.Message.Content
}

func (u TokenUsage) Normalize() TokenUsage {
	if u.TotalTokens <= 0 {
		u.TotalTokens = u.InputTokens + u.OutputTokens
	}

	return u
}
