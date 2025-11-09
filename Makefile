BIN_DIR ?= bin
BINARY ?= $(BIN_DIR)/ipcounter

.PHONY: build run test fmt clean

build: ## Build the ipcounter binary.
	mkdir -p $(BIN_DIR)
	go build -o $(BINARY) ./cmd/ipcounter

run: ## Run the ipcounter application.
	go run ./cmd/ipcounter

test: ## Execute all Go tests.
	go test ./...

test-race: ## Execute all Go tests with race detection.
	go test -race ./...

bench: ## Execute all Go benchmarks.
	go test -bench=. ./...

test-all: ## Execute all Go tests and benchmarks.
	test
	test-race
	bench

fmt: ## Format the ipcounter codebase.
	go fmt ./cmd/ipcounter ./internal/...

clean: ## Remove built artifacts.
	rm -rf $(BIN_DIR)

