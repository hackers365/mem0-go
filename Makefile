.PHONY: build test clean release snapshot

build:
	mkdir -p bin
	go build -o bin/mem0 ./cmd/mem0

test:
	go test -v ./...

clean:
	rm -rf bin/

install: build
	cp bin/mem0 /usr/local/bin/

all: clean build test

deps:
	go mod tidy
	go mod download

# Release commands
release:
	goreleaser release --clean

snapshot:
	goreleaser release --snapshot --clean --skip-publish 