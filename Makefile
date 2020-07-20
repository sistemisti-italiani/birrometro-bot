# DO NOT MODIFY

# ################################
# Load configuration
SHELL := /bin/bash

GITORIGIN := $(shell git remote get-url origin | sed 's/.git$$//' | sed 's/git@github.com://' | sed 's$$https://.*git@github.com/$$$$')
PROJECT_NAME := $(shell echo ${GITORIGIN} | cut -d '/' -f 2 | tr '[:upper:]' '[:lower:]')
GROUP_NAME := sysadminita
# GROUP_NAME was $(shell echo ${GITORIGIN} | cut -d '/' -f 1 | tr '[:upper:]' '[:lower:]')
VERSION := $(shell git fetch --unshallow >/dev/null 2>&1 && git describe --tags 2>/dev/null)

DOCKER_IMAGE_PATH := ${GROUP_NAME}/${PROJECT_NAME}

ifeq (${VERSION},)
VERSION := no-version
endif

# ################################
# Targets

.PHONY: all
all: docker

.PHONY: info
info:
	$(info ** Environment info **)
	$(info Project name: ${PROJECT_NAME})
	$(info Group name: ${GROUP_NAME})
	$(info Version: ${VERSION})
	$(info Docker image path: ${DOCKER_IMAGE_PATH})
	@:

.PHONY: env
env:
	$(info PROJECT_NAME="${PROJECT_NAME}")
	$(info GROUP_NAME="${GROUP_NAME}")
	$(info VERSION="${VERSION}")
	@:

.PHONY: docker
docker:
	@docker build \
		-t ${DOCKER_IMAGE_PATH}:${VERSION} \
		--build-arg PROJECT_NAME=${PROJECT_NAME} \
		--build-arg GROUP_NAME=${GROUP_NAME} \
		--build-arg APP_VERSION="${shell git describe --tags --dirty}" \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

.PHONY: push
push:
	docker push ${DOCKER_IMAGE_PATH}:${VERSION}
	docker tag ${DOCKER_IMAGE_PATH}:${VERSION} ${DOCKER_IMAGE_PATH}:latest
	docker push ${DOCKER_IMAGE_PATH}:latest

.PHONY: clean
clean:
	docker rmi -f ${DOCKER_IMAGE_PATH}:latest ${DOCKER_IMAGE_PATH}:${VERSION}
	rm -rf bot

.PHONY: prune
prune:
	docker system prune -f

.PHONY: deps-reset
deps-reset:
	git checkout -- go.mod
	go mod tidy

.PHONY: deps-upgrade
deps-upgrade:
	go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)

.PHONY: deps-cleancache
deps-cleancache:
	go clean -modcache

.PHONY: deploy
deploy:
	kubectl -n ${GROUP_NAME} set image deploy/${PROJECT_NAME} ${PROJECT_NAME}=$(shell docker inspect --format='{{index .RepoDigests 0}}' ${DOCKER_IMAGE_PATH}:${VERSION})

.PHONY: deploy-skopeo
deploy-skopeo:
	kubectl -n ${GROUP_NAME} set image deploy/${PROJECT_NAME} ${PROJECT_NAME}=${DOCKER_IMAGE_PATH}@$(shell skopeo inspect docker://${DOCKER_IMAGE_PATH}:${VERSION} | jq -r ".Digest")

.PHONY: check
check:
	go test ./...
	go vet ./...
	gosec -quiet ./...
	staticcheck -tests=false ./...
	ineffassign .
	errcheck ./...

SRCCMD=$(wildcard cmd/bot/*.go)
SERVICECMD=$(wildcard service/**/*.go)

bot: $(SRCCMD) $(SERVICECMD)
	CGO_ENABLED=0 go build -mod=readonly -ldflags "-extldflags \"-static\"" -a -installsuffix cgo -o bot $(shell grep module go.mod | cut -f 2 -d ' ')/cmd/bot/
