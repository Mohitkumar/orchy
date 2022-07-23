package redis

import (
	"fmt"
	"strings"

	rd "github.com/go-redis/redis/v9"
)

type baseDao struct {
	redisClient rd.UniversalClient
	namespace   string
}

func newBaseDao(conf Config) *baseDao {
	redisClient := rd.NewUniversalClient(&rd.UniversalOptions{
		Addrs: conf.Addrs,
	})
	return &baseDao{
		redisClient: redisClient,
		namespace:   conf.Namespace,
	}
}

func (bs *baseDao) getNamespaceKey(args ...string) string {
	return fmt.Sprintf("%s:%s", bs.namespace, strings.Join(args, ":"))
}
