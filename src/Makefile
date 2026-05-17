.PHONY: build build-win test run

# build for linux and macos
build:
	mkdir -p ../bin
	go build -o ../bin/agent0 .
	go test -c -o /dev/null ./...
	cp env-example ../bin/env-example

# build for windows
build-win:
	@if not exist "..\bin" mkdir "..\bin"
	go build -o ..\bin\agent0.exe .
	go test -c -o NUL ./...
	copy env-example ..\bin\env-example

test:
	go test ./...

run:
	go run main.go