.PHONY: build run clean lint test test-e2e

# Default path to the a2a-go SDK repository, required for conformance tests.
A2A_GO_SRC ?= ../../github/a2a-go

build:
	go build -o bin/a2acli ./cmd/a2acli

run: build
	./bin/a2acli

lint:
	golangci-lint run

test:
	go test ./...

test-e2e:
	A2A_GO_SRC=$(A2A_GO_SRC) go test -v ./e2e/...

clean:
	rm -rf bin/