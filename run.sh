#!/bin/bash
set -B                  # enable brace expansion
for i in {1..10000}; do
  echo $i	
  curl --location --request POST 'http://localhost:8080/flow/execute' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name":"wf7",
     "input":{
        "k1":1,
        "userId":1234,
        "k2":"90",
        "nest":{
            "k1":8909
        },
        "list":[1,2,3]
    }
}'
done
