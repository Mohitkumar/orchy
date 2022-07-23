package model

type FlowState int

const RUNNING FlowState = 1
const FAILED FlowState = 2
const COMPLETED FlowState = 3
const WAITING_DELAY FlowState = 4
const WAITING_EVENT FlowState = 5

type FlowContext struct {
	Id            string         `json:"id"`
	CurrentAction int            `json:"currentAction"`
	NextAction    int            `json:"nextAction"`
	Data          map[string]any `json:"data"`
	State         FlowState      `json:"flowState"`
	TTL           uint64         `json:"ttl"`
}
