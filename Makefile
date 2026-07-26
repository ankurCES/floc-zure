BINARY := azfloci
PKG := github.com/ankurCES/floc-zure
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X $(PKG)/internal/cli.Version=$(VERSION)"

.PHONY: build test lint lint-ci e2e-test install clean fmt vet sim-build sim-test test-all test-sim-unit all

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/azfloci

install:
	go install $(LDFLAGS) ./cmd/azfloci

test:
	go test -race -cover ./...

test-sim-unit:
	go test -race -cover ./simulator/...

test-all: test test-sim-unit

e2e-test:
	go test -race -tags=e2e -timeout 300s -v ./tests/e2e/...

lint:
	golangci-lint run ./...

lint-ci:
	golangci-lint run --out-format=github-actions ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

sim-build:
	go build -o bin/az-simulator ./simulator/cmd/az

sim-test: build sim-build
	AZFLOCI_AZ_PATH=./bin/az-simulator AZFLOCI_SIM_STATE=$$(mktemp) go test -race -tags=e2e -timeout 120s -v ./tests/e2e/...

clean:
	rm -rf bin/ dist/

all: fmt vet lint test build sim-build
