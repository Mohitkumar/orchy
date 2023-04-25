package cassandra

import (
	"fmt"
	"strings"

	"github.com/gocql/gocql"
)

type baseDao struct {
	Session   *gocql.Session
	namespace string
}

func NewBaseDao(conf Config) *baseDao {
	cluster := gocql.NewCluster(conf.Addrs...)
	cluster.Keyspace = conf.KeySpace
	Session, err := cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	return &baseDao{
		Session:   Session,
		namespace: conf.KeySpace,
	}
}

func (bs *baseDao) getNamespaceKey(args ...string) string {
	return fmt.Sprintf("%s:%s", bs.namespace, strings.Join(args, ":"))
}
