.DEFAULT_GOAL := start
.PHONY: start test

start:
	go run cmd/api/main.go --config=config/config.yaml

test:
	go test -v -race -timeout 30s ./...

up-migrations:
	go run cmd/migrator/main.go --migrations-path=./migrations --migrations-type=up

down-migrations:
	go run cmd/migrator/main.go --migrations-path=./migrations --migrations-type=down