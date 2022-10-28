## System Action
System actions are predefined actions which runs on server itself instead of worker.

### Delay Action
```
{
    "id":7,
    "type":"system",
    "name":"delay",
    "delaySeconds":10,
    "next":{
        "default": 2
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
    "expression":"$.1.output.key1",
    "next":{
        "200": 3,
        "300": 4,
        "default":6
    }
}
 ```   
Switch action evalute the value of ```expression``` and that value is matched against the ```next``` json object, if the value match the action against that is executed next else default is executed. Expression is represented as a jsonpath expression.

### Javasctipt Action
```
{
    "id":6,
    "type":"system",
    "name":"javascript",
    "expression":"if($['1'].output.key1 == '200') {$['x']='newValue'};",
    "next":{"default":7}
}
```
Javascript action can run any javascript expression. Input parameters and output of the previous action can be accessed using \$ alias inside javascipt expression. In above example ```$['1'].output.key1``` gets the key1 from the output of action id 1 and ```$['x']='newValue'``` adds a new key ```x``` with value ```newValue```

### Wait Action
```
{
    "id":3,
    "type":"system",
    "name":"wait",
    "event":"test",
    "next":{
        "default": 4
    }
}
```
Wait action waits for an external event. For example above waits for an external event with name ```test```. On receiving the event it moves forward and run action with id 4.