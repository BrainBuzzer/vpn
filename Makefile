.PHONY: build docker up docker-down
BINARY_NAME=simple-vpn

build:
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME) main.go

docker:
	docker-compose up -d --force-recreate --build --no-deps

docker-down:
	docker-compose down --volumes

up:	build docker

down: docker-down
