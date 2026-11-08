BINARY := api
BIN_DIR := bin
CMD_DIR := ./cmd/api
 
.PHONY: build run test lint tidy
 
build:
	go build -o $(BIN_DIR)/$(BINARY) $(CMD_DIR)
 
run:
	set -a; [ -f .env ] && . ./.env; set +a; go run $(CMD_DIR)
 
test:
	go test -race ./...
 
lint:
	golangci-lint run
 
tidy:
	go mod tidy
 