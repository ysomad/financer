include .env
export

export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING=${PG_URL}

.PHONY:
generate:
	rm -rf internal/gen/proto/
	buf generate proto/

.PHONY:
run:
	go mod tidy && go mod download && \
	go run ./cmd/tgbot \
	go run ./cmd/server


.PHONY:
run-bot:
	go mod tidy && go mod download && \
	go run ./cmd/tgbot

.PHONY:
run-server:
	go mod tidy && go mod download && \
	go run ./cmd/server

.PHONY: run-server-migrate
run-server-migrate:
	go mod tidy && go mod download && \
	go run ./cmd/server -migrate

.PHONY: dry-run-server
dry-run-server: goose-reset run-server-migrate

.PHONY: compose-up
compose-up:
	docker-compose up --build -d postgres && docker-compose logs -f

.PHONY: compose-down
compose-down:
	docker-compose down --remove-orphans

.PHONY: goose-new
goose-new:
	@read -p "Enter the name of the new migration: " name; \
	goose -dir migrations create $${name// /_} sql

.PHONY: goose-up
goose-up:
	@echo "Running all new database migrations..."
	goose -dir migrations validate
	goose -dir migrations up

.PHONY: goose-down
goose-down:
	@echo "Running all down database migrations..."
	goose -dir migrations down

.PHONY: goose-reset
goose-reset:
	@echo "Dropping everything in database..."
	goose -dir migrations reset

.PHONY: goose-status
goose-status:
	goose -dir migrations status
