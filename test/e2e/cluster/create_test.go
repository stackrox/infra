//go:build e2e
// +build e2e

package cluster_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stackrox/infra/test/utils/mock"
	"github.com/stretchr/testify/assert"
)

func TestCanCreateWorkflowWithoutName(t *testing.T) {
	clusterID, err := mock.InfractlCreateCluster(
		"simulate",
		"--lifespan=20s",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, clusterID)
	assert.Contains(t, clusterID, time.Now().Format("01-02"))
}

func TestDefaultNamesDoNotConflict(t *testing.T) {
	firstClusterID, err := mock.InfractlCreateCluster(
		"simulate",
		"--lifespan=20s",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, firstClusterID)
	secondClusterID, err := mock.InfractlCreateCluster(
		"simulate",
		"--lifespan=20s",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, secondClusterID)

	today := time.Now().Format("01-02")
	pattern := fmt.Sprintf(`-%s(-\d)?$`, today)
	assert.NotEqual(t, firstClusterID, secondClusterID)
	assert.Regexp(t, pattern, firstClusterID)
}

//	@test "provided name failed validation because too short" {
//		run infractl create test-qa-demo ab
//		assert_failure
//		assert_output --partial "Error: cluster name too short"
//	  }
func TestNameValidationTooShort(t *testing.T) {
	_, err := mock.InfractlCreateCluster(
		"simulate", "ab",
		"--lifespan=20s",
	)
	assert.ErrorContains(t, err, "cluster name too short")
}

func TestNameValidationTooLong(t *testing.T) {
	_, err := mock.InfractlCreateCluster(
		"simulate", "this-name-will-be-too-loooooooooooooooooooong",
		"--lifespan=20s",
	)
	assert.ErrorContains(t, err, "cluster name too long")
}

func TestNameValidationNoRegexMatch(t *testing.T) {
	_, err := mock.InfractlCreateCluster(
		"simulate", "THIS-IN-INVALID",
		"--lifespan=20s",
	)
	assert.ErrorContains(t, err, "The name does not match the requirements")
}

func TestCannotCreateClusterWithInvalidLifespan(t *testing.T) {
	_, err := mock.InfractlCreateCluster(
		"simulate",
		"--lifespan=3w",
	)
	assert.ErrorContains(t, err, "invalid argument \"3w\" for \"--lifespan\" flag")
}
