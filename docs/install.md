### How to install
Require go version 1.18 or above
```
git clone github.com/mohitkumar/orchy
make
```
### Start single node server
```
bin/orchy --bind-addr 127.0.0.1:8400 --grpc-port 8099 --http-port 8080 --node-name node1 --storage-impl redis --queue-impl redis --redis-addr localhost:6379
```

### Starting 3 nodes cluster of orchy server
```
bin/orchy --bind-addr 127.0.0.1:8400 --grpc-port 8099 --http-port 8080 --node-name node1 --storage-impl redis --queue-impl redis --redis-addr localhost:6379 --cluster-address localhost:8400

bin/orchy --bind-addr 127.0.0.1:8401 --grpc-port 8098 --http-port 8081 --node-name node2 --storage-impl redis --queue-impl redis --redis-addr localhost:6379 --cluster-address localhost:8400

bin/orchy --bind-addr 127.0.0.1:8402 --grpc-port 8097 --http-port 8082 --node-name node3 --storage-impl redis --queue-impl redis --redis-addr localhost:6379 --cluster-address localhost:8400
```
### Options Description
*bind-addr* - Used to communicate the cluster events- join/leave of nodes. This is used by serf to send/receive the cluster gossip events. <br />

*grpc-port* - Workers communicates with server using grpc. Workers poll and push action execution result using grpc.

*http-port* - Various rest endpoints are exposed on http. These endpoints defines CRUD for workflow, workflow execution, event etc.

*node-name* - Unique idnetifier of a node in multi node cluster setup.

*storage-impl* - Storge implementation decides where workflow metadata and workflow run context is stored, currently supports redis only.

*queue-impl* - A queue of ready to run actions is maintained which is polled by the workers for action execution. This decides the backend used for queue implementation currently supports redis only.

*redis-addr* - As our storage and queue usese redis to store data. This specifies the address of redis instance/cluster. Redis cluster nodes can be specified in comma separated format.

*cluster-address* - Address of a node of a existing orchy cluster. New node with this option will join the cluster.

*namespace* - Used at persistence layer. IF redis is used at persistence layer prefix all the keys with this namespace while storing the data.

*partitions* - Number of partition used by consistent hash ring. A fixed number of partitions are created when the server starts. All nodes in the clsuter should keep this value same, if value is different on each node the behaviour of cluster is undefined.