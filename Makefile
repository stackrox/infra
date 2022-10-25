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

IMAGE=us.gcr.io/stackrox-infra/infra-server:$(TAG)
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
image: server cli ui clean-image
	@echo "+ $@"
	@cp -f bin/infra-server-linux-amd64 image/infra-server
	@mkdir -p image/static/downloads
	@cp -R ui/build/* image/static/
	@cp bin/infractl-darwin-amd64 image/static/downloads
	@cp bin/infractl-darwin-arm64 image/static/downloads
	@cp bin/infractl-linux-amd64 image/static/downloads
	docker build -t $(IMAGE) image

.PHONY: clean-image
clean-image:
	@echo "+ $@"
	@rm -rf image/infra-server image/static

#############
## Testing ##
#############

.PHONY: unit-test
unit-test:
	@echo "+ $@"
	@go test -v ./...

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

.PHONY: push
push:
	docker push $(IMAGE) | cat

.PHONY: clean-render
clean-render:
	@rm -rf chart-rendered

.PHONY: render-local
render-local: clean-render create-consolidated-values
	@if [[ ! -e chart/infra-server/configuration ]]; then \
		echo chart/infra-server/configuration is absent. Try:; \
		echo make configuration-download; \
		exit 1; \
	fi
	@mkdir -p chart-rendered
	helm template chart/infra-server \
	    --output-dir chart-rendered \
		--set deployment="local" \
		--set tag="$(TAG)" \
		--values chart/infra-server/configuration/development-values.yaml \
		--values chart/infra-server/configuration/development-values-from-files.yaml

.PHONY: render-development
render-development: clean-render create-consolidated-values
	@mkdir -p chart-rendered
	helm template chart/infra-server \
	    --output-dir chart-rendered \
		--set deployment="development" \
		--set tag="$(TAG)" \
		--values chart/infra-server/configuration/development-values.yaml \
		--values chart/infra-server/configuration/development-values-from-files.yaml

.PHONY: render-production
render-production: clean-render create-consolidated-values
	@mkdir -p chart-rendered
	helm template chart/infra-server \
	    --output-dir chart-rendered \
		--set deployment="production" \
		--set tag="$(TAG)" \
		--values chart/infra-server/configuration/production-values.yaml \
		--values chart/infra-server/configuration/production-values-from-files.yaml

dev_context = gke_stackrox-infra_us-west2_infra-development
prod_context = gke_stackrox-infra_us-west2_infra-production
this_context = $(shell kubectl config current-context)
kcdev = kubectl --context $(dev_context)
kcprod = kubectl --context $(prod_context)

.PHONY: install-local-common
install-local-common:
	@if [[ "$(this_context)" == "$(dev_context)" ]]; then \
		echo Your kube context is set to development infra, should be a local cluster; \
		exit 1; \
	fi
	@if [[ "$(this_context)" == "$(prod_context)" ]]; then \
		echo Your kube context is set to production infra, should be a local cluster; \
		exit 1; \
	fi
	@if ! kubectl get ns argo 2> /dev/null; then \
		kubectl create namespace argo; \
		kubectl apply -n argo -f https://github.com/argoproj/argo-workflows/releases/download/v3.3.9/install.yaml; \
	fi
	@if ! kubectl get ns infra 2> /dev/null; then \
		kubectl apply \
			-f chart/infra-server/templates/namespace.yaml; \
		sleep 10; \
	fi
	kubectl apply -f workflows/*

.PHONY: install-local
install-local: install-local-common
	kubectl apply -R \
	    -f chart-rendered/infra-server

.PHONY: install-local-without-write
install-local-without-write: install-local-common
	gsutil cat gs://infra-configuration/latest/configuration/development-values.yaml \
               gs://infra-configuration/latest/configuration/development-values-from-files.yaml | \
	helm template chart/infra-server \
		--set deployment="local" \
		--set tag="$(TAG)" \
		--values - | \
	kubectl apply -R \
	    -f -
	# Bounce the infra-server to ensure proper update
	@sleep 5
	kubectl -n infra delete pods -l app=infra-server --wait
	@sleep 5

.PHONY: local-data-dev-cycle
local-data-dev-cycle: render-local install-local
	# Bounce the infra-server to ensure proper update
	@sleep 5
	kubectl -n infra delete pods -l app=infra-server --wait
	@sleep 5

.PHONY: diff-development
diff-development: render-development
	$(kcdev) diff -R \
		-f chart-rendered/infra-server

.PHONY: install-development
install-development: render-development
	@if ! $(kcdev) get ns argo; then \
		$(kcdev) create namespace argo; \
		$(kcdev) apply -n argo -f https://github.com/argoproj/argo-workflows/releases/download/v3.3.9/install.yaml; \
	fi
	@if ! $(kcdev) get ns infra; then \
		$(kcdev) apply \
			-f chart-rendered/infra-server/templates/namespace.yaml; \
		sleep 10; \
	fi
	$(kcdev) apply -f workflows/*
	$(kcdev) apply -R \
	    -f chart-rendered/infra-server

.PHONY: diff-production
diff-production: render-production
	$(kcprod) diff -R \
		--context $(prod_context) \
		-f chart-rendered/infra-server

.PHONY: install-production
install-production: render-production
	@if ! $(kcprod) get ns argo; then \
		$(kcprod) create namespace argo; \
		$(kcprod) apply -n argo -f https://github.com/argoproj/argo-workflows/releases/download/v3.3.9/install.yaml; \
	fi
	@if ! $(kcprod) get ns infra; then \
		$(kcprod) apply \
			-f chart-rendered/infra-server/templates/namespace.yaml; \
		sleep 10; \
	fi
	$(kcprod) apply -f workflows/*
	$(kcprod) apply -R \
	    --context $(prod_context) \
	    -f chart-rendered/infra-server

.PHONY: deploy-local
deploy-local: push install-local
	@echo "All done!"

.PHONY: clean-local
clean-local:
	kubectl delete namespace infra || true
	kubectl delete namespace argo || true

.PHONY: deploy-development
deploy-development: push install-development
	@echo "All done!"

.PHONY: clean-development
clean-development:
	kubectl delete namespace infra || true

.PHONY: deploy-production
deploy-production: push install-production
	@echo "All done!"

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
	curl --retry 3 --insecure --silent --show-error --fail --location https://localhost:8443/v1/cli/linux/amd64/upgrade \
          | jq -r ".result.fileChunk" \
          | base64 -d \
          > bin/infractl
	chmod +x bin/infractl
	bin/infractl -k -e localhost:8443 version

.PHONY: e2e-tests
e2e-tests:
	@bats -r .
