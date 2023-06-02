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
test: fmt
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

######## PG_DB ########
## pg: starts postgres db inside docker container
pg:
	@echo "Starting postgres container..."
	@docker run --name tdexd-pg -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=tdexd -d postgres

## droppg: stop and remove postgres container
droppg:
	@echo "Stopping and removing postgres container..."
	@docker stop tdexd-pg
	@docker rm tdexd-pg

## createdb: create db inside docker container
createdb:
	@echo "Creating db..."
	@docker exec tdexd-pg createdb --username=root --owner=root tdexd

## createtestdb: create test db inside docker container
createtestdb:
	@echo "Creating test db..."
	@docker exec tdexd-pg createdb --username=root --owner=root tdexd-test

## recreatedb: drop and create main db
recreatedb: dropdb createdb

## recreatetestdb: drop and create main and test db
recreatetestdb: droptestdb createtestdb

## pgcreatetestdb: starts docker container and creates test db, used in CI
pgcreatetestdb: pg sleep createtestdb
	@echo "Starting postgres container with test db..."

## dropdb: drops db inside docker container
dropdb:
	@echo "Dropping db..."
	@docker exec tdexd-pg dropdb tdexd

## droptestdb: drops test db inside docker container
droptestdb:
	@echo "Dropping test db..."
	@docker exec tdexd-pg dropdb tdexd-test

## mig_file: creates pg migration file(eg. make FILE=init mig_file)
mig_file:
	@echo "creating migration file..."
	@migrate create -ext sql -dir ./internal/infrastructure/storage/db/pg/migration $(FILE)

## mig_up_test: creates test db schema
mig_up_test:
	@echo "creating test db schema..."
	@echo "creating db schema..."
	@migrate -database "postgres://root:secret@localhost:5432/tdexd-test?sslmode=disable" -path ./internal/infrastructure/storage/db/pg/migration up

## mig_up: creates db schema
mig_up:
	@echo "creating db schema..."
	@migrate -database "postgres://root:secret@localhost:5432/tdexd?sslmode=disable" -path ./internal/infrastructure/storage/db/pg/migration up

## mig_down_test: apply down migration on test db
mig_down_test:
	@echo "migration down on test db..."
	@migrate -database "postgres://root:secret@localhost:5432/tdexd-test?sslmode=disable" -path ./internal/infrastructure/storage/db/pg/migration down

## mig_down: apply down migration without prompt
mig_down:
	@echo "migration down..."
	@"yes" | migrate -database "postgres://root:secret@localhost:5432/tdexd?sslmode=disable" -path ./internal/infrastructure/storage/db/pg/migration down

## vet_db: check if mig_up and mig_down are ok
vet_db: recreatedb mig_up mig_down
	@echo "vet db migration scripts..."

## sqlc: gen sql
sqlc:
	@echo "gen sql..."
	@cd ./internal/infrastructure/storage/db/pg; sqlc generate

sleep:
	@echo "sleeping for 3 seconds..."
	@sleep 3