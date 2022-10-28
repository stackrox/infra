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

IMAGE=us.gcr.io/stackrox-infra/infra-server:$(VERSION)
.PHONY: image-name
image-name:
	@echo $(IMAGE)

###########
## Build ##
###########

# server - Builds the infra-server binary
# When run locally, a Darwin binary is built and installed into the user's GOPATH bin.
# When run in CI, a Darwin and Linux binary is built.
.PHONY: server
server: proto-generated-srcs
	@echo "+ $@"
	GOARCH=amd64 GOOS=linux ./scripts/go-build -o bin/infra-server-linux-amd64 ./cmd/infra-server

# cli - Builds the infractl client binary
# When run in CI or when preparing an image, a Darwin and Linux binary is built.
.PHONY: cli
cli: proto-generated-srcs
	@echo "+ $@"
	GOARCH=amd64 GOOS=darwin ./scripts/go-build -o bin/infractl-darwin-amd64 ./cmd/infractl
	GOARCH=arm64 GOOS=darwin ./scripts/go-build -o bin/infractl-darwin-arm64 ./cmd/infractl
	GOARCH=amd64 GOOS=linux  ./scripts/go-build -o bin/infractl-linux-amd64  ./cmd/infractl

# cli-local - Builds the infractl client binary
# When run locally, a Darwin binary is built and installed into the user's GOPATH bin.
.PHONY: cli-local
cli-local: proto-generated-srcs
	@echo "+ $@"
	./scripts/go-build -o $(GOPATH)/bin/infractl  ./cmd/infractl

.PHONY: ui
ui:
	@echo "+ $@"
	@make -C ui all

