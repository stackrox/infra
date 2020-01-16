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
	GOARCH=amd64 GOOS=linux ./scripts/go-build -o bin/infra-server-linux-amd64 ./cmd/infra-server

# cli - Builds the infractl client binary
# When run locally, a Darwin binary is built and installed into the user's GOPATH bin.
# When run in CI, a Darwin and Linux binary is built.
.PHONY: cli
cli: proto-generated-srcs
ifdef CI
	GOARCH=amd64 GOOS=darwin ./scripts/go-build -o bin/infractl-darwin-amd64 ./cmd/infractl
	GOARCH=amd64 GOOS=linux  ./scripts/go-build -o bin/infractl-linux-amd64  ./cmd/infractl
else
	./scripts/go-build -o $(GOPATH)/bin/infractl  ./cmd/infractl
endif

.PHONY: image
image: server
	@cp -f bin/infra-server-linux-amd64 image/infra-server
	docker build -t us.gcr.io/ultra-current-825/infra-server:$(TAG) image

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
	@echo "Installing protoc $(protoc-version) to $(protoc)"
	@wget -q $(PROTOC_ZIP) -O /tmp/protoc.zip
	@unzip -o -q -d /tmp /tmp/protoc.zip bin/protoc
	@install /tmp/bin/protoc $(protoc)

# This target installs the protoc-gen-go binary.
$(protoc-gen-go):
	@echo "Installing protoc-gen-go $(protoc-gen-go-version) to $(protoc-gen-go)"
	@cd /tmp; go get -u github.com/golang/protobuf/protoc-gen-go@v$(protoc-gen-go-version)

# This target installs the protoc-gen-grpc-gateway binary.
$(protoc-gen-grpc-gateway):
	@echo "Installing protoc-gen-grpc-gateway $(protoc-gen-grpc-gateway-version) to $(protoc-gen-grpc-gateway)"
	@cd /tmp; go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v$(protoc-gen-grpc-gateway-version)

# This target installs the protoc-gen-swagger binary.
$(protoc-gen-swagger):
	@echo "Installing protoc-gen-swagger $(protoc-gen-swagger-version) to $(protoc-gen-swagger)"
	@cd /tmp; go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v$(protoc-gen-swagger-version)

# This target installs all of the protoc related binaries.
.PHONY: protoc-tools
protoc-tools: $(protoc) $(protoc-gen-go) $(protoc-gen-grpc-gateway) $(protoc-gen-swagger)

PROTO_INPUT_DIR   = proto/api/v1
PROTO_FILES       = service.proto
PROTO_OUTPUT_DIR  = generated/api/v1

# This target compiles proto files into:
# - Go gRPC bindings
# - Go gRPC-Gateway bindings
# - JSON Swagger definitions file
.PHONY: proto-generated-srcs
proto-generated-srcs: protoc-tools
	GO111MODULE=off go get github.com/gogo/protobuf/protobuf || true
	GO111MODULE=off go get github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis || true
	GO111MODULE=off go get github.com/protocolbuffers/protobuf || true
	@mkdir -p $(PROTO_OUTPUT_DIR)
	# Generate gRPC bindings
	$(protoc) -I$(PROTO_INPUT_DIR) \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		-I${GOPATH}/src/github.com/protocolbuffers/protobuf/src \
		-I${GOPATH}/src/github.com/gogo/protobuf/protobuf \
		--go_out=plugins=grpc:$(PROTO_OUTPUT_DIR) \
		$(PROTO_FILES)

	# Generate gRPC-Gateway bindings
	$(protoc) -I$(PROTO_INPUT_DIR) \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		-I${GOPATH}/src/github.com/protocolbuffers/protobuf/src \
		-I${GOPATH}/src/github.com/gogo/protobuf/protobuf \
		--grpc-gateway_out=logtostderr=true:$(PROTO_OUTPUT_DIR) \
		$(PROTO_FILES)

	# Generate JSON Swagger manifest
	$(protoc) -I$(PROTO_INPUT_DIR) \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		-I${GOPATH}/src/github.com/protocolbuffers/protobuf/src \
		-I${GOPATH}/src/github.com/gogo/protobuf/protobuf \
		--swagger_out=logtostderr=true:$(PROTO_OUTPUT_DIR) \
		$(PROTO_FILES)

##########
## Kube ##
##########
.PHONY: push
push: image
	docker push us.gcr.io/ultra-current-825/infra-server:$(TAG) | cat

.PHONY: render
render:
	@mkdir -p chart-rendered
	helm template chart/infra-server --output-dir chart-rendered \
		--name infra-server --namespace infra \
		--set tag=$(TAG) \
		--set host=test1.demo.stackrox.com \
		--set ip=34.94.91.159

.PHONY: sanity
sanity:
	@test -d chart/infra-server/configs || { echo "Deployment configs missing"; exit 1; }

.PHONY: deploy
deploy: sanity render push
	kubectl apply -R -f chart-rendered
