#!/bin/bash

GO_BIN_DIR="$(go env GOPATH)/bin"

BUF_VERSION="1.16.0"
INSTALLED_BUF_VERSION=$(buf --version || echo "not installed")

if [ "$BUF_VERSION" != "${INSTALLED_BUF_VERSION}" ]; then
    mkdir -p "${GO_BIN_DIR}"
    curl -sSL \
        "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-$(uname -s)-$(uname -m)" \
        -o "${GO_BIN_DIR}/buf" && \
    chmod +x "${GO_BIN_DIR}/buf"
fi
