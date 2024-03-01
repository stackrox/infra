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

IMAGE=quay.io/rhacs-eng/infra-certifier:$(VERSION)
.PHONY: image-name
image-name:
	@echo $(IMAGE)

.PHONY: image
image:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o certifier -ldflags='-s -w' *.go
	docker build -t $(IMAGE) .

.PHONY: push
push:
	docker push $(IMAGE) | cat
