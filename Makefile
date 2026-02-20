.PHONY: build run clean

build:
	go build -o bin/a2acli ./cmd/a2acli

run: build
	./bin/a2acli

clean:
	rm -rf bin/