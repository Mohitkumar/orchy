package redis

import (
	"fmt"
	"strings"

	rd "github.com/go-redis/redis/v9"
	"github.com/mohitkumar/orchy/server/cluster"
)

type baseDao struct {
	redisClient rd.UniversalClient
	namespace   string
	ring        *cluster.Ring
	membership  *cluster.Membership
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

func (bs *baseDao) getPartition(flowId string) int {
	return bs.ring.GetPartition(flowId)
}

func (bs *baseDao) getLocalPartitions() []int {
	member := bs.membership.GetLocalMemebr()
	return bs.ring.GetPartitions(member)
}
