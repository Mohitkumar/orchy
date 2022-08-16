package model

type ActionExecutionRequest struct {
	WorkflowName string `json:"wfName"`
	TaskName     string `json:"taskName"`
	FlowId       string `json:"flowId"`
	ActionId     int    `json:"actionId"`
	TryNumber    int    `json:"tryNumber"`
}
