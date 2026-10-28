package workflowruntime

import (
	"encoding/json"
	"testing"

	"agentflow-studio/services/api/internal/airuntime"

	"github.com/google/uuid"
)

func TestBuildToolCallBridge(t *testing.T) {
	schema := &WorkflowSchema{
		SchemaVersion: WorkflowSchemaVersion,
		Nodes: []WorkflowSchemaNode{
			{
				ID:   "llm_1",
				Type: WorkflowNodeTypeLLM,
			},
			{
				ID:   "tool_1",
				Type: WorkflowNodeTypeTool,
				Config: WorkflowNodeConfig{
					"tool_name":  "search_docs",
					"timeout_ms": 30000,
				},
			},
		},
	}

	executionContext := NewExecutionContext(ExecutionContextInput{
		WorkspaceID: uuid.New(),
		WorkflowID:  uuid.New(),
		RunID:       uuid.New(),
		UserID:      uuid.New(),
		Input:       JSONMap{},
		Variables:   JSONMap{},
	})

	llmResult := &NodeExecutionResult{
		NodeID:   "llm_1",
		NodeType: WorkflowNodeTypeLLM,
		Output: JSONMap{
			"tool_calls": []airuntime.ToolCall{
				{
					ID:   "call_1",
					Type: "function",
					Function: airuntime.ToolCallFunction{
						Name:      "search_docs",
						Arguments: `{"query":"hello"}`,
					},
				},
			},
		},
	}

	bridge, err := BuildToolCallBridge(schema, executionContext, schema.Nodes[0], llmResult)
	if err != nil {
		t.Fatalf("BuildToolCallBridge returned error: %v", err)
	}

	if bridge == nil {
		t.Fatalf("BuildToolCallBridge returned nil bridge")
	}

	if len(bridge.Results) != 1 {
		t.Fatalf("expected 1 tool call result, got %d", len(bridge.Results))
	}

	if len(bridge.ToolMessages) != 1 {
		t.Fatalf("expected 1 tool message, got %d", len(bridge.ToolMessages))
	}

	message := bridge.ToolMessages[0]
	if message.Role != "tool" {
		t.Fatalf("expected tool role, got %s", message.Role)
	}

	if message.ToolCallID != "call_1" {
		t.Fatalf("expected tool_call_id call_1, got %s", message.ToolCallID)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(message.Content), &payload); err != nil {
		t.Fatalf("failed to decode tool message content: %v", err)
	}

	if payload["tool_name"] != "search_docs" {
		t.Fatalf("expected tool_name search_docs, got %v", payload["tool_name"])
	}

	if payload["status"] != toolCallStatusSucceeded {
		t.Fatalf("expected status %s, got %v", toolCallStatusSucceeded, payload["status"])
	}
}

func TestBuildToolCallBridgeFromAssistantMessageToolCalls(t *testing.T) {
	schema := &WorkflowSchema{
		SchemaVersion: WorkflowSchemaVersion,
		Nodes: []WorkflowSchemaNode{
			{
				ID:   "llm_1",
				Type: WorkflowNodeTypeLLM,
			},
			{
				ID:   "tool_1",
				Type: WorkflowNodeTypeTool,
				Config: WorkflowNodeConfig{
					"tool_name": "search_docs",
				},
			},
		},
	}

	executionContext := NewExecutionContext(ExecutionContextInput{
		WorkspaceID: uuid.New(),
		WorkflowID:  uuid.New(),
		RunID:       uuid.New(),
		UserID:      uuid.New(),
		Input:       JSONMap{},
		Variables:   JSONMap{},
	})

	llmResult := &NodeExecutionResult{
		NodeID:   "llm_1",
		NodeType: WorkflowNodeTypeLLM,
		Output: JSONMap{
			"message": JSONMap{
				"role":    "assistant",
				"content": "tool call ready",
				"tool_calls": []airuntime.ToolCall{
					{
						ID:   "call_2",
						Type: "function",
						Function: airuntime.ToolCallFunction{
							Name:      "search_docs",
							Arguments: `{"query":"phase5"}`,
						},
					},
				},
			},
		},
	}

	bridge, err := BuildToolCallBridge(schema, executionContext, schema.Nodes[0], llmResult)
	if err != nil {
		t.Fatalf("BuildToolCallBridge returned error: %v", err)
	}

	if bridge == nil {
		t.Fatalf("BuildToolCallBridge returned nil bridge")
	}

	if !bridge.Pending {
		t.Fatalf("expected pending bridge result")
	}

	if len(bridge.Results) != 1 {
		t.Fatalf("expected 1 tool result, got %d", len(bridge.Results))
	}
}

