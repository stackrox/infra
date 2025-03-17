package flavor_test

import (
	"os"
	"testing"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/stackrox/infra/pkg/flavor"
	"github.com/stretchr/testify/assert"
)

func TestWorkflow2TemplateExpectsDescription(t *testing.T) {
	template, err := readWorkflowTemplateFromFixture("testdata/missing-flavor-description-in-annotation.yaml")
	assert.NoError(t, err)
	f, validationErrors := flavor.WorkflowTemplate2Flavor(template)
	assert.Nil(t, f, "flavor must be nil, because a validation error occurred")
	assert.Len(t, validationErrors, 1)
	assert.Equal(t, "ignoring a workflow template without infra.stackrox.io/description annotation", validationErrors[0].Error())
}

func TestWorkflow2TemplateNeedsValidAvailability(t *testing.T) {
	template, err := readWorkflowTemplateFromFixture("testdata/invalid-availability.yaml")
	assert.NoError(t, err)
	f, validationErrors := flavor.WorkflowTemplate2Flavor(template)
	assert.Nil(t, f, "flavor must be nil, because a validation error occurred")
	assert.Len(t, validationErrors, 1)
	assert.Equal(t, "ignoring a workflow template with an unknown infra.stackrox.io/availability annotation", validationErrors[0].Error())
}

func TestWorkflow2TemplateParametersNeedDescription(t *testing.T) {
	template, err := readWorkflowTemplateFromFixture("testdata/missing-parameter-description.yaml")
	assert.NoError(t, err)
	f, validationErrors := flavor.WorkflowTemplate2Flavor(template)
	assert.Nil(t, f, "flavor must be nil, because a validation error occurred")
	assert.Len(t, validationErrors, 3)
	assert.Equal(t, "ignoring a workflow template with a parameter that has no description: pod-security-policy", validationErrors[0].Error())
}

func readWorkflowTemplateFromFixture(path string) (*v1alpha1.WorkflowTemplate, error) {
	template := v1alpha1.WorkflowTemplate{}
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(file, &template)
	if err != nil {
		return nil, err
	}

	return &template, nil
}
