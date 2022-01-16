ARCHLINUX  = archlinux:latest
IMG_BASE   = grokloc/grokloc-server:base
IMG_DEV    = grokloc/grokloc-server:dev
DOCKER     = docker
DOCKER_RUN = $(DOCKER) run --rm -it
GO         = go
UNIT_ENVS  = --env-file ./env/unit.env
PORTS      = -p 3000:3000
CWD        = $(shell pwd)
BASE       = /grokloc
RUN        = $(DOCKER_RUN) -v $(CWD):$(BASE) -w $(BASE) $(UNIT_ENVS) $(PORTS) $(IMG_DEV)

LINT       = golangci-lint --timeout=24h run pkg/... && staticcheck ./... && go vet ./...
TEST       = $(GO) test -v -race ./...

.DEFAULT_GOAL := build

.PHONY: build
build:
	$(GO) build ./...

.PHONY: docker
docker:
	$(DOCKER) pull $(ARCHLINUX)
	$(DOCKER) build . -f Dockerfile.base -t $(IMG_BASE)
	$(DOCKER) build . -f Dockerfile.dev -t $(IMG_DEV)

.PHONY: mod
mod:
	$(GO) mod tidy
	$(GO) mod download
	$(GO) mod vendor
	$(GO) build ./...

.PHONY: shell
shell:
	$(RUN) /bin/bash

.PHONY: local-check
local-check:
	$(LINT)

.PHONY: check
check:
	$(RUN) $(LINT)

.PHONY: local-test
local-test:
	$(TEST)

.PHONY: test
test:
	$(RUN) $(TEST)
