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

assert_status_becomes() {
  id="$1"
  desired_status="$2"
  limit="${3:-40}"

  tries=0
  while [[ "$((tries++))" -le "$limit" ]]; do
    status="$(infractl get "$id" --json | jq -r '.Status')"
    # diag "$id $status"
    assert_success
    if [[ "$status" == "$desired_status" ]]; then
      break
    fi
    if [[ "$tries" -eq "$limit" ]]; then
      assert_equal "$status" "$desired_status"
    fi
    sleep 1
  done
}

assert_status_remains() {
  id="$1"
  status="$2"
  try_for="$3"

  tries=0
  limit="$try_for"
  while [[ "$((tries++))" -le "$limit" ]]; do
    currently="$(infractl get "$id" --json | jq -r '.Status')"
    # diag "$id $currently"
    assert_success
    assert_equal "$status" "$currently"
    sleep 1
  done
}
