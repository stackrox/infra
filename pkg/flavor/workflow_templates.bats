#!/usr/bin/env bats

# shellcheck disable=SC1091
source "$BATS_TEST_DIRNAME/../../test/bats-lib.sh"
load_bats_support

setup_file() {
  e2e_setup
  kubectl apply -f "$BATS_TEST_DIRNAME/testdata/*.yaml"
}

@test "expects a name" {
  run kubectl apply -f "$BATS_TEST_DIRNAME/testdata/missing-name.yaml"
  assert_failure
  assert_output --partial "cannot use generate name with apply"
}

@test "expects a description" {
  expect_count_flavor_id "missing-annotations" 0
  run kubectl -n infra logs -l app=infra-server
  assert_output --partial '"msg":"ignoring a workflow template without infra.stackrox.io/description annotation","template-name":"missing-annotations"'
}


@test "invalid availability workflows are dropped" {
  expect_count_flavor_id "invalid-availability" 0
  run kubectl -n infra logs -l app=infra-server
  assert_output --partial '"msg":"ignoring a workflow template with an unknown infra.stackrox.io/availability annotation","template-name":"invalid-availability","template-availability":"woot!"'
}

# Parameters
@test "parameters must have descriptions" {
  expect_count_flavor_id "missing-parameter-descriptions" 0
  run kubectl -n infra logs -l app=infra-server
  assert_output --partial '"msg":"ignoring a workflow template with a parameter that has no description","template-name":"missing-parameter-descriptions","parameter":"gcp-zone"'
}

infractl() {
  bin/infractl -e localhost:8443 -k "$@"
}

expect_count_flavor_id() {
  local expect_ID="$1"
  local expect_count="$2"
  local listing count

  listing="$(infractl flavor list --all --json)"
  assert_success
  count="$(echo "$listing" | jq '.Flavors[] | select(.ID == "'"$expect_ID"'")' | jq -s 'length')"
  assert_equal "$count" "$expect_count"
}
