#!/usr/bin/env bats

# shellcheck disable=SC1091
source "$BATS_TEST_DIRNAME/../../../../test/bats-lib.sh"

load_bats_support
e2e_setup

kubectl delete workflowtemplates --all --wait
kubectl apply -f "$BATS_TEST_DIRNAME/testdata/*.yaml"

tag_suffix="$(make --quiet tag | sed 's/\./-/g')"

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
  assert_output --regexp "ID\: ...?.?"
}

@test "names include a date by default" {
  run infractl create test-hello-world
  assert_success
  date_suffix="$(date '+%m-%d')"
  assert_output --regexp "ID\: ...?.?-${date_suffix}"
}

@test "names do not conflict" {
  run infractl create test-hello-world
  run infractl create test-hello-world
  assert_success
  date_suffix="$(date '+%m-%d')"
  assert_output --regexp "ID\: ...?.?-${date_suffix}-2"
}

@test "qa-demo names use the tag" {
  run infractl create test-qa-demo
  assert_success
  assert_output --regexp "ID\: ...?.?-${tag_suffix}-1"
}

@test "qa-demo names use the tag - subdirs are OK" {
  pushd "$BATS_TEST_DIRNAME"
  run infractl create test-qa-demo
  assert_success
  assert_output --regexp "ID\: ...?.?-${tag_suffix}-1"
  popd
}

infractl() {
  "$(git rev-parse --show-toplevel)"/bin/infractl -e localhost:8443 -k "$@"
}
