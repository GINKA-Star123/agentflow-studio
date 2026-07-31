package workflowruntime

import "fmt"

type DAGValidator struct{}

func NewDAGValidator() *DAGValidator {
	return &DAGValidator{}
}

func ValidateWorkflowDAG(schema *WorkflowSchema) DAGValidationResult {
	return NewDAGValidator().Validate(schema)
}

func ValidateWorkflowDAGOrError(schema *WorkflowSchema) error {
	result := ValidateWorkflowDAG(schema)
	return result.RuntimeErrorOrNil()
}

func (v *DAGValidator) Validate(schema *WorkflowSchema) DAGValidationResult {
	result := DAGValidationResult{
		Issues: []DAGValidationIssue{},
	}

	if schema == nil {
		result.addIssue(
			DAGValidationSeverityError,
			DAGValidationCodeSchemaEmpty,
			"Workflow Schema 不能为空",
			"",
			"",
			nil,
		)
		result.finalize()
		return result
	}

	schema.Normalize()

	nodeByID := v.buildNodeIndex(schema, &result)
	startNodes := schema.StartNodes()
	endNodes := schema.EndNodes()

	if len(startNodes) != 1 {
		result.addIssue(
			DAGValidationSeverityError,
			DAGValidationCodeStartCountInvalid,
			fmt.Sprintf("Start 节点必须且只能有 1 个，当前有 %d 个", len(startNodes)),
			"",
			"",
			map[string]any{
				"actual":   len(startNodes),
				"expected": 1,
			},
		)
	}

	if len(endNodes) != 1 {
		result.addIssue(
			DAGValidationSeverityError,
			DAGValidationCodeEndCountInvalid,
			fmt.Sprintf("End 节点必须且只能有 1 个，当前有 %d 个", len(endNodes)),
			"",
			"",
			map[string]any{
				"actual":   len(endNodes),
				"expected": 1,
			},
		)
	}

	graph, reverseGraph := v.buildGraphs(schema, nodeByID, &result)

	if cyclePath, ok := detectCycle(graph, schema.NodeIDs()); ok {
		result.addIssue(
			DAGValidationSeverityError,
			DAGValidationCodeCycleDetected,
			"Workflow DAG 中存在环",
			"",
			"",
			map[string]any{
				"path": cyclePath,
			},
		)
	}

	if len(startNodes) == 1 {
		v.checkReachableFromStart(schema, graph, startNodes[0].ID, &result)
	}

	if len(endNodes) == 1 {
		v.checkCanReachEnd(schema, reverseGraph, endNodes[0].ID, &result)
	}

	result.finalize()
	return result
}

func (v *DAGValidator) buildNodeIndex(
	schema *WorkflowSchema,
	result *DAGValidationResult,
) map[string]WorkflowSchemaNode {
	nodeByID := map[string]WorkflowSchemaNode{}

	for index, node := range schema.Nodes {
		if node.ID == "" {
			result.addIssue(
				DAGValidationSeverityError,
				DAGValidationCodeNodeIDEmpty,
				"Workflow 节点 ID 不能为空",
				"",
				"",
				map[string]any{
					"index": index,
				},
			)
			continue
		}

		if _, exists := nodeByID[node.ID]; exists {
			result.addIssue(
				DAGValidationSeverityError,
				DAGValidationCodeDuplicateNodeID,
				fmt.Sprintf("Workflow 节点 ID 重复：%s", node.ID),
				node.ID,
				"",
				map[string]any{
					"index": nodeIndexList(schema.Nodes, node.ID),
				},
			)
			continue
		}

		nodeByID[node.ID] = node
	}

	return nodeByID
}

