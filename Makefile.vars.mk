## These are some common variables for Make

PROJECT_ROOT_DIR = .
PROJECT_NAME ?= appcat-comp-functions
PROJECT_OWNER ?= vshn

WORK_DIR = $(PWD)/.work

# For alpine image it is required the following env before building the application
DOCKER_IMAGE_GOOS = linux
DOCKER_IMAGE_GOARCH = amd64

## BUILD:go
BIN_FILENAME ?= $(PROJECT_NAME)
go_bin ?= $(WORK_DIR)/bin
$(go_bin):
	@mkdir -p $@

## BUILD:docker
DOCKER_CMD ?= docker

## BUILD:docker VSHN Postgres
IMG_TAG ?= latest
# Image URL to use all building/pushing image targets
CONTAINER_IMG ?= ghcr.io/$(PROJECT_OWNER)/$(PROJECT_NAME):$(IMG_TAG)


## KIND:setup

# https://hub.docker.com/r/kindest/node/tags
KIND_NODE_VERSION ?= v1.24.0
KIND_IMAGE ?= docker.io/kindest/node:$(KIND_NODE_VERSION)
KIND ?= go run sigs.k8s.io/kind
KIND_KUBECONFIG ?= $(kind_dir)/kind-kubeconfig
KIND_CLUSTER ?= $(PROJECT_NAME)



## Stackgres CRDs
STACKGRES_VERSION ?= 1.4.3
STACKGRES_CRD_URL ?= https://gitlab.com/ongresinc/stackgres/-/raw/${STACKGRES_VERSION}/stackgres-k8s/src/common/src/main/resources/crds
