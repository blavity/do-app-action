.PHONY: lint fmt test build check

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## fmt: format code with gofumpt and goimports
fmt:
	gofumpt -l -w .
	goimports -local github.com/blavity/do-app-action -w .

## test: run tests with race detector
test:
	go test -race ./...

## build: build all action binaries
build:
	go build ./deploy ./archive ./delete ./unarchive

## check: fmt + lint + test (run before pushing)
check: fmt lint test

## help: print this help
help:
	@grep -E '^## ' Makefile | sed 's/## //'
