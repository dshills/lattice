.PHONY: all build build-go build-frontend test test-go test-frontend lint lint-go lint-frontend run run-api run-ui install migrate migrate-down migrate-status clean

# Load .env if present
ifneq (,$(wildcard .env))
  include .env
  export
endif

BINARY     := lattice
BUILD_DIR  := bin
GO_PKG     := ./cmd/lattice
DB_DSN     := mysql://$(LATTICE_DB_USER):$(LATTICE_DB_PASSWORD)@tcp($(LATTICE_DB_HOST):$(LATTICE_DB_PORT))/$(LATTICE_DB_NAME)
MIGRATIONS := migrations

## all: build everything
all: build

## build: build Go binary and frontend
build: build-go build-frontend

## build-go: compile the Go server
build-go:
	go build -o $(BUILD_DIR)/$(BINARY) $(GO_PKG)

## build-frontend: build the React frontend
build-frontend:
	cd frontend && npm run build

## test: run all tests
test: test-go test-frontend

## test-go: run Go tests
test-go:
	go test ./...

## test-frontend: run frontend tests
test-frontend:
	cd frontend && npm test

## lint: lint everything
lint: lint-go lint-frontend

## lint-go: run golangci-lint
lint-go:
	golangci-lint run ./...

## lint-frontend: run eslint
lint-frontend:
	cd frontend && npm run lint

## run: start the Go API server
run: run-api

## run-api: start the Go API server with hot reload
run-api:
	air

## run-ui: start the Vite dev server
run-ui:
	cd frontend && npm run dev

## install: install all dependencies
install:
	go mod download
	cd frontend && npm install

## migrate: run database migrations up
migrate:
	go run ./cmd/migrate up

## migrate-down: roll back the last migration
migrate-down:
	go run ./cmd/migrate down 1

## migrate-status: show current migration version
migrate-status:
	@MYSQL_PWD=$(LATTICE_DB_PASSWORD) mysql -u$(LATTICE_DB_USER) -h$(LATTICE_DB_HOST) -P$(LATTICE_DB_PORT) $(LATTICE_DB_NAME) -e "SELECT version FROM schema_migrations ORDER BY version;"

## clean: remove build artifacts
clean:
	rm -rf $(BUILD_DIR) frontend/dist

## help: show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## //' | column -t -s ':'
