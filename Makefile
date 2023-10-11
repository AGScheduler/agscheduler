SHELL=/bin/bash

.PHONY: install format format-check lint test check-all \
	up-ci-services down-ci-services

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
		-v . ./stores
	go tool cover -func=coverage.out

check-all: format-check lint test

up-ci-services:
	docker compose -f ci/docker-compose.ci.yml up -d

down-ci-services:
	docker compose -f ci/docker-compose.ci.yml down
