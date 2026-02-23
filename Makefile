VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -s -w -X main.Version=$(VERSION)

## Build the project
build:
	go build -ldflags "$(LDFLAGS)" -o skit .

## Run all tests
test:
	go test -race ./...

## Run static analysis
vet:
	go vet ./...

## Clean build artifacts
clean:
	rm -f skit
	rm -rf dist/

## Build for all platforms
dist: clean
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/skit-linux-amd64 .
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/skit-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/skit-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/skit-windows-amd64.exe .

## Format source code
fmt:
	go fmt ./...

## Run tests with coverage
coverage:
	go test -race -cover ./...

## Install to GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" .

## Run vet, test, and build
all: vet test build
