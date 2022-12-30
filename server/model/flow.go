package model

type FlowExecutionRequest struct {
	WorkflowName string
	FlowId       string
	Event        string
	ActionId     int
	DataMap      map[string]any
	FirstAction  bool
}

type FlowStateChangeRequest struct {
	WorkflowName string
	FlowId       string
	State        FlowState
}