.PHONY: image
image: server cli ui clean-image
	@echo "+ $@"
	@cp -f bin/infra-server-linux-amd64 image/infra-server
	@mkdir -p image/static/downloads
	@cp -R ui/build/* image/static/
	@cp bin/infractl-darwin-amd64 image/static/downloads
	@cp bin/infractl-darwin-arm64 image/static/downloads
	@cp bin/infractl-linux-amd64 image/static/downloads
	docker build -t $(IMAGE) image

.PHONY: push
push:
	docker push $(IMAGE) | cat

.PHONY: clean-image
clean-image:
	@echo "+ $@"
	@rm -rf image/infra-server image/static

#############
## Testing ##
#############

.PHONY: unit-test
unit-test: proto-generated-srcs
	@echo "+ $@"
	@go test -v ./...

.PHONY: e2e-tests
e2e-tests:
	@bats -r .

##############
## Protobuf ##
##############
# Tool versions.
protoc-version = 3.11.2
protoc-gen-go-version = 1.3.2
protoc-gen-grpc-gateway-version = 1.12.1
protoc-gen-swagger-version = 1.12.1

# Tool binary paths
protoc = $(GOPATH)/bin/protoc
protoc-gen-go = $(GOPATH)/bin/protoc-gen-go
protoc-gen-grpc-gateway = $(GOPATH)/bin/protoc-gen-grpc-gateway
protoc-gen-swagger = $(GOPATH)/bin/protoc-gen-swagger

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
	@cd /tmp; go install github.com/golang/protobuf/protoc-gen-go@v$(protoc-gen-go-version)

# This target installs the protoc-gen-grpc-gateway binary.
$(protoc-gen-grpc-gateway):
	@echo "+ $@"
	@echo "Installing protoc-gen-grpc-gateway $(protoc-gen-grpc-gateway-version) to $(protoc-gen-grpc-gateway)"
	@cd /tmp; go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v$(protoc-gen-grpc-gateway-version)

# This target installs the protoc-gen-swagger binary.
$(protoc-gen-swagger):
	@echo "+ $@"
	@echo "Installing protoc-gen-swagger $(protoc-gen-swagger-version) to $(protoc-gen-swagger)"
	@cd /tmp; go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v$(protoc-gen-swagger-version)

# This target installs all of the protoc related binaries.
.PHONY: protoc-tools
protoc-tools: $(protoc) $(protoc-gen-go) $(protoc-gen-grpc-gateway) $(protoc-gen-swagger)

PROTO_INPUT_DIR   = proto/api/v1
PROTO_THIRD_PARTY = proto/third_party
PROTO_FILES       = service.proto
PROTO_OUTPUT_DIR  = generated/api/v1

# This target compiles proto files into:
# - Go gRPC bindings
# - Go gRPC-Gateway bindings
# - JSON Swagger definitions file
.PHONY: proto-generated-srcs
proto-generated-srcs: protoc-tools
	@echo "+ $@"
	@mkdir -p $(PROTO_OUTPUT_DIR)
	# Generate gRPC bindings
	$(protoc) -I$(PROTO_INPUT_DIR) \
		-I$(PROTO_THIRD_PARTY) \
		--go_out=plugins=grpc:$(PROTO_OUTPUT_DIR) \
		$(PROTO_FILES)

	# Generate gRPC-Gateway bindings
	$(protoc) -I$(PROTO_INPUT_DIR) \
		-I$(PROTO_THIRD_PARTY) \
		--grpc-gateway_out=logtostderr=true:$(PROTO_OUTPUT_DIR) \
		$(PROTO_FILES)

	# Generate JSON Swagger manifest
	$(protoc) -I$(PROTO_INPUT_DIR) \
		-I$(PROTO_THIRD_PARTY) \
		--swagger_out=logtostderr=true:$(PROTO_OUTPUT_DIR) \
		$(PROTO_FILES)

##########
## Kube ##
##########
## Meta
.PHONY: pre-check
pre-check:
ifndef DEPLOYMENT
	$(error DEPLOYMENT is undefined)
endif
ifndef ENVIRONMENT
	$(error ENVIRONMENT is undefined)
endif

.PHONY: setup-kc
setup-kc: pre-check
	$(info DEPLOYMENT: ${DEPLOYMENT}, ENVIRONMENT: ${ENVIRONMENT})
ifeq ($(DEPLOYMENT), local)
kc=kubectl
else ifeq ($(DEPLOYMENT), development)
kc=kubectl --context gke_stackrox-infra_us-west2_infra-development
else ifeq ($(DEPLOYMENT), production)
kc=kubectl --context gke_stackrox-infra_us-west2_infra-production
endif

## Configuration
.PHONY: configuration-download
configuration-download:
	@echo "Downloading configuration from gs://infra-configuration"
	gsutil -m cp -R "gs://infra-configuration/latest/configuration" "chart/infra-server/"

.PHONY: configuration-upload
configuration-upload: CONST_DATESTAMP := $(shell date '+%Y-%m-%d-%H-%M-%S')
configuration-upload:
	@echo "Uploading configuration to gs://infra-configuration/${CONST_DATESTAMP}"
	gsutil -m cp -R chart/infra-server/configuration "gs://infra-configuration/${CONST_DATESTAMP}/"
	@echo "Uploading configuration to gs://infra-configuration/latest/"
	gsutil -m cp -R chart/infra-server/configuration "gs://infra-configuration/latest/"

# Combines configuration/{development,production} files into single helm value.yaml files
# (configuration/{development,production}-values-from-files.yaml) that can be used in template
# rendering.
.PHONY: create-consolidated-values
create-consolidated-values:
	@./scripts/create-consolidated-values.sh

## Render
.PHONY: clean-render
clean-render:
	@rm -rf chart-rendered

.PHONY: render
render: pre-check clean-render create-consolidated-values
	@if [[ ! -e chart/infra-server/configuration ]]; then \
		echo chart/infra-server/configuration is absent. Try:; \
		echo make configuration-download; \
		exit 1; \
	fi
	@mkdir -p chart-rendered
	helm template chart/infra-server \
	    --output-dir chart-rendered \
		--set deployment="${DEPLOYMENT}" \
		--set tag="$(VERSION)" \
		--values chart/infra-server/configuration/${ENVIRONMENT}-values.yaml \
		--values chart/infra-server/configuration/${ENVIRONMENT}-values-from-files.yaml

.PHONY: render-local
render-local:
	DEPLOYMENT=local ENVIRONMENT=development make render

.PHONY: render-development
render-development:
	DEPLOYMENT=development ENVIRONMENT=development make render

.PHONY: render-production
render-production:
	DEPLOYMENT=production ENVIRONMENT=production make render

## Common install targets
bounce-infra-pods: setup-kc
	$(kc) -n infra rollout restart deploy/infra-server-deployment
	$(kc) -n infra rollout status deploy/infra-server-deployment --watch --timeout=3m

install-common: setup-kc
	@if ! $(kc) get ns argo 2> /dev/null; then \
		$(kc) create namespace argo; \
	fi
	$(kc) apply -n argo -f https://github.com/argoproj/argo-workflows/releases/download/v3.3.9/install.yaml;
	@if ! $(kc) get ns infra 2> /dev/null; then \
		$(kc) apply -f chart/infra-server/templates/namespace.yaml; \
	fi

## Install
.PHONY: install
install: setup-kc install-common
	$(kc) apply -R \
	    -f chart-rendered/infra-server

.PHONY: install-local
install-local:
	DEPLOYMENT=local ENVIRONMENT=development make render install

.PHONY: install-development
install-development:
	DEPLOYMENT=development ENVIRONMENT=development make render install

.PHONY: install-production
install-production:
	DEPLOYMENT=production ENVIRONMENT=production make render install


## Install without write
install-without-write: setup-kc install-common
	gsutil cat gs://infra-configuration/latest/configuration/$(ENVIRONMENT)-values.yaml \
               gs://infra-configuration/latest/configuration/$(ENVIRONMENT)-values-from-files.yaml | \
	helm template chart/infra-server \
		--set deployment="$(DEPLOYMENT)" \
		--set tag="$(VERSION)" \
		--values - | \
	$(kc) apply -R \
	    -f -
	@sleep 5
	make bounce-infra-pods

.PHONY: install-local-without-write
install-local-without-write:
	DEPLOYMENT=local ENVIRONMENT=development make install-common install-without-write

.PHONY: install-development-without-write
install-development-without-write:
	DEPLOYMENT=development ENVIRONMENT=development make install-common install-without-write

.PHONY: install-production-without-write
install-production-without-write:
	DEPLOYMENT=production ENVIRONMENT=production make install-common install-without-write

## Diff
.PHONY: diff
diff: setup-kc
	gsutil cat gs://infra-configuration/latest/configuration/$(ENVIRONMENT)-values.yaml \
               gs://infra-configuration/latest/configuration/$(ENVIRONMENT)-values-from-files.yaml | \
	helm template chart/infra-server \
		--set deployment="$(DEPLOYMENT)" \
		--set tag="$(VERSION)" \
		--values - | \
	$(kc) diff -R -f -

.PHONY: diff-local
diff-local:
	DEPLOYMENT=local ENVIRONMENT=development make diff

.PHONY: diff-development
diff-development:
	DEPLOYMENT=development ENVIRONMENT=development make diff

.PHONY: diff-production
diff-production:
	DEPLOYMENT=production ENVIRONMENT=production make diff

## Clean
.PHONY: clean-infra
clean-infra:
	$(kc) delete namespace infra || true

.PHONY: clean-argo
clean-argo:
	$(kc) delete namespace argo || true

.PHONY: clean-local
clean-local: DEPLOYMENT := local
clean-local: # setup-kc
	DEPLOYMENT=local ENVIRONMENT=development make setup-kc clean-infra clean-argo

.PHONY: clean-development
clean-development: setup-kc
	DEPLOYMENT=development ENVIRONMENT=development make setup-kc clean-infra

## Deploy
.PHONY: deploy-local
deploy-local: push install-local
	@echo "All done!"

.PHONY: deploy-development
deploy-development: push install-development
	@echo "All done!"

.PHONY: deploy-production
deploy-production: push install-production
	@echo "All done!"

##########
## Misc ##
##########
.PHONY: gotags
gotags:
	@gotags -R . > tags
	@echo "GoTags written to $(PWD)/tags"

.PHONY: argo-workflow-lint
argo-workflow-lint:
	@argo lint ./chart/infra-server/static/workflow*.yaml

.PHONY: update-version
update-version: image_regex   := gcr.io/stackrox-infra/automation-flavors/.*
update-version: image_version := 0.2.16
update-version:
	@echo 'Updating automation-flavor image versions to "${image_version}"'
	@perl -p -i -e 's#image: (${image_regex}):(.*)#image: \1:${image_version}#g' \
		./chart/infra-server/static/*.yaml
	@git diff --name-status ./chart/infra-server/static/*.yaml

# Assuming a local dev infra server is running and accessible via a port-forward
# i.e. nohup kubectl -n infra port-forward svc/infra-server-service 8443:8443 &
.PHONY: pull-infractl-from-dev-server
pull-infractl-from-dev-server:
	@mkdir -p bin
	@rm -f bin/infractl
	set -o pipefail; \
	curl --retry 3 --insecure -v --show-error --fail --location https://localhost:8443/v1/cli/linux/amd64/upgrade \
          | jq -r ".result.fileChunk" \
          | base64 -d \
          > bin/infractl
	chmod +x bin/infractl
	bin/infractl -k -e localhost:8443 version
