SHELL := /usr/bin/env bash
export GO111MODULE=on

.PHONY: all
all: image

TAG=$(shell git describe --tags --abbrev=10 --long)
TAGGED=$(shell git tag --sort=-creatordate --contains | head --lines=1)
ifneq (,$(TAGGED))
	# We're tagged. Use the tag explicitly.
	VERSION := $(TAGGED)
else
	# We're on a dev/PR branch
	VERSION := $(TAG)
endif

LOCAL_VALUES_FILE=chart/infra-server/configuration/infra-values-${ENVIRONMENT}.yaml
LOCAL_COMBINED_VALUES_FILE=chart/infra-server/configuration/infra-values-from-files-${ENVIRONMENT}.yaml

ifeq '$(SECRET_VERSION)' ''
SECRET_VERSION := latest
endif

.PHONY: tag
tag:
	@echo $(VERSION)

IMAGE=quay.io/rhacs-eng/infra-server:$(VERSION)
.PHONY: image-name
image-name:
	@echo $(IMAGE)

##############
## Protobuf ##
##############
# Tool versions.
protoc-version = 29.2
protoc-gen-go-version = 1.36.1
protoc-gen-grpc-gateway-version = 2.25.1
protoc-gen-openapiv2-version = 2.25.1

# Tool binary paths
protoc = $(GOPATH)/bin/protoc
protoc-gen-go = $(GOPATH)/bin/protoc-gen-go
protoc-gen-grpc-gateway = $(GOPATH)/bin/protoc-gen-grpc-gateway
protoc-gen-openapiv2 = $(GOPATH)/bin/protoc-gen-openapiv2

# The protoc zip url changes depending on if we're running in CI or not.
ifeq ($(shell uname -s),Linux)
PROTOC_ZIP = https://github.com/protocolbuffers/protobuf/releases/download/v$(protoc-version)/protoc-$(protoc-version)-linux-x86_64.zip
endif
ifeq ($(shell uname -s),Darwin)
PROTOC_ZIP = https://github.com/protocolbuffers/protobuf/releases/download/v$(protoc-version)/protoc-$(protoc-version)-osx-x86_64.zip
endif

# This target installs the protoc binary.
$(protoc):
	@echo "+ $@"
	@echo "Installing protoc $(protoc-version) to $(protoc)"
	@mkdir -p $(GOPATH)/bin
	@wget -q $(PROTOC_ZIP) -O /tmp/protoc.zip
	@unzip -o -q -d /tmp /tmp/protoc.zip bin/protoc
	@install /tmp/bin/protoc $(protoc)

# This target installs the protoc-gen-go binary.
$(protoc-gen-go):
	@echo "+ $@"
	@echo "Installing protoc-gen-go $(protoc-gen-go-version) to $(protoc-gen-go)"
	@cd /tmp; go install google.golang.org/protobuf/cmd/protoc-gen-go@v$(protoc-gen-go-version)

# This target installs the protoc-gen-grpc-gateway binary.
$(protoc-gen-grpc-gateway):
	@echo "+ $@"
	@echo "Installing protoc-gen-grpc-gateway $(protoc-gen-grpc-gateway-version) to $(protoc-gen-grpc-gateway)"
	@cd /tmp; go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v$(protoc-gen-grpc-gateway-version)

# This target installs the protoc-gen-openapiv2 binary.
$(protoc-gen-openapiv2):
	@echo "+ $@"
	@echo "Installing protoc-gen-openapiv2 $(protoc-gen-openapiv2-version) to $(protoc-gen-openapiv2)"
	@cd /tmp; go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v$(protoc-gen-openapiv2-version)

# This target installs all of the protoc related binaries.
.PHONY: protoc-tools
protoc-tools: $(protoc) $(protoc-gen-go) $(protoc-gen-grpc-gateway) $(protoc-gen-openapiv2)

PROTO_INPUT_DIR   = proto/api/v1
PROTO_THIRD_PARTY = proto/third_party
PROTO_FILES       = service.proto
PROTO_OUTPUT_DIR  = generated

