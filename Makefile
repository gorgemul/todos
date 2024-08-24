include .env

build:
	go build -o ./bin/todo ./cmd/todos/main.go

run:
	go run ./cmd/todos/main.go

test:
	go test ./test/...

migrate_up:
	migrate -database ${TODO_DB} -path internal/db/migrations up

migrate_down:
	migrate -database ${TODO_DB} -path internal/db/migrations down

clean:
	go clean
	rm ./bin/todo

.PHONY: build run test clean migrate_up migrate_down
