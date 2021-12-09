export GO111MODULE=on

.PHONY: all
all: image

TAG=$(shell git describe --tags)
.PHONY: tag
tag:
	@echo $(TAG)

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
	docker build -t us.gcr.io/stackrox-infra/infra-server:$(TAG) image

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
	@go test ./...

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
	@wget -q $(PROTOC_ZIP) -O /tmp/protoc.zip
	@unzip -o -q -d /tmp /tmp/protoc.zip bin/protoc
	@install /tmp/bin/protoc $(protoc)

# This target installs the protoc-gen-go binary.
$(protoc-gen-go):
	@echo "+ $@"
	@echo "Installing protoc-gen-go $(protoc-gen-go-version) to $(protoc-gen-go)"
	@cd /tmp; go get -u github.com/golang/protobuf/protoc-gen-go@v$(protoc-gen-go-version)

# This target installs the protoc-gen-grpc-gateway binary.
$(protoc-gen-grpc-gateway):
	@echo "+ $@"
	@echo "Installing protoc-gen-grpc-gateway $(protoc-gen-grpc-gateway-version) to $(protoc-gen-grpc-gateway)"
	@cd /tmp; go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v$(protoc-gen-grpc-gateway-version)

# This target installs the protoc-gen-swagger binary.
$(protoc-gen-swagger):
	@echo "+ $@"
	@echo "Installing protoc-gen-swagger $(protoc-gen-swagger-version) to $(protoc-gen-swagger)"
	@cd /tmp; go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v$(protoc-gen-swagger-version)

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

.PHONY: push
push: image
	docker push us.gcr.io/stackrox-infra/infra-server:$(TAG) | cat

.PHONY: clean-render
clean-render:
	@rm -rf chart-rendered

.PHONY: render-local
render-local: clean-render
	@mkdir -p chart-rendered
	helm template chart/infra-server \
	    --output-dir chart-rendered \
		--set deployment="local" \
		--set tag="$(TAG)" \
		--values chart/infra-server/configuration/development-values.yaml

.PHONY: render-development
render-development: clean-render
	@mkdir -p chart-rendered
	helm template chart/infra-server \
	    --output-dir chart-rendered \
		--set deployment="development" \
		--set tag="$(TAG)" \
		--values chart/infra-server/configuration/development-values.yaml

.PHONY: render-production
render-production: clean-render
	@mkdir -p chart-rendered
	helm template chart/infra-server \
	    --output-dir chart-rendered \
		--set deployment="production" \
		--set tag="$(TAG)" \
		--values chart/infra-server/configuration/production-values.yaml

dev_context = gke_stackrox-infra_us-west2_infra-development
prod_context = gke_stackrox-infra_us-west2_infra-production
this_context = $(shell kubectl config current-context)

.PHONY: install-local
install-local: render-local
	@if [[ "$(this_context)" == "$(dev_context)" ]]; then \
		echo Your kube context is set to development infra, should be a local cluster; \
		exit 1; \
	fi
	@if [[ "$(this_context)" == "$(prod_context)" ]]; then \
		echo Your kube context is set to production infra, should be a local cluster; \
		exit 1; \
	fi
	helm upgrade --install \
		--repo https://argoproj.github.io/argo-helm \
		--create-namespace \
		--namespace argo \
		argo argo
	@if ! kubectl get crd applications.app.k8s.io; then \
		kubectl apply \
			-f chart-rendered/infra-server/templates/application-crd.yaml; \
		sleep 10; \
	fi
	@if ! kubectl get ns infra; then \
		kubectl apply \
			-f chart-rendered/infra-server/templates/namespace.yaml; \
		sleep 10; \
	fi
	kubectl apply -R \
	    -f chart-rendered/infra-server

.PHONY: install-development
install-development: render-development
	helm upgrade --install \
	    --kube-context $(dev_context) \
		--repo https://argoproj.github.io/argo-helm \
		--create-namespace \
		--namespace argo \
		argo argo
	kubectl apply -R \
	    --context $(dev_context) \
	    -f chart-rendered/infra-server

.PHONY: install-production
install-production: render-production
	helm upgrade --install \
	    --kube-context $(prod_context) \
		--repo https://argoproj.github.io/argo-helm \
		--create-namespace \
		--namespace argo \
		argo argo
	kubectl apply -R \
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
