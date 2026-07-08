ifeq ($(OS),Windows_NT)
GO ?= C:/Program\ Files/Go/bin/go.exe
else
GO ?= go
endif
DATABASE_URL ?= postgres://aipass:aipass@localhost:5432/aipass?sslmode=disable

.PHONY: tidy fmt test run-access-api compose-up compose-down migrate-up

tidy:
	$(GO) mod tidy

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

run-access-api:
	$(GO) run ./cmd/access-api

compose-up:
	docker compose up --build -d postgres redis minio redpanda access-api

compose-down:
	docker compose down

migrate-up:
	$(GO) run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.1 -path migrations -database "$(DATABASE_URL)" up
