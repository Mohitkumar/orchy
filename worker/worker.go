package worker

type Worker interface {
	Execute(map[string]any) (map[string]any, error)
	GetName() string
	GetPollInterval() int
}
