SHELL=/bin/bash

.PHONY: install
install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3
	go mod tidy

	pip3 install grpcio-tools

.PHONY: format
format:
	gofmt -l -s -w .

.PHONY: format-check
format-check:
	diff -u <(echo -n) <(gofmt -d .)

.PHONY: lint
lint:
	go vet .

.PHONY: up-cluster-ci-service
up-cluster-ci-service:
	go run examples/cluster/cluster_main.go -e 127.0.0.1:36680 -eh 127.0.0.1:36690 -se 127.0.0.1:36660 -seh 127.0.0.1:36670 &
	sleep 2s

.PHONY: down-cluster-ci-service
down-cluster-ci-service:
	ps -ef | grep "cluster_main -e 127.0.0.1:36680 -eh 127.0.0.1:36690 -se 127.0.0.1:36660 -seh 127.0.0.1:36670" \
	| grep -v grep | awk '{print $$2}' | xargs kill 2>/dev/null | echo "down-cluster-ci-service"

.PHONY: down-cluster-ci-service_second
down-cluster-ci-service_second:
	ps -ef | grep "cluster_main -e 127.0.0.1:36680 -eh 127.0.0.1:36690 -se 127.0.0.1:36660 -seh 127.0.0.1:36670" \
	| grep -v grep | awk '{print $$2}' | xargs kill 2>/dev/null | echo "down-cluster-ci-service"


.PHONY: test
test: down-cluster-ci-service up-cluster-ci-service
	go test \
		-timeout 120s \
		-covermode=set \
		-coverprofile=coverage.out \
		. \
		./services \
		./stores \
		-v
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

.PHONY: check-all
check-all: format-check lint test down-cluster-ci-service_second

.PHONY: up-ci-services
up-ci-services:
	docker compose -f ci/docker-compose.ci.yml up -d

.PHONY: down-ci-services
down-ci-services:
	docker compose -f ci/docker-compose.ci.yml down

.PHONY: protobuf
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

.PHONY: examples
examples:
	go run examples/stores/base.go examples/stores/memory.go
	go run examples/stores/base.go examples/stores/gorm.go
	go run examples/stores/base.go examples/stores/redis.go
	go run examples/stores/base.go examples/stores/mongodb.go
	go run examples/stores/base.go examples/stores/etcd.go
	go run examples/rpc/rpc.go
	go run examples/http/http.go
