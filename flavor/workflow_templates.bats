#!/usr/bin/env bats

# shellcheck disable=SC1091
source "$BATS_TEST_DIRNAME/../test/bats-lib.sh"
load_bats_support

setup_file() {
  e2e_setup
  delete_all_workflows_by_flavor "test-gke-lite"
  kubectl apply -f "$BATS_TEST_DIRNAME/testdata/*.yaml"
}

# setup() {
#   # kubectl delete workflowtemplates --all --wait
# }

@test "a flavor from workflow template" {
  expect_count_flavor_id "test-gke-lite" 1
}

@test "expects a name" {
  run kubectl apply -f "$BATS_TEST_DIRNAME/testdata/negative/missing-name.yaml"
  assert_failure
  assert_output --partial "cannot use generate name with apply"
}

@test "gets a name from metadata.name" {
  flavor="$(infractl flavor get test-gke-lite --json)"
  assert_success
  name="$(echo "$flavor" | jq -r '.Name')"
  assert_equal "$name" "test-gke-lite"
}

@test "expects a description" {
  expect_count_flavor_id "missing-annotations" 0
  run kubectl -n infra logs -l app=infra-server
  assert_output --partial "[WARN] Ignoring a workflow template without infra.stackrox.io/description annotation: missing-annotations"
}

@test "availability is alpha by default" {
  flavor="$(infractl flavor get default-availability --json)"
  assert_success
  availability="$(echo "$flavor" | jq -r '.Availability')"
  assert_equal "$availability" "alpha"
}

@test "availability can be set" {
  flavor="$(infractl flavor get test-gke-lite --json)"
  assert_success
  availability="$(echo "$flavor" | jq -r '.Availability')"
  assert_equal "$availability" "stable"
}

@test "invalid availability workflows are dropped" {
  expect_count_flavor_id "invalid-availability" 0
  run kubectl -n infra logs -l app=infra-server
  assert_output --partial "[WARN] Ignoring a workflow template with an unknown infra.stackrox.io/availability annotation: invalid-availability, woot!"
}

# Parameters

@test "parameters must have descriptions" {
  expect_count_flavor_id "missing-parameter-descriptions" 0
  run kubectl -n infra logs -l app=infra-server
  assert_output --partial "[WARN] Ignoring a workflow template with a parameter (pod-security-policy) that has no description: missing-parameter-descriptions"
}

@test "a required parameter shows as such" {
  name_parm="$(infractl flavor get test-gke-lite --json | jq '.Parameters[] | select(.Name == "name")')"
  optionality="$(echo "$name_parm" | jq -r '.Optional')"
  assert_equal "$optionality" "false"
  internality="$(echo "$name_parm" | jq -r '.Internal')"
  assert_equal "$internality" "false"
}

@test "a parameter may have a description" {
  name_parm="$(infractl flavor get test-gke-lite --json | jq '.Parameters[] | select(.Name == "name")')"
  description="$(echo "$name_parm" | jq -r '.Description')"
  assert_equal "$description" "The name for the GKE cluster (tests required parameters)"
}

@test "an optional parameter shows as such" {
  nodes_param="$(infractl flavor get test-gke-lite --json | jq '.Parameters[] | select(.Name == "nodes")')"
  optionality="$(echo "$nodes_param" | jq -r '.Optional')"
  assert_equal "$optionality" "true"
  internality="$(echo "$nodes_param" | jq -r '.Internal')"
  assert_equal "$internality" "false"
}

@test "an optional parameter may have a default value" {
  nodes_param="$(infractl flavor get test-gke-lite --json | jq '.Parameters[] | select(.Name == "nodes")')"
  value="$(echo "$nodes_param" | jq -r '.Value')"
  assert_equal "$value" "1"
}

@test "an optional parameter may not have a default value" {
  k8s_param="$(infractl flavor get test-gke-lite --json | jq '.Parameters[] | select(.Name == "k8s-version")')"
  value="$(echo "$k8s_param" | jq -r '.Value')"
  assert_equal "$value" ""
}

@test "hardcoded (internal) parameters are hidden" {
  machine_param="$(infractl flavor get test-gke-lite --json | jq '.Parameters[] | select(.Name == "machine-type")')"
  assert_equal "$machine_param" ""
}

@test "parameters order follow workflow template order" {
  name_parm="$(infractl flavor get test-gke-lite --json | jq '.Parameters[] | select(.Name == "name")')"
  order="$(echo "$name_parm" | jq -r '.Order')"
  assert_equal "$order" "1"
  gcp_zone_parm="$(infractl flavor get test-gke-lite --json | jq '.Parameters[] | select(.Name == "gcp-zone")')"
  order="$(echo "$gcp_zone_parm" | jq -r '.Order')"
  assert_equal "$order" "7"
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
