#!/usr/bin/env bash

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
source "$ROOT/scripts/lib.sh"

set -euo pipefail

create_consolidated_values() {
    if [[ "$#" -ne 1 ]]; then
        die "missing args. usage: create_consolidated_values <environment>"
    fi
    local environment="$1"

    info "Creating a combined values file for chart/infra-server/configuration/$environment files"

    if [[ ! -e "$ROOT/chart/infra-server/configuration" ]]; then
        die "chart/infra-server/configuration is missing. Download the configuration with 'make configuration-download'"
    fi

    local values_file="$ROOT/chart/infra-server/configuration/$environment-values-from-files.yaml"
    rm -f "$values_file"

    pushd "$ROOT/chart/infra-server/configuration/$environment" > /dev/null
    shopt -s globstar nullglob
    for cfg_file in **; do
        if [[ -d "$cfg_file" ]]; then
            continue
        fi
        if [[ "$cfg_file" =~ (README|DS_Store) ]]; then
            continue
        fi

        local helm_safe_key="${cfg_file//[.-]/_}"
        helm_safe_key="${helm_safe_key////__}"

        echo "$helm_safe_key: $(base64 -w0 < "$cfg_file")" >> "$values_file"
        echo >> "$values_file"
    done
    popd > /dev/null
}

create_consolidated_values "development"
create_consolidated_values "production"
