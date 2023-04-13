
antora_preview_version ?= 3.0.1.1
antora_preview_cmd ?= $(DOCKER_CMD) run --rm --publish 2020:2020 --volume "${PWD}":/preview/antora ghcr.io/vshn/antora-preview:3.1.2.3 --style=vshn

.PHONY: docs-preview
docs-preview: ## Preview the documentation
	$(antora_preview_cmd)
