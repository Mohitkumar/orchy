package model

type FlowState string

const RUNNING FlowState = "R"
const FAILED FlowState = "F"
const COMPLETED FlowState = "C"
const WAITING_DELAY FlowState = "WD"
const WAITING_EVENT FlowState = "WE"
const PAUSED FlowState = "P"

type FlowContext struct {
	Id               string         `json:"id"`
	CurrentActionIds map[int]int    `json:"currentActionIds"`
	Data             map[string]any `json:"data"`
	State            FlowState      `json:"flowState"`
	Event            string         `json:"event"`
	ExecutedActions  map[int]bool   `json:"executedActions"`
}