func TestBuildToolCallBridgeWithoutToolCalls(t *testing.T) {
	schema := &WorkflowSchema{
		SchemaVersion: WorkflowSchemaVersion,
		Nodes: []WorkflowSchemaNode{
			{
				ID:   "llm_1",
				Type: WorkflowNodeTypeLLM,
			},
		},
	}

	executionContext := NewExecutionContext(ExecutionContextInput{
		WorkspaceID: uuid.New(),
		WorkflowID:  uuid.New(),
		RunID:       uuid.New(),
		UserID:      uuid.New(),
		Input:       JSONMap{},
		Variables:   JSONMap{},
	})

	llmResult := &NodeExecutionResult{
		NodeID:   "llm_1",
		NodeType: WorkflowNodeTypeLLM,
		Output: JSONMap{
			"message": JSONMap{
				"role":    "assistant",
				"content": "plain response",
			},
		},
	}

	bridge, err := BuildToolCallBridge(schema, executionContext, schema.Nodes[0], llmResult)
	if err != nil {
		t.Fatalf("BuildToolCallBridge returned error: %v", err)
	}

	if bridge == nil {
		t.Fatalf("BuildToolCallBridge returned nil bridge")
	}

	if bridge.Pending {
		t.Fatalf("expected pending=false when no tool calls exist")
	}

	if len(bridge.Results) != 0 {
		t.Fatalf("expected 0 tool results, got %d", len(bridge.Results))
	}

	if len(bridge.ToolMessages) != 0 {
		t.Fatalf("expected 0 tool messages, got %d", len(bridge.ToolMessages))
	}
}

func TestBuildToolCallBridgeInvalidArguments(t *testing.T) {
	schema := &WorkflowSchema{
		SchemaVersion: WorkflowSchemaVersion,
		Nodes: []WorkflowSchemaNode{
			{
				ID:   "llm_1",
				Type: WorkflowNodeTypeLLM,
			},
			{
				ID:   "tool_1",
				Type: WorkflowNodeTypeTool,
				Config: WorkflowNodeConfig{
					"tool_name": "search_docs",
				},
			},
		},
	}

	executionContext := NewExecutionContext(ExecutionContextInput{
		WorkspaceID: uuid.New(),
		WorkflowID:  uuid.New(),
		RunID:       uuid.New(),
		UserID:      uuid.New(),
		Input:       JSONMap{},
		Variables:   JSONMap{},
	})

	llmResult := &NodeExecutionResult{
		NodeID:   "llm_1",
		NodeType: WorkflowNodeTypeLLM,
		Output: JSONMap{
			"tool_calls": []airuntime.ToolCall{
				{
					ID:   "call_3",
					Type: "function",
					Function: airuntime.ToolCallFunction{
						Name:      "search_docs",
						Arguments: `{"query":`,
					},
				},
			},
		},
	}

	bridge, err := BuildToolCallBridge(schema, executionContext, schema.Nodes[0], llmResult)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if bridge != nil {
		t.Fatalf("expected nil bridge, got %#v", bridge)
	}
}

func TestBuildToolCallBridgeMissingToolNode(t *testing.T) {
	schema := &WorkflowSchema{
		SchemaVersion: WorkflowSchemaVersion,
		Nodes: []WorkflowSchemaNode{
			{
				ID:   "llm_1",
				Type: WorkflowNodeTypeLLM,
			},
		},
	}

	executionContext := NewExecutionContext(ExecutionContextInput{
		WorkspaceID: uuid.New(),
		WorkflowID:  uuid.New(),
		RunID:       uuid.New(),
		UserID:      uuid.New(),
		Input:       JSONMap{},
		Variables:   JSONMap{},
	})

	llmResult := &NodeExecutionResult{
		NodeID:   "llm_1",
		NodeType: WorkflowNodeTypeLLM,
		Output: JSONMap{
			"tool_calls": []airuntime.ToolCall{
				{
					ID:   "call_1",
					Type: "function",
					Function: airuntime.ToolCallFunction{
						Name:      "search_docs",
						Arguments: `{"query":"hello"}`,
					},
				},
			},
		},
	}

	bridge, err := BuildToolCallBridge(schema, executionContext, schema.Nodes[0], llmResult)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if bridge != nil {
		t.Fatalf("expected nil bridge, got %#v", bridge)
	}
}
