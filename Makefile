APP_NAME=sym

.PHONY: help build test clean

help:
	@echo "Welcome to sym build"

build:
	go build -o bin/

test:
	go test -v ./...

clean:
	go clean
	rm -rf bin/
