#!/bin/bash

set -xeuo pipefail

go env GOPATH
go env GOARCH
go env GOOS

uname -s
uname -m

BUF_VERSION="1.16.0"
INSTALLED_BUF_VERSION=$(buf --version || echo "not installed")
if [ "$BUF_VERSION" != "${INSTALLED_BUF_VERSION}" ]; then
    curl -SL \
        "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-$(uname -s)-$(uname -m)" \
        -o "$(go env GOPATH)/bin/buf" && \
    chmod +x "$(go env GOPATH)/bin/buf"
fi
