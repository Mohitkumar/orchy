## User Defined Action
User action are defined inside the worker process which then executed remotely by the worker process.

```
{
    "id":4,
    "type":"user",
    "name":"logAction",
    "parameters" :{
        "k1" :23,
        "k2" : "$1.output.key1",
        "k3" :"$6.output.x"
    }
}
```
User action are defined with ```type``` -```user```. Any unique name could be given to the action name. parameters are passed as a ```map[string]any``` to the input of the function defining the action.<br />
Every user action should return a ```map[string]any``` which becomes available for the actions which comes after this action in workflow DAG.