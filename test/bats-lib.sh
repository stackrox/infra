#!/usr/bin/env bash

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
