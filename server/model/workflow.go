package model

type Workflow struct {
	Name       string      `json:"name"`
	RootAction int         `json:"rootAction"`
	Actions    []ActionDef `json:"actions"`
	OnFailure  string      `json:"onFailure"`
	OnSuccess  string      `json:"onSuccess"`
}

type ActionDef struct {
	Id           int              `json:"id"`
	Type         string           `json:"type"`
	Name         string           `json:"name"`
	InputParams  map[string]any   `json:"parameters"`
	Next         map[string][]int `json:"next"`
	Expression   string           `json:"expression"`
	Join         int              `json:"join"`
	DelaySeconds int              `json:"delaySeconds"`
	Event        string           `json:"event"`
}

type WorkflowRunRequest struct {
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

type WorkflowEvent struct {
	Name   string `json:"name"`
	FlowId string `json:"flowId"`
	Event  string `json:"event"`
}
