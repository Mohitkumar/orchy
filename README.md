# Orchy
Distributed, scalable and faault tolerant workflow orchestrator backed by Redis.

Multiple workers can run simultanously polling the orchy server for work distribution. Task/Work can be distributed at worker level by providing thread count, each worker runs N number of threads and tasks are executed by these threads.


Orchy server itself runs as a cluster of n number of servers. Each server in cluster is responsible of handling the request and hosts some partitions. Partitions are automaticaly distributed equally among servers. If any server leaves or joins the cluster the partitiions are automatically redistributed.

Redis is used to store the workflow and task definition and workflow run context data. Workflow context data is the input and output of each nodes in the worklow.
## Requirements
* go 1.18 or above
* Redis for storage

## How to Install
``` go install github.com/mohitkumar/orchy ```

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
Option    | Description 
| :---    | :--- |
| bind-addr | Used to communicate the cluster events- join/leave of nodes. This is used by underline gossip protocol to send/receive events. |
| grpc-port | Workers communicates with server using grpc. This port is required to receive grpc call from workers. |
| http-port | This port is used to receive rest request to make CRUD for workflow, workflow execution, event etc. |
| node-name | Unique idnetifier of a node in multi node cluster setup. |
| redis-addr | This specifies the address of redis instance/cluster. Redis cluster nodes can be specified in comma separated format. |
| cluster-address | Address of a node of a existing orchy cluster. New node with this option will join the cluster. |
| namespace | Used at persistence layer. All redis keys are prefixed with this namespace while storing the data.|
| partitions| Number of partition used by consistent hash ring. A fixed number of partitions are created when the server starts. All nodes in the clsuter should keep this value same, if value is different on each node the behaviour of cluster is undefined. |
## Workflow

<img src="docs/use_case_1.svg?raw=true" width="600" height="450">

Workflow assembles actions in a DAG. Output of each action becomes the input for next action.<br />
## Workflow Definiton
Workflows are represented using json, above worklfow is represented as below
```
    {
	"name" :"notificationWorkflow",
	"rootAction":1,
	"actions":[
		{
			"id":1,
			"type":"user",
			"name":"query-db",
			"parameters":{
                "userId" : "{$.input.userId}"
            },
			"next":{"default":[2]}
		},
        {
			"id":2,
			"type":"system",
			"name":"javascript",
            "expression":"$['user']=$['1'].output.user;if($['1'].output.user.age >= 20) {$['user']['discount']=20} else{$['user']['discount']=10};",
			"next":{"default":[3]}
		},
		{
			"id":3,
			"type":"system",
			"name":"switch",
			"expression":"{$.1.output.user.gender}",
			"next":{
				"MALE": [4],
				"FEMALE": [5]
			}
		},
		{
			"id":4,
			"type":"user",
			"name":"sendSms",
            "parameters" :{
                "phoneNumber" : "{$.1.output.user.phone}",
                "message" : "Hi {$.input.sms.first.message} get discount {$.2.output.user.discount} %"
            },
			"next":{
				"default": [6]
			}
		},
		{
			"id":5,
			"type":"user",
			"name":"sendEmail",
            "parameters" :{
                "to" :"{$.1.output.email}",
                "subject" : "{$.input.email.first.subject}",
                "message" :"{$.input.email.first.message}"
            },
			"next":{
				"default": [7]
			}
		},
        {
			"id":6,
			"type":"system",
			"name":"delay",
			"delaySeconds":10,
			"next":{
				"default": [8]
			}
		},
		
        {
			"id":7,
			"type":"system",
			"name":"wait",
			"event":"open",
			"timeoutSeconds" : 20,
			"next":{
				"default":[10]
				"open": [9]
			}
		},
		{
			"id":8,
			"type":"user",
			"name":"sendWahtsapp",
            "parameters" :{
               "phoneNumber" : "{$.1.output.phone}",
               "message" : "{$.input.whatsapp.message}"
            }
		},
		{
			"id":9,
			"type":"user",
			"name":"sendSms",
            "parameters" :{
                "phoneNumber" : "{$.1.output.phone}",
                "message" : "{$.input.sms.second.message}"
            }
		},
		{
			"id":10,
			"type":"user",
			"name":"sendEmail",
            "parameters" :{
                "to" :"{$.1.output.email}",
                "subject" : "{$.input.email.second.subject}",
                "message" :"{$.input.email.second.message}"
            }
		},
	]
}
```
### Workflow json

