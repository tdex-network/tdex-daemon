.PHONY: build build-cli build-migration clean cov help integrationtest lint lint-fix mock proto proto-lint run test trade-cert vet

## build: build tdexd
build: clean
	@echo "Building tdexd binary..."
	@bash ./scripts/build

## build-cli: build cli
build-cli: clean
	@echo "Building tdex binary..." 
	@bash ./scripts/build-cli

## build-migration: build migration
build-migration: clean
	@echo "Building migration binary..."
	@bash ./scripts/build-migration

## clean: clean files and cached files
clean:
	@echo "Cleaning..."
	@go clean

## cov: generate coverage report
cov:
	@echo "Coverage..."
	@go test -cover ./...

## help: print this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## integrationtest: run e2e test
integrationtest:
	@echo "Running integration tests..."
	@go run test/e2e/main.go

## lint: check code format
lint: 
	@echo "Check linting..."
	@golangci-lint run

## lint-fix: check & fix code format if possible
lint-fix:
	@echo "Linting code..."
	@golangci-lint run --fix

## mock: generate mocks for unit tests
mock:
	@echo "Generating mocks for unit tests..."
	@mockery --dir=internal/core/domain --name=SwapParser --structname=MockSwapParser --filename=swap.go --output=internal/core/domain/mocks

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

## run: run tdexd in dev mode
run: clean
	@export TDEX_WALLET_ADDR=localhost:18000; \
	export TDEX_LOG_LEVEL=5; \
	export TDEX_FEE_ACCOUNT_BALANCE_THRESHOLD=1000; \
	export TDEX_NO_MACAROONS=true; \
	export TDEX_NO_OPERATOR_TLS=true; \
	export TDEX_CONNECT_PROTO=http; \
	go run ./cmd/tdexd

## test: run unit and component tests
test:
	@echo "Running unit tests..."
	@go test -v -count=1 -race ./...

## trade-cert: generate self-signed cert for the trade interface
trade-cert:
	@echo "Creating self-signed cert for trade interface..."
	@bash ./scripts/tlscert

## vet: code analysis
vet:
	@echo "Running code analysis..."
	@go vet ./...