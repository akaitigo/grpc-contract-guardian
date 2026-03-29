.PHONY: build test lint format check clean tidy install e2e

# Go build settings
BINARY_NAME := guardian
GOFLAGS ?= -trimpath
LDFLAGS ?= -s -w
BUILD_DIR := bin

build:
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/guardian/

test: build
	go test -v -race -count=1 -coverprofile=coverage.out ./...
	@if command -v bats >/dev/null 2>&1; then \
		bats test/cli.bats; \
	else \
		echo "bats not found, skipping E2E tests"; \
	fi

e2e: build
	bats test/cli.bats

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
	@INSTALL_DIR="$${GOPATH:-$$(go env GOPATH)}/bin"; \
	mkdir -p "$$INSTALL_DIR" && \
	cp $(BUILD_DIR)/$(BINARY_NAME) "$$INSTALL_DIR/$(BINARY_NAME)" && \
	echo "Installed to $$INSTALL_DIR/$(BINARY_NAME)"
