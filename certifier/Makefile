SHELL := /usr/bin/env bash
export GO111MODULE=on

.PHONY: all
all: image

TAG=$(shell git describe --tags --abbrev=10 --long)
TAGGED=$(shell git tag --contains | head)
ifneq (,$(TAGGED))
	# We're tagged. Use the tag explicitly.
	VERSION := $(TAGGED)
else
	# We're on a dev/PR branch
	VERSION := $(TAG)
endif

.PHONY: tag
tag:
	@echo $(VERSION)

IMAGE=us.gcr.io/stackrox-infra/certifier:$(VERSION)
.PHONY: image-name
image-name:
	@echo $(IMAGE)

.PHONY: image
image:
	GOOS=linux GOARCH=amd64 go build -o certifier -ldflags='-s -w' *.go
	docker build . -t $(IMAGE)

.PHONY: push
push:
	docker push $(IMAGE) | cat
