BINARY := azfloci
PKG := github.com/ankurCES/floc-zure
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X $(PKG)/internal/cli.Version=$(VERSION)"

.PHONY: build test lint e2e-test install clean fmt vet

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/azfloci

install:
	go install $(LDFLAGS) ./cmd/azfloci

test:
	go test -race -cover ./...

e2e-test:
	go test -race -tags=e2e -timeout 300s ./tests/e2e/...

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	rm -rf bin/ dist/

all: fmt vet lint test build
