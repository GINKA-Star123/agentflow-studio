package workflowruntime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"agentflow-studio/services/api/internal/airuntime"
)

const (
	ToolBridgeOutputKey     = "tool_bridge"
	ToolMessagesOutputKey   = "tool_messages"
	toolCallStatusSucceeded = "succeeded"
)

type ToolCallRequest struct {
	ToolCallID    string  `json:"tool_call_id"`
	ToolName      string  `json:"tool_name"`
	ToolNodeID    string  `json:"tool_node_id"`
	ToolNodeLabel string  `json:"tool_node_label,omitempty"`
	LLMNodeID     string  `json:"llm_node_id"`
	WorkspaceID   string  `json:"workspace_id"`
	WorkflowID    string  `json:"workflow_id"`
	RunID         string  `json:"run_id"`
	TimeoutMS     int     `json:"timeout_ms"`
	Arguments     JSONMap `json:"arguments"`
	Metadata      JSONMap `json:"metadata,omitempty"`
}

type ToolCallResult struct {
	Request   ToolCallRequest `json:"request"`
	Status    string          `json:"status"`
	Mock      bool            `json:"mock"`
	Output    JSONMap         `json:"output,omitempty"`
	Error     string          `json:"error,omitempty"`
	LatencyMS int64           `json:"latency_ms"`
}

type ToolCallBridgeResult struct {
	Results      []ToolCallResult        `json:"results"`
	ToolMessages []airuntime.ChatMessage `json:"tool_messages,omitempty"`
	Pending      bool                    `json:"pending"`
}

func BuildToolCallBridge(
	schema *WorkflowSchema,
	executionContext *ExecutionContext,
	llmNode WorkflowSchemaNode,
	llmResult *NodeExecutionResult,
) (*ToolCallBridgeResult, error) {
	if schema == nil {
		return nil, NewRuntimeError(
			ErrorCodeInvalidSchema,
			"Workflow Schema 不能为空",
			ErrInvalidSchema,
		)
	}

	if executionContext == nil {
		return nil, NewRuntimeError(
			ErrorCodeInvalidExecutionContext,
			"执行上下文不能为空",
			ErrInvalidExecutionContext,
		)
	}

	if err := executionContext.Validate(); err != nil {
		return nil, err
	}

	if llmResult == nil {
		return nil, NewRuntimeError(
			ErrorCodeNodeExecutionFailed,
			"LLM 节点执行结果不能为空",
			ErrNodeExecutionFailed,
		)
	}

	if llmNode.ID == "" {
		return nil, NewRuntimeErrorWithDetails(
			ErrorCodeInvalidInput,
			"LLM 节点缺少 node_id",
			ErrInvalidInput,
			map[string]any{
				"field": "node.id",
			},
		)
	}

	toolCalls, err := extractToolCallsFromNodeResult(llmResult)
	if err != nil {
		return nil, err
	}

	if len(toolCalls) == 0 {
		return &ToolCallBridgeResult{
			Results:      []ToolCallResult{},
			ToolMessages: []airuntime.ChatMessage{},
			Pending:      false,
		}, nil
	}

	toolNodeIndex, err := buildToolNodeIndex(schema)
	if err != nil {
		return nil, err
	}

	bridgeResult := &ToolCallBridgeResult{
		Results:      make([]ToolCallResult, 0, len(toolCalls)),
		ToolMessages: make([]airuntime.ChatMessage, 0, len(toolCalls)),
		Pending:      true,
	}

	for index, toolCall := range toolCalls {
		toolName := strings.TrimSpace(toolCall.Function.Name)
		if toolName == "" {
			return nil, NewRuntimeErrorWithDetails(
				ErrorCodeNodeExecutionFailed,
				"Tool 调用缺少 function.name",
				ErrNodeExecutionFailed,
				map[string]any{
					"tool_call_id": toolCall.ID,
					"index":        index,
				},
			)
		}

		toolNode, exists := toolNodeIndex[toolName]
		if !exists {
			return nil, NewRuntimeErrorWithDetails(
				ErrorCodeNodeExecutionFailed,
				"未找到匹配的 Tool 节点",
				ErrNodeExecutionFailed,
				map[string]any{
					"tool_call_id": toolCall.ID,
					"tool_name":    toolName,
					"llm_node_id":  llmNode.ID,
				},
			)
		}

		arguments, err := parseToolCallArguments(toolCall.Function.Arguments, toolCall.ID, toolName)
		if err != nil {
			return nil, err
		}

		request := ToolCallRequest{
			ToolCallID:    strings.TrimSpace(toolCall.ID),
			ToolName:      toolName,
			ToolNodeID:    toolNode.ID,
			ToolNodeLabel: toolNode.Label,
			LLMNodeID:     llmNode.ID,
			WorkspaceID:   executionContext.WorkspaceID.String(),
			WorkflowID:    executionContext.WorkflowID.String(),
			RunID:         executionContext.RunID.String(),
			TimeoutMS:     readToolNodeTimeoutMS(toolNode),
			Arguments:     CloneJSONMap(arguments),
			Metadata: JSONMap{
				"bridge":           "mock",
				"tool_match_mode":  "exact",
				"tool_node_id":     toolNode.ID,
				"tool_node_label":  toolNode.Label,
				"tool_node_type":   toolNode.Type.String(),
				"tool_node_config": CloneJSONMap(JSONMap(toolNode.Config)),
				"tool_call_index":  index,
				"tool_call_name":   toolName,
				"tool_call_source": llmNode.ID,
			},
		}

		result := mockExecuteToolCall(request, toolNode)
		bridgeResult.Results = append(bridgeResult.Results, result)
		bridgeResult.ToolMessages = append(bridgeResult.ToolMessages, result.ToToolMessage())
	}

	return bridgeResult, nil
}

