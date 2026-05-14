.PHONY: build test run

build:
	mkdir -p ../bin
	go build -o ../bin/agent0 .
	go test -c -o /dev/null ./...
	cp env-example ../bin/env-example

test:
	go test ./...

run:
	go run main.go