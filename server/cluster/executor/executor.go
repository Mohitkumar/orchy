package executor

type Executor interface {
	Start()
	Stop()
	Name() string
}
