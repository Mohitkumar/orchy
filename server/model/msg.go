package model

type ActionExecutionRequest struct {
	WorkflowName string `json:"wfName"`
	FlowId       string `json:"flowId"`
	ActionId     int    `json:"actionId"`
}
