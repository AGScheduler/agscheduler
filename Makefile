SHELL=/bin/bash

.PHONY: install
install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.1
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.58.2
	go install golang.org/x/tools/cmd/goimports@v0.21.0
	go mod tidy

	pip3 install grpcio-tools

.PHONY: format
format:
	gofmt -l -s -w .

.PHONY: format-check
format-check:
	diff -u <(echo -n) <(gofmt -d .)

.PHONY: isort
isort:
	find . -type f -name '*.go' -not -name '*.pb.go' | xargs goimports -l -w -local github.com/agscheduler/agscheduler

.PHONY: isort-check
isort-check:
	diff -u <(echo -n) <(find . -type f -name '*.go' -not -name '*.pb.go' | xargs goimports -d -local github.com/agscheduler/agscheduler)

.PHONY: lint
lint:
	golangci-lint run --timeout=5m

.PHONY: up-cluster-ci-service
up-cluster-ci-service:
	go run examples/cluster/cluster_node.go -e 127.0.0.1:36680 -egr 127.0.0.1:36660 -eh 127.0.0.1:36670 &
	sleep 2s

.PHONY: down-cluster-ci-service
down-cluster-ci-service:
	ps -ef | grep "cluster_node -e 127.0.0.1:36680 -egr 127.0.0.1:36660 -eh 127.0.0.1:36670" \
	| grep -v grep | awk '{print $$2}' | xargs kill 2>/dev/null | echo "down-cluster-ci-service"

.PHONY: down-cluster-ci-service_second
down-cluster-ci-service_second:
	ps -ef | grep "cluster_node -e 127.0.0.1:36680 -egr 127.0.0.1:36660 -eh 127.0.0.1:36670" \
	| grep -v grep | awk '{print $$2}' | xargs kill 2>/dev/null | echo "down-cluster-ci-service"

.PHONY: test
test: down-cluster-ci-service up-cluster-ci-service
	go test \
		-timeout 120s \
		-covermode=set \
		-coverprofile=coverage.out \
		. \
		./stores \
		./services \
		./queues \
		./backends \
		-v
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

.PHONY: check-all
check-all: format-check isort-check lint test down-cluster-ci-service_second

.PHONY: up-ci-services
up-ci-services:
	docker compose -f ci/docker-compose.store.ci.yml up -d
	docker compose -f ci/docker-compose.queue.ci.yml up -d

.PHONY: down-ci-services
down-ci-services:
	docker compose -f ci/docker-compose.store.ci.yml down
	docker compose -f ci/docker-compose.queue.ci.yml down

.PHONY: protobuf
protobuf:
	protoc \
		-I services/proto/ \
		--go_out=services/proto --go_opt=paths=source_relative \
		--go-grpc_out=services/proto --go-grpc_opt=paths=source_relative \
		services/proto/*.proto

	python3 \
		-m grpc_tools.protoc \
		-I services/proto/ \
		--python_out=examples/grpc/python/proto \
		--pyi_out=examples/grpc/python/proto \
		--grpc_python_out=examples/grpc/python/proto \
		services/proto/*.proto && \
	sed -i 's/^\(import.*pb2\)/from proto \1/g' examples/grpc/python/proto/*pb2_grpc.py

.PHONY: examples-store
examples-store:
	go run examples/stores/base.go examples/stores/memory.go
	go run examples/stores/base.go examples/stores/gorm.go
	go run examples/stores/base.go examples/stores/redis.go
	go run examples/stores/base.go examples/stores/mongodb.go
	go run examples/stores/base.go examples/stores/etcd.go
	go run examples/stores/base.go examples/stores/elasticsearch.go

.PHONY: examples-api
examples-api:
	go run examples/grpc/grpc.go
	go run examples/http/http.go

.PHONY: examples-queue
examples-queue:
	go run examples/queues/base.go examples/queues/memory.go
	go run examples/queues/base.go examples/queues/nsq.go
	go run examples/queues/base.go examples/queues/rabbitmq.go
	go run examples/queues/base.go examples/queues/redis.go
	go run examples/queues/base.go examples/queues/mqtt.go
	go run examples/queues/base.go examples/queues/kafka.go

.PHONY: examples-backend
examples-backend:
	go run examples/backends/base.go examples/backends/memory.go
	go run examples/backends/base.go examples/backends/gorm.go
	go run examples/backends/base.go examples/backends/mongodb.go

.PHONY: examples-all
examples-all: examples-store examples-api examples-queue examples-backend