# This target compiles proto files into:
# - Go gRPC bindings
# - Go gRPC-Gateway bindings
# - JSON OpenAPI v2 (Swagger) definitions file
.PHONY: proto-generated-srcs
proto-generated-srcs: protoc-tools
	@echo "+ $@"
	@mkdir -p $(PROTO_OUTPUT_DIR)
	# Generate gRPC bindings
	$(protoc) -I$(PROTO_INPUT_DIR) \
		-I$(PROTO_THIRD_PARTY) \
		--go_out=$(PROTO_OUTPUT_DIR) \
		--go-grpc_out=$(PROTO_OUTPUT_DIR) \
		$(PROTO_FILES)

	# Generate gRPC-Gateway bindings
	$(protoc) -I$(PROTO_INPUT_DIR) \
		-I$(PROTO_THIRD_PARTY) \
		--grpc-gateway_out=logtostderr=true:$(PROTO_OUTPUT_DIR) \
		$(PROTO_FILES)

	# Generate JSON OpenAPI manifest
	$(protoc) -I$(PROTO_INPUT_DIR) \
		-I$(PROTO_THIRD_PARTY) \
		--openapiv2_out=logtostderr=true:$(PROTO_OUTPUT_DIR)/api/v1 \
		$(PROTO_FILES)

###########
## Build ##
###########

# server - Builds the infra-server binary
# When run locally, a Darwin binary is built and installed into the user's GOPATH bin.
# When run in CI, a Darwin and Linux binary is built.
.PHONY: server
server:
	@echo "+ $@"
	GOARCH=amd64 GOOS=linux ./scripts/go-build -o bin/infra-server-linux-amd64 ./cmd/infra-server

# cli - Builds the infractl client binary
# When run in CI or when preparing an image, a Darwin and Linux binary is built.
.PHONY: cli
cli:
	@echo "+ $@"
	GOARCH=amd64 GOOS=darwin ./scripts/go-build -o bin/infractl-darwin-amd64 ./cmd/infractl
	GOARCH=arm64 GOOS=darwin ./scripts/go-build -o bin/infractl-darwin-arm64 ./cmd/infractl
	GOARCH=amd64 GOOS=linux  ./scripts/go-build -o bin/infractl-linux-amd64  ./cmd/infractl

# cli-local - Builds the infractl client binary
# When run locally, a Darwin binary is built and installed into the user's GOPATH bin.
.PHONY: cli-local
cli-local:
	@echo "+ $@"
	./scripts/go-build -o $(GOPATH)/bin/infractl  ./cmd/infractl

.PHONY: ui
ui:
	@echo "+ $@"
	@make -C ui all

.PHONY: image
image:
	docker build . -t $(IMAGE) -f image/Dockerfile --secret id=npmrc,src=${HOME}/.npmrc

.PHONY: push
push:
	docker push $(IMAGE) | cat

#############
## Linting ##
#############

.PHONY: argo-workflow-lint
argo-workflow-lint:
	@argo lint ./chart/infra-server/static/workflow*.yaml

