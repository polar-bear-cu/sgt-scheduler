.PHONY: run build test lint format tidy

run:
	go run .

build:
	go build -o bin/server .

test:
	go test ./... -cover

format:
	golangci-lint fmt

lint:
	golangci-lint run

tidy:
	go mod tidy
