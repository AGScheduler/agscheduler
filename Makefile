SHELL=/bin/bash

.PHONY: install format format-check lint test check-all

install:
	go mod tidy

format:
	gofmt -l -s -w .

format-check:
	diff -u <(echo -n) <(gofmt -d .)

lint:
	go vet .

test:
	go test .

check-all: format-check lint test
