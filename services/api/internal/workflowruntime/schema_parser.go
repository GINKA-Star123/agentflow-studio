package workflowruntime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/datatypes"
)

func ParseWorkflowSchemaJSON(raw []byte) (*WorkflowSchema, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil, NewRuntimeError(
			ErrorCodeInvalidSchema,
			"Workflow Schema 不能为空",
			ErrInvalidSchema,
		)
	}

	if bytes.Equal(raw, []byte("null")) {
		return nil, NewRuntimeError(
			ErrorCodeInvalidSchema,
			"Workflow Schema 不能为 null",
			ErrInvalidSchema,
		)
	}

	var schema WorkflowSchema
	if err := json.Unmarshal(raw, &schema); err != nil {
		return nil, NewRuntimeErrorWithDetails(
			ErrorCodeInvalidSchema,
			"Workflow Schema JSON 格式错误",
			err,
			map[string]any{
				"reason": err.Error(),
			},
		)
	}

	schema.Normalize()

	if err := schema.ValidateShape(); err != nil {
		return nil, err
	}

	return &schema, nil
}

func ParseWorkflowSchemaString(value string) (*WorkflowSchema, error) {
	return ParseWorkflowSchemaJSON([]byte(value))
}

func ParseWorkflowSchemaDatatypes(raw datatypes.JSON) (*WorkflowSchema, error) {
	return ParseWorkflowSchemaJSON([]byte(raw))
}

func (s *WorkflowSchema) ToDatatypesJSON() (datatypes.JSON, error) {
	if s == nil {
		return nil, NewRuntimeError(
			ErrorCodeInvalidSchema,
			"Workflow Schema 不能为空",
			ErrInvalidSchema,
		)
	}

	s.Normalize()

	if err := s.ValidateShape(); err != nil {
		return nil, err
	}

	raw, err := json.Marshal(s)
	if err != nil {
		return nil, NewRuntimeErrorWithDetails(
			ErrorCodeInvalidSchema,
			"Workflow Schema 序列化失败",
			err,
			map[string]any{
				"reason": err.Error(),
			},
		)
	}

	return datatypes.JSON(raw), nil
}

func (s *WorkflowSchema) Clone() (*WorkflowSchema, error) {
	raw, err := s.ToDatatypesJSON()
	if err != nil {
		return nil, err
	}

	return ParseWorkflowSchemaDatatypes(raw)
}

func (s *WorkflowSchema) Normalize() {
	if s == nil {
		return
	}

	s.SchemaVersion = strings.TrimSpace(s.SchemaVersion)
	s.Name = strings.TrimSpace(s.Name)

	if s.Nodes == nil {
		s.Nodes = []WorkflowSchemaNode{}
	}

	for i := range s.Nodes {
		node := &s.Nodes[i]

		node.ID = strings.TrimSpace(node.ID)
		node.Type = WorkflowNodeType(strings.TrimSpace(string(node.Type)))
		node.Label = strings.TrimSpace(node.Label)
		node.Description = strings.TrimSpace(node.Description)

		if node.Config == nil {
			node.Config = WorkflowNodeConfig{}
		}
	}

	if s.Edges == nil {
		s.Edges = []WorkflowSchemaEdge{}
	}

	for i := range s.Edges {
		edge := &s.Edges[i]

		edge.ID = strings.TrimSpace(edge.ID)
		edge.Source = strings.TrimSpace(edge.Source)
		edge.Target = strings.TrimSpace(edge.Target)
		edge.SourceHandle = normalizeOptionalString(edge.SourceHandle)
		edge.TargetHandle = normalizeOptionalString(edge.TargetHandle)
		edge.Type = normalizeOptionalString(edge.Type)
	}

	s.Summary = s.BuildSummary()
}

func (s *WorkflowSchema) ValidateShape() error {
	if s == nil {
		return NewRuntimeError(
			ErrorCodeInvalidSchema,
			"Workflow Schema 不能为空",
			ErrInvalidSchema,
		)
	}

	if s.SchemaVersion == "" {
		return NewRuntimeErrorWithDetails(
			ErrorCodeInvalidSchema,
			"Workflow Schema 缺少 schema_version",
			ErrInvalidSchema,
			map[string]any{
				"field": "schema_version",
			},
		)
	}

	if s.SchemaVersion != WorkflowSchemaVersion {
		return NewRuntimeErrorWithDetails(
			ErrorCodeInvalidSchema,
			"Workflow Schema 版本不支持",
			ErrInvalidSchema,
			map[string]any{
				"field":    "schema_version",
				"expected": WorkflowSchemaVersion,
				"actual":   s.SchemaVersion,
			},
		)
	}

	for index, node := range s.Nodes {
		if node.ID == "" {
			return NewRuntimeErrorWithDetails(
				ErrorCodeInvalidSchema,
				"Workflow 节点缺少 id",
				ErrInvalidSchema,
				map[string]any{
					"field": fmt.Sprintf("nodes[%d].id", index),
				},
			)
		}

		if !node.Type.IsSupported() {
			return NewRuntimeErrorWithDetails(
				ErrorCodeUnsupportedNodeType,
				"不支持的 Workflow 节点类型",
				ErrUnsupportedNodeType,
				map[string]any{
					"field": fmt.Sprintf("nodes[%d].type", index),
					"value": node.Type.String(),
				},
			)
		}
	}

	for index, edge := range s.Edges {
		if edge.ID == "" {
			return NewRuntimeErrorWithDetails(
				ErrorCodeInvalidSchema,
				"Workflow 连线缺少 id",
				ErrInvalidSchema,
				map[string]any{
					"field": fmt.Sprintf("edges[%d].id", index),
				},
			)
		}

		if edge.Source == "" {
			return NewRuntimeErrorWithDetails(
				ErrorCodeInvalidSchema,
				"Workflow 连线缺少 source",
				ErrInvalidSchema,
				map[string]any{
					"field": fmt.Sprintf("edges[%d].source", index),
				},
			)
		}

		if edge.Target == "" {
			return NewRuntimeErrorWithDetails(
				ErrorCodeInvalidSchema,
				"Workflow 连线缺少 target",
				ErrInvalidSchema,
				map[string]any{
					"field": fmt.Sprintf("edges[%d].target", index),
				},
			)
		}
	}

	return nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
