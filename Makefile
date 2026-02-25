.PHONY: proto build test lint up down migrate seed reset-db demo demo-down demo-logs demo-reset

SERVICES := core module-hr module-subject module-timetable
export PATH := $(HOME)/go/bin:$(PATH)

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

lint:
	go tool buf lint
	@for svc in $(SERVICES); do \
		cd services/$$svc && go vet ./... && cd ../..; \
	done

up:
	docker compose -f deploy/docker/compose.yml up -d

down:
	docker compose -f deploy/docker/compose.yml down

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
	docker compose -f deploy/docker/compose.yml down -v --remove-orphans 2>/dev/null || true
	docker compose -f deploy/docker/compose.yml up --build -d
	@echo ""
	@echo "Myrmex is starting up..."
	@echo "  Frontend: http://localhost:3000"
	@echo "  API:      http://localhost:8080"
	@echo ""
	@echo "Run 'make demo-logs' to see logs"
	@echo "Run 'make demo-down' to stop"

demo-down:
	docker compose -f deploy/docker/compose.yml down

demo-logs:
	docker compose -f deploy/docker/compose.yml logs -f

demo-reset:
	docker compose -f deploy/docker/compose.yml down -v
	$(MAKE) demo
