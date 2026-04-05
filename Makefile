-include .env
PG_URL=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/$(POSTGRES_DB)?sslmode=disable
MIGRATIONS_PATH=migrations

migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(PG_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(PG_URL)" down 1

migrate-drop:
	migrate -path $(MIGRATIONS_PATH) -database "$(PG_URL)" drop

.PHONY: migrate-create migrate-up migrate-down migrate-drop