func (v *DAGValidator) buildGraphs(
	schema *WorkflowSchema,
	nodeByID map[string]WorkflowSchemaNode,
	result *DAGValidationResult,
) (map[string][]string, map[string][]string) {
	graph := map[string][]string{}
	reverseGraph := map[string][]string{}
	edgeIDSet := map[string]struct{}{}

	for nodeID := range nodeByID {
		graph[nodeID] = []string{}
		reverseGraph[nodeID] = []string{}
	}

	for index, edge := range schema.Edges {
		if edge.ID == "" {
			result.addIssue(
				DAGValidationSeverityError,
				DAGValidationCodeEdgeIDEmpty,
				"Workflow 连线 ID 不能为空",
				"",
				"",
				map[string]any{
					"index": index,
				},
			)
		} else {
			if _, exists := edgeIDSet[edge.ID]; exists {
				result.addIssue(
					DAGValidationSeverityError,
					DAGValidationCodeDuplicateEdgeID,
					fmt.Sprintf("Workflow 连线 ID 重复：%s", edge.ID),
					"",
					edge.ID,
					map[string]any{
						"index": edgeIndexList(schema.Edges, edge.ID),
					},
				)
			}

			edgeIDSet[edge.ID] = struct{}{}
		}

		sourceNode, sourceExists := nodeByID[edge.Source]
		targetNode, targetExists := nodeByID[edge.Target]
		edgeCanEnterGraph := true

		if edge.Source == "" || !sourceExists {
			result.addIssue(
				DAGValidationSeverityError,
				DAGValidationCodeEdgeSourceMissing,
				fmt.Sprintf("连线 %s 的 source 节点不存在：%s", edge.ID, edge.Source),
				"",
				edge.ID,
				map[string]any{
					"source": edge.Source,
				},
			)
			edgeCanEnterGraph = false
		}

		if edge.Target == "" || !targetExists {
			result.addIssue(
				DAGValidationSeverityError,
				DAGValidationCodeEdgeTargetMissing,
				fmt.Sprintf("连线 %s 的 target 节点不存在：%s", edge.ID, edge.Target),
				"",
				edge.ID,
				map[string]any{
					"target": edge.Target,
				},
			)
			edgeCanEnterGraph = false
		}

		if edge.Source != "" && edge.Source == edge.Target {
			result.addIssue(
				DAGValidationSeverityError,
				DAGValidationCodeSelfLoop,
				fmt.Sprintf("连线 %s 不能连接到自身", edge.ID),
				edge.Source,
				edge.ID,
				nil,
			)
			edgeCanEnterGraph = false
		}

		if sourceExists && sourceNode.Type == WorkflowNodeTypeEnd {
			result.addIssue(
				DAGValidationSeverityError,
				DAGValidationCodeEdgeFromEnd,
				fmt.Sprintf("End 节点 %s 不能作为连线 source", sourceNode.ID),
				sourceNode.ID,
				edge.ID,
				nil,
			)
			edgeCanEnterGraph = false
		}

		if targetExists && targetNode.Type == WorkflowNodeTypeStart {
			result.addIssue(
				DAGValidationSeverityError,
				DAGValidationCodeEdgeToStart,
				fmt.Sprintf("Start 节点 %s 不能作为连线 target", targetNode.ID),
				targetNode.ID,
				edge.ID,
				nil,
			)
			edgeCanEnterGraph = false
		}

		if !edgeCanEnterGraph {
			continue
		}

		graph[edge.Source] = append(graph[edge.Source], edge.Target)
		reverseGraph[edge.Target] = append(reverseGraph[edge.Target], edge.Source)
	}

	return graph, reverseGraph
}

func (v *DAGValidator) checkReachableFromStart(
	schema *WorkflowSchema,
	graph map[string][]string,
	startNodeID string,
	result *DAGValidationResult,
) {
	reachable := traverseGraph(graph, startNodeID)

	for _, node := range schema.Nodes {
		if node.ID == "" {
			continue
		}

		if !reachable[node.ID] {
			result.addIssue(
				DAGValidationSeverityError,
				DAGValidationCodeUnreachableFromStart,
				fmt.Sprintf("节点 %s 无法从 Start 节点到达", node.ID),
				node.ID,
				"",
				map[string]any{
					"start_node_id": startNodeID,
				},
			)
		}
	}
}

