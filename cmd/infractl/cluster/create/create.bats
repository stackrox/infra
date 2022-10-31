#!/usr/bin/env bats

# shellcheck disable=SC1091
source "$BATS_TEST_DIRNAME/../../../../test/bats-lib.sh"

load_bats_support
e2e_setup

kubectl delete workflowtemplates --all --wait
kubectl apply -f "$BATS_TEST_DIRNAME/testdata/test-hello-world.yaml"

setup() {
  kubectl delete workflows --all --wait
}

@test "can create a workflow" {
  run infractl create test-hello-world this-is-a-test
  assert_success
  assert_output "ID: this-is-a-test"
}

@test "can create a workflow without a name" {
  run infractl create test-hello-world
  assert_success
  assert_output --regex "ID\: ...?.?"
}

infractl() {
  bin/infractl -e localhost:8443 -k "$@"
}
