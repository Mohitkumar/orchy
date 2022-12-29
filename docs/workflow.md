## Workflow
A workflow is a DAG of connected actions.
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
			"next":{"default":[6]}
		},
        {
			"id":6,
			"type":"system",
			"name":"javascript",
            "expression":"if($['1'].output.key1 == '200') {$['x']='nesX'};",
			"next":{"default":[7]}
		},
        {
			"id":7,
			"type":"system",
			"name":"delay",
			"delaySeconds":10,
			"next":{
				"default": [2]
			}
		},
		{
			"id":2,
			"type":"system",
			"name":"switch",
			"expression":"$.1.output.key1",
			"next":{
				"200": [3]
			}
		},
        {
			"id":3,
			"type":"system",
			"name":"wait",
			"event":"test",
			"next":{
				"default": [4,9]
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
		},
		{
			"id":9,
			"type":"user",
			"name":"log-action",
            "parameters" :{
                "k1" :24,
                "k2" : "$1.output.key2",
                "k3" :"$6.output.x",
				"k4" :"$.input.age"
            }
		}
	]
}
```
*name* - name of the workflow, it should be unique.

*rootAction* - Action from which the execution of workfow starts

*actions* - actions is an array of actions which describe the DAG of actions(system/user) as workflow. It defines which action would run after the current action execution finishes.

### Action
```
{
    "id":2,
    "type":"user",
    "name":"enhanceDataAction",
    "parameters":{
        "key1" : "200",
        "key2" :"$.1.output.key1",
    },
    "next":{"default":6}
}
```
*id* - Unique id of action inside a workflow, this id is used inside *next* object as reference.

*type* - Type of action it can be either ```user``` or ```system``` .
Refer to [System Action](docs/system-action.md) or [User Action](docs/user-action.md).

*parameters* - paramters are passed as input to the action definition. We can access input to the workflow as jsonpath expression- ```$.input.param1``` also we can access the output of previous action using jsonPath - ```$.1.output.key1```, this gets the key1 from the output of action with id 1
