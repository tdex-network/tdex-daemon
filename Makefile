.PHONY: build-arm build-linux build-mac clean cov fmt help vet test

## build-arm: build binary for ARM
build-arm:
	export GO111MODULE=on
	chmod u+x ./scripts/build
	./scripts/build linux arm

## build-linux: build binary for Linux
build-linux:
	export GO111MODULE=on
	chmod u+x ./scripts/build
	./scripts/build linux amd64

## build-mac: build binary for Mac
build-mac:
	export GO111MODULE=on
	chmod u+x ./scripts/build
	./scripts/build darwin amd64

## build-cli-arm: build CLI for ARM
build-cli-arm:
	export GO111MODULE=on
	chmod u+x ./scripts/build
	./scripts/build-cli linux arm

## build-linux: build CLI for Linux
build-cli-linux:
	export GO111MODULE=on
	chmod u+x ./scripts/build
	./scripts/build-cli linux amd64

## build-cli-mac: build CLI for Mac
build-cli-mac:
	export GO111MODULE=on
	chmod u+x ./scripts/build
	./scripts/build-cli darwin amd64


## build: build for all platforms
build: build-arm build-linux build-mac

## build-cli: build CLI for all platforms
build-cli: build-cli-arm build-cli-linux build-cli-mac


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

## run-linux: Run locally with default configuration
run-linux: clean build-linux
	./build/tdexd-linux-amd64

## run-mac: Run locally with default configuration
run-mac: clean build-mac
	./build/tdexd-darwin-amd64

## vet: code analysis
vet:
	@echo "Vet..."
	@go vet ./...

## clean-test: remove test folders
clean-test:
	@echo "Deleting test folders..."
	rm -rf ./internal/core/application/testDatadir
	rm -rf ./internal/infrastructure/storage/db/badger/testdb

## test: runs go unit test with default values
test: clean-test fmt shorttest

## shorttest: runs unit tests by skipping those that are time expensive
shorttest:
	@echo "Testing..."
	go test -v -count=1 -race -short ./...

## integrationtest: runs e2e tests by
integrationtest:
	@echo "E2E Testing..."
	go test -v -count=1 -race ./cmd/tdexd

