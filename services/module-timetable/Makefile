MODULE_DIR := $(shell pwd)
WORKSPACE_ROOT := $(MODULE_DIR)/../..

.PHONY: build run vet tidy migrate-up migrate-down

## build: compile the timetable server binary
build:
	cd $(WORKSPACE_ROOT) && go build -o $(MODULE_DIR)/bin/timetable-server ./services/module-timetable/cmd/server

## run: start the timetable server locally
run:
	cd $(MODULE_DIR) && go run ./cmd/server

## vet: run go vet from workspace root
vet:
	cd $(WORKSPACE_ROOT) && go vet ./services/module-timetable/...

## tidy: tidy go.mod (with GOWORK=off for isolated module tidy)
tidy:
	cd $(MODULE_DIR) && GOWORK=off go mod tidy

## migrate-up: apply all pending migrations
migrate-up:
	goose -dir migrations postgres "$$DATABASE_URL" up

## migrate-down: roll back the last migration
migrate-down:
	goose -dir migrations postgres "$$DATABASE_URL" down
