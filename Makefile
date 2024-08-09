.PHONY: build run test clean

build:
	go build -o ./bin/todo ./cmd/todos/main.go

run: 
	go run ./cmd/todos/main.go

test:
	go test ./test/...

clean:
	go clean
	rm ./bin/todo
