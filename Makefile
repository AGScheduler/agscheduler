SHELL=/bin/bash

MODULES := $(shell awk '!/^[[:space:]]*(#|$$)/ { print ($$0 == "." ? "." : "./" $$0) }' scripts/modules.txt)
COVERAGE_DIR := $(CURDIR)/coverage

.PHONY: install
install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0
	go install golang.org/x/tools/cmd/goimports@v0.53.0
	./scripts/tidy-modules.sh

	pip3 install grpcio-tools==1.85.3

.PHONY: tidy
tidy:
	./scripts/tidy-modules.sh

.PHONY: add-module
add-module:
	@test -n "$(MODULE)" || (echo "usage: make add-module MODULE=stores/cassandra [DEPS='vendor/pkg@v1.2.3 ...']" >&2; exit 1)
	./scripts/add-module.sh "$(MODULE)" $(DEPS)

.PHONY: bump-version
bump-version:
	@test -n "$(VERSION)" || (echo "usage: make bump-version VERSION=0.21.0" >&2; exit 1)
	./scripts/bump-version.sh "$(VERSION)"

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
	for module in $(MODULES); do \
		echo "Linting $$module"; \
		(cd $$module && golangci-lint run --timeout=15m -v) || exit $$?; \
	done

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
test: down-cluster-ci-service up-cluster-ci-service wait-ci-services
	mkdir -p $(COVERAGE_DIR)
	for module in $(MODULES); do \
		echo "Testing $$module"; \
		profile="$(COVERAGE_DIR)/$$(printf '%s' "$$module" | tr '/.' '__').out"; \
		(cd $$module && go test -count=1 -timeout 120s -covermode=set -coverprofile=$$profile ./... -v) || exit $$?; \
		if [ -f $$profile ]; then \
			go tool cover -func=$$profile; \
			go tool cover -html=$$profile -o $${profile%.out}.html; \
		fi; \
	done
	$(MAKE) down-cluster-ci-service_second

.PHONY: check-deps
check-deps:
	go run ./scripts/check-module-deps.go

.PHONY: check-all
check-all: format-check isort-check lint check-deps test

.PHONY: up-ci-services
up-ci-services:
	docker compose -f ci/docker-compose.store.ci.yml up -d
	docker compose -f ci/docker-compose.queue.ci.yml up -d

.PHONY: wait-ci-services
wait-ci-services:
	./ci/wait-for-ci-services.sh

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
