SHELL=/bin/bash

.PHONY: install
install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.9
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.5.0
	go install golang.org/x/tools/cmd/goimports@v0.37.0
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
	golangci-lint run --timeout=5m -v

.PHONY: up-cluster-ci-service
up-cluster-ci-service:
	go run examples/cluster/cluster_node/main.go -e 127.0.0.1:36680 -egr 127.0.0.1:36660 -eh 127.0.0.1:36670 &
	sleep 2s

.PHONY: down-cluster-ci-service
down-cluster-ci-service:
	ps -ef | grep "main -e 127.0.0.1:36680 -egr 127.0.0.1:36660 -eh 127.0.0.1:36670" \
	| grep -v grep | awk '{print $$2}' | xargs kill 2>/dev/null | echo "down-cluster-ci-service"

.PHONY: down-cluster-ci-service_second
down-cluster-ci-service_second:
	ps -ef | grep "main -e 127.0.0.1:36680 -egr 127.0.0.1:36660 -eh 127.0.0.1:36670" \
	| grep -v grep | awk '{print $$2}' | xargs kill 2>/dev/null | echo "down-cluster-ci-service-second"

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
	$(MAKE) down-cluster-ci-service_second

.PHONY: check-all
check-all: format-check isort-check lint test

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
	go run examples/stores/memory/main.go
	go run examples/stores/gorm/main.go
	go run examples/stores/redis/main.go
	go run examples/stores/mongodb/main.go
	go run examples/stores/etcd/main.go
	go run examples/stores/elasticsearch/main.go

.PHONY: examples-api
examples-api:
	go run examples/grpc/main.go
	go run examples/http/main.go

.PHONY: examples-queue
examples-queue:
	go run examples/queues/memory/main.go
	go run examples/queues/nsq/main.go
	go run examples/queues/rabbitmq/main.go
	go run examples/queues/redis/main.go
	go run examples/queues/mqtt/main.go
	go run examples/queues/kafka/main.go

.PHONY: examples-backend
examples-backend:
	go run examples/backends/memory/main.go
	go run examples/backends/gorm/main.go
	go run examples/backends/mongodb/main.go

.PHONY: examples-event
examples-event:
	go run examples/event/main.go

.PHONY: examples-all
examples-all: examples-store examples-api examples-queue examples-backend examples-event
