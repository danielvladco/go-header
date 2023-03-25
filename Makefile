SHELL=/usr/bin/env bash
PROJECTNAME=$(shell basename "$(PWD)")
LDFLAGS=-ldflags="-X 'main.buildTime=$(shell date)' -X 'main.lastCommit=$(shell git rev-parse HEAD)' -X 'main.semanticVersion=$(shell git describe --tags --dirty=-dev)'"
ifeq (${PREFIX},)
	PREFIX := /usr/local
endif
## help: Get more info on make commands.
help: Makefile
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
.PHONY: help

## deps: install dependencies.
deps:
	@echo "--> Installing Dependencies"
	@go mod download
.PHONY: deps

## fmt: Formats only *.go (excluding *.pb.go *pb_test.go). Runs `gofmt & goimports` internally.
fmt: sort-imports
	@find . -name '*.go' -type f -not -path "*.git*" -not -name '*.pb.go' -not -name '*pb_test.go' | xargs gofmt -w -s
	@find . -name '*.go' -type f -not -path "*.git*"  -not -name '*.pb.go' -not -name '*pb_test.go' | xargs goimports -w -local github.com/celestiaorg
	@go mod tidy -compat=1.17
	@cfmt -w -m=100 ./...
	@markdownlint --fix --quiet --config .markdownlint.yaml .
.PHONY: sort-imports

## lint: Linting *.go files using golangci-lint. Look for .golangci.yml for the list of linters.
lint: lint-imports
	@echo "--> Running linter"
	@golangci-lint run
	@markdownlint --config .markdownlint.yaml '**/*.md'
	@cfmt -m=100 ./...
.PHONY: lint

## test-all: Running both unit and swamp tests
test:
	@echo "--> Running all tests without data race detector"
	@go test ./...
	@echo "--> Running all tests with data race detector"
	@go test -race ./...
.PHONY: test

## cover: generate to code coverage report.
cover:
	@echo "--> Generating Code Coverage"
	@go install github.com/ory/go-acc@latest
	@go-acc -o coverage.txt `go list ./... | grep -v nodebuilder/tests` -- -v
.PHONY: cover

## benchmark: Running all benchmarks
benchmark:
	@echo "--> Running benchmarks"
	@go test -run="none" -bench=. -benchtime=100x -benchmem ./...
.PHONY: benchmark

PB_PKGS=$(shell find . -name 'pb' -type d)
PB_CORE=$(shell go list -f {{.Dir}} -m github.com/tendermint/tendermint)
PB_GOGO=$(shell go list -f {{.Dir}} -m github.com/gogo/protobuf)
PB_CELESTIA_APP=$(shell go list -f {{.Dir}} -m github.com/celestiaorg/celestia-app)

lint-imports:
	@echo "--> Running imports linter"
	@for file in `find . -type f -name '*.go'`; \
		do goimports-reviser -list-diff -set-exit-status -company-prefixes "github.com/celestiaorg"  -project-name "github.com/celestiaorg/celestia-node" -output stdout $$file \
		 || exit 1;  \
    done;
.PHONY: lint-imports

sort-imports:
	@goimports-reviser -company-prefixes "github.com/celestiaorg"  -project-name "github.com/celestiaorg/celestia-node" -output stdout ./...
.PHONY: sort-imports
