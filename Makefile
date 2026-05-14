-include .env
export

export PROJECT_ROOT=$(shell pwd)

PG_URL=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable
MIGRATIONS_PATH=migrations

env-up:
	docker-compose up -d postgres redis minio

env-down:
	docker-compose down

migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(PG_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(PG_URL)" down 1

migrate-drop:
	migrate -path $(MIGRATIONS_PATH) -database "$(PG_URL)" drop

run:
	go run cmd/gopher-drive/main.go

.PHONY: migrate-create migrate-up migrate-down migrate-drop