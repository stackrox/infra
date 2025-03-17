package flavor_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/stackrox/infra/pkg/flavor"
	"github.com/stretchr/testify/assert"
)

type workflowTemplateTest struct {
	title                               string
	pathToWorkflowTemplate              string
	expectedValidationErrorLen          int
	expectedFirstValidationErrorMessage string
}

func TestWorkflow2TemplateTransformation(t *testing.T) {
	tests := []workflowTemplateTest{
		{
			title:                               "workflow template must have an annotation for the flavor description",
			pathToWorkflowTemplate:              "testdata/missing-flavor-description-in-annotation.yaml",
			expectedValidationErrorLen:          1,
			expectedFirstValidationErrorMessage: "ignoring a workflow template without infra.stackrox.io/description annotation",
		},
		{
			title:                               "workflow template must have an annotation for a valid availability",
			pathToWorkflowTemplate:              "testdata/invalid-availability.yaml",
			expectedValidationErrorLen:          1,
			expectedFirstValidationErrorMessage: "ignoring a workflow template with an unknown infra.stackrox.io/availability annotation",
		},
		{
			title:                               "workflow template parameter must have a description",
			pathToWorkflowTemplate:              "testdata/missing-parameter-description.yaml",
			expectedValidationErrorLen:          3,
			expectedFirstValidationErrorMessage: "ignoring a workflow template with a parameter that has no description: pod-security-policy",
		},
	}

	for index, tc := range tests {
		name := fmt.Sprintf("%d %s", index+1, tc.title)
		t.Run(name, func(t *testing.T) {
			template, err := readWorkflowTemplateFromFixture(tc.pathToWorkflowTemplate)
			assert.NoError(t, err)
			f, validationErrors := flavor.WorkflowTemplate2Flavor(template)
			assert.Nil(t, f, "flavor must be nil, because a validation error occurred")
			assert.Len(t, validationErrors, tc.expectedValidationErrorLen)
			assert.Equal(t, tc.expectedFirstValidationErrorMessage, validationErrors[0].Error())
		})
	}
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
