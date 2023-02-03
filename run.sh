#!/bin/bash
set -B                  # enable brace expansion
for i in {1..10000}; do
  sleep 0.01
  curl --location --request POST 'http://localhost:8080/flow/execute' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name":"wf6",
     "input":{
        "k1":1,
        "k2":"90",
        "nest":{
            "k1":8909
        }
    }
}'
done
