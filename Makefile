.PHONY: build build-cli build-unlocker proto proto-lint clean cov fmt help install integrationtest mock run test trade-cert vet

install:
	@echo "Installing deps..."
	@go mod download
	@go mod tidy

## build: build for all platforms
build:
	@echo "Building tdexd binary..."
	@bash ./scripts/build

## build-cli: build CLI for all platforms
build-cli:
	@echo "Building tdex binary..." 
	@bash ./scripts/build-cli

## proto: compile proto stubs
proto: proto-lint
	@echo "Compiling stubs..."
	@buf generate buf.build/tdex-network/tdex-protobuf
	@buf generate buf.build/vulpemventures/ocean
	@buf generate

## proto-lint: lint protos & detect breaking changes
proto-lint:
	@echo "Linting protos..."
	@buf lint

## clean: cleans the binary
clean:
	@echo "Cleaning..."
	@go clean

## 
trade-cert:
	@echo "Creating self-signed cert for trade interface..."
	@bash ./scripts/tlscert

## cov: generates coverage report
cov:
	@echo "Coverage..."
	go test -cover ./...

## fmt: Go Format
fmt:
	@echo "Checking code format..."
	@if [ -n "$(gofmt -l .)" ]; then echo "Go code is not formatted"; exit 1; fi


## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## run: Run locally with default configuration in regtest
run: clean
	@export TDEX_NETWORK=regtest; \
	export TDEX_WALLET_ADDR=localhost:18000; \
	export TDEX_LOG_LEVEL=5; \
	export TDEX_FEE_ACCOUNT_BALANCE_THRESHOLD=1000; \
	export TDEX_NO_MACAROONS=true; \
	export TDEX_NO_OPERATOR_TLS=true; \
	export TDEX_CONNECT_PROTO=http; \
	go run ./cmd/tdexd


## vet: code analysis
vet:
	@echo "Running code analysis..."
	@go vet ./...

## test: runs unit and component tests
test: fmt mock
	@echo "Running unit tests..."
	@go test -v -count=1 -race ./...

## integrationtest: runs e2e test
integrationtest:
	@echo "Running integration tests..."
	@go run test/e2e/main.go

## mock: generates mocks for unit tests
mock:
	@echo "Generating mocks for unit tests..."
	@mockery --dir=internal/core/domain --name=SwapParser --structname=MockSwapParser --filename=swap.go --output=internal/core/domain/mocks