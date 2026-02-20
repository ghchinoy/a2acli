.PHONY: build run clean lint test test-e2e

build:
	go build -o bin/a2acli ./cmd/a2acli

run: build
	./bin/a2acli

lint:
	golangci-lint run

test:
	go test ./...

test-e2e:
	go test -v ./e2e/...

clean:
	rm -rf bin/