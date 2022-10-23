# Orchy
Distributed, Fault tolerant workflow orchestrator. Scalable at all layers-
* Workers can scale independently
* Servers can scale independently
* Storage can scale independently

To distribute workload any number of workres can be run independently polling the orchy servers for task.

To mitigate single point of failure and distribute load a cluster of n number servers can be started

Storage layer is backed by redis which can be scaled running in cluster mode.
## Requirements
* go 1.18 or above

## Storage
Currently supports redis for storing workflow definitions, metadata
### TODO-
- [ ] DynamoDB as storage for workflows and metadata
- [ ] Cassandra as storage for workflows and metadata

## Task Queue
Currently supports redis for managing tasks(actions).
### TODO
- [ ] SQS as TaskQueue implementation
- [ ] Kafka as TaskQueue implementation

## Workflow

Workflow assembles actions in a DAG. Output of each action becomes the input for next action.<br />
Example Workflow Definiton-
```
    {
	"name" :"test_workflow",
	"rootAction":1,
	"actions":[
		{
			"id":1,
			"type":"user",
			"name":"add-params-action",
			"parameters":{
                "key1" : "200",
                "key2" :39
            },
			"next":{"default":6}
		},
        {
			"id":6,
			"type":"system",
			"name":"javascript",
            "expression":"if($['1'].output.key1 == '200') {$['x']='nesX'};",
			"next":{"default":7}
		},
        {
			"id":7,
			"type":"system",
			"name":"delay",
			"delaySeconds":10,
			"next":{
				"default": 2
			}
		},
		{
			"id":2,
			"type":"system",
			"name":"switch",
			"expression":"$.1.output.key1",
			"next":{
				"200": 3
			}
		},
        {
			"id":3,
			"type":"system",
			"name":"wait",
			"event":"test",
			"next":{
				"default": 4
			}
		},
		{
			"id":4,
			"type":"user",
			"name":"log-action",
            "parameters" :{
                "k1" :23,
                "k2" : "$1.output.key1",
                "k3" :"$6.output.x"
            }
		}
	]
}
```
A workflow is executed by hitting a rest endpoint to the orch server. When a workfow start executing an instance of workflow is created and a UUID is assigned to each execution.<br />
Endpoint-
```
curl --location --request POST 'http://localhost:8080/flow/execute' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name":"test-workflow",
    "input":{
        "age": 89
    }
}'
```
input provided to execute the workflow flows through each action any action can reference input using json-path expression i.e. ```$.input.age```.<br />
Also output of previous action can be referenced in next action parameters. i.e ```$.1.output.key1``` . This reference the parameter ```key1``` from the output of action with id 1.
## Action(Task)
There could be two types of actions in a workflow user defined action or system action.

### System Action- 
System actions are type of action which runs inside the server itself instead of worker. Currently supported system actions are- ```wait```,```delay```,```switch``` and ```javascript```.

### User Action- 
User actions are defined in the worker liberary which runs as a independent process/system and access the orchy server using grpc for available actions to execute.<br />
Any number of actions can be defined in the worker and their definition is registerd on server by the worker itself.

## Table of Content
* [Design](docs/design.md)
* [installing]()
* [Starting worker]()
* [System action]()
* [User action]()