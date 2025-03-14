//go:build e2e
// +build e2e

package cluster_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/test/utils"
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

func TestQaDemoDefaultsOverrideMainImage(t *testing.T) {
	clusterID, err := mock.InfractlCreateCluster(
		"test-qa-demo", utils.GetUniqueClusterName("qa-demo-override"),
		"--arg=main-image=a.b.c",
		"--lifespan=20s",
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, clusterID)

	cluster, err := mock.InfractlGetCluster(clusterID)
	assert.NoError(t, err)

	p, err := findParameter(cluster.Parameters, "main-image")
	assert.NoError(t, err)
	assert.Equal(t, "a.b.c", p.GetValue())
}

func findParameter(parameters []v1.Parameter, name string) (v1.Parameter, error) {
	for _, p := range parameters {
		if p.GetName() == name {
			return p, nil
		}
	}
	return v1.Parameter{}, errors.New("parameter not found in cluster")
}
