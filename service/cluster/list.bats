#!/usr/bin/env bats

# Tests cluster list behavior.

# shellcheck disable=SC1091
source "$BATS_TEST_DIRNAME/../../test/bats-lib.sh"
load_bats_support

setup_file() {
  e2e_setup
  delete_all_workflows_by_name_prefix "list-"
  kubectl apply -f "workflows/*.yaml"

  ROOT="$(git rev-parse --show-toplevel)"
  export ROOT
}

@test "lists created" {
  id="$(infractl create simulate list-created --lifespan=30s --arg create-delay-seconds=5 --arg destroy-delay-seconds=5)"
  assert_success
  id="$(grep -E ^ID: <<<"$id")"
  id="${id//ID: /}"
  length="$(infractl list --json --prefix=list-created | jq -r '.Clusters|length')"
  assert_success
  assert_equal "$length" "1"
}

@test "lists expired (or not)" {
  id="$(infractl create simulate list-expired --lifespan=5s --arg create-delay-seconds=5 --arg destroy-delay-seconds=5)"
  assert_success
  id="$(grep -E ^ID: <<<"$id")"
  id="${id//ID: /}"
  assert_status_becomes "$id" "FINISHED" 60
  length="$(infractl list --json --prefix=list-expired | jq -r '.Clusters|length')"
  assert_success
  assert_equal "$length" "0"
  length="$(infractl list --json --prefix=list-expired --expired | jq -r '.Clusters|length')"
  assert_success
  assert_equal "$length" "1"
}

infractl() {
  "$ROOT"/bin/infractl -e localhost:8443 -k "$@"
}
