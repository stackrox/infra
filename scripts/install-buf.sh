#!/bin/bash

BUF_VERSION="1.16.0"
INSTALLED_BUF_VERSION=$(buf --version || echo "not installed")
if [ "$BUF_VERSION" != "${INSTALLED_BUF_VERSION}" ]; then
    curl -sSL \
        "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-$(uname -s)-$(uname -m)" \
        -o "$(go env GOPATH)/bin/buf" && \
    chmod +x "$(go env GOPATH)/bin/buf"
fi
