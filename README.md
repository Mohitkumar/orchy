# Orchy
Distributed, scalable and fault tolerant workflow orchestrator backed by Redis.

Multiple workers run simultaneously polling the Orchy server for work distribution. Tasks are distributed at worker level by providing thread count, each worker runs N number of threads and tasks are executed by these threads.


Orchy server itself runs as a cluster of n number of servers. Each server in cluster is responsible of handling the request and hosts some partitions. Partitions are automatically distributed equally among servers. If any server leaves or joins the cluster the partitions are automatically redistributed.

Redis is used to store the workflow and task definition and workflow run context data. Workflow context data is the input and output of each nodes in the workflow.

# Table Of Contents
* [Design](docs/design.md)
* [REST API](docs/rest.md)
## Requirements
* go 1.18 or above
* Redis for storage

## How to Install
``` go install github.com/mohitkumar/orchy ```

### Start single node server
```
bin/orchy --bind-address 127.0.0.1:8400 --grpc-port 8099 --http-port 8080 --node-name node1 --redis-address localhost:6379
```

### Starting 3 nodes cluster of orchy server
```
bin/orchy --bind-address 127.0.0.1:8400 --grpc-port 8099 --http-port 8080 --node-name node1 --redis-address localhost:6379 --cluster-address localhost:8400

bin/orchy --bind-address 127.0.0.1:8401 --grpc-port 8098 --http-port 8081 --node-name node2 --redis-address localhost:6379 --cluster-address localhost:8400

bin/orchy --bind-address 127.0.0.1:8402 --grpc-port 8097 --http-port 8082 --node-name node3 --redis-address localhost:6379 --cluster-address localhost:8400
```
Option    | Description 
| :---    | :--- |
| bind-address | Used to communicate the cluster events- join/leave of nodes. This is used by underline gossip protocol to send/receive events. |
| grpc-port | Workers communicates with server using grpc. This port is required to receive grpc call from workers. |
| http-port | This port is used to receive rest request to make CRUD for workflow, workflow execution, event etc. |
| node-name | Unique identifier of a node in multi node cluster setup. |
| redis-address | This specifies the address of redis instance/cluster. Redis cluster nodes can be specified in comma separated format. |
| cluster-address | Address of a node of a existing orchy cluster. New node with this option will join the cluster. |
| namespace | Used at persistence layer. All redis keys are prefixed with this namespace while storing the data.|
| partitions| Number of partition used by consistent hash ring. A fixed number of partitions are created when the server starts. All nodes in the cluster should keep this value same, if value is different on each node the behavior of cluster is undefined. |
## Workflow

<img src="docs/use_case_1.svg?raw=true" width="800" height="500">

Workflow assembles actions in a DAG. Output of each action becomes the input for next action.<br />
## Workflow Definition
Workflows are represented using json, above workflow is represented as below
```
    {
  "name" :"notificationWorkflow",
  "rootAction":1,
  "onSuccess" :"DELETE",
  "onFailure" :"DELETE",
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
                "message" : "Hi {$.input.sms.first.message} get discount {$.2.output.user.discount}%"
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
                "to" :"{$.1.output.user.email}",
                "subject" : "{$.input.email.first.subject}",
                "message" :"{$.input.email.first.message} get discount {$.2.output.user.discount}%"
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
			"timeoutSeconds" : 10,
			"next":{
				"default":[10],
				"open": [9]
			}
		},
		{
			"id":8,
			"type":"user",
			"name":"sendWahtsapp",
            "parameters" :{
               "phoneNumber" : "{$.1.output.user.phone}",
               "message" : "{$.input.whatsapp.message} get discount {$.2.output.user.discount}%"
            }
		},
		{
			"id":9,
			"type":"user",
			"name":"sendSms",
            "parameters" :{
                "phoneNumber" : "{$.1.output.user.phone}",
                "message" : "{$.input.sms.second.message} get discount {$.2.output.user.discount}% {$.1.output.user.address.country}"
            }
		},
		{
			"id":10,
			"type":"user",
			"name":"sendEmail",
            "parameters" :{
                "to" :"{$.1.output.user.email}",
                "subject" : "{$.input.email.second.subject}",
                "message" :"{$.input.email.second.message} get discount {$.2.output.user.discount}% {$.1.output.user.address.country}"
            }
		}
	]
}
```
### Workflow json

Field| Description
| :---    | :--- |
|name|Name of the workflow, it should be unique.|
|rootAction|Action from which the execution of workflow starts|
|actions|It is an array of actions which describe the DAG of actions(system/user) as workflow. It defines which action would run after the current action execution finishes.|

### Action json

Field| Description
| :---    | :--- |
|id|Unique id of action inside a workflow, this id is used inside ```next``` object as reference.|
|type|Type of action it can be either ```user``` or ```system``` .|
|parameters|parameters are passed as input to the action definition. We can access input to the workflow as jsonpath expression- ```{$.input.param1}``` also we can access the output of previous action using jsonPath - ```{$.1.output.key1}```, this gets the key1 from the output of action with id 1|

A workflow is executed by hitting a rest endpoint to the orch server. When a workflow start executing an instance of workflow is created and a UUID is assigned to each execution.

