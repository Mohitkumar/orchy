package model

type RetryPolicy string

const RETRY_POLICY_FIXED RetryPolicy = "FIXED"
const RETRY_POLICY_BACKOFF RetryPolicy = "BACKOFF"

type TaskDef struct {
	Name              string      `json:"name"`
	RetryCount        int         `json:"retry_count"`
	RetryAfterSeconds int         `json:"retry_after_seconds"`
	RetryPolicy       RetryPolicy `json:"retry_policy"`
	TimeoutSeconds    int         `json:"timeout_seconds"`
}
