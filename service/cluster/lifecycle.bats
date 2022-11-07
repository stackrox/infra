#!/usr/bin/env bats

# Tests infra through some typical workflow lifecycles.

# Notes:
# These tests can run in parallel (bats --jobs #).
# The 5 second create/destroy times are typically longer due to argo container
# overhead. Hence the need for a 30 second lifespan - to test the workflow enters
# the READY state and does not move to DESTROYING before that is detected. Overall
# run time is typically < 60 seconds.
# If you make changes, run a repeat look to get confidence e.g.:
# run=0; while time bats --jobs 5 service/**/*.bats; do echo running "$((run++))"; done

# shellcheck disable=SC1091
source "$BATS_TEST_DIRNAME/../../test/bats-lib.sh"
load_bats_support

setup_file() {
  e2e_setup
  kubectl apply -f "workflows/*.yaml"
  delete_all_workflows_by_flavor "simulate"

  ROOT="$(git rev-parse --show-toplevel)"
  export ROOT
}

@test "can run through the infra standard lifecycle" {
  id="$(infractl create simulate standard-"$(date '+%s')" --lifespan=30s --arg create-delay-seconds=5 --arg destroy-delay-seconds=5)"
  assert_success
  id="$(grep -E ^ID: <<<"$id")"
  id="${id//ID: /}"
  status="$(infractl get "$id" --json | jq -r '.Status')"
  assert_success
  assert_equal "$status" "CREATING"
  assert_status_becomes "$id" "READY"
  assert_status_becomes "$id" "DESTROYING"
  assert_status_becomes "$id" "FINISHED"
}

@test "can fail in create" {
  id="$(infractl create simulate create-fails-"$(date '+%s')" --lifespan=30s --arg create-delay-seconds=5 --arg create-outcome=fail)"
  assert_success
  id="$(grep -E ^ID: <<<"$id")"
  id="${id//ID: /}"
  status="$(infractl get "$id" --json | jq -r '.Status')"
  assert_success
  assert_equal "$status" "CREATING"
  assert_status_becomes "$id" "FAILED"
}

@test "can fail in destroy" {
  id="$(infractl create simulate destroy-fails-"$(date '+%s')" --lifespan=30s --arg create-delay-seconds=5 --arg destroy-delay-seconds=5 --arg destroy-outcome=fail)"
  assert_success
  id="$(grep -E ^ID: <<<"$id")"
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
    if [[ "$status" == "$desired_status" ]]; then
      break
    fi
    if [[ "$tries" -eq "$limit" ]]; then
      assert_equal "$status" "$desired_status"
    fi
    sleep 1
  done
}