Field| Description
| :---    | :--- |
|name|name of the workflow, it should be unique.|
|rootAction|Action from which the execution of workfow starts|
|actions|actions is an array of actions which describe the DAG of actions(system/user) as workflow. It defines which action would run after the current action execution finishes.|

### Action json

Field| Description
| :---    | :--- |
|id|Unique id of action inside a workflow, this id is used inside ```next``` object as reference.|
|type|Type of action it can be either ```user``` or ```system``` .|
|parameters|paramters are passed as input to the action definition. We can access input to the workflow as jsonpath expression- ```{$.input.param1}``` also we can access the output of previous action using jsonPath - ```{$.1.output.key1}```, this gets the key1 from the output of action with id 1|

A workflow is executed by hitting a rest endpoint to the orch server. When a workfow start executing an instance of workflow is created and a UUID is assigned to each execution.

Endpoint-
```
curl --location --request POST 'http://localhost:8080/flow/execute' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name":"notificationWorkflow",
    "input":{
        "userId": 1234,
		"sms":{
			"first":{
				"message":"hi first sms"
			},
			"second":{
				"message":"hi second sms"
			}
		},
		"email":{
			"first":{
				"subject":"first email subject"
				"message":"first email subject
			},
			"second":{
				"subject":"second email subject"
				"message":"second email subject
			}
		},
		"whatsapp":{
			"message":"wahtsapp message"
		}
    }
}'
```
input provided to execute the workflow flows through each action any action can reference input using json-path expression enclosed within {} i.e. ```{$.input.userId}```.<br />
Also output of previous action can be referenced in next action parameters. i.e ```{$.1.output.user}``` . This reference the parameter ```user``` from the output of action with id 1.
## Action/Task
There could be two types of action in a workflow user defined action or system action.

### System Action
System actions are type of action which runs inside the server itself instead of worker. Currently supported system actions are- ```wait```,```delay```,```switch``` and ```javascript```.
### Delay Action
```
{
    "id":7,
    "type":"system",
    "name":"delay",
    "delaySeconds":10,
    "next":{
        "default": [2]
    }
}
 ```
 Delay Action is used to introduce a delay in between of two action's execution.

 ### Switch Action
 ```
 {
    "id":2,
    "type":"system",
    "name":"switch",
    "expression":"{$.1.output.key1}",
    "next":{
        "200": [3],
        "300": [4],
        "default":[6]
    }
}
 ```   
Switch action evalute the value of ```expression``` and that value is matched against the ```next``` json object, if the value match the action against that is executed next else default is executed. Expression is represented as a jsonpath expression.

### Javasctipt Action
```
{
    "id":2,
    "type":"system",
    "name":"javascript",
    "expression":"$['user']=$['1'].output.user;if($['1'].output.user.age >= 20) {$['user']['discount']=20} else{$['user']['discount']=10};",
    "next":{"default":[3]}
}
```
Javascript action can run any javascript expression. Input parameters and output of the previous action can be accessed using \$ alias inside javascipt expression. In above example ```$['1'].output.user``` gets the user from the output of action id 1 and ```$['user']['discount']=20``` adds a property to user object ```discount``` with value ```20``` which will be accessible to next nodes using ```{$.2.output.user.discount}```

### Wait Action
```
{
    "id":3,
    "type":"system",
    "name":"wait",
    "event":"test",
    "timeoutSeconds" : 20,
    "next":{
        "default": [4],
        "test" :[6]
    }
}
```
Wait action waits for an external event until timeout. If the event is received by workflow within 20 seconds then action against that event is executed, if it timesout then action against default is executed. For example above waits for an external event with name ```test```. On receiving the event it moves forward and run action with id 6. If workflow does not receive event within 20 seconds then action with id 4 is executed.
### User Defined Action
User actions are defined in the worker liberary which runs as a independent process/system and access the orchy server using grpc for available actions to execute.<br />
Any number of actions can be defined in the worker and their definition is registerd on server by the worker itself.
```
{
    "id":4,
    "type":"user",
    "name":"logAction",
    "parameters" :{
        "k1" :23,
        "k2" : "{$1.output.key1}",
        "k3" :"{$6.output.x}"
    },
    "next":{
        "default":[3]
    }
}
```
User action are defined with ```type``` -```user```. Any unique name could be given to the action name. parameters are passed as a ```map[string]any``` to the input of the function defining the action.<br />
Every user action should return a ```map[string]any``` which becomes available for the actions which comes after this action in workflow DAG.

#
* [Design](docs/design.md)
* [Starting worker](docs/worker.md)