Endpoint-
```
curl --location --request POST 'http://localhost:8080/execution' \
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
Switch action evaluates the value of ```expression``` and that value is matched against the ```next``` json object, if the value match the action against that is executed next else default is executed. Expression is represented as a jsonpath expression.

### JavaScript Action
```
{
    "id":2,
    "type":"system",
    "name":"javascript",
    "expression":"$['user']=$['1'].output.user;if($['1'].output.user.age >= 20) {$['user']['discount']=20} else{$['user']['discount']=10};",
    "next":{"default":[3]}
}
```
Javascript action can run any javascript expression. Input parameters and output of the previous action can be accessed using \$ alias inside javascript expression. In above example ```$['1'].output.user``` gets the user from the output of action id 1 and ```$['user']['discount']=20``` adds a property to user object ```discount``` with value ```20``` which will be accessible to next nodes using ```{$.2.output.user.discount}```

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
Wait action waits for an external event until timeout. If the event is received by workflow within 20 seconds then action against that event is executed, if it times-out then action against default is executed. For example above waits for an external event with name ```test```. On receiving the event it moves forward and run action with id 6. If workflow does not receive event within 20 seconds then action with id 4 is executed.
### User Defined Action
User actions are defined in the worker library which runs as a independent process/system and access the orchy server using grpc for available actions to execute.<br />
Any number of actions can be defined in the worker and their definition is registered on server by the worker itself.
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

## Worker
* Worker runs as a separate process remotely, any number of workers can run in parallel. <br />
* Worker polls orchy server on a fixed (configurable) interval for the available actions to execute. Orchy server hands over an action for execution to a worker and worker push back the result to the orchy server after execution.<br />
* Worker polls each node in server in round robin fashion and execute the available actions <br />
* if a new node joins or an existing node leaves the cluster, worker is capable to detect that change.

### Worker example(Golang).
```
package main

import (
	"fmt"
	"time"

	worker "github.com/mohitkumar/orchy-worker"
)

func main() {
	config := &worker.WorkerConfiguration{
		ServerUrl:                "localhost:8099",
		MaxRetryBeforeResultPush: 1,
		RetryIntervalSecond:      1,
	}

	addParamActionFn := func(data map[string]any) (map[string]any, error) {
		data["newKey"] = 22
		return data, nil
	}
	logActionFn := func(data map[string]any) (map[string]any, error) {
		fmt.Println(data)
		return data, nil
	}
	tp := worker.NewWorkerConfigurer(*config)

	addParamWorker := worker.NewDefaultWorker(addParamActionFn).WithRetryCount(1).WithTimeoutSeconds(20)
	logWorker := worker.NewDefaultWorker(logActionFn)

	err := tp.RegisterWorker(addParamWorker, "add-data-worker", 1*time.Second, 2, 1)
	fmt.Print(err)
	tp.RegisterWorker(logWorker, "print-worker", 1*time.Second, 2, 1)
	tp.Start()
}

```
Each worker is a golang function ```func(data map[string]any) (map[string]any, error)``` with additional configuration.
Each worker configuration can be created using a chain of With Function ie- ```addParamWorker := worker.NewDefaultWorker(addParamActionFn).WithRetryCount(1).WithTimeoutSeconds(20)```.
<br /> Once a worker is defined it has to be registered with the server. ```tp.RegisterWorker(addParamWorker, "add-params-action", 1*time.Second, 2, 1)```, which takes worker name, worker function, poll interval, batch size and number of threads.
<br /><br />
Working example in golang - (https://github.com/Mohitkumar/orchy-worker-example)

### Worker example(Java).
```
package io.github.orchy.example;

import io.github.mohitkumar.orchy.worker.RetryPolicy;
import io.github.mohitkumar.orchy.worker.Worker;
import io.github.mohitkumar.orchy.worker.WorkerManager;
import io.github.orchy.example.action.EnhanceDataAction;
import io.github.orchy.example.action.SmsAction;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.CommandLineRunner;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.ComponentScan;

import java.util.Collections;
import java.util.Map;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicLong;
import java.util.function.Function;

@SpringBootApplication
public class Main implements CommandLineRunner
{
    private static final Logger LOGGER = LoggerFactory.getLogger(Main.class);

    @Autowired
    private WorkerManager manager;

    @Autowired
    private SmsAction smsAction;

    @Autowired
    private EnhanceDataAction enhanceDataAction;

    public static void main(String[] args) {
        SpringApplication.run(Main.class, args);
    }

    @Override
    public void run(String... args) throws Exception {
        Function<Map<String, Object>, Map<String, Object>> logAction= (Map<String, Object> input) ->{
            LOGGER.info("output ={}", input);
            return Collections.emptyMap();
        };

        Worker logWorker = Worker
                .newBuilder()
                .DefaultWorker(logAction,"logAction", 2,100,TimeUnit.MILLISECONDS).build();

        manager.registerWorker(logWorker,10);
        manager.registerWorker(Worker.newBuilder()
                .DefaultWorker(smsAction,"smsAction",2,100,TimeUnit.MILLISECONDS)
                .build(), 10);

        manager.registerWorker(Worker.newBuilder()
                .DefaultWorker(enhanceDataAction,"enhanceData",2,100,TimeUnit.MILLISECONDS)
                        .withRetryCount(3).withRetryPolicy(RetryPolicy.FIXED).withTimeout(100)
                .build(), 10);

        manager.start();
    }
}

```

Working example Java- (https://github.com/Mohitkumar/orchy-worker-example-java)