package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"agentflow-studio/services/api/internal/model"
	"agentflow-studio/services/api/internal/repository"
	"agentflow-studio/services/api/internal/workflowruntime"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type WorkflowRunnerService struct {
	workflowRepo     *repository.WorkflowDefinitionRepository
	runRepo          *repository.WorkflowRunRepository
	workspaceService *WorkspaceService
	executorRegistry *workflowruntime.ExecutorRegistry
}

func NewWorkflowRunnerService(
	workflowRepo *repository.WorkflowDefinitionRepository,
	runRepo *repository.WorkflowRunRepository,
	workspaceService *WorkspaceService,
	executorRegistry *workflowruntime.ExecutorRegistry,
) *WorkflowRunnerService {
	return &WorkflowRunnerService{
		workflowRepo:     workflowRepo,
		runRepo:          runRepo,
		workspaceService: workspaceService,
		executorRegistry: executorRegistry,
	}
}

type StartWorkflowRunInput struct {
	WorkspaceID uuid.UUID
	WorkflowID  uuid.UUID
	UserID      uuid.UUID
	Input       workflowruntime.JSONMap
	TraceID     string
}

type WorkflowRunResult struct {
	ID          uuid.UUID               `json:"id"`
	WorkspaceID uuid.UUID               `json:"workspace_id"`
	WorkflowID  uuid.UUID               `json:"workflow_id"`
	Status      string                  `json:"status"`
	Output      workflowruntime.JSONMap `json:"output,omitempty"`
	StartedAt   *time.Time              `json:"started_at,omitempty"`
	FinishedAt  *time.Time              `json:"finished_at,omitempty"`
}

func (s *WorkflowRunnerService) RunWorkflow(
	ctx context.Context,
	input StartWorkflowRunInput,
) (*WorkflowRunResult, error) {
	if err := s.validateRunInput(input); err != nil {
		return nil, err
	}

	if err := s.requireRunPermission(ctx, input.WorkspaceID, input.UserID); err != nil {
		return nil, err
	}

	workflowDefinition, err := s.workflowRepo.FindRunnableByID(
		ctx,
		input.WorkspaceID,
		input.WorkflowID,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, workflowruntime.NewRuntimeError(
				workflowruntime.ErrorCodeWorkflowNotFound,
				"Workflow 不存在",
				workflowruntime.ErrWorkflowNotFound,
			)
		}

		return nil, err
	}

	schema, err := workflowruntime.ParseWorkflowSchemaDatatypes(workflowDefinition.SchemaJSON)
	if err != nil {
		return nil, err
	}

	schemaSnapshot, err := schema.ToDatatypesJSON()
	if err != nil {
		return nil, err
	}

	run := &model.WorkflowRun{
		WorkspaceID:    input.WorkspaceID,
		WorkflowID:     input.WorkflowID,
		TriggeredBy:    input.UserID,
		Status:         model.WorkflowRunStatusPending,
		SchemaSnapshot: schemaSnapshot,
		Input:          toDatatypesJSON(workflowruntime.CloneJSONMap(input.Input)),
	}

	if err := s.runRepo.CreateRun(ctx, nil, run); err != nil {
		return nil, workflowruntime.NewRuntimeError(
			workflowruntime.ErrorCodeCreateRunFailed,
			"Workflow Run 创建失败",
			err,
		)
	}

	if dagResult := workflowruntime.ValidateWorkflowDAG(schema); !dagResult.Valid {
		err := dagResult.RuntimeErrorOrNil()
		return nil, s.markRunFailed(ctx, run, err)
	}

	if err := s.executorRegistry.ValidateSchemaExecutors(schema); err != nil {
		return nil, s.markRunFailed(ctx, run, err)
	}

	now := time.Now().UTC()
	run.MarkRunning(now)

	if err := s.runRepo.UpdateRun(ctx, nil, run); err != nil {
		return nil, workflowruntime.NewRuntimeError(
			workflowruntime.ErrorCodeUpdateRunFailed,
			"Workflow Run 状态更新为 running 失败",
			err,
		)
	}

	executionContext := workflowruntime.NewExecutionContext(workflowruntime.ExecutionContextInput{
		WorkspaceID: input.WorkspaceID,
		WorkflowID:  input.WorkflowID,
		RunID:       run.ID,
		UserID:      input.UserID,
		Input:       workflowruntime.CloneJSONMap(input.Input),
		Variables:   workflowruntime.JSONMap{},
		TraceID:     input.TraceID,
	})

	if err := executionContext.Validate(); err != nil {
		return nil, s.markRunFailed(ctx, run, err)
	}

	executionOrder, err := buildWorkflowExecutionOrder(schema)
	if err != nil {
		return nil, s.markRunFailed(ctx, run, err)
	}

	for _, node := range executionOrder {
		if _, err := s.executeNode(ctx, executionContext, schema, node); err != nil {
			return nil, s.markRunFailed(ctx, run, err)
		}
	}

	runOutput := buildWorkflowRunOutput(executionContext, schema)
	finishedAt := time.Now().UTC()
	run.MarkSucceeded(finishedAt, toDatatypesJSON(runOutput))

	if err := s.runRepo.UpdateRun(ctx, nil, run); err != nil {
		return nil, workflowruntime.NewRuntimeError(
			workflowruntime.ErrorCodeUpdateRunFailed,
			"Workflow Run 状态更新为 succeeded 失败",
			err,
		)
	}

	return &WorkflowRunResult{
		ID:          run.ID,
		WorkspaceID: run.WorkspaceID,
		WorkflowID:  run.WorkflowID,
		Status:      string(run.Status),
		Output:      runOutput,
		StartedAt:   run.StartedAt,
		FinishedAt:  run.FinishedAt,
	}, nil
}

