//go:build e2e
// +build e2e

package cluster_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stackrox/infra/cmd/infractl/cluster/create"
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

// #######
// QA DEMO
// #######
func mockRootDir(path string) func() string {
	return func() string {
		return path
	}
}

func mockMakeTag(tag string) func(string) string {
	return func(rootDir string) string {
		return tag
	}
}

func TestQaDemoDefaultToTag(t *testing.T) {
	origGetRootDir := create.GetRootDir
	create.GetRootDir = mockRootDir("stackrox/stackrox")
	defer func() { create.GetRootDir = origGetRootDir }()
	origGetMakeTag := create.GetMakeTag
	create.GetMakeTag = mockMakeTag("4.6.0")
	defer func() { create.GetMakeTag = origGetMakeTag }()

	clusterID, err := mock.InfractlCreateCluster(
		"test-qa-demo",
		"--lifespan=20s",
	)
	assert.NoError(t, err)
	assert.Contains(t, clusterID, "4-6-0")

	cluster, err := mock.InfractlGetCluster(clusterID)
	assert.NoError(t, err)

	p, err := findParameter(cluster.Parameters, "main-image")
	assert.NoError(t, err)
	assert.Equal(t, "quay.io/stackrox-io/main:4.6.0", p.GetValue())
}

func TestQaDemoDefaultToTagWithoutDirty(t *testing.T) {
	origGetRootDir := create.GetRootDir
	create.GetRootDir = mockRootDir("stackrox/stackrox")
	defer func() { create.GetRootDir = origGetRootDir }()
	origGetMakeTag := create.GetMakeTag
	create.GetMakeTag = mockMakeTag("4.6.0")
	defer func() { create.GetMakeTag = origGetMakeTag }()

	clusterID, err := mock.InfractlCreateCluster(
		"test-qa-demo",
		"--lifespan=20s",
	)
	assert.NoError(t, err)
	assert.Contains(t, clusterID, "4-6-0")
	assert.NotContains(t, clusterID, "-dirty")
}

func TestQaDemoDefaultToRhacsMainImageFromTag(t *testing.T) {
	origGetRootDir := create.GetRootDir
	create.GetRootDir = mockRootDir("stackrox/stackrox")
	defer func() { create.GetRootDir = origGetRootDir }()
	origGetMakeTag := create.GetMakeTag
	create.GetMakeTag = mockMakeTag("4.6.0")
	defer func() { create.GetMakeTag = origGetMakeTag }()

	clusterID, err := mock.InfractlCreateCluster(
		"test-qa-demo",
		"--rhacs",
		"--lifespan=20s",
	)
	assert.NoError(t, err)
	assert.Contains(t, clusterID, "4-6-0")

	cluster, err := mock.InfractlGetCluster(clusterID)
	assert.NoError(t, err)

	p, err := findParameter(cluster.Parameters, "main-image")
	assert.NoError(t, err)
	assert.Equal(t, "quay.io/rhacs-eng/main:4.6.0", p.GetValue())
}

func TestQaDemoDefaultToDateNotGit(t *testing.T) {
	clusterID, err := mock.InfractlCreateCluster(
		"test-qa-demo",
		"--lifespan=20s",
		"--arg=main-image=test",
	)
	assert.NoError(t, err)
	assert.Contains(t, clusterID, time.Now().Format("01-02"))
}

func TestQaDemoNotGitMustHaveMainImage(t *testing.T) {
	_, err := mock.InfractlCreateCluster(
		"test-qa-demo",
		"--lifespan=20s",
	)
	assert.ErrorContains(t, err, "parameter \"main-image\" was not provided")
}

func TestQaDemoOtherGitUseDate(t *testing.T) {
	create.GetRootDir = mockRootDir("stackrox/collector")
	create.GetMakeTag = mockMakeTag("a.b.c")

	clusterID, err := mock.InfractlCreateCluster(
		"test-qa-demo",
		"--lifespan=20s",
		"--arg=main-image=test",
	)
	assert.NoError(t, err)
	assert.NotContains(t, clusterID, "a.b.c")
	assert.Contains(t, clusterID, time.Now().Format("01-02"))
}

func TestQaDemoOtherGitMustHaveMainImage(t *testing.T) {
	create.GetRootDir = mockRootDir("stackrox/collector")
	create.GetMakeTag = mockMakeTag("a.b.c")

	_, err := mock.InfractlCreateCluster(
		"test-qa-demo",
		"--lifespan=20s",
	)
	assert.ErrorContains(t, err, "parameter \"main-image\" was not provided")
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
