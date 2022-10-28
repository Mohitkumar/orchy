## Worker
Worker runs as a spearate process remotly, any number of workers can run in parallel. <br />
Worker polls orchy server on a fixed interval for the available actions to execute. Orchy server hands over an action for exection to a worker and worker push back the result to the orchy server after execution.<br />
Worker polls each node in server in round robin fashion, it is capable of detecting a change in the cluster config. <br />
if a new node joins or an existing node leaves the cluster, worker is capable to detect that change.

### How to define a worker.
```
package main

import (
	"fmt"
	"time"

	"github.com/mohitkumar/orchy/worker"
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

	tp.RegisterWorker(addParamWorker, "add-params-action", 1*time.Second, 2, 1)
	tp.RegisterWorker(logWorker, "print-worker", 1*time.Second, 2, 1)
	tp.Start()
}

```
Each worker is a golang function ```func(data map[string]any) (map[string]any, error)``` with additional configuration.
Each worker configration can be created using a chain of With Function ie- ```addParamWorker := worker.NewDefaultWorker(addParamActionFn).WithRetryCount(1).WithTimeoutSeconds(20)```.
<br /> Once  aworker is defined it has to be registerd with the server. ```tp.RegisterWorker(addParamWorker, "add-params-action", 1*time.Second, 2, 1)```, which takes worker name, worker function, poll interval, batch size and number of threads.
<br />
For working example refer to - (https://github.com/Mohitkumar/orchy-worker-example)