func (s *WorkflowRunnerService) validateRunInput(input StartWorkflowRunInput) error {
	if s.workflowRepo == nil {
		return workflowruntime.NewRuntimeError(
			workflowruntime.ErrorCodeInvalidInput,
			"WorkflowDefinitionRepository 未初始化",
			workflowruntime.ErrInvalidInput,
		)
	}

	if s.runRepo == nil {
		return workflowruntime.NewRuntimeError(
			workflowruntime.ErrorCodeInvalidInput,
			"WorkflowRunRepository 未初始化",
			workflowruntime.ErrInvalidInput,
		)
	}

	if s.workspaceService == nil {
		return workflowruntime.NewRuntimeError(
			workflowruntime.ErrorCodeInvalidInput,
			"WorkspaceService 未初始化",
			workflowruntime.ErrInvalidInput,
		)
	}

	if s.executorRegistry == nil {
		return workflowruntime.NewRuntimeError(
			workflowruntime.ErrorCodeExecutorNotFound,
			"ExecutorRegistry 未初始化",
			workflowruntime.ErrExecutorNotFound,
		)
	}

	if input.WorkspaceID == uuid.Nil {
		return workflowruntime.NewRuntimeErrorWithDetails(
			workflowruntime.ErrorCodeInvalidInput,
			"workspace_id 不能为空",
			workflowruntime.ErrInvalidInput,
			map[string]any{"field": "workspace_id"},
		)
	}

	if input.WorkflowID == uuid.Nil {
		return workflowruntime.NewRuntimeErrorWithDetails(
			workflowruntime.ErrorCodeInvalidInput,
			"workflow_id 不能为空",
			workflowruntime.ErrInvalidInput,
			map[string]any{"field": "workflow_id"},
		)
	}

	if input.UserID == uuid.Nil {
		return workflowruntime.NewRuntimeErrorWithDetails(
			workflowruntime.ErrorCodeInvalidInput,
			"user_id 不能为空",
			workflowruntime.ErrInvalidInput,
			map[string]any{"field": "user_id"},
		)
	}

	return nil
}

func (s *WorkflowRunnerService) requireRunPermission(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
) error {
	member, err := s.workspaceService.RequireMember(ctx, workspaceID, userID)
	if err != nil {
		return err
	}

	if member.Role == model.WorkspaceRoleViewer {
		return workflowruntime.NewRuntimeError(
			workflowruntime.ErrorCodePermissionDenied,
			"viewer 无权执行 Workflow",
			workflowruntime.ErrPermissionDenied,
		)
	}

	return nil
}

