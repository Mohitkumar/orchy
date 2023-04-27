## Worker
* Worker runs as a spearate process remotly, any number of workers can run in parallel. <br />
* Worker polls orchy server on a fixed interval for the available actions to execute. Orchy server hands over an action for exection to a worker and worker push back the result to the orchy server after execution.<br />
* Worker polls each node in server in round robin fashion, it is capable of detecting a change in the cluster config. <br />
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
Each worker configration can be created using a chain of With Function ie- ```addParamWorker := worker.NewDefaultWorker(addParamActionFn).WithRetryCount(1).WithTimeoutSeconds(20)```.
<br /> Once  aworker is defined it has to be registerd with the server. ```tp.RegisterWorker(addParamWorker, "add-params-action", 1*time.Second, 2, 1)```, which takes worker name, worker function, poll interval, batch size and number of threads.
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

    private static AtomicLong counter = new AtomicLong(0);

    public static void main(String[] args) {
        SpringApplication.run(Main.class, args);
    }

    @Override
    public void run(String... args) throws Exception {
        System.out.println("run");
        Function<Map<String, Object>, Map<String, Object>> logAction= (Map<String, Object> input) ->{
            LOGGER.info("output ={}", input);
            System.out.println(counter.incrementAndGet());
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