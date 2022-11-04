#!/usr/bin/env bats

# shellcheck disable=SC1091
source "$BATS_TEST_DIRNAME/../../test/bats-lib.sh"
load_bats_support

setup_file() {
  e2e_setup
  kubectl apply -f "workflows/*.yaml"
  kubectl delete workflows --all --wait

  ROOT="$(git rev-parse --show-toplevel)"
  export ROOT
}

@test "can run through the infra standard lifecycle" {
  id="$(infractl create simulate --lifespan=15s --arg create-delay-seconds=5 --arg destroy-delay-seconds=5)"
  assert_success
  id="${id//ID: /}"
  status="$(infractl get "$id" --json | jq -r '.Status')"
  assert_success
  assert_equal "$status" "CREATING"
  assert_status_becomes "$id" "READY"
  assert_status_becomes "$id" "DESTROYING"
  assert_status_becomes "$id" "FINISHED"
}

@test "can fail in create" {
  id="$(infractl create simulate --lifespan=15s --arg create-delay-seconds=5 --arg create-outcome=fail)"
  assert_success
  id="${id//ID: /}"
  status="$(infractl get "$id" --json | jq -r '.Status')"
  assert_success
  assert_equal "$status" "CREATING"
  assert_status_becomes "$id" "FAILED"
}

@test "can fail in destroy" {
  id="$(infractl create simulate --lifespan=15s --arg create-delay-seconds=5 --arg destroy-delay-seconds=5 --arg destroy-outcome=fail)"
  assert_success
  id="${id//ID: /}"
  status="$(infractl get "$id" --json | jq -r '.Status')"
  assert_success
  assert_equal "$status" "CREATING"
  assert_status_becomes "$id" "READY"
  assert_status_becomes "$id" "DESTROYING"
  assert_status_becomes "$id" "FAILED"
}

infractl() {
  "$ROOT"/bin/infractl -e localhost:8443 -k "$@"
}

assert_status_becomes() {
  id="$1"
  desired_status="$2"

  tries=0
  limit=30
  while [[ "$((tries++))" -le "$limit" ]]; do
    status="$(infractl get "$id" --json | jq -r '.Status')"
    assert_success
    diag "$status at iteration $tries"
    if [[ "$status" == "$desired_status" ]]; then
      break
    fi
    if [[ "$tries" -eq "$limit" ]]; then
      assert_equal "$status" "$desired_status"
    fi
    sleep 1
  done
}
