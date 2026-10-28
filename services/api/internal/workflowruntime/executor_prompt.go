package workflowruntime

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"text/template"
)

type PromptExecutor struct{}

func NewPromptExecutor() *PromptExecutor {
	return &PromptExecutor{}
}

func (e *PromptExecutor) Type() WorkflowNodeType {
	return WorkflowNodeTypePrompt
}

func (e *PromptExecutor) Validate(config WorkflowNodeConfig) error {
	templateText, err := readPromptTemplate(config)
	if err != nil {
		return err
	}

	_, err = parsePromptTemplate("prompt", templateText)
	return err
}

func (e *PromptExecutor) Execute(
	ctx context.Context,
	input NodeExecutionInput,
) (*NodeExecutionResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	templateText, err := readPromptTemplate(input.Config())
	if err != nil {
		return nil, err
	}

	tmpl, err := parsePromptTemplate(input.Node.ID, templateText)
	if err != nil {
		return nil, err
	}

	templateData := buildPromptTemplateData(input)

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, templateData); err != nil {
		return nil, NewRuntimeErrorWithDetails(
			ErrorCodePromptRenderFailed,
			"Prompt 模板渲染失败",
			err,
			map[string]any{
				"node_id":   input.Node.ID,
				"node_type": input.Node.Type.String(),
				"reason":    err.Error(),
			},
		)
	}

	renderedPrompt := buffer.String()
	variableSpec, err := readOptionalStringConfig(input.Config(), "variables")
	if err != nil {
		return nil, err
	}

	output := JSONMap{
		"prompt":          renderedPrompt,
		"rendered_prompt": renderedPrompt,
	}

	if variableSpec != "" {
		output["variable_spec"] = variableSpec
	}

	result := NewNodeExecutionResult(input.Node, output)
	result.Metadata = JSONMap{
		"executor":        "PromptExecutor",
		"template_length": len(templateText),
		"rendered_length": len(renderedPrompt),
	}

	return &result, nil
}

func readPromptTemplate(config WorkflowNodeConfig) (string, error) {
	value, exists := config["promptTemplate"]
	if !exists || value == nil {
		return "", NewRuntimeErrorWithDetails(
			ErrorCodeInvalidInput,
			"Prompt 节点缺少 promptTemplate",
			ErrInvalidInput,
			map[string]any{
				"field": "config.promptTemplate",
			},
		)
	}

	templateText, ok := value.(string)
	if !ok {
		return "", NewRuntimeErrorWithDetails(
			ErrorCodeInvalidInput,
			"Prompt 节点 promptTemplate 必须是字符串",
			ErrInvalidInput,
			map[string]any{
				"field": "config.promptTemplate",
			},
		)
	}

	if strings.TrimSpace(templateText) == "" {
		return "", NewRuntimeErrorWithDetails(
			ErrorCodeInvalidInput,
			"Prompt 节点 promptTemplate 不能为空",
			ErrInvalidInput,
			map[string]any{
				"field": "config.promptTemplate",
			},
		)
	}

	return templateText, nil
}

func readOptionalStringConfig(config WorkflowNodeConfig, key string) (string, error) {
	value, exists := config[key]
	if !exists || value == nil {
		return "", nil
	}

	text, ok := value.(string)
	if !ok {
		return "", NewRuntimeErrorWithDetails(
			ErrorCodeInvalidInput,
			"节点配置字段必须是字符串",
			ErrInvalidInput,
			map[string]any{
				"field": "config." + key,
			},
		)
	}

	return strings.TrimSpace(text), nil
}

func parsePromptTemplate(name string, templateText string) (*template.Template, error) {
	templateName := strings.TrimSpace(name)
	if templateName == "" {
		templateName = "prompt"
	}

	tmpl, err := template.New(templateName).
		Option("missingkey=error").
		Funcs(promptTemplateFuncs()).
		Parse(templateText)

	if err != nil {
		return nil, NewRuntimeErrorWithDetails(
			ErrorCodePromptRenderFailed,
			"Prompt 模板解析失败",
			err,
			map[string]any{
				"template_name": templateName,
				"reason":        err.Error(),
			},
		)
	}

	return tmpl, nil
}

func promptTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"json": func(value any) (string, error) {
			raw, err := json.Marshal(value)
			if err != nil {
				return "", err
			}

			return string(raw), nil
		},
		"jsonPretty": func(value any) (string, error) {
			raw, err := json.MarshalIndent(value, "", "  ")
			if err != nil {
				return "", err
			}

			return string(raw), nil
		},
	}
}

func buildPromptTemplateData(input NodeExecutionInput) JSONMap {
	nodeOutputs := input.ExecutionContext.SnapshotNodeOutputs()

	return JSONMap{
		"input":        input.ExecutionContext.SnapshotInput(),
		"variables":    input.ExecutionContext.SnapshotVariables(),
		"nodeOutputs":  nodeOutputs,
		"node_outputs": nodeOutputs,
		"inbound":      buildPromptInboundResults(input.InboundResults),
		"current":      CloneJSONMap(input.Input),
		"config":       CloneJSONMap(JSONMap(input.Config())),
	}
}

func buildPromptInboundResults(inboundResults []NodeExecutionResult) []JSONMap {
	items := make([]JSONMap, 0, len(inboundResults))

	for _, inbound := range inboundResults {
		items = append(items, JSONMap{
			"node_id":   inbound.NodeID,
			"node_type": inbound.NodeType.String(),
			"output":    CloneJSONMap(inbound.Output),
			"variables": CloneJSONMap(inbound.Variables),
		})
	}

	return items
}
