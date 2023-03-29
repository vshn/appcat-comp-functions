# Set Shell to bash, otherwise some targets fail with dash/zsh etc.
SHELL := /bin/bash

# Disable built-in rules
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --no-builtin-variables
.SUFFIXES:
.SECONDARY:
.DEFAULT_GOAL := help

# General variables
include Makefile.vars.mk

# Following includes do not print warnings or error if files aren't found
# Optional Documentation module.
-include docs/antora-preview.mk docs/antora-build.mk
# Optional kind module
-include kind/kind.mk

golangci_bin = $(go_bin)/golangci-lint

.PHONY: help
help: ## Show this help
	@grep -E -h '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

## BUILD an instance
.PHONY: build
build: build-bin build-docker ## All-in-one build

.PHONY: build-bin
build-bin: export CGO_ENABLED = 0
build-bin: export GOOS = linux
build-bin: export GARCH = amd64
build-bin: fmt vet ## Build binary
	@go build -o $(PROJECT_ROOT_DIR)/cmd/$(instance)/$(BIN_FILENAME) $(PROJECT_ROOT_DIR)/cmd/$(instance)

.PHONY: build-docker
build-docker: build-bin ## Build docker image
	$(DOCKER_CMD) build -t $(CONTAINER_IMG) $(PROJECT_ROOT_DIR)/cmd/$(instance)

## BUILD all instances
.PHONY: build-all
build-all: build-bin-all build-docker-all ## All-in-one build for all instances

.PHONY: build-bin-all
build-bin-all: recursive_target=build-bin
build-bin-all: $(instances) ## Build binaries

.PHONY: build-docker-all
build-docker-all: recursive_target=build-docker
build-docker-all: $(instances) ## Build docker images

.PHONY: test
test: test-go ## All-in-one test

.PHONY: test-go
test-go: ## Run unit tests against code
	go test -race -covermode atomic ./...

.PHONY: fmt
fmt: ## Run 'go fmt' against code
	go fmt ./...

.PHONY: vet
vet: ## Run 'go vet' against code
	go vet ./...

.PHONY: lint
lint: generate fmt golangci-lint git-diff ## All-in-one linting

.PHONY: golangci-lint
golangci-lint: $(golangci_bin) ## Run golangci linters
	$(golangci_bin) run --timeout 5m --out-format colored-line-number ./...

.PHONY: git-diff
git-diff:
	@echo 'Check for uncommitted changes ...'
	git diff --exit-code

.PHONY: generate
generate: ## Generate additional code and artifacts
	@go generate ./...

.PHONY: clean
clean: kind-clean ## Cleans local build artifacts
clean: recursive_target=.clean-build-img
clean: $(instances)
	rm -rf docs/node_modules $(docs_out_dir) dist .cache $(WORK_DIR)

.PHONY: .clean-build-img
.clean-build-img:
	docker rmi $(CONTAINER_IMG) -f || true
	rm $(PROJECT_ROOT_DIR)/cmd/$(instance)/$(BIN_FILENAME)

$(golangci_bin): | $(go_bin)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go_bin)"

.PHONY: $(instances)
$(instances):
	$(MAKE) $(recursive_target) -e instance=$(basename $(@F))
