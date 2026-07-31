package workflowruntime

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"agentflow-studio/services/api/internal/airuntime"
)

type LLMExecutor struct {
	client *airuntime.Client
}

type LLMNodeConfig struct {
	Provider     string
	Model        string
	Temperature  *float64
	MaxTokens    *int
	SystemPrompt string
}

func NewLLMExecutor(client *airuntime.Client) *LLMExecutor {
	return &LLMExecutor{
		client: client,
	}
}

func (e *LLMExecutor) Type() WorkflowNodeType {
	return WorkflowNodeTypeLLM
}

func (e *LLMExecutor) Validate(config WorkflowNodeConfig) error {
	_, err := parseLLMNodeConfig(config)
	return err
}

func (e *LLMExecutor) Execute(
	ctx context.Context,
	input NodeExecutionInput,
) (*NodeExecutionResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if e.client == nil {
		return nil, NewRuntimeError(
			ErrorCodeAIRuntimeError,
			"AI Runtime Client 未初始化",
			ErrAIRuntimeError,
		)
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	config, err := parseLLMNodeConfig(input.Config())
	if err != nil {
		return nil, err
	}

	renderedPrompt, err := findRenderedPrompt(input)
	if err != nil {
		return nil, err
	}

	messages := buildLLMMessages(config, renderedPrompt)

	response, err := e.client.Chat(ctx, airuntime.ChatRequest{
		Provider:    config.Provider,
		Model:       config.Model,
		Messages:    messages,
		Temperature: config.Temperature,
		MaxTokens:   config.MaxTokens,
		Metadata: map[string]any{
			"workspace_id": input.ExecutionContext.WorkspaceID.String(),
			"workflow_id":  input.ExecutionContext.WorkflowID.String(),
			"run_id":       input.ExecutionContext.RunID.String(),
			"node_id":      input.Node.ID,
			"node_type":    input.Node.Type.String(),
			"trace_id":     input.ExecutionContext.TraceID,
		},
	})
	if err != nil {
		return nil, NewRuntimeErrorWithDetails(
			ErrorCodeAIRuntimeError,
			"AI Runtime Chat 调用失败",
			err,
			map[string]any{
				"node_id":  input.Node.ID,
				"provider": config.Provider,
				"model":    config.Model,
				"reason":   err.Error(),
			},
		)
	}

	responseText := response.ResponseText()
	tokenUsage := TokenUsage{
		Provider:     config.Provider,
		Model:        config.Model,
		InputTokens:  response.TokenUsage.InputTokens,
		OutputTokens: response.TokenUsage.OutputTokens,
		TotalTokens:  response.TokenUsage.TotalTokens,
	}.Normalize()

	output := JSONMap{
		"response_text": responseText,
		"text":          responseText,
		"message": JSONMap{
			"role":    "assistant",
			"content": responseText,
		},
	}

	result := NewNodeExecutionResult(input.Node, output)
	result.TokenUsage = &tokenUsage
	result.Metadata = JSONMap{
		"executor": "LLMExecutor",
		"provider": config.Provider,
		"model":    config.Model,
	}

	return &result, nil
}

func parseLLMNodeConfig(config WorkflowNodeConfig) (*LLMNodeConfig, error) {
	provider, err := readLLMRequiredStringConfig(config, "provider")
	if err != nil {
		return nil, err
	}

	model, err := readLLMRequiredStringConfig(config, "model")
	if err != nil {
		return nil, err
	}

	temperature, err := readLLMOptionalFloatConfig(config, "temperature")
	if err != nil {
		return nil, err
	}

	if temperature != nil {
		if *temperature < 0 || *temperature > 2 {
			return nil, NewRuntimeErrorWithDetails(
				ErrorCodeInvalidLLMConfig,
				"LLM 节点 temperature 必须在 0 到 2 之间",
				ErrInvalidLLMConfig,
				map[string]any{
					"field": "config.temperature",
					"value": *temperature,
				},
			)
		}
	}

	maxTokens, err := readLLMOptionalIntConfig(config, "maxTokens", "max_tokens")
	if err != nil {
		return nil, err
	}

	if maxTokens != nil && *maxTokens <= 0 {
		return nil, NewRuntimeErrorWithDetails(
			ErrorCodeInvalidLLMConfig,
			"LLM 节点 maxTokens 必须大于 0",
			ErrInvalidLLMConfig,
			map[string]any{
				"field": "config.maxTokens",
				"value": *maxTokens,
			},
		)
	}

	systemPrompt, err := readLLMOptionalStringConfig(config, "systemPrompt", "system_prompt")
	if err != nil {
		return nil, err
	}

	return &LLMNodeConfig{
		Provider:     provider,
		Model:        model,
		Temperature:  temperature,
		MaxTokens:    maxTokens,
		SystemPrompt: systemPrompt,
	}, nil
}

func readLLMRequiredStringConfig(config WorkflowNodeConfig, key string) (string, error) {
	value, exists := config[key]
	if !exists || value == nil {
		return "", NewRuntimeErrorWithDetails(
			ErrorCodeInvalidLLMConfig,
			"LLM 节点缺少必要配置",
			ErrInvalidLLMConfig,
			map[string]any{
				"field": "config." + key,
			},
		)
	}

	text, ok := value.(string)
	if !ok {
		return "", NewRuntimeErrorWithDetails(
			ErrorCodeInvalidLLMConfig,
			"LLM 节点配置字段必须是字符串",
			ErrInvalidLLMConfig,
			map[string]any{
				"field": "config." + key,
			},
		)
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return "", NewRuntimeErrorWithDetails(
			ErrorCodeInvalidLLMConfig,
			"LLM 节点配置字段不能为空",
			ErrInvalidLLMConfig,
			map[string]any{
				"field": "config." + key,
			},
		)
	}

	return text, nil
}

func readLLMOptionalStringConfig(config WorkflowNodeConfig, keys ...string) (string, error) {
	for _, key := range keys {
		value, exists := config[key]
		if !exists || value == nil {
			continue
		}

		text, ok := value.(string)
		if !ok {
			return "", NewRuntimeErrorWithDetails(
				ErrorCodeInvalidLLMConfig,
				"LLM 节点配置字段必须是字符串",
				ErrInvalidLLMConfig,
				map[string]any{
					"field": "config." + key,
				},
			)
		}

		return strings.TrimSpace(text), nil
	}

	return "", nil
}

func readLLMOptionalFloatConfig(config WorkflowNodeConfig, key string) (*float64, error) {
	value, exists := config[key]
	if !exists || value == nil {
		return nil, nil
	}

	switch typedValue := value.(type) {
	case float64:
		return &typedValue, nil

	case float32:
		result := float64(typedValue)
		return &result, nil

	case int:
		result := float64(typedValue)
		return &result, nil

	case int64:
		result := float64(typedValue)
		return &result, nil

	case json.Number:
		result, err := typedValue.Float64()
		if err != nil {
			return nil, err
		}

		return &result, nil

	default:
		return nil, NewRuntimeErrorWithDetails(
			ErrorCodeInvalidLLMConfig,
			"LLM 节点 temperature 必须是数字",
			ErrInvalidLLMConfig,
			map[string]any{
				"field": "config." + key,
			},
		)
	}
}

func readLLMOptionalIntConfig(config WorkflowNodeConfig, keys ...string) (*int, error) {
	for _, key := range keys {
		value, exists := config[key]
		if !exists || value == nil {
			continue
		}

		result, err := configValueToInt(value)
		if err != nil {
			return nil, NewRuntimeErrorWithDetails(
				ErrorCodeInvalidLLMConfig,
				"LLM 节点 maxTokens 必须是整数",
				ErrInvalidLLMConfig,
				map[string]any{
					"field": "config." + key,
				},
			)
		}

		return &result, nil
	}

	return nil, nil
}

func configValueToInt(value any) (int, error) {
	switch typedValue := value.(type) {
	case int:
		return typedValue, nil

	case int64:
		return int(typedValue), nil

	case float64:
		if typedValue != float64(int(typedValue)) {
			return 0, fmt.Errorf("value is not integer")
		}

		return int(typedValue), nil

	case json.Number:
		result, err := typedValue.Int64()
		if err != nil {
			return 0, err
		}

		return int(result), nil

	default:
		return 0, fmt.Errorf("unsupported integer type")
	}
}

func findRenderedPrompt(input NodeExecutionInput) (string, error) {
	for index := len(input.InboundResults) - 1; index >= 0; index-- {
		inbound := input.InboundResults[index]
		if inbound.NodeType != WorkflowNodeTypePrompt {
			continue
		}

		if prompt, ok := readStringFromJSONMap(inbound.Output, "rendered_prompt"); ok {
			return prompt, nil
		}

		if prompt, ok := readStringFromJSONMap(inbound.Output, "prompt"); ok {
			return prompt, nil
		}
	}

	if prompt, ok := readStringFromJSONMap(input.Input, "rendered_prompt"); ok {
		return prompt, nil
	}

	if prompt, ok := readStringFromJSONMap(input.Input, "prompt"); ok {
		return prompt, nil
	}

	return "", NewRuntimeErrorWithDetails(
		ErrorCodeInvalidInput,
		"LLM 节点没有找到上游 Prompt rendered_prompt",
		ErrInvalidInput,
		map[string]any{
			"node_id": input.Node.ID,
		},
	)
}

func readStringFromJSONMap(value JSONMap, key string) (string, bool) {
	if value == nil {
		return "", false
	}

	rawValue, exists := value[key]
	if !exists || rawValue == nil {
		return "", false
	}

	text, ok := rawValue.(string)
	if !ok {
		return "", false
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return "", false
	}

	return text, true
}

func buildLLMMessages(config *LLMNodeConfig, renderedPrompt string) []airuntime.ChatMessage {
	messages := []airuntime.ChatMessage{}

	if strings.TrimSpace(config.SystemPrompt) != "" {
		messages = append(messages, airuntime.ChatMessage{
			Role:    "system",
			Content: config.SystemPrompt,
		})
	}

	messages = append(messages, airuntime.ChatMessage{
		Role:    "user",
		Content: renderedPrompt,
	})

	return messages
}
