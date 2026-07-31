package workflowruntime

type DAGValidationSeverity string

const (
	DAGValidationSeverityError   DAGValidationSeverity = "error"
	DAGValidationSeverityWarning DAGValidationSeverity = "warning"
)

type DAGValidationCode string

const (
	DAGValidationCodeSchemaEmpty          DAGValidationCode = "DAG_SCHEMA_EMPTY"
	DAGValidationCodeStartCountInvalid    DAGValidationCode = "DAG_START_COUNT_INVALID"
	DAGValidationCodeEndCountInvalid      DAGValidationCode = "DAG_END_COUNT_INVALID"
	DAGValidationCodeNodeIDEmpty          DAGValidationCode = "DAG_NODE_ID_EMPTY"
	DAGValidationCodeDuplicateNodeID      DAGValidationCode = "DAG_DUPLICATE_NODE_ID"
	DAGValidationCodeEdgeIDEmpty          DAGValidationCode = "DAG_EDGE_ID_EMPTY"
	DAGValidationCodeDuplicateEdgeID      DAGValidationCode = "DAG_DUPLICATE_EDGE_ID"
	DAGValidationCodeEdgeSourceMissing    DAGValidationCode = "DAG_EDGE_SOURCE_MISSING"
	DAGValidationCodeEdgeTargetMissing    DAGValidationCode = "DAG_EDGE_TARGET_MISSING"
	DAGValidationCodeEdgeFromEnd          DAGValidationCode = "DAG_EDGE_FROM_END"
	DAGValidationCodeEdgeToStart          DAGValidationCode = "DAG_EDGE_TO_START"
	DAGValidationCodeSelfLoop             DAGValidationCode = "DAG_SELF_LOOP"
	DAGValidationCodeCycleDetected        DAGValidationCode = "DAG_CYCLE_DETECTED"
	DAGValidationCodeUnreachableFromStart DAGValidationCode = "DAG_UNREACHABLE_FROM_START"
	DAGValidationCodeCannotReachEnd       DAGValidationCode = "DAG_CANNOT_REACH_END"
)

type DAGValidationIssue struct {
	ID       string                `json:"id"`
	Severity DAGValidationSeverity `json:"severity"`
	Code     DAGValidationCode     `json:"code"`
	Message  string                `json:"message"`
	NodeID   string                `json:"nodeId,omitempty"`
	EdgeID   string                `json:"edgeId,omitempty"`
	Details  map[string]any        `json:"details,omitempty"`
}

type DAGValidationResult struct {
	Valid        bool                 `json:"valid"`
	ErrorCount   int                  `json:"error_count"`
	WarningCount int                  `json:"warning_count"`
	Issues       []DAGValidationIssue `json:"issues"`
}

func (r DAGValidationResult) RuntimeErrorOrNil() error {
	if r.Valid {
		return nil
	}

	return NewRuntimeErrorWithDetails(
		ErrorCodeInvalidDAG,
		"Workflow DAG 校验失败",
		ErrInvalidDAG,
		r,
	)
}
