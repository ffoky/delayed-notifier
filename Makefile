ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: test fmt vet lint format gen migrate-up migrate-down migrate-status migrate-create up down

GOOSE_DRIVER ?= postgres
GOOSE_DBSTRING ?= $(DATABASE_URL)
GOOSE_MIGRATION_DIR ?= ./migrations

test:
	@echo "go test -race ./..."
	@go test -race ./...

fmt:
	@echo "go fmt ./..."
	@go fmt ./...

vet:
	@echo "go vet ./..."
	@go vet ./...

lint:
	@echo "golangci-lint run -j8 --fix"
	@golangci-lint run -j8 --fix

format: fmt vet lint

gen:
	@echo "go generate ./..."
	@go generate ./...
	@oapi-codegen -config api/oapi-codegen.yaml api/web-server.yaml

migrate-up:
	@echo "goose up"
	@GOOSE_DRIVER=$(GOOSE_DRIVER) \
	GOOSE_DBSTRING='$(GOOSE_DBSTRING)' \
	GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) \
	goose up

migrate-down:
	@echo "goose down"
	@GOOSE_DRIVER=$(GOOSE_DRIVER) \
	GOOSE_DBSTRING='$(GOOSE_DBSTRING)' \
	GOOSE_MIGRATION_DIR=$(GOOSE_MIGRATION_DIR) \
	goose down

up:
	@echo "docker compose up --build"
	@docker compose up --build

down:
	@echo "docker compose down -v"
	@docker compose down -v
