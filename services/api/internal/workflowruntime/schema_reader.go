package workflowruntime

func (s *WorkflowSchema) BuildSummary() WorkflowSchemaSummary {
	if s == nil {
		return WorkflowSchemaSummary{}
	}

	return WorkflowSchemaSummary{
		NodeCount:  len(s.Nodes),
		EdgeCount:  len(s.Edges),
		StartCount: s.CountNodeType(WorkflowNodeTypeStart),
		EndCount:   s.CountNodeType(WorkflowNodeTypeEnd),
	}
}

func (s *WorkflowSchema) NodeMap() map[string]WorkflowSchemaNode {
	result := map[string]WorkflowSchemaNode{}

	if s == nil {
		return result
	}

	for _, node := range s.Nodes {
		result[node.ID] = node
	}

	return result
}

func (s *WorkflowSchema) EdgeMap() map[string]WorkflowSchemaEdge {
	result := map[string]WorkflowSchemaEdge{}

	if s == nil {
		return result
	}

	for _, edge := range s.Edges {
		result[edge.ID] = edge
	}

	return result
}

func (s *WorkflowSchema) NodeByID(nodeID string) (WorkflowSchemaNode, bool) {
	if s == nil {
		return WorkflowSchemaNode{}, false
	}

	for _, node := range s.Nodes {
		if node.ID == nodeID {
			return node, true
		}
	}

	return WorkflowSchemaNode{}, false
}

func (s *WorkflowSchema) EdgeByID(edgeID string) (WorkflowSchemaEdge, bool) {
	if s == nil {
		return WorkflowSchemaEdge{}, false
	}

	for _, edge := range s.Edges {
		if edge.ID == edgeID {
			return edge, true
		}
	}

	return WorkflowSchemaEdge{}, false
}

func (s *WorkflowSchema) HasNode(nodeID string) bool {
	_, ok := s.NodeByID(nodeID)
	return ok
}

func (s *WorkflowSchema) CountNodeType(nodeType WorkflowNodeType) int {
	if s == nil {
		return 0
	}

	count := 0
	for _, node := range s.Nodes {
		if node.Type == nodeType {
			count++
		}
	}

	return count
}

func (s *WorkflowSchema) StartNodes() []WorkflowSchemaNode {
	return s.NodesByType(WorkflowNodeTypeStart)
}

func (s *WorkflowSchema) EndNodes() []WorkflowSchemaNode {
	return s.NodesByType(WorkflowNodeTypeEnd)
}

func (s *WorkflowSchema) NodesByType(nodeType WorkflowNodeType) []WorkflowSchemaNode {
	result := []WorkflowSchemaNode{}

	if s == nil {
		return result
	}

	for _, node := range s.Nodes {
		if node.Type == nodeType {
			result = append(result, node)
		}
	}

	return result
}

func (s *WorkflowSchema) IncomingEdges(nodeID string) []WorkflowSchemaEdge {
	result := []WorkflowSchemaEdge{}

	if s == nil {
		return result
	}

	for _, edge := range s.Edges {
		if edge.Target == nodeID {
			result = append(result, edge)
		}
	}

	return result
}

func (s *WorkflowSchema) OutgoingEdges(nodeID string) []WorkflowSchemaEdge {
	result := []WorkflowSchemaEdge{}

	if s == nil {
		return result
	}

	for _, edge := range s.Edges {
		if edge.Source == nodeID {
			result = append(result, edge)
		}
	}

	return result
}

func (s *WorkflowSchema) NodeIDs() []string {
	result := []string{}

	if s == nil {
		return result
	}

	for _, node := range s.Nodes {
		result = append(result, node.ID)
	}

	return result
}

func (s *WorkflowSchema) EdgeIDs() []string {
	result := []string{}

	if s == nil {
		return result
	}

	for _, edge := range s.Edges {
		result = append(result, edge.ID)
	}

	return result
}
