include .env
export

DB_URL = postgres://$(PG_DB_USER):$(PG_DB_PASSWORD)@$(PG_DB_HOST):$(PG_DB_PORT)/$(PG_DB_NAME)?sslmode=disable

.PHONY: run build migrate-up migrate-down migrate-create clean dev test docs lint docker-build docker-run

run:
	go run cmd/main.go

build:
	go build -o bin/neurolab-service cmd/main.go

build-docker:
	echo "Building checker-python:1.0..."
	docker build -t checker-python:1.0 -f data/docker/checker-python/Dockerfile .
	echo "Building checker-golang:1.0..."
	docker build -t checker-golang:1.0 -f data/docker/checker-golang/Dockerfile .

dev:
	docker-compose up -d postgres

dev-down:
	docker-compose down

migrate-up:
	goose -dir migrations postgres "$(DB_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DB_URL)" down

test-integration:
	go test ./tests/integration/... -v -tags=integration

docs: clean
	swag init -g ./cmd/main.go --output ./docs --parseDependency --parseInternal

clean:
	rm -rf docs/ bin/