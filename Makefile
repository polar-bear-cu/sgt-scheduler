.PHONY: run build image container test lint format tidy

IMAGE ?= sgt-scheduler
PORT ?= 8090

run:
	go run .

build:
	go build -o bin/server .

image:
	docker build -t $(IMAGE) .

container: image
	docker run --rm --env-file .env -p $(PORT):8080 $(IMAGE)

test:
	go test ./... -cover

format:
	golangci-lint fmt

lint:
	golangci-lint run

tidy:
	go mod tidy
