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
build-bin: fmt vet ## Build binary
	@go build -o $(BIN_FILENAME) .

.PHONY: build-docker
build-docker: ## Build docker image
	env CGO_ENABLED=0 GOOS=$(DOCKER_IMAGE_GOOS) GOARCH=$(DOCKER_IMAGE_GOARCH) \
		go build -o ${BIN_FILENAME}
	$(DOCKER_CMD) build -t $(CONTAINER_IMG) .

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
	docker rmi $(CONTAINER_IMG) -f || true
	rm -rf docs/node_modules $(docs_out_dir) dist .cache $(WORK_DIR)

.PHONY: push-docker
push-docker: build-docker ## Push docker image with the manager.
	docker push $(CONTAINER_IMG)

$(golangci_bin): | $(go_bin)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go_bin)"


.PHONY: generate-stackgres-crds
generate-stackgres-crds:
	curl ${STACKGRES_CRD_URL}/SGDbOps.yaml?inline=false -o apis/stackgres/v1/sgdbops_crd.yaml
	yq -i e apis/stackgres/v1/sgdbops.yaml --expression ".components.schemas.SGDbOpsSpec=load(\"apis/stackgres/v1/sgdbops_crd.yaml\").spec.versions[0].schema.openAPIV3Schema.properties.spec"
	yq -i e apis/stackgres/v1/sgdbops.yaml --expression ".components.schemas.SGDbOpsStatus=load(\"apis/stackgres/v1/sgdbops_crd.yaml\").spec.versions[0].schema.openAPIV3Schema.properties.status"
	go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --package=v1 -generate=types -o apis/stackgres/v1/sgdbops.gen.go apis/stackgres/v1/sgdbops.yaml
	perl -i -0pe 's/\*struct\s\{\n\s\sAdditionalProperties\smap\[string\]string\s`json:"-"`\n\s}/map\[string\]string/gms' apis/stackgres/v1/sgdbops.gen.go

	curl ${STACKGRES_CRD_URL}/SGCluster.yaml?inline=false -o apis/stackgres/v1/sgcluster_crd.yaml
	yq -i e apis/stackgres/v1/sgcluster.yaml --expression ".components.schemas.SGClusterSpec=load(\"apis/stackgres/v1/sgcluster_crd.yaml\").spec.versions[0].schema.openAPIV3Schema.properties.spec"
	yq -i e apis/stackgres/v1/sgcluster.yaml --expression ".components.schemas.SGClusterStatus=load(\"apis/stackgres/v1/sgcluster_crd.yaml\").spec.versions[0].schema.openAPIV3Schema.properties.status"
	go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --package=v1 -generate=types -o apis/stackgres/v1/sgcluster.gen.go apis/stackgres/v1/sgcluster.yaml
	perl -i -0pe 's/\*struct\s\{\n\s\sAdditionalProperties\smap\[string\]string\s`json:"-"`\n\s}/map\[string\]string/gms' apis/stackgres/v1/sgcluster.gen.go

	go run sigs.k8s.io/controller-tools/cmd/controller-gen object paths=./apis/stackgres/v1/...

