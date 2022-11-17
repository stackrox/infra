#!/usr/bin/env bash

# (this does not appear to be BATs compatible)
# set -euo pipefail

load_bats_support() {
  # Find the location for bats packages. Centos, local install, CI.
  for helpers_root in "/usr/share/toolbox/test/system/libs" "${HOME}/bats-core" "/usr/lib/node_modules"; do
    if [[ -f "$helpers_root/bats-support/load.bash" ]]; then
      break
    fi
  done
  if [[ ! -f "${helpers_root}/bats-support/load.bash" ]]; then
    echo "Cannot find bats packages. Quitting test."
    exit 1
  fi
  load "${helpers_root}/bats-support/load.bash"
  load "${helpers_root}/bats-assert/load.bash"
}

e2e_setup() {
  # safety check, must be an infra-pr cluster
  context="$(kubectl config current-context)"
  if ! [[ "$context" =~ infra-pr-[[:digit:]]+ ]]; then
    echo "kubectl config current-context: $context"
    echo "Quitting test. This is not an infra PR development cluster."
    exit 1
  fi
}

delete_all_workflows_by_flavor() {
  flavor="$1"
  kubectl get workflows -o json | jq -r '.items[] | select( .metadata.annotations."infra.stackrox.com/flavor" == "'"$flavor"'" ) | .metadata.name' | \
    xargs kubectl delete workflow || true
}

diag() {
  # shellcheck disable=SC2001
  echo "$(date '+%H:%M:%S') " "$@" | sed -e 's/^/# /' >&3 ;
}