func (s *WorkflowRunnerService) executeNode(
	ctx context.Context,
	executionContext *workflowruntime.ExecutionContext,
	schema *workflowruntime.WorkflowSchema,
	node workflowruntime.WorkflowSchemaNode,
) (*workflowruntime.NodeExecutionResult, error) {
	inboundResults, err := collectInboundResults(executionContext, schema, node.ID)
	if err != nil {
		return nil, err
	}

	nodeInput := buildNodeRuntimeInput(executionContext, node, inboundResults)

	nodeExecution := &model.NodeExecution{
		WorkspaceID: executionContext.WorkspaceID,
		RunID:       executionContext.RunID,
		NodeID:      node.ID,
		NodeType:    node.Type.String(),
		Status:      model.NodeExecutionStatusPending,
	}

	startedAt := time.Now().UTC()
	nodeExecution.MarkRunning(startedAt, toDatatypesJSON(nodeInput))

	if err := s.runRepo.CreateNodeExecution(ctx, nil, nodeExecution); err != nil {
		return nil, err
	}

	executionInput := workflowruntime.NewNodeExecutionInput(
		executionContext,
		node,
		nodeInput,
		inboundResults,
	)

	result, err := s.executorRegistry.ExecuteNode(ctx, executionInput)
	if err != nil {
		finishedAt := time.Now().UTC()
		latencyMS := finishedAt.Sub(startedAt).Milliseconds()

		nodeExecution.MarkFailed(finishedAt, errorToDatatypesJSON(err), latencyMS)

		_ = s.runRepo.UpdateNodeExecution(ctx, nil, nodeExecution)
		return nil, err
	}

	if node.Type == workflowruntime.WorkflowNodeTypeLLM {
		toolBridge, bridgeErr := workflowruntime.BuildToolCallBridge(schema, executionContext, node, result)
		if bridgeErr != nil {
			finishedAt := time.Now().UTC()
			latencyMS := finishedAt.Sub(startedAt).Milliseconds()

			nodeExecution.MarkFailed(finishedAt, errorToDatatypesJSON(bridgeErr), latencyMS)
			_ = s.runRepo.UpdateNodeExecution(ctx, nil, nodeExecution)
			return nil, bridgeErr
		}

		if toolBridge != nil && len(toolBridge.Results) > 0 {
			if result.Output == nil {
				result.Output = workflowruntime.JSONMap{}
			}

			result.Output[workflowruntime.ToolBridgeOutputKey] = toolBridge
			result.Output[workflowruntime.ToolMessagesOutputKey] = toolBridge.ToolMessages
		}
	}

	finishedAt := time.Now().UTC()
	latencyMS := finishedAt.Sub(startedAt).Milliseconds()

	nodeExecution.MarkSucceeded(finishedAt, toDatatypesJSON(result.Output), latencyMS)

	if result.TokenUsage != nil {
		nodeExecution.TokenUsage = toDatatypesJSON(result.TokenUsage)
	}

	if err := s.runRepo.UpdateNodeExecution(ctx, nil, nodeExecution); err != nil {
		return nil, err
	}

	if err := result.ApplyToContext(executionContext); err != nil {
		return nil, err
	}

	return result, nil
}

func buildWorkflowExecutionOrder(
	schema *workflowruntime.WorkflowSchema,
) ([]workflowruntime.WorkflowSchemaNode, error) {
	nodeByID := schema.NodeMap()
	nodeOrder := map[string]int{}
	indegree := map[string]int{}
	outgoing := map[string][]string{}

	for index, node := range schema.Nodes {
		if _, exists := nodeOrder[node.ID]; !exists {
			nodeOrder[node.ID] = index
		}

		indegree[node.ID] = 0
		outgoing[node.ID] = []string{}
	}

	for _, edge := range schema.Edges {
		outgoing[edge.Source] = append(outgoing[edge.Source], edge.Target)
		indegree[edge.Target]++
	}

	ready := []string{}
	for nodeID, degree := range indegree {
		if degree == 0 {
			ready = append(ready, nodeID)
		}
	}

	sortNodeIDsBySchemaOrder(ready, nodeOrder)

	result := []workflowruntime.WorkflowSchemaNode{}

	for len(ready) > 0 {
		current := ready[0]
		ready = ready[1:]

		node, exists := nodeByID[current]
		if !exists {
			return nil, workflowruntime.NewRuntimeErrorWithDetails(
				workflowruntime.ErrorCodeInvalidDAG,
				"执行顺序中发现不存在的节点",
				workflowruntime.ErrInvalidDAG,
				map[string]any{"node_id": current},
			)
		}

		result = append(result, node)

		for _, target := range outgoing[current] {
			indegree[target]--
			if indegree[target] == 0 {
				ready = append(ready, target)
			}
		}

		sortNodeIDsBySchemaOrder(ready, nodeOrder)
	}

	if len(result) != len(indegree) {
		return nil, workflowruntime.NewRuntimeError(
			workflowruntime.ErrorCodeInvalidDAG,
			"Workflow DAG 无法生成完整执行顺序",
			workflowruntime.ErrInvalidDAG,
		)
	}

	return result, nil
}

