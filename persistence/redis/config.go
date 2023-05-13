package redis

type Config struct {
	Addrs     []string
	Namespace string
	PoolSize  int
	Password  string
}