.PHONY: shellcheck
shellcheck:
	@shellcheck -x -- **/*.{bats,sh}

#############
## Testing ##
#############

.PHONY: unit-test
unit-test: proto-generated-srcs
	@echo "+ $@"
	@go test -v ./...

.PHONY: bats-e2e-tests
bats-e2e-tests:
	@kubectl apply -f "workflows/*.yaml"
	@bats --recursive .

.PHONY: go-e2e-tests
go-e2e-tests: proto-generated-srcs
	@kubectl apply -f workflows/
	@go test ./test/e2e/... -tags=e2e -v -parallel 5 -count 1 -cover -timeout 1h

# Assuming a local dev infra server is running and accessible via a port-forward
# i.e. nohup kubectl -n infra port-forward svc/infra-server-service 8443:8443 &
.PHONY: pull-infractl-from-dev-server
pull-infractl-from-dev-server:
	@mkdir -p bin
	@rm -f bin/infractl
	set -o pipefail; \
	curl --retry 3 --insecure --silent --show-error --fail --location https://localhost:8443/v1/cli/$(shell go env GOOS)/$(shell go env GOARCH)/upgrade \
          | jq -r ".result.fileChunk" \
          | base64 -d \
          > bin/infractl
	chmod +x bin/infractl
	bin/infractl -k -e localhost:8443 version

##########
## Kube ##
##########
dev_context = gke_acs-team-automation_us-west2_infra-development
prod_context = gke_acs-team-automation_us-west2_infra-production
this_context = $(shell kubectl config current-context)

## Meta
.PHONY: pre-check
pre-check:
ifndef ENVIRONMENT
	$(error ENVIRONMENT is undefined)
endif
	@if [[ "${ENVIRONMENT}" == "development" && "${this_context}" == "${prod_context}" ]]; then \
		echo -e "Your kube context is not set to a development infra. Use the following for dev cluster or set it to your PR cluster\n\tkubectl config use-context ${dev_context}\n"; \
		exit 1; \
	fi
	@if [[ "${ENVIRONMENT}" == "production" && "${this_context}" != "${prod_context}" ]]; then \
		echo -e "Your kube context is not set to production infra:\n\tkubectl config use-context ${prod_context}"; \
		exit 1; \
	fi

## Render template
.PHONY: helm-template
helm-template: pre-check
	@./scripts/deploy/helm.sh template $(VERSION) $(ENVIRONMENT) $(SECRET_VERSION)

## Deploy
.PHONY: helm-deploy
helm-deploy: pre-check
	@./scripts/deploy/helm.sh deploy $(VERSION) $(ENVIRONMENT) $(SECRET_VERSION)
	# Pick up any eventual changes to the workflow controller configmap
	@make bounce-argo-pods

## Diff
.PHONY: helm-diff
helm-diff: pre-check
	@./scripts/deploy/helm.sh diff $(VERSION) $(ENVIRONMENT) $(SECRET_VERSION)

## Bounce pods
.PHONY: bounce-infra-pods
bounce-infra-pods:
	kubectl -n infra rollout restart deploy/infra-server-deployment
	kubectl -n infra rollout status deploy/infra-server-deployment --watch --timeout=3m

.PHONY: bounce-argo-pods
bounce-argo-pods:
	kubectl rollout restart deploy/argo-workflows-workflow-controller -n argo
	kubectl rollout status deploy/argo-workflows-workflow-controller -n argo --watch --timeout=3m
	kubectl rollout restart deploy/argo-workflows-server -n argo
	kubectl rollout status deploy/argo-workflows-server -n argo --watch --timeout=3m

#############
## Secrets ##
#############
.PHONY: secrets-download
secrets-download:
	@./scripts/deploy/secrets.sh download_secrets $(ENVIRONMENT)

.PHONY: secrets-upload
secrets-upload:
	@./scripts/deploy/secrets.sh upload_secrets $(ENVIRONMENT) $(SECRET_VERSION)

.PHONY: secrets-show
secrets-show:
	@./scripts/deploy/secrets.sh show $(ENVIRONMENT) $(SECRET_VERSION)

.PHONY: secrets-edit
secrets-edit:
	@./scripts/deploy/secrets.sh edit $(ENVIRONMENT) $(SECRET_VERSION)

.PHONY: secrets-revert
secrets-revert:
	@./scripts/deploy/secrets.sh revert $(ENVIRONMENT) $(SECRET_VERSION)

##################
## Dependencies ##
##################
.PHONY: install-argo
install-argo: pre-check
	helm repo add argo https://argoproj.github.io/argo-helm
	helm upgrade \
		argo-workflows \
		argo/argo-workflows \
		--version 0.16.9 \
		--install \
		--create-namespace \
		--namespace argo

.PHONY: clean-argo-config
clean-argo-config: pre-check
	kubectl delete configmap argo-workflows-workflow-controller-configmap -n argo || true

.PHONY: install-monitoring
install-monitoring: pre-check
	helm dependency update chart/infra-monitoring
	helm upgrade prometheus-stack chart/infra-monitoring \
		--install \
		--namespace monitoring \
		--create-namespace \
		--values chart/infra-monitoring/values.yaml \
		--wait

###############
## Debugging ##
###############
.PHONY: prepare-local-server-debugging
prepare-local-server-debugging:
	@./scripts/local-dev/prepare.sh
