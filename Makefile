SHELL := /bin/bash
BINARY_NAME=orchy
GOPATH ?= $(shell go env GOPATH)

GO                  := GO111MODULE=on go
build:
	go build -o bin/${BINARY_NAME} server/main.go
	GOARCH=amd64 GOOS=linux go build -o bin/${BINARY_NAME}-linux server/main.go

run:
	./${BINARY_NAME}

test:
	go test ./server/...

compile:
	protoc api/v1/*.proto \
		--go_out=. \
		--go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		--proto_path=.