func (v *DAGValidator) checkCanReachEnd(
	schema *WorkflowSchema,
	reverseGraph map[string][]string,
	endNodeID string,
	result *DAGValidationResult,
) {
	canReachEnd := traverseGraph(reverseGraph, endNodeID)

	for _, node := range schema.Nodes {
		if node.ID == "" || node.Type == WorkflowNodeTypeEnd {
			continue
		}

		if !canReachEnd[node.ID] {
			result.addIssue(
				DAGValidationSeverityError,
				DAGValidationCodeCannotReachEnd,
				fmt.Sprintf("节点 %s 无法到达 End 节点", node.ID),
				node.ID,
				"",
				map[string]any{
					"end_node_id": endNodeID,
				},
			)
		}
	}
}

func detectCycle(graph map[string][]string, nodeIDs []string) ([]string, bool) {
	const (
		unvisited = 0
		visiting  = 1
		visited   = 2
	)

	color := map[string]int{}
	stack := []string{}
	stackIndex := map[string]int{}

	var visit func(nodeID string) ([]string, bool)

	visit = func(nodeID string) ([]string, bool) {
		color[nodeID] = visiting
		stackIndex[nodeID] = len(stack)
		stack = append(stack, nodeID)

		for _, nextNodeID := range graph[nodeID] {
			switch color[nextNodeID] {
			case unvisited:
				if cyclePath, ok := visit(nextNodeID); ok {
					return cyclePath, true
				}

			case visiting:
				startIndex := stackIndex[nextNodeID]
				cyclePath := append([]string{}, stack[startIndex:]...)
				cyclePath = append(cyclePath, nextNodeID)
				return cyclePath, true
			}
		}

		stack = stack[:len(stack)-1]
		delete(stackIndex, nodeID)
		color[nodeID] = visited

		return nil, false
	}

	for _, nodeID := range nodeIDs {
		if nodeID == "" {
			continue
		}

		if color[nodeID] != unvisited {
			continue
		}

		if cyclePath, ok := visit(nodeID); ok {
			return cyclePath, true
		}
	}

	return nil, false
}

func traverseGraph(graph map[string][]string, startNodeID string) map[string]bool {
	visited := map[string]bool{}

	if startNodeID == "" {
		return visited
	}

	stack := []string{startNodeID}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if visited[current] {
			continue
		}

		visited[current] = true

		for _, next := range graph[current] {
			if !visited[next] {
				stack = append(stack, next)
			}
		}
	}

	return visited
}

func nodeIndexList(nodes []WorkflowSchemaNode, nodeID string) []int {
	indexes := []int{}

	for index, node := range nodes {
		if node.ID == nodeID {
			indexes = append(indexes, index)
		}
	}

	return indexes
}

func edgeIndexList(edges []WorkflowSchemaEdge, edgeID string) []int {
	indexes := []int{}

	for index, edge := range edges {
		if edge.ID == edgeID {
			indexes = append(indexes, index)
		}
	}

	return indexes
}

func (r *DAGValidationResult) addIssue(
	severity DAGValidationSeverity,
	code DAGValidationCode,
	message string,
	nodeID string,
	edgeID string,
	details map[string]any,
) {
	scope := "global"
	if nodeID != "" {
		scope = "node-" + nodeID
	} else if edgeID != "" {
		scope = "edge-" + edgeID
	}

	r.Issues = append(r.Issues, DAGValidationIssue{
		ID:       fmt.Sprintf("%03d-%s-%s", len(r.Issues)+1, code, scope),
		Severity: severity,
		Code:     code,
		Message:  message,
		NodeID:   nodeID,
		EdgeID:   edgeID,
		Details:  details,
	})
}

func (r *DAGValidationResult) finalize() {
	r.ErrorCount = 0
	r.WarningCount = 0

	for _, issue := range r.Issues {
		switch issue.Severity {
		case DAGValidationSeverityError:
			r.ErrorCount++
		case DAGValidationSeverityWarning:
			r.WarningCount++
		}
	}

	r.Valid = r.ErrorCount == 0
}
