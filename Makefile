SHELL=/bin/bash

.PHONY: install format format-check lint test check-all \
	up-ci-services down-ci-services protobuf

install:
	go mod tidy

format:
	gofmt -l -s -w .

format-check:
	diff -u <(echo -n) <(gofmt -d .)

lint:
	go vet .

test:
	go test \
		-covermode=set \
		-coverprofile=coverage.out \
		-v . ./services ./stores
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

check-all: format-check lint test

up-ci-services:
	docker compose -f ci/docker-compose.ci.yml up -d

down-ci-services:
	docker compose -f ci/docker-compose.ci.yml down

protobuf:
	protoc -I services/proto/ --go_out=plugins=grpc:services/proto services/proto/scheduler.proto
	protoc -I services/proto/ --python_out=./examples/rpc/python services/proto/scheduler.proto
