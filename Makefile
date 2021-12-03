# set default shell
SHELL = bash -e -o pipefail

# Variables
VERSION                  ?= $(shell cat ./VERSION)

default: build-win

help:
	@echo "Usage: make [<target>]"
	@echo "where available targets are:"
	@echo
	@echo "help              : Print this help"
	@echo "test              : Run unit tests, if any"
	@echo "sca               : Run SCA"
	@echo

test:
	go test -v ./...

sca:
	golangci-lint run

fmt:
	 goimports -w .
	 gofmt -w .
