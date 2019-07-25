.PHONY: all
all: deps gazelle

TAG=$(shell git describe --tags)
.PHONY: tag
tag:
	@echo $(TAG)

deps: Gopkg.toml Gopkg.lock
ifdef CI
	@# `dep check` exits with a nonzero code if there is a toml->lock mismatch.
	dep check -skip-vendor
endif
	dep ensure
	@touch deps

.PHONY: clean-deps
clean-deps:
	@rm -f deps

###########
## Build ##
###########
BAZEL_FLAGS := --cpu=k8 --features=pure --features=race --workspace_status_command=scripts/bazel-workspace-status.sh

cleanup:
	@git status --ignored --untracked-files=all --porcelain | grep '^\(!!\|??\) ' | cut -d' ' -f 2- | grep '\(/\|^\)BUILD\.bazel$$' | xargs rm

.PHONY: gazelle
gazelle: proto-generated-srcs deps cleanup
	bazel run //:gazelle

# server - Builds the infra-server binary
# When run locally, a Darwin binary is built and installed into the user's GOPATH bin.
# When run in CI, a Darwin and Linux binary is built.
.PHONY: server
server: gazelle
ifdef CI
	bazel build $(BAZEL_FLAGS) --platforms=@io_bazel_rules_go//go/toolchain:darwin_amd64 -- //cmd/infra-server
	bazel build $(BAZEL_FLAGS) --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64  -- //cmd/infra-server
else
	bazel build $(BAZEL_FLAGS) --platforms=@io_bazel_rules_go//go/toolchain:darwin_amd64 -- //cmd/infra-server
	@install bazel-bin/cmd/infra-server/darwin_amd64_pure_stripped/infra-server $(GOPATH)/bin
endif

# cli - Builds the infractl client binary
# When run locally, a Darwin binary is built and installed into the user's GOPATH bin.
# When run in CI, a Darwin and Linux binary is built.
.PHONY: cli
cli: gazelle
ifdef CI
	bazel build $(BAZEL_FLAGS) --platforms=@io_bazel_rules_go//go/toolchain:darwin_amd64 -- //cmd/infractl
	bazel build $(BAZEL_FLAGS) --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64  -- //cmd/infractl
else
	bazel build $(BAZEL_FLAGS) --platforms=@io_bazel_rules_go//go/toolchain:darwin_amd64 -- //cmd/infractl
	@install bazel-bin/cmd/infractl/darwin_amd64_pure_stripped/infractl $(GOPATH)/bin
endif

.PHONY: image
image: gazelle
	bazel build $(BAZEL_FLAGS) --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64  -- //cmd/infra-server
	@cp -f bazel-bin/cmd/infra-server/linux_amd64_pure_stripped/infra-server image/infra-server
	docker build -t us.gcr.io/ultra-current-825/infra-server:$(TAG) image

##############
## Protobuf ##
##############
# The protoc zip url changes depending on if we're running in CI or not.
ifeq ($(shell uname -s),Linux)
PROTOC_ZIP = https://github.com/protocolbuffers/protobuf/releases/download/v3.9.0/protoc-3.9.0-linux-x86_64.zip
endif
ifeq ($(shell uname -s),Darwin)
PROTOC_ZIP = https://github.com/protocolbuffers/protobuf/releases/download/v3.9.0/protoc-3.9.0-osx-x86_64.zip
endif

# This target installs the protoc binary.
$(GOPATH)/bin/protoc:
	@echo "Installing protoc 3.9.0 to $(GOPATH)/bin/protoc"
	@wget -q $(PROTOC_ZIP) -O /tmp/protoc.zip
	@unzip -o -q -d /tmp /tmp/protoc.zip bin/protoc
	@install /tmp/bin/protoc $(GOPATH)/bin/protoc

# This target installs the protoc-gen-go binary.
$(GOPATH)/bin/protoc-gen-go:
	@echo "Installing protoc-gen-go to $(GOPATH)/bin/protoc-gen-go"
	@go get github.com/golang/protobuf/protoc-gen-go

# This target installs the protoc-gen-grpc-gateway binary.
$(GOPATH)/bin/protoc-gen-grpc-gateway:
	@echo "Installing protoc-gen-grpc-gateway to $(GOPATH)/bin/protoc-gen-grpc-gateway"
	@go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway

# This target installs the protoc-gen-swagger binary.
$(GOPATH)/bin/protoc-gen-swagger:
	@echo "Installing protoc-gen-swagger to $(GOPATH)/bin/protoc-gen-swagger"
	@go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger

# This target installs all of the protoc related binaries.
.PHONY: protoc-tools
protoc-tools: $(GOPATH)/bin/protoc $(GOPATH)/bin/protoc-gen-go $(GOPATH)/bin/protoc-gen-grpc-gateway $(GOPATH)/bin/protoc-gen-swagger

PROTO_INPUT_DIR   = proto/api/v1
PROTO_FILES       = service.proto
PROTO_OUTPUT_DIR  = generated/api/v1

# This target compiles proto files into:
# - Go gRPC bindings
# - Go gRPC-Gateway bindings
# - JSON Swagger definitions file
.PHONY: proto-generated-srcs
proto-generated-srcs: protoc-tools
	go get github.com/gogo/protobuf/protobuf || true
	go get github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis || true
	go get github.com/protocolbuffers/protobuf || true
	@mkdir -p $(PROTO_OUTPUT_DIR)
	# Generate gRPC bindings
	protoc -I$(PROTO_INPUT_DIR) \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		-I${GOPATH}/src/github.com/protocolbuffers/protobuf/src \
		-I${GOPATH}/src/github.com/gogo/protobuf/protobuf \
		--go_out=plugins=grpc:$(PROTO_OUTPUT_DIR) \
		$(PROTO_FILES)

	# Generate gRPC-Gateway bindings
	protoc -I$(PROTO_INPUT_DIR) \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		-I${GOPATH}/src/github.com/protocolbuffers/protobuf/src \
		-I${GOPATH}/src/github.com/gogo/protobuf/protobuf \
		--grpc-gateway_out=logtostderr=true:$(PROTO_OUTPUT_DIR) \
		$(PROTO_FILES)

	# Generate JSON Swagger manifest
	protoc -I$(PROTO_INPUT_DIR) \
		-I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		-I${GOPATH}/src/github.com/protocolbuffers/protobuf/src \
		-I${GOPATH}/src/github.com/gogo/protobuf/protobuf \
		--swagger_out=logtostderr=true:$(PROTO_OUTPUT_DIR) \
		$(PROTO_FILES)
