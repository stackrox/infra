// Package flavor is augmented with available argo.WorkflowTemplates
package flavor

import (
	"context"
	"fmt"
	"strings"
	"time"

	argov3client "github.com/argoproj/argo-workflows/v3/cmd/argo/commands/client"
	workflowtemplatepkg "github.com/argoproj/argo-workflows/v3/pkg/apiclient/workflowtemplate"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/pkg/errors"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/logging"
)

func (r *Registry) initWorkflowTemplatesClient() error {
	ctx, argoClient, err := argov3client.NewAPIClient(context.Background())
	if err != nil {
		return err
	}

	argoWorkflowTemplatesClient, err := argoClient.NewWorkflowTemplateServiceClient()
	if err != nil {
		return err
	}

	r.argoClientCtx = ctx
	r.argoWorkflowTemplatesClient = argoWorkflowTemplatesClient
	r.workflowTemplateNamespace = "default"

	return nil
}

func (r *Registry) addWorkflowTemplates(results []v1.Flavor) []v1.Flavor {
	templates, err := r.argoWorkflowTemplatesClient.ListWorkflowTemplates(r.argoClientCtx, &workflowtemplatepkg.WorkflowTemplateListRequest{
		Namespace: r.workflowTemplateNamespace,
	})
	if err != nil {
		log.Log(logging.ERROR, "failed to list argo workflow templates", "error", err)
		return results
	}

	for i := range templates.Items {
		flavor, errs := WorkflowTemplate2Flavor(&templates.Items[i])
		if len(errs) == 0 {
			results = append(results, *flavor)
		}
	}

	return results
}

func (r *Registry) getPairFromWorkflowTemplate(id string) (*v1.Flavor, *v1alpha1.Workflow) {
	// This short lived cache is useful for performance of list operations when
	// there are large numbers of workflows whose flavor details need to be
	// resolved. The GetWorkflowTemplate() call can be relatively expensive.
	nowTimestamp := time.Now().Unix()
	if r.workflowTemplateCacheTimestamp != nowTimestamp {
		// invalidate the short lived templace cache
		r.workflowTemplateCache = make(map[string]*v1alpha1.WorkflowTemplate)
		r.workflowTemplateCacheTimestamp = nowTimestamp
	}

	if _, found := r.workflowTemplateCache[id]; !found {
		template, err := r.argoWorkflowTemplatesClient.GetWorkflowTemplate(r.argoClientCtx, &workflowtemplatepkg.WorkflowTemplateGetRequest{
			Name:      id,
			Namespace: r.workflowTemplateNamespace,
		})
		if err != nil {
			log.Log(logging.WARN, "failed to get an argo workflow template", "id", id, "error", err)
			return nil, nil
		}
		r.workflowTemplateCache[id] = template
	}
	template := r.workflowTemplateCache[id]

	workflow := &v1alpha1.Workflow{}
	workflow.APIVersion = template.APIVersion
	workflow.Kind = "Workflow"
	workflow.ObjectMeta.GenerateName = template.ObjectMeta.GenerateName
	for _, annotation := range template.ObjectMeta.GetAnnotations() {
		if strings.HasPrefix(annotation, "infra.stackrox.io/") {
			workflow.ObjectMeta.Annotations[annotation] = template.ObjectMeta.Annotations[annotation]
		}
	}
	workflow.Spec = *template.Spec.DeepCopy()

	flavor, _ := WorkflowTemplate2Flavor(template)
	return flavor, workflow
}

func WorkflowTemplate2Flavor(template *v1alpha1.WorkflowTemplate) (*v1.Flavor, []error) {
	var validationErrors []error
	if template.ObjectMeta.Annotations["infra.stackrox.io/description"] == "" {
		err := errors.New("ignoring a workflow template without infra.stackrox.io/description annotation")
		validationErrors = append(validationErrors, err)
		log.Log(logging.WARN, err.Error(),
			"template-name", template.GetObjectMeta().GetName(),
		)
	}
	availability := v1.Flavor_alpha
	if template.ObjectMeta.Annotations["infra.stackrox.io/availability"] != "" {
		value, ok := v1.FlavorAvailability_value[template.ObjectMeta.Annotations["infra.stackrox.io/availability"]]
		if !ok {
			err := errors.New("ignoring a workflow template with an unknown infra.stackrox.io/availability annotation")
			validationErrors = append(validationErrors, err)
			log.Log(logging.WARN, err.Error(),
				"template-name", template.GetObjectMeta().GetName(),
				"template-availability", template.GetObjectMeta().GetAnnotations()["infra.stackrox.io/availability"],
			)
		}
		availability = v1.FlavorAvailability(value)
	}
	for _, wfParameter := range template.Spec.Arguments.Parameters {
		if wfParameter.Description == nil {
			err := fmt.Errorf("ignoring a workflow template with a parameter that has no description: %s", wfParameter.Name)
			validationErrors = append(validationErrors, err)
			log.Log(logging.WARN, err.Error(),
				"template-name", template.GetObjectMeta().GetName(),
				"parameter", wfParameter.Name,
			)
		}
	}
	if len(validationErrors) > 0 {
		return nil, validationErrors
	}

	flavor := &v1.Flavor{
		ID:           template.ObjectMeta.Name,
		Name:         template.ObjectMeta.Name,
		Description:  template.ObjectMeta.Annotations["infra.stackrox.io/description"],
		Availability: availability,
		Parameters:   getParametersFromWorkflowTemplate(template),
		Artifacts:    make(map[string]*v1.FlavorArtifact),
	}

	return flavor, nil
}

func getParametersFromWorkflowTemplate(template *v1alpha1.WorkflowTemplate) map[string]*v1.Parameter {
	parameters := make(map[string]*v1.Parameter)

	for idx, wfParameter := range template.Spec.Arguments.Parameters {
		if wfParameter.Description == nil {
			log.Log(logging.WARN, "ignoring a workflow template with a parameter that has no description",
				"template-name", template.GetObjectMeta().GetName(),
				"parameter", wfParameter.Name,
			)
			continue
		}
		parameter := &v1.Parameter{
			Name:  wfParameter.Name,
			Order: int32(idx + 1),
		}
		parameter.Description = wfParameter.Description.String()
		//nolint:gocritic
		if wfParameter.Default == nil && wfParameter.Value == nil {
			// Required
			parameter.Optional = false
			parameter.Internal = false
			parameter.Value = ""
		} else if wfParameter.Default != nil {
			// Optional
			parameter.Optional = true
			parameter.Internal = false
			parameter.Value = wfParameter.Default.String()
		} else if wfParameter.Default == nil && wfParameter.Value != nil {
			// Hardcoded
			parameter.Optional = true
			parameter.Internal = true
			parameter.Value = wfParameter.Value.String()
		}

		parameters[parameter.Name] = parameter
	}

	return parameters
}
