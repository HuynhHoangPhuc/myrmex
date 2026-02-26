.PHONY: proto build test lint up down migrate seed reset-db demo demo-down demo-logs demo-reset

SERVICES := core module-hr module-subject module-timetable module-analytics
export PATH := $(HOME)/go/bin:$(PATH)

# Pass root .env to docker compose (compose file lives in a subdirectory)
COMPOSE := docker compose -f deploy/docker/compose.yml $(if $(wildcard .env),--env-file .env,)

proto:
	go tool buf generate

proto-lint:
	go tool buf lint

build:
	@for svc in $(SERVICES); do \
		echo "Building $$svc..."; \
		cd services/$$svc && go build ./... && cd ../..; \
	done

test:
	@for svc in $(SERVICES); do \
		echo "Testing $$svc..."; \
		cd services/$$svc && go test ./... && cd ../..; \
	done

test-cover:
	@for svc in $(SERVICES); do \
		echo "Testing $$svc with coverage..."; \
		cd services/$$svc && go test -coverprofile=coverage.out -covermode=atomic ./... && \
		go tool cover -func=coverage.out | tail -1 && cd ../..; \
	done

lint:
	go tool buf lint
	@for svc in $(SERVICES); do \
		cd services/$$svc && go vet ./... && cd ../..; \
	done

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

migrate:
	@for svc in $(SERVICES); do \
		echo "Migrating $$svc..."; \
		cd services/$$svc && go tool goose -dir migrations postgres "$$DATABASE_URL" up && cd ../..; \
	done

seed:
	@echo "Seeding demo data..."
	psql "$$DATABASE_URL" -f deploy/docker/seed.sql
	@echo "Seed complete."

reset-db:
	@for svc in $(SERVICES); do \
		echo "Resetting $$svc..."; \
		cd services/$$svc && go tool goose -dir migrations postgres "$$DATABASE_URL" reset && cd ../..; \
	done
	$(MAKE) migrate
	$(MAKE) seed

demo:
	$(COMPOSE) down -v --remove-orphans 2>/dev/null || true
	$(COMPOSE) up --build -d
	@echo ""
	@echo "Myrmex is starting up..."
	@echo "  Frontend: http://localhost:3000"
	@echo "  API:      http://localhost:8080"
	@echo ""
	@echo "Run 'make demo-logs' to see logs"
	@echo "Run 'make demo-down' to stop"

demo-down:
	$(COMPOSE) down

demo-logs:
	$(COMPOSE) logs -f

demo-reset:
	$(COMPOSE) down -v
	$(MAKE) demo
