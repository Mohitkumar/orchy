# Orchy
Distributed, scalable and faault tolerant workflow orchestrator backed by Redis.

Multiple workers can run simultanously polling the orchy server for work distribution. Task/Work can be distributed at worker level by providing thread count, each worker runs N number of threads and tasks are executed by these threads.


Orchy server itself runs as a cluster of n number of servers. Each server in cluster is responsible of handling the request and hosts some partitions. Partitions are automaticaly distributed equally among servers. If any server leaves or joins the cluster the partitiions are automatically redistributed.

Redis is used to store the workflow and task definition and workflow run context data. Workflow context data is the input and output of each nodes in the worklow.
## Requirements
* go 1.18 or above
* Redis for storage

## Workflow

<img src="docs/use_case_1.svg?raw=true" width="600" height="450">

Workflow assembles actions in a DAG. Output of each action becomes the input for next action.<br />
## Workflow Definiton-
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
                "userId" : "$.input.userId"
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
			"expression":"$.user.gender",
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
                "phoneNumber" : "$.user.phone",
                "message" : "$.input.sms.first.message"
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
                "to" :"$.user.email",
                "subject" : "$.input.email.first.subject",
                "message" :"$.input.email.first.message"
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
               "phoneNumber" : "$.user.phone",
               "message" : "$.input.whatsapp.message"
            }
		},
		{
			"id":9,
			"type":"user",
			"name":"sendSms",
            "parameters" :{
                "phoneNumber" : "$.user.phone",
                "message" : "$.input.sms.second.message"
            }
		},
		{
			"id":10,
			"type":"user",
			"name":"sendEmail",
            "parameters" :{
                "to" :"$.user.email",
                "subject" : "$.input.email.second.subject",
                "message" :"$.input.email.second.message"
            }
		},
	]
}
```
A workflow is executed by hitting a rest endpoint to the orch server. When a workfow start executing an instance of workflow is created and a UUID is assigned to each execution.<br />
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
input provided to execute the workflow flows through each action any action can reference input using json-path expression i.e. ```$.input.userId```.<br />
Also output of previous action can be referenced in next action parameters. i.e ```$.1.output.user``` . This reference the parameter ```user``` from the output of action with id 1.
## Action(Task)
There could be two types of actions in a workflow user defined action or system action.

### System Action- 
System actions are type of action which runs inside the server itself instead of worker. Currently supported system actions are- ```wait```,```delay```,```switch``` and ```javascript```.

### User Defined Action- 
User actions are defined in the worker liberary which runs as a independent process/system and access the orchy server using grpc for available actions to execute.<br />
Any number of actions can be defined in the worker and their definition is registerd on server by the worker itself.

## Table of Content
* [Design](docs/design.md)
* [installing](docs/install.md)
* [Starting worker](docs/worker.md)
* [Workflow & Actions](docs/workflow.md)
* [System action](docs/system-action.md)
* [User action](docs/user-action.md)