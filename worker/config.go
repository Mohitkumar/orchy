package worker

type WorkerConfiguration struct {
	ServerUrl                string
	MaxRetryBeforeResultPush int
	RetryIntervalSecond      int
}
