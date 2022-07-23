package executor

type Executor interface {
	Start() error
	Stop() error
	Name() string
}
