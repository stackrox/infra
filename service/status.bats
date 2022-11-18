#!/usr/bin/env bats

# shellcheck disable=SC1091
source "$BATS_TEST_DIRNAME/../test/bats-lib.sh"
load_bats_support

delete_status_configmap() {
  kubectl delete configmap/status -n infra || true
}

infractl() {
  bin/infractl -e localhost:8443 -k "$@"
}

setup_file() {
  e2e_setup
  delete_status_configmap
}

@test "reset returns no active maintenance" {
  status="$(infractl status reset --json | jq -r '.Status')"
  assert_success
  assert_equal "$status" "{}"

  run kubectl -n infra logs -l app=infra-server
  assert_output --partial "[INFO] Status was reset"
}

@test "set returns active maintenance with maintainer" {
  whoami="$(infractl whoami --json | jq -r '.Principal.ServiceAccount.Email')"
  status="$(infractl status set --json | jq -r '.Status')"
  maintenanceActive="$(echo "$status" | jq -r '.MaintenanceActive')"
  maintainer="$(echo "$status" | jq -r '.Maintainer')"
  assert_success
  assert_equal "$maintenanceActive" "true"
  assert_equal "$maintainer" "$whoami"

  run kubectl -n infra logs -l app=infra-server
  assert_output --partial "[INFO] New Status was set by maintainer $maintainer"
}

@test "get returns no active maintenance after lazy initialization" {
    delete_status_configmap
    status="$(infractl status get --json | jq -r '.Status')"
    assert_success
    assert_equal "$status" "{}"

  run kubectl -n infra logs -l app=infra-server
  assert_output --partial "[INFO] Initialized infra status lazily"
}
