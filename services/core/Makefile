.PHONY: build test lint migrate run sqlc

build:
	go build -o bin/server ./cmd/server

test:
	go test ./...

lint:
	go vet ./...

migrate:
	go tool goose -dir migrations postgres "$(DB_URL)" up

run:
	go tool air

sqlc:
	go tool sqlc generate
