#!/usr/bin/env bash

set -euo pipefail

info() {
    echo "INFO: $(date): $*"
}

die() {
    echo >&2 "$@"
    exit 1
}