func extractToolCallsFromNodeResult(result *NodeExecutionResult) ([]airuntime.ToolCall, error) {
	if result == nil {
		return []airuntime.ToolCall{}, nil
	}

	if rawCalls, ok := result.OutputValue("tool_calls"); ok {
		toolCalls, err := normalizeToolCalls(rawCalls)
		if err != nil {
			return nil, err
		}

		if len(toolCalls) > 0 || rawCalls != nil {
			return toolCalls, nil
		}
	}

	if rawMessage, ok := result.OutputValue("message"); ok {
		switch message := rawMessage.(type) {
		case map[string]any:
			if rawCalls, exists := message["tool_calls"]; exists {
				return normalizeToolCalls(rawCalls)
			}

		case JSONMap:
			if rawCalls, exists := message["tool_calls"]; exists {
				return normalizeToolCalls(rawCalls)
			}
		}
	}

	return []airuntime.ToolCall{}, nil
}

func normalizeToolCalls(raw any) ([]airuntime.ToolCall, error) {
	if raw == nil {
		return []airuntime.ToolCall{}, nil
	}

	payload, err := json.Marshal(raw)
	if err != nil {
		return nil, NewRuntimeErrorWithDetails(
			ErrorCodeNodeExecutionFailed,
			"Tool 调用结果序列化失败",
			ErrNodeExecutionFailed,
			map[string]any{
				"reason": err.Error(),
			},
		)
	}

	payload = bytes.TrimSpace(payload)
	if len(payload) == 0 || bytes.Equal(payload, []byte("null")) {
		return []airuntime.ToolCall{}, nil
	}

	var toolCalls []airuntime.ToolCall
	if err := json.Unmarshal(payload, &toolCalls); err != nil {
		return nil, NewRuntimeErrorWithDetails(
			ErrorCodeNodeExecutionFailed,
			"Tool 调用结果格式无效",
			ErrNodeExecutionFailed,
			map[string]any{
				"reason": err.Error(),
			},
		)
	}

	return toolCalls, nil
}

