.PHONY: db-up db-down db-logs db-psql migrate migrate-info migrate-repair clean deps build run sqlc sqlc-docker up down logs api-build

# Use docker compose for orchestration
DC := docker compose

db-up:
	$(DC) up -d db

db-down:
	$(DC) down --volumes --remove-orphans

db-logs:
	$(DC) logs -f db

db-psql:
	$(DC) exec db psql -U $$POSTGRES_USER -d $$POSTGRES_DB

migrate:
	$(DC) run --rm flyway

migrate-info:
	$(DC) run --rm flyway info -url=jdbc:postgresql://db:5432/$$POSTGRES_DB -user=$$POSTGRES_USER -password=$$POSTGRES_PASSWORD -connectRetries=60 -locations=filesystem:/flyway/sql

migrate-repair:
	$(DC) run --rm flyway repair -url=jdbc:postgresql://db:5432/$$POSTGRES_DB -user=$$POSTGRES_USER -password=$$POSTGRES_PASSWORD -connectRetries=60 -locations=filesystem:/flyway/sql

clean:
	$(DC) down --volumes --remove-orphans

deps:
	go mod tidy

build:
	go build -o bin/server ./cmd/api

run:
	go run ./cmd/api

sqlc:
	sqlc generate

sqlc-docker:
	docker run --rm -v ${PWD}:/src -w /src --network host kjconroy/sqlc:1.27.0 generate

up:
	docker compose up -d --build db api

down:
	docker compose down --volumes --remove-orphans

logs:
	docker compose logs -f api db

api-build:
	docker compose build api

