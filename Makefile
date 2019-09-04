PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)

.PHONY: all dep build clean test coverage coverhtml lint

all: build

test: ## Run unittests
	go test -short ./...

race: ## Run data race detector
	go test -race -short ./...

msan: ## Run memory sanitizer
	go test -msan -short ./...

build: ## Build the binary file
	go build -i -v ./...