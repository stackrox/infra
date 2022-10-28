#!/usr/bin/env bats

# shellcheck disable=SC1091
source "$BATS_TEST_DIRNAME/../../../../test/bats-lib.sh"

load_bats_support

setup() {
  e2e_setup
  kubectl apply -f "$BATS_TEST_DIRNAME/testdata/test-hello-world.yaml"
}

@test "can create a workflow" {
  run infractl create test-hello-world this-is-a-test
  assert_success
  assert_output "ID: this-is-a-test"
}

infractl() {
  bin/infractl -e localhost:8443 -k "$@"
}
