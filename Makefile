.PHONY: build test lint format check clean tidy install

# Go build settings
BINARY_NAME := guardian
GOFLAGS ?= -trimpath
LDFLAGS ?= -s -w
BUILD_DIR := bin

build:
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/guardian/

test:
	go test -v -race -count=1 -coverprofile=coverage.out ./...

lint:
	golangci-lint run ./...

format:
	gofumpt -w .
	goimports -w .

tidy:
	go mod tidy

check: format tidy lint test build
	@echo "All checks passed."

clean:
	go clean -cache -testcache
	rm -f coverage.out
	rm -rf $(BUILD_DIR)

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME) 2>/dev/null || \
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
