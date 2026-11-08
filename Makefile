BINARY := app
BIN_DIR := bin

.PHONY: build run test lint tidy

build:
	go build -o $(BIN_DIR)/$(BINARY) .

run:
	set -a; [ -f .env ] && . ./.env; set +a; go run .

test:
	go test -race ./...

lint:
	golangci-lint run

tidy:
	go mod tidy