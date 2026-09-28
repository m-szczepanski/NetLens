GO := go
LDFLAGS := -ldflags "-X main.version=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev)"

.PHONY: all build test run lint fmt vet clean

all: build

build:
	$(GO) build $(LDFLAGS) -o ../bin/netlens ./cmd/netlens

test:
	$(GO) test ./...

run:
	$(GO) run ./cmd/netlens

lint:
	$(GO) vet ./...
	@test -z "$(shell gofmt -l .)" || { echo "gofmt needed:"; gofmt -l .; exit 1; }
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipped (CI runs it via golangci-lint-action)"

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

clean:
	rm -rf ../bin
