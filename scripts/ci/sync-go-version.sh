#!/usr/bin/env bash

set -euo pipefail

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly NC='\033[0m' # No Color

# Script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Files to manage
readonly MAIN_GO_MOD="${PROJECT_ROOT}/go.mod"
readonly LOCALDEV_GO_MOD="${PROJECT_ROOT}/scripts/local-dev/go.mod"
readonly GOLANGCI_CONFIG="${PROJECT_ROOT}/.golangci.yml"
readonly DOCKERFILE="${PROJECT_ROOT}/image/Dockerfile"

usage() {
    cat << EOF
Usage: $0 [COMMAND]

Commands:
    sync    Synchronize Go versions across all files (uses go.mod as source of truth)
    check   Check if Go versions are synchronized (exits 1 if not)
    help    Show this help message

Files managed:
    - go.mod (source of truth)
    - .golangci.yml
    - image/Dockerfile
    - scripts/local-dev/go.mod
EOF
}

get_go_mod_version() {
    local file=$1
    grep '^go ' "$file" | cut -d' ' -f2
}

get_golangci_version() {
    grep 'go: ' "$GOLANGCI_CONFIG" | sed 's/.*go: "\([^"]*\)".*/\1/'
}

get_dockerfile_version() {
    grep '^FROM golang:' "$DOCKERFILE" | sed 's/FROM golang:\([^ ]*\).*/\1/'
}

cleanup_backups() {
    rm -f "$GOLANGCI_CONFIG.bak" "$DOCKERFILE.bak" "$LOCALDEV_GO_MOD.bak"
}

sync_versions() {
    local go_version
    go_version=$(get_go_mod_version "${MAIN_GO_MOD}")

    echo "Syncing Go version $go_version across all files..."

    # Set up cleanup trap
    trap cleanup_backups EXIT INT TERM

    sed -i.bak 's/go: "[^"]*"/go: "'"$go_version"'"/' "$GOLANGCI_CONFIG"
    sed -i.bak 's/FROM golang:[^ ]*/FROM golang:'"$go_version"'/' "$DOCKERFILE"
    sed -i.bak 's/^go .*/go '"$go_version"'/' "$LOCALDEV_GO_MOD"

    echo -e "${GREEN}✅ Go version synced to $go_version${NC}"
}

check_versions() {
    local go_version golangci_version dockerfile_version localdev_version

    go_version=$(get_go_mod_version "${MAIN_GO_MOD}")
    golangci_version=$(get_golangci_version)
    dockerfile_version=$(get_dockerfile_version)
    localdev_version=$(get_go_mod_version "${LOCALDEV_GO_MOD}")

    echo "go.mod: $go_version"
    echo ".golangci.yml: $golangci_version"
    echo "Dockerfile: $dockerfile_version"
    echo "local-dev: $localdev_version"

    if [[ "$go_version" == "$golangci_version" && "$go_version" == "$dockerfile_version" && "$go_version" == "$localdev_version" ]]; then
        echo -e "${GREEN}✅ All versions synchronized${NC}"
        return 0
    else
        echo -e "${RED}❌ Versions out of sync. Run 'make sync-go-version' to synchronize them.${NC}"
        return 1
    fi
}

main() {
    cd "$PROJECT_ROOT"

    case "${1:-}" in
        sync) sync_versions ;;
        check) check_versions ;;
        help|--help|-h) usage ;;
        *) [[ -z "${1:-}" ]] && echo -e "${RED}ERROR: No command specified${NC}" >&2 || echo -e "${RED}ERROR: Unknown command: $1${NC}" >&2
           echo; usage; exit 1 ;;
    esac
}

main "$@"
