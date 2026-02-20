MODULE_NAME := module-subject
BINARY      := bin/$(MODULE_NAME)
MAIN        := ./cmd/server/main.go

.PHONY: build run migrate sqlc deps tidy vet

build:
	go build -o $(BINARY) $(MAIN)

run:
	go run $(MAIN)

deps:
	go get github.com/jackc/pgx/v5@latest
	go get go.uber.org/zap@latest
	go get github.com/spf13/viper@latest
	go get github.com/google/uuid@latest
	go get google.golang.org/grpc@latest
	go get google.golang.org/protobuf@latest

tidy:
	go mod tidy

vet:
	go vet ./...

sqlc:
	go tool sqlc generate

migrate-up:
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DATABASE_URL)" down
