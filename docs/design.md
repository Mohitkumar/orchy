![alt orchy](https://github.com/Mohitkumar/orchy/blob/main/docs/orchy.png?raw=true)

* Metadata Service is responsible for handling workflow and action definition CRUD operations. All opertions related to workflow and action definition are forwarded to persistence layer which stores the data to actual storage used.<br />
* Execution Service is responsible for handling the workflow execution request. When it receive a workflow execution request it creates an instance of that workflow with a UUID. This workflow instance is stored on a partition of the storage layer. Workflow instances are distributed among the partitions.
* Queue service is responsible for maintaining the action instances. The instances of the actions are also partitioned. Worker polls the action queues and execute the actions.
* Orchy uses hashicorp serf, which is an implementation of gossip protocol, to discover the new nodes and remove inactive and unreachable nodes from the cluster.<br />
* Consistent hasing is used to divide the workflow instance and its task among the partitions. Partitions are distributed among the nodes in the cluster.<br />
* Each partition is responsible for maintaining the workflow state and task queue for particular workflow. Persistence layer accepts the parition id which can be used by the acutal sotrage to keep the data in different partition for each partition id.<br />
* Each worker polls the cluster nodes in round robin fashion on fixed intervals. Queue Service checks for the ready to run task inside partitions owend by this node and returns that task. Worker executes the task and push the result using grpc to the server.<br />
* Execution service is also responsible for maintaining the workflow state machine which reacts to the execution request from user and results pushed by the workers and moves the state according to the definition of workflow. On each event workflow state machine is stored on the persistence layer.
