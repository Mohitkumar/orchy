package worker

type RetryPolicy string

const RETRY_POLICY_FIXED RetryPolicy = "FIXED"
const RETRY_POLICY_BACKOFF RetryPolicy = "BACKOFF"

type WorkerWrapper struct {
	worker            func(map[string]any) (map[string]any, error)
	retryCount        int
	retryAfterSeconds int
	retryPolicy       RetryPolicy
	timeoutSeconds    int
}

func NewDefaultWorker(w func(map[string]any) (map[string]any, error)) *WorkerWrapper {
	return &WorkerWrapper{
		worker:            w,
		retryCount:        1,
		retryAfterSeconds: 5,
		retryPolicy:       RETRY_POLICY_FIXED,
		timeoutSeconds:    20,
	}
}

func (w *WorkerWrapper) WithRetryCount(count int) *WorkerWrapper {
	w.retryCount = count
	return w
}

func (w *WorkerWrapper) WithRetryInterval(retryInterval int) *WorkerWrapper {
	w.retryAfterSeconds = retryInterval
	return w
}

func (w *WorkerWrapper) WithRetryPolicy(policy string) *WorkerWrapper {
	w.retryPolicy = RetryPolicy(policy)
	return w
}

func (w *WorkerWrapper) WithTimeoutSeconds(timeout int) *WorkerWrapper {
	w.timeoutSeconds = timeout
	return w
}
