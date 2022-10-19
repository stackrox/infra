#!/usr/bin/env bats

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

setup() {
  # safety check, must be an infra-pr cluster
  context="$(kubectl config current-context)"
  if ! [[ "$context" =~ infra-pr-[[:digit:]]+ ]]; then
    echo "kubectl config current-context: $context"
    echo "Quitting test. This is not an infra PR development cluster."
    exit 1
  fi
  kubectl delete workflowtemplates --all --wait
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

@test "can add a workflow template" {
  run kubectl apply -f "$BATS_TEST_DIRNAME/testdata/test-gke-lite.yaml"
  expect_count_flavor_id "test-gke-lite" 1
}

@test "expects a name" {
  run kubectl apply -f "$BATS_TEST_DIRNAME/testdata/missing-annotations.yaml"
  expect_count_flavor_id "missing-annotations" 0
  run kubectl -n infra logs -l app=infra-server
  assert_output --partial "[WARN] Ignoring a workflow template without infra.stackrox.io/name annotation: missing-annotations"
}

@test "expects a description" {
  run kubectl apply -f "$BATS_TEST_DIRNAME/testdata/missing-annotations.yaml"
  expect_count_flavor_id "missing-annotations" 0
  run kubectl -n infra logs -l app=infra-server
  assert_output --partial "[WARN] Ignoring a workflow template without infra.stackrox.io/description annotation: missing-annotations"
}

@test "availability is alpha by default" {
  run kubectl apply -f "$BATS_TEST_DIRNAME/testdata/default-availability.yaml"
  flavor="$(infractl flavor get default-availability --json)"
  assert_success
  availability="$(echo "$flavor" | jq -r '.Availability')"
  assert_equal "$availability" "alpha"
}

@test "availability can be set" {
  run kubectl apply -f "$BATS_TEST_DIRNAME/testdata/test-gke-lite.yaml"
  flavor="$(infractl flavor get test-gke-lite --json)"
  assert_success
  availability="$(echo "$flavor" | jq -r '.Availability')"
  assert_equal "$availability" "stable"
}

@test "invalid availability workflows are dropped" {
  run kubectl apply -f "$BATS_TEST_DIRNAME/testdata/invalid-availability.yaml"
  expect_count_flavor_id "invalid-availability" 0
  run kubectl -n infra logs -l app=infra-server
  assert_output --partial "[WARN] Ignoring a workflow template with an unknown infra.stackrox.io/availability annotation: invalid-availability, woot!"
}

@test "a required parameter shows as such" {
  run kubectl apply -f "$BATS_TEST_DIRNAME/testdata/test-gke-lite.yaml"
  name_parm="$(infractl flavor get test-gke-lite --json | jq '.Parameters[] | select(.Name == "name")')"
  optionality="$(echo "$name_parm" | jq -r '.Optional')"
  assert_equal "$optionality" "false"
  internality="$(echo "$name_parm" | jq -r '.Internal')"
  assert_equal "$internality" "false"
}

@test "a parameter may have a description" {
  run kubectl apply -f "$BATS_TEST_DIRNAME/testdata/test-gke-lite.yaml"
  name_parm="$(infractl flavor get test-gke-lite --json | jq '.Parameters[] | select(.Name == "name")')"
  description="$(echo "$name_parm" | jq -r '.Description')"
  assert_equal "$description" "The name for the GKE cluster (tests required parameters)"
}

@test "an optional parameter shows as such" {
  run kubectl apply -f "$BATS_TEST_DIRNAME/testdata/test-gke-lite.yaml"
  nodes_param="$(infractl flavor get test-gke-lite --json | jq '.Parameters[] | select(.Name == "nodes")')"
  optionality="$(echo "$nodes_param" | jq -r '.Optional')"
  assert_equal "$optionality" "true"
  internality="$(echo "$nodes_param" | jq -r '.Internal')"
  assert_equal "$internality" "false"
}

@test "an optional parameter may have a default value" {
  run kubectl apply -f "$BATS_TEST_DIRNAME/testdata/test-gke-lite.yaml"
  nodes_param="$(infractl flavor get test-gke-lite --json | jq '.Parameters[] | select(.Name == "nodes")')"
  value="$(echo "$nodes_param" | jq -r '.Value')"
  assert_equal "$value" "1"
}

@test "an optional parameter may not have a default value" {
  run kubectl apply -f "$BATS_TEST_DIRNAME/testdata/test-gke-lite.yaml"
  k8s_param="$(infractl flavor get test-gke-lite --json | jq '.Parameters[] | select(.Name == "k8s-version")')"
  value="$(echo "$k8s_param" | jq -r '.Value')"
  assert_equal "$value" ""
}

@test "hardcoded (internal) parameters are hidden" {
  run kubectl apply -f "$BATS_TEST_DIRNAME/testdata/test-gke-lite.yaml"
  machine_param="$(infractl flavor get test-gke-lite --json | jq '.Parameters[] | select(.Name == "machine-type")')"
  assert_equal "$machine_param" ""
}
