.PHONY: all
all: deps gazelle

.PHONY: tag
tag:
	@git describe --tags

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

.PHONY: gazelle
gazelle: deps
	bazel run //:gazelle

# server - Builds the infra server binary
# When run locally, a Darwin binary is built and installed into the user's GOPATH bin.
# When run in CI, a Darwin and Linux binary is built.
.PHONY: server
server: gazelle
ifdef CI
	bazel build $(BAZEL_FLAGS) --platforms=@io_bazel_rules_go//go/toolchain:darwin_amd64 -- //cmd/infra
	bazel build $(BAZEL_FLAGS) --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64  -- //cmd/infra
else
	bazel build $(BAZEL_FLAGS) --platforms=@io_bazel_rules_go//go/toolchain:darwin_amd64 -- //cmd/infra
	@install bazel-bin/cmd/infra/darwin_amd64_pure_stripped/infra $(GOPATH)/bin
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
