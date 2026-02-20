.PHONY: build run clean lint

build:
	go build -o bin/a2acli ./cmd/a2acli

run: build
	./bin/a2acli

lint:
	golangci-lint run

clean:
	rm -rf bin/