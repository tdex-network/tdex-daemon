.PHONY: build build-cli proto clean create-cert cov fmt help install integrationtest run test vet

install:
	go mod download
	go mod tidy

## build: build for all platforms
build: 
	chmod u+x ./scripts/build
	./scripts/build

## build-cli: build CLI for all platforms
build-cli: 
	chmod u+x ./scripts/build-cli
	./scripts/build-cli

## proto: compile proto files
proto: install
	chmod u+x ./scripts/compile_proto
	./scripts/compile_proto

## clean: cleans the binary
clean:
	@echo "Cleaning..."
	@go clean

## cov: generates coverage report
cov:
	@echo "Coverage..."
	go test -cover ./...

## fmt: Go Format
fmt:
	@echo "Gofmt..."
	@if [ -n "$(gofmt -l .)" ]; then echo "Go code is not formatted"; exit 1; fi


## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## run: Run locally with default configuration in regtest
run: clean
	export TDEX_NETWORK=regtest; \
	export TDEX_EXPLORER_ENDPOINT=http://127.0.0.1:3001; \
	export TDEX_LOG_LEVEL=5; \
	export TDEX_BASE_ASSET=5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225; \
	export TDEX_FEE_ACCOUNT_BALANCE_THRESHOLD=1000; \
	export TDEX_NO_TLS=true; \
	export TDEX_NO_MACAROONS=true; \
	go run ./cmd/tdexd


## vet: code analysis
vet:
	@echo "Vet..."
	@go vet ./...

## test: runs go unit test with default values
test: fmt shorttest

## shorttest: runs unit tests by skipping those that are time expensive
shorttest:
	@echo "Testing..."
	export TDEX_NETWORK=regtest; \
	export TDEX_EXPLORER_ENDPOINT=http://127.0.0.1:3001; \
	export TDEX_LOG_LEVEL=5; \
	export TDEX_BASE_ASSET=5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225; \
	export TDEX_FEE_ACCOUNT_BALANCE_THRESHOLD=1000; \
	go test -v -count=1 -race -short ./...

## integrationtest: runs e2e test
integrationtest:
	go run test/e2e/main.go