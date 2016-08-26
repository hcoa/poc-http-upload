SHELL := /bin/bash
# APP := $(shell basename "$(CURDIR)")
APP := app
CONTAINER_NAME := http-load-poc
default: build

help:
	@echo "  vendor             populate application dependecies"
	@echo "  build              build application"
	@echo "  build-mac          build application for darwin amd64"
	@echo "  build-container    build container with application"

vendor:
	@export GOPATH="$(CURDIR)" && pushd ./src/$(APP)/ && glide up --strip-vcs --update-vendored; popd

build:
	@export GOPATH="$(CURDIR)" && pushd ./src/$(APP)/ && GOOS=linux GOARCH=amd64 go build -o ../../app . ; popd

build-mac:
	@export GOPATH="$(CURDIR)" && pushd ./src/$(APP)/ && GOOS=darwin GOARCH=amd64 go build -o ../../app . ; popd


build-container: build
	@docker build -t $(CONTAINER_NAME) .

cleanup:
	@rm app
