GOLANG_BASE = golang:bullseye
IMG_BASE    = grokloc/grokloc-server:base
IMG_DEV     = grokloc/grokloc-server:dev
DOCKER      = docker
DOCKER_RUN  = $(DOCKER) run --rm -it
GO          = go
UNIT_ENVS   = --env-file ./env/unit.env
PORTS       = -p 3000:3000
CWD         = $(shell pwd)
BASE        = /grokloc
RUN         = $(DOCKER_RUN) -v $(CWD):$(BASE) -w $(BASE) $(UNIT_ENVS) $(PORTS) $(IMG_DEV)
LINT        = golangci-lint --timeout=24h run pkg/... && staticcheck ./... && go vet ./...
TEST        = $(GO) test -v -race ./...

.DEFAULT_GOAL := build

.PHONY: build
build:
	$(GO) build ./...

.PHONY: golang-base
golang-base:
	$(DOCKER) pull $(GOLANG_BASE)

.PHONY: docker-base
docker-base:
	$(DOCKER) build . -f Dockerfile.base -t $(IMG_BASE)
	$(DOCKER) system prune -f
	$(DOCKER) system prune -f

.PHONY: docker-dev
docker-dev:
	$(DOCKER) tag $(IMG_BASE) $(IMG_DEV)

.PHONY: docker
docker: golang-base docker-base docker-dev

.PHONY: docker-push
docker-push:
	$(DOCKER) push $(IMG_BASE)
	$(DOCKER) push $(IMG_DEV)

.PHONY: docker-pull
docker-pull: golang-base
	$(DOCKER) pull $(IMG_BASE)
	$(DOCKER) pull $(IMG_DEV)
	$(DOCKER) system prune -f
	$(DOCKER) system prune -f

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

.PHONY: all
all: check test
