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

    {
        echo "# This is a helm values file that combines the contents of the $environment configuration files."
        echo "# It is updated by each render-* make target. Changes made here will be lost."
        echo
    } >> "$values_file"

    pushd "$ROOT/chart/infra-server/configuration/$environment" > /dev/null
    while IFS='' read -r cfg_file; do
      local helm_safe_key="${cfg_file//[.-]/_}"
      helm_safe_key="${helm_safe_key////__}"

      echo "$helm_safe_key: $(base64 < "$cfg_file" | tr -d '\n')" >> "$values_file"
      echo >> "$values_file"
    done < <(find . -type f -not -name '*.md' -not -name '*.DS_Store' | cut -c3-)
    popd > /dev/null
}

create_consolidated_values "development"
create_consolidated_values "production"