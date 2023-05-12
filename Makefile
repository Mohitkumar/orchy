SHELL := /bin/bash
BINARY_NAME=orchy
GOPATH ?= $(shell go env GOPATH)
TAG ?= 1.0.0

GO                  := GO111MODULE=on go
build:
	go build -o bin/${BINARY_NAME} main.go

run:
	./bin/${BINARY_NAME}

test:
	go test ./...

compile:
	protoc api/v1/*.proto \
		--go_out=. \
		--go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		--proto_path=.

build-docker:
	docker build -t github.com/mohitkumar/orchy:$(TAG) .