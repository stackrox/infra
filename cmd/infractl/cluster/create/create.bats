#!/usr/bin/env bats

# shellcheck disable=SC1091
source "$BATS_TEST_DIRNAME/../../../../test/bats-lib.sh"
load_bats_support

setup_file() {
  e2e_setup

  # These cannot run in parallel because the mocks in particular are per test
  # dependent.
  export BATS_NO_PARALLELIZE_WITHIN_FILE=true

  kubectl apply -f "$BATS_TEST_DIRNAME/testdata/*.yaml"

  ROOT="$(git rev-parse --show-toplevel)"
  export ROOT
  mkdir -p "$ROOT/test/mocks/stackrox/collector"
  mkdir -p "$ROOT/test/mocks/stackrox/stackrox"
  export PATH="$ROOT/test/mocks:$PATH}"
  date_suffix="$(date '+%m-%d')"
  export date_suffix
  test_tag="0.5.2-23-g2e873a9145"
  export test_tag
  tag_suffix="0-5-2-23-g2e873a9145"
  export tag_suffix
}

setup() {
  create_mock_make_for_tag "${test_tag}"
  create_mock_git_for_toplevel "$ROOT/test/mocks/stackrox/stackrox"
  delete_all_workflows_by_flavor "test-hello-world"
  delete_all_workflows_by_flavor "test-qa-demo"
}

@test "can create a workflow" {
  run infractl create test-hello-world this-is-a-test
  assert_success
  assert_output --partial "ID: this-is-a-test"
}

@test "can create a workflow without a name" {
  run infractl create test-hello-world
  assert_success
  assert_output --regexp "ID: ...?.?"
}

@test "default names include a date" {
  run infractl create test-hello-world
  assert_success
  assert_output --regexp "ID: ...?.?-${date_suffix}"
}

@test "default names do not conflict" {
  run infractl create test-hello-world
  run infractl create test-hello-world
  assert_success
  assert_output --regexp "ID: ...?.?-${date_suffix}-2"
}

@test "default qa-demo names use the tag" {
  run infractl create test-qa-demo
  assert_success
  assert_output --regexp "ID: ...?.?-${tag_suffix}-1"
}

@test "default qa-demo names strip any -dirty suffix" {
  create_mock_make_for_tag "${test_tag}-dirty"
  run infractl create test-qa-demo
  assert_success
  assert_output --regexp "ID: ...?.?-${tag_suffix}-1"
}

@test "qa-demo defaults main-image from the tag" {
  create_mock_make_for_tag "${test_tag}-dirty"
  run infractl create test-qa-demo
  assert_success
  arg="$(kubectl get workflows -o json | jq -r '.items[].spec.arguments.parameters[] | select(.name=="main-image") | .value')"
  assert_equal "$arg" "quay.io/stackrox-io/main:${test_tag}"
}

@test "qa-demo defaults main-image from --rhacs" {
  run infractl create test-qa-demo --rhacs
  assert_success
  arg="$(kubectl get workflows -o json | jq -r '.items[].spec.arguments.parameters[] | select(.name=="main-image") | .value')"
  assert_equal "$arg" "quay.io/rhacs-eng/main:${test_tag}"
}

@test "does not override main-image if passed" {
  run infractl create test-qa-demo --arg main-image=a.b.c
  assert_success
  arg="$(kubectl get workflows -o json | jq -r '.items[].spec.arguments.parameters[] | select(.name=="main-image") | .value')"
  assert_equal "$arg" "a.b.c"
}

@test "default qa-demo names use the date when not in a git context" {
  create_mock_git_that_fails
  run infractl create test-qa-demo --arg main-image=test
  assert_success
  assert_output --regexp "ID: ...?.?-${date_suffix}-1"
}

@test "qa-demo must supply a main-image when not in a git context" {
  create_mock_git_that_fails
  run infractl create test-qa-demo
  assert_failure
  assert_output --partial "parameter \"main-image\" was not provided"
}

@test "default qa-demo names use the date when in a git context other than stackrox" {
  create_mock_git_for_toplevel "$ROOT/test/mocks/stackrox/collector"
  create_mock_make_for_tag_without_pwd_check "${test_tag}"
  run infractl create test-qa-demo --arg main-image=test
  assert_success
  assert_output --regexp "ID: ...?.?-${date_suffix}-1"
}

@test "qa-demo must supply a main-image when in a git context other than stackrox" {
  create_mock_git_for_toplevel "$ROOT/test/mocks/stackrox/collector"
  create_mock_make_for_tag_without_pwd_check "${test_tag}"
  run infractl create test-qa-demo
  assert_failure
  assert_output --partial "parameter \"main-image\" was not provided"
}

@test "provided name failed validation because too short" {
  run infractl create test-qa-demo ab
  assert_failure
  assert_output --partial "Error: cluster name too short"
}

@test "provided name failed validation because too long" {
  run infractl create test-qa-demo this-name-will-be-too-loooooooooooooooooooong
  assert_failure
  assert_output --partial "Error: cluster name too long"
}

@test "provided name failed validation because does not match regex" {
  run infractl create test-qa-demo THIS-IN-INVALID
  assert_failure
  assert_output --partial "Error: The name does not match its requirements."
}

infractl() {
  "$ROOT"/bin/infractl -e localhost:8443 -k "$@"
}

create_mock_make_for_tag() {
  cat <<_EOD_ > "$ROOT/test/mocks/make"
#!/usr/bin/env bash
# this should really be an @test that make runs in the right context, but
# I cannot figure that one out.
if [[ "\$(pwd)" != "$ROOT/test/mocks/stackrox/stackrox" ]]; then
  echo "make should run in the mock root"
  exit 1
fi
echo $1
_EOD_
  chmod 0755 "$ROOT/test/mocks/make"
}

create_mock_make_for_tag_without_pwd_check() {
  cat <<_EOD_ > "$ROOT/test/mocks/make"
#!/usr/bin/env bash
echo $1
_EOD_
  chmod 0755 "$ROOT/test/mocks/make"
}

create_mock_git_for_toplevel() {
  cat <<_EOD_ > "$ROOT/test/mocks/git"
#!/usr/bin/env bash
echo $1
_EOD_
  chmod 0755 "$ROOT/test/mocks/git"
}

create_mock_git_that_fails() {
  cat <<_EOD_ > "$ROOT/test/mocks/git"
#!/usr/bin/env bash
exit 1
_EOD_
  chmod 0755 "$ROOT/test/mocks/git"
}
