.PHONY: build run clean lint test test-e2e install help
.DEFAULT_GOAL := help

# Default path to the a2a-go SDK repository, required for conformance tests.
A2A_GO_SRC ?= ../../github/a2a-go

# Version details for binary
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT  ?= $(shell git rev-parse --short HEAD)
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the a2acli binary
	go build -ldflags="$(LDFLAGS)" -o bin/a2acli ./cmd/a2acli

run: build ## Build and run the a2acli binary
	./bin/a2acli

lint: ## Run golangci-lint
	golangci-lint run

test: ## Run unit tests
	go test ./...

test-e2e: ## Run end-to-end conformance tests (requires a2a-go SDK)
	A2A_GO_SRC=$(A2A_GO_SRC) go test -v ./e2e/...

install: ## Install the binary to GOBIN
	go install -ldflags="$(LDFLAGS)" ./cmd/a2acli

clean: ## Remove build artifacts
	rm -rf bin/
