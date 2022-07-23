package model

type Workflow struct {
	Name       string      `json:"name"`
	RootAction int         `json:"rootAction"`
	Actions    []ActionDef `json:"actions"`
}

type ActionDef struct {
	Id           int            `json:"id"`
	Type         string         `json:"type"`
	Name         string         `json:"name"`
	InputParams  map[string]any `json:"parameters"`
	Next         int            `json:"next"`
	Expression   string         `json:"expression"`
	Cases        map[string]int `json:"cases"`
	Forks        []int          `json:"forks"`
	Join         int            `json:"join"`
	DelaySeconds int            `json:"delaySeconds"`
}

type WorkflowRunRequest struct {
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}