func sortNodeIDsBySchemaOrder(nodeIDs []string, nodeOrder map[string]int) {
	sort.Slice(nodeIDs, func(i int, j int) bool {
		return nodeOrder[nodeIDs[i]] < nodeOrder[nodeIDs[j]]
	})
}

func collectInboundResults(
	executionContext *workflowruntime.ExecutionContext,
	schema *workflowruntime.WorkflowSchema,
	nodeID string,
) ([]workflowruntime.NodeExecutionResult, error) {
	inboundEdges := schema.IncomingEdges(nodeID)
	results := make([]workflowruntime.NodeExecutionResult, 0, len(inboundEdges))

	for _, edge := range inboundEdges {
		result, exists := executionContext.GetNodeOutput(edge.Source)
		if !exists {
			return nil, workflowruntime.NewRuntimeErrorWithDetails(
				workflowruntime.ErrorCodeInvalidExecutionContext,
				"上游节点输出不存在",
				workflowruntime.ErrInvalidExecutionContext,
				map[string]any{
					"source_node_id": edge.Source,
					"target_node_id": edge.Target,
					"edge_id":        edge.ID,
				},
			)
		}

		results = append(results, result)
	}

	return results, nil
}

func buildNodeRuntimeInput(
	executionContext *workflowruntime.ExecutionContext,
	node workflowruntime.WorkflowSchemaNode,
	inboundResults []workflowruntime.NodeExecutionResult,
) workflowruntime.JSONMap {
	if node.Type == workflowruntime.WorkflowNodeTypeStart {
		return executionContext.SnapshotInput()
	}

	if len(inboundResults) == 0 {
		return workflowruntime.JSONMap{}
	}

	if len(inboundResults) == 1 {
		return workflowruntime.CloneJSONMap(inboundResults[0].Output)
	}

	input := workflowruntime.JSONMap{}
	for _, inbound := range inboundResults {
		input[inbound.NodeID] = workflowruntime.CloneJSONMap(inbound.Output)
	}

	return input
}

func buildWorkflowRunOutput(
	executionContext *workflowruntime.ExecutionContext,
	schema *workflowruntime.WorkflowSchema,
) workflowruntime.JSONMap {
	for _, endNode := range schema.EndNodes() {
		result, exists := executionContext.GetNodeOutput(endNode.ID)
		if exists {
			return workflowruntime.CloneJSONMap(result.Output)
		}
	}

	return workflowruntime.JSONMap{
		"node_outputs": executionContext.SnapshotNodeOutputs(),
	}
}

func (s *WorkflowRunnerService) markRunFailed(
	ctx context.Context,
	run *model.WorkflowRun,
	cause error,
) error {
	finishedAt := time.Now().UTC()
	run.MarkFailed(finishedAt, errorToDatatypesJSON(cause))

	if err := s.runRepo.UpdateRun(ctx, nil, run); err != nil {
		return fmt.Errorf("%w; additionally failed to mark run as failed: %v", cause, err)
	}

	return cause
}

func toDatatypesJSON(value any) datatypes.JSON {
	if value == nil {
		return datatypes.JSON([]byte("null"))
	}

	raw, err := json.Marshal(value)
	if err != nil {
		return datatypes.JSON([]byte(`{"code":"JSON_MARSHAL_FAILED"}`))
	}

	return datatypes.JSON(raw)
}

func errorToDatatypesJSON(err error) datatypes.JSON {
	return toDatatypesJSON(errorToJSONMap(err))
}

func errorToJSONMap(err error) workflowruntime.JSONMap {
	if err == nil {
		return workflowruntime.JSONMap{}
	}

	payload := workflowruntime.JSONMap{
		"code":    "RUNTIME_ERROR",
		"message": err.Error(),
	}

	var runtimeErr *workflowruntime.RuntimeError
	if errors.As(err, &runtimeErr) {
		payload["code"] = string(runtimeErr.Code)

		if runtimeErr.Details != nil {
			payload["details"] = runtimeErr.Details
		}
	}

	return payload
}
