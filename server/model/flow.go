package model

type FlowExecutionType string

const NEW_FLOW_EXECUTION FlowExecutionType = "NEW"
const RETRY_FLOW_EXECUTION FlowExecutionType = "RETRY"
const RESUME_FLOW_EXECUTION FlowExecutionType = "RESUME"

type FlowExecutionRequest struct {
	WorkflowName string
	FlowId       string
	Event        string
	ActionId     int
	DataMap      map[string]any
	TryCount     int
	RequestType  FlowExecutionType
}

type FlowStateChangeRequest struct {
	WorkflowName string
	FlowId       string
	State        FlowState
}