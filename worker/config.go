package worker

type WorkerConfiguration struct {
	ServerUrl                string
	PollInterval             int
	MaxRetryBeforeResultPush int
	RetryIntervalSecond      int
}