func buildToolNodeIndex(schema *WorkflowSchema) (map[string]WorkflowSchemaNode, error) {
	result := map[string]WorkflowSchemaNode{}

	if schema == nil {
		return result, nil
	}

	for index, node := range schema.Nodes {
		if node.Type != WorkflowNodeTypeTool {
			continue
		}

		toolName, err := readToolNodeName(node.Config)
		if err != nil {
			return nil, NewRuntimeErrorWithDetails(
				ErrorCodeInvalidInput,
				"Tool 节点缺少 tool_name",
				ErrInvalidInput,
				map[string]any{
					"node_id": node.ID,
					"index":   index,
				},
			)
		}

		if _, exists := result[toolName]; exists {
			return nil, NewRuntimeErrorWithDetails(
				ErrorCodeInvalidInput,
				"Tool 节点 tool_name 重复",
				ErrInvalidInput,
				map[string]any{
					"tool_name": toolName,
				},
			)
		}

		result[toolName] = node
	}

	return result, nil
}

func readToolNodeName(config WorkflowNodeConfig) (string, error) {
	value, exists := config["tool_name"]
	if !exists || value == nil {
		return "", fmt.Errorf("tool_name is required")
	}

	text, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("tool_name must be string")
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("tool_name cannot be empty")
	}

	return text, nil
}

func readToolNodeTimeoutMS(node WorkflowSchemaNode) int {
	if node.Config == nil {
		return 0
	}

	for _, key := range []string{"timeout_ms", "timeoutMs"} {
		value, exists := node.Config[key]
		if !exists || value == nil {
			continue
		}

		timeout, err := configValueToInt(value)
		if err != nil || timeout <= 0 {
			continue
		}

		return timeout
	}

	return 0
}

func parseToolCallArguments(raw string, toolCallID string, toolName string) (JSONMap, error) {
	text := strings.TrimSpace(raw)
	if text == "" || text == "null" {
		return JSONMap{}, nil
	}

	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.UseNumber()

	var arguments map[string]any
	if err := decoder.Decode(&arguments); err != nil {
		return nil, NewRuntimeErrorWithDetails(
			ErrorCodeNodeExecutionFailed,
			"Tool 调用 arguments 解析失败",
			ErrNodeExecutionFailed,
			map[string]any{
				"tool_call_id": toolCallID,
				"tool_name":    toolName,
				"arguments":    raw,
				"reason":       err.Error(),
			},
		)
	}

	if arguments == nil {
		return JSONMap{}, nil
	}

	return JSONMap(arguments), nil
}

func mockExecuteToolCall(request ToolCallRequest, toolNode WorkflowSchemaNode) ToolCallResult {
	startedAt := time.Now().UTC()

	output := JSONMap{
		"mock":            true,
		"tool_call_id":    request.ToolCallID,
		"tool_name":       request.ToolName,
		"tool_node_id":    toolNode.ID,
		"tool_node_label": toolNode.Label,
		"tool_node_type":  toolNode.Type.String(),
		"workspace_id":    request.WorkspaceID,
		"workflow_id":     request.WorkflowID,
		"run_id":          request.RunID,
		"llm_node_id":     request.LLMNodeID,
		"timeout_ms":      request.TimeoutMS,
		"arguments":       CloneJSONMap(request.Arguments),
		"config":          CloneJSONMap(JSONMap(toolNode.Config)),
	}

	latencyMS := time.Since(startedAt).Milliseconds()
	if latencyMS <= 0 {
		latencyMS = 1
	}

	return ToolCallResult{
		Request:   request,
		Status:    toolCallStatusSucceeded,
		Mock:      true,
		Output:    output,
		LatencyMS: latencyMS,
	}
}

func (r ToolCallResult) ToToolMessage() airuntime.ChatMessage {
	payload := JSONMap{
		"tool_call_id": r.Request.ToolCallID,
		"tool_name":    r.Request.ToolName,
		"tool_node_id": r.Request.ToolNodeID,
		"status":       r.Status,
		"mock":         r.Mock,
		"latency_ms":   r.LatencyMS,
		"output":       CloneJSONMap(r.Output),
	}

	if r.Error != "" {
		payload["error"] = r.Error
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		raw = []byte(`{"error":"marshal tool message failed"}`)
	}

	return airuntime.ChatMessage{
		Role:       "tool",
		ToolCallID: r.Request.ToolCallID,
		Content:    string(raw),
	}
}
