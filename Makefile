BINARY=pma-up

.PHONY: all build clean test lint e2e release local-release

all: lint test build

build:
	go build -o $(BINARY) ./cmd/pma-up

clean:
	rm -f $(BINARY)
	rm -rf dist/

test:
	go test ./internal/...

e2e:
	go test ./e2e

lint:
	golangci-lint run

local-release:
	goreleaser release --snapshot --clean

release:
	goreleaser release --clean

