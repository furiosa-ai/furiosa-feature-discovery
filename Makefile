SHELL := /bin/bash

.PHONY: all
all: build fmt lint vet test tidy

.PHONY: build
build:
	CGO_ENABLED=0 go build -o furiosa-feature-discovery cmd/main.go

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	golangci-lint run --timeout=30m

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test:
	go test ./...

.PHONY: cover
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	rm coverage.out

.PHONY: tidy
tidy:
	go mod tidy
