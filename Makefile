SHELL=/bin/bash

.PHONY: install format format-check lint test check-all \
	up-ci-services down-ci-services protobuf examples

install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3
	go mod tidy

	pip3 install grpcio-tools

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
	protoc \
		-I services/proto/ \
		--go_out=services/proto --go_opt=paths=source_relative \
		--go-grpc_out=services/proto --go-grpc_opt=paths=source_relative \
		services/proto/scheduler.proto

	python3 \
		-m grpc_tools.protoc \
		-I services/proto/ \
		--python_out=examples/rpc/python \
		--pyi_out=examples/rpc/python \
		--grpc_python_out=examples/rpc/python \
		services/proto/scheduler.proto

examples:
	go run examples/stores/base.go examples/stores/memory.go
	go run examples/stores/base.go examples/stores/gorm.go
	go run examples/stores/base.go examples/stores/redis.go
	go run examples/stores/base.go examples/stores/mongodb.go
	go run examples/stores/base.go examples/stores/etcd.go
	go run examples/rpc/rpc.go
	go run examples/http/http.go
