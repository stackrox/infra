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
    - image/Dockerfile
EOF
}

get_go_mod_version() {
    local file=$1
    grep '^go ' "$file" | cut -d' ' -f2
}

get_dockerfile_version() {
    grep '^FROM golang:' "$DOCKERFILE" | sed 's/FROM golang:\([^ ]*\).*/\1/'
}

cleanup_backups() {
    rm -f "$DOCKERFILE.bak"
}

sync_versions() {
    local go_version
    go_version=$(get_go_mod_version "${MAIN_GO_MOD}")

    echo "Syncing Go version $go_version across all files..."

    # Set up cleanup trap
    trap cleanup_backups EXIT INT TERM

    sed -i.bak 's/FROM golang:[^ ]*/FROM golang:'"$go_version"'/' "$DOCKERFILE"

    echo -e "${GREEN}✅ Go version synced to $go_version${NC}"
}

check_versions() {
    local go_version dockerfile_version

    go_version=$(get_go_mod_version "${MAIN_GO_MOD}")
    dockerfile_version=$(get_dockerfile_version)

    echo "go.mod: $go_version"
    echo "Dockerfile: $dockerfile_version"

    if [[ "$go_version" == "$dockerfile_version" ]]; then
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
