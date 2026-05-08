.PHONY: build test run


build:
	go build -o ../bin/agent0 .
	go test -c -o /dev/null ./...

test:
	go test ./...

run:
	go run main.go