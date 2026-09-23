package airuntime

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type ChatMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

type ToolFunctionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type ToolDefinition struct {
	Type     string                 `json:"type,omitempty"`
	Function ToolFunctionDefinition `json:"function"`
}

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type,omitempty"`
	Function ToolCallFunction `json:"function"`
}

type ToolCallDelta struct {
	Index          int    `json:"index"`
	ID             string `json:"id,omitempty"`
	Type           string `json:"type,omitempty"`
	FunctionName   string `json:"function_name,omitempty"`
	ArgumentsDelta string `json:"arguments_delta,omitempty"`
}

type ChatRequest struct {
	Provider string        `json:"provider"`
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`

	Temperature      *float64 `json:"temperature,omitempty"`
	MaxTokens        *int     `json:"max_tokens,omitempty"`
	TopP             *float64 `json:"top_p,omitempty"`
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64 `json:"presence_penalty,omitempty"`
	Stop             any      `json:"stop,omitempty"`

	Tools      []ToolDefinition `json:"tools,omitempty"`
	ToolChoice any              `json:"tool_choice,omitempty"`

	Metadata map[string]any `json:"metadata,omitempty"`
}

type ChatResponse struct {
	Text       string         `json:"text"`
	Message    ChatMessage    `json:"message"`
	ToolCalls  []ToolCall     `json:"tool_calls,omitempty"`
	TokenUsage TokenUsage     `json:"token_usage"`
	Raw        map[string]any `json:"raw,omitempty"`
}

type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type ChatStreamEventType string

const (
	ChatStreamEventTypeStart         ChatStreamEventType = "start"
	ChatStreamEventTypeDelta         ChatStreamEventType = "delta"
	ChatStreamEventTypeToolCallDelta ChatStreamEventType = "tool_call_delta"
	ChatStreamEventTypeUsage         ChatStreamEventType = "usage"
	ChatStreamEventTypeDone          ChatStreamEventType = "done"
	ChatStreamEventTypeError         ChatStreamEventType = "error"
)

type ChatStreamError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
	Details   any    `json:"details,omitempty"`
}

type ChatStreamEvent struct {
	Type ChatStreamEventType `json:"type"`

	Delta         string           `json:"delta,omitempty"`
	Text          string           `json:"text,omitempty"`
	FinishReason  string           `json:"finish_reason,omitempty"`
	TokenUsage    *TokenUsage      `json:"token_usage,omitempty"`
	ToolCallDelta *ToolCallDelta   `json:"tool_call_delta,omitempty"`
	ToolCalls     []ToolCall       `json:"tool_calls,omitempty"`
	Error         *ChatStreamError `json:"error,omitempty"`

	Metadata map[string]any `json:"metadata,omitempty"`

	RequestID        string `json:"request_id,omitempty"`
	GatewayRequestID string `json:"gateway_request_id,omitempty"`
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
		if err := validateChatMessage(index, message); err != nil {
			return err
		}
	}

	toolNames, err := validateTools(r.Tools)
	if err != nil {
		return err
	}

	if err := validateToolChoice(r.ToolChoice, toolNames); err != nil {
		return err
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

	if r.TopP != nil {
		if *r.TopP <= 0 || *r.TopP > 1 {
			return errors.New("top_p must be greater than 0 and less than or equal to 1")
		}
	}

	if r.FrequencyPenalty != nil {
		if *r.FrequencyPenalty < -2 || *r.FrequencyPenalty > 2 {
			return errors.New("frequency_penalty must be between -2 and 2")
		}
	}

	if r.PresencePenalty != nil {
		if *r.PresencePenalty < -2 || *r.PresencePenalty > 2 {
			return errors.New("presence_penalty must be between -2 and 2")
		}
	}

	if err := validateStop(r.Stop); err != nil {
		return err
	}

	return nil
}

func validateChatMessage(index int, message ChatMessage) error {
	role := strings.TrimSpace(message.Role)
	content := strings.TrimSpace(message.Content)
	toolCallID := strings.TrimSpace(message.ToolCallID)

	if role == "" {
		return fmt.Errorf("message role is required at index %d", index)
	}

	switch role {
	case "system", "user":
		if content == "" {
			return fmt.Errorf("message content is required at index %d", index)
		}

		if toolCallID != "" || len(message.ToolCalls) > 0 {
			return fmt.Errorf("system/user message cannot include tool fields at index %d", index)
		}

	case "assistant":
		if toolCallID != "" {
			return fmt.Errorf("assistant message cannot include tool_call_id at index %d", index)
		}

		if content == "" && len(message.ToolCalls) == 0 {
			return fmt.Errorf("assistant message requires content or tool_calls at index %d", index)
		}

	case "tool":
		if toolCallID == "" {
			return fmt.Errorf("tool message requires tool_call_id at index %d", index)
		}

		if content == "" {
			return fmt.Errorf("tool message content is required at index %d", index)
		}

		if len(message.ToolCalls) > 0 {
			return fmt.Errorf("tool message cannot include tool_calls at index %d", index)
		}

	default:
		return fmt.Errorf("unsupported message role at index %d", index)
	}

	return nil
}

func validateTools(tools []ToolDefinition) (map[string]struct{}, error) {
	names := map[string]struct{}{}

	for index, tool := range tools {
		toolType := strings.TrimSpace(tool.Type)
		if toolType != "" && toolType != "function" {
			return nil, fmt.Errorf("tool type must be function at index %d", index)
		}

		name := strings.TrimSpace(tool.Function.Name)
		if name == "" {
			return nil, fmt.Errorf("tool function name is required at index %d", index)
		}

		if !isValidToolName(name) {
			return nil, fmt.Errorf("invalid tool function name at index %d", index)
		}

		if _, exists := names[name]; exists {
			return nil, fmt.Errorf("duplicate tool function name: %s", name)
		}

		names[name] = struct{}{}
	}

	return names, nil
}

func validateToolChoice(value any, toolNames map[string]struct{}) error {
	if value == nil {
		return nil
	}

	switch typedValue := value.(type) {
	case string:
		choice := strings.TrimSpace(typedValue)
		switch choice {
		case "none":
			return nil
		case "auto", "required":
			if len(toolNames) == 0 {
				return errors.New("tool_choice requires tools")
			}
			return nil
		default:
			return errors.New("tool_choice must be none, auto, required, or function object")
		}

	case map[string]any:
		return validateToolChoiceObject(typedValue, toolNames)

	default:
		rawValue, err := json.Marshal(typedValue)
		if err != nil {
			return fmt.Errorf("marshal tool_choice failed: %w", err)
		}

		var objectValue map[string]any
		if err := json.Unmarshal(rawValue, &objectValue); err != nil {
			return errors.New("tool_choice must be none, auto, required, or function object")
		}

		return validateToolChoiceObject(objectValue, toolNames)
	}
}

func validateToolChoiceObject(value map[string]any, toolNames map[string]struct{}) error {
	if len(toolNames) == 0 {
		return errors.New("tool_choice requires tools")
	}

	toolType, _ := value["type"].(string)
	if strings.TrimSpace(toolType) != "function" {
		return errors.New("tool_choice object type must be function")
	}

	functionValue, ok := value["function"].(map[string]any)
	if !ok {
		return errors.New("tool_choice function is required")
	}

	name, _ := functionValue["name"].(string)
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("tool_choice function name is required")
	}

	if _, exists := toolNames[name]; !exists {
		return fmt.Errorf("tool_choice function name does not exist in tools: %s", name)
	}

	return nil
}

func isValidToolName(name string) bool {
	if len(name) == 0 || len(name) > 64 {
		return false
	}

	for _, char := range name {
		if char >= 'a' && char <= 'z' {
			continue
		}

		if char >= 'A' && char <= 'Z' {
			continue
		}

		if char >= '0' && char <= '9' {
			continue
		}

		if char == '_' || char == '-' {
			continue
		}

		return false
	}

	return true
}

func (r ChatResponse) ResponseText() string {
	if strings.TrimSpace(r.Text) != "" {
		return r.Text
	}

	return r.Message.Content
}

func (r ChatResponse) HasToolCalls() bool {
	return len(r.ToolCalls) > 0 || len(r.Message.ToolCalls) > 0
}

func (r ChatResponse) AllToolCalls() []ToolCall {
	if len(r.ToolCalls) > 0 {
		return r.ToolCalls
	}

	return r.Message.ToolCalls
}

func (u TokenUsage) Normalize() TokenUsage {
	if u.TotalTokens <= 0 {
		u.TotalTokens = u.InputTokens + u.OutputTokens
	}

	return u
}

func (e ChatStreamEvent) SSEEventName() string {
	if e.Type == "" {
		return "message"
	}

	return string(e.Type)
}

func EncodeSSE(event ChatStreamEvent) ([]byte, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("marshal sse event failed: %w", err)
	}

	return []byte(
		"event: " + event.SSEEventName() + "\n" +
			"data: " + string(payload) + "\n\n",
	), nil
}

func NewStreamErrorEvent(code string, message string, retryable bool, details any) ChatStreamEvent {
	return ChatStreamEvent{
		Type: ChatStreamEventTypeError,
		Error: &ChatStreamError{
			Code:      code,
			Message:   message,
			Retryable: retryable,
			Details:   details,
		},
	}
}

func validateStop(value any) error {
	if value == nil {
		return nil
	}

	switch typedValue := value.(type) {
	case string:
		return nil

	case []string:
		if len(typedValue) > 4 {
			return errors.New("stop supports at most 4 items")
		}

		return nil

	case []any:
		if len(typedValue) > 4 {
			return errors.New("stop supports at most 4 items")
		}

		for index, item := range typedValue {
			if _, ok := item.(string); !ok {
				return fmt.Errorf("stop item at index %d must be string", index)
			}
		}

		return nil

	default:
		return errors.New("stop must be string or string array")
	}
}
