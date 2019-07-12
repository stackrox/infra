.PHONY: all
all: deps gazelle

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

.PHONY: gazelle
gazelle: deps
	bazel run //:gazelle

.PHONY: tag
tag:
	@git describe --tags
