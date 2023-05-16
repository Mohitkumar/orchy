# Rest API

API    | URL | Method |
| :--- | :--- | :--- |
|Create Workflow| /metadata/workflow | POST |
|Get Workflow|/metadata/workflow/{name}| GET |
|Create Action Definition|/metadata/action| POST |
|Get Action Definition|/metadata/action/{name}| GET |
|Run Workflow|/execution| POST |
|Get Workflow Execution|/execution/{name}/{id}| GET |

# Postman Collection Link
(https://api.postman.com/collections/409310-c9653250-bb25-4f6a-822f-822eaa1089e1?access_key=PMAT-01H0H53QK8J0PT9H09D2HWK825)

# API Examples-

## Create Workflow -

```
curl --location 'http://localhost/metadata/workflow' \
--header 'Content-Type: application/json' \
--data '{
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
            "expression":"$['\''user'\'']=$['\''1'\''].output.user;if($['\''1'\''].output.user.age >= 20) {$['\''user'\'']['\''discount'\'']=20} else{$['\''user'\'']['\''discount'\'']=10};",
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
}' 
```

## Get Workflow -
``` 
curl --location 'http://localhost/metadata/workflow/notificationWorkflow' 
```

## Run Workflow - 
```
curl --location 'http://localhost/execution' \
--header 'Content-Type: application/json' \
--data '{
    "name":"notificationWorkflow",
    "input":{
        "userId": "2",
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
				"subject":"first email subject",
				"message":"first email subject"
			},
			"second":{
				"subject":"second email subject",
				"message":"second email subject"
			}
		},
		"whatsapp":{
			"message":"wahtsapp message"
		}
    }
}'
```
## Create Action Definition - 
```
curl --location 'http://localhost:8080/metadata/action' \
--header 'Content-Type: application/json' \
--data '{
    "name": "testAction",
    "retry_count": 3,
    "retry_after_seconds": 5,
    "retry_policy": "FIXED",
    "timeout_seconds": 10
}'
```

## Get Action Definition
```
curl --location 'http://localhost/metadata/action/sendSms' 
```

## Get Workflow Execution -
``` 
curl --location 'http://localhost/execution/notificationWorkflow/52b8a8bf-3329-46be-a7f0-e895cf9ac634' 
```

## Send Event - 
```
curl --location 'http://localhost/event' \
--header 'Content-Type: application/json' \
--data '{
    "name":"notificationWorkflow",
    "flowId":"e421752a-c0f1-4aec-9e4f-f0ada1624281",
    "event":"open"
}'
```