// Package flavor is augmented with available argo.WorkflowTemplates
package flavor

import (
	"context"
	"log"

	argov3client "github.com/argoproj/argo-workflows/v3/cmd/argo/commands/client"
	workflowtemplatepkg "github.com/argoproj/argo-workflows/v3/pkg/apiclient/workflowtemplate"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	v1 "github.com/stackrox/infra/generated/api/v1"
)

func (r *Registry) initWorkflowTemplatesClient() error {
	ctx, argoClient := argov3client.NewAPIClient(context.Background())

	argoWorkflowTemplatesClient, err := argoClient.NewWorkflowTemplateServiceClient()
	if err != nil {
		return err
	}

	r.argoClientCtx = ctx
	r.argoWorkflowTemplatesClient = argoWorkflowTemplatesClient
	r.workflowTemplateNamespace = "default"

	return nil
}

func (r *Registry) appendFromWorkflowTemplates(results []v1.Flavor) []v1.Flavor {
	templates, err := r.argoWorkflowTemplatesClient.ListWorkflowTemplates(r.argoClientCtx, &workflowtemplatepkg.WorkflowTemplateListRequest{
		Namespace: r.workflowTemplateNamespace,
	})
	if err != nil {
		log.Printf("[ERROR] failed to list argo workflow templates: %v", err)
		return results
	}

	for _, template := range templates.Items {
		flavor := workflowTemplate2Flavor(&template)
		if flavor != nil {
			results = append(results, *flavor)
		}
	}

	return results
}

func (r *Registry) getFromWorkflowTemplate(id string) (*v1.Flavor, *v1alpha1.WorkflowTemplate) {
	template, err := r.argoWorkflowTemplatesClient.GetWorkflowTemplate(r.argoClientCtx, &workflowtemplatepkg.WorkflowTemplateGetRequest{
		Name:      id,
		Namespace: r.workflowTemplateNamespace,
	})
	if err != nil {
		log.Printf("Failed to get an argo workflow template: %s, %v", id, err)
		return nil, nil
	}

	return workflowTemplate2Flavor(template), template
}

func workflowTemplate2Flavor(template *v1alpha1.WorkflowTemplate) *v1.Flavor {
	valid := true
	log.Printf("Found workflow template: %v\n", template.ObjectMeta.Name)
	if template.ObjectMeta.Annotations["infra.stackrox.io/name"] == "" {
		log.Printf("[WARN] Ignoring a workflow template without infra.stackrox.io/name annotation: %v\n", template.ObjectMeta.Name)
		valid = false
	}
	if template.ObjectMeta.Annotations["infra.stackrox.io/description"] == "" {
		log.Printf("[WARN] Ignoring a workflow template without infra.stackrox.io/description annotation: %v\n", template.ObjectMeta.Name)
		valid = false
	}
	availability := v1.Flavor_alpha
	if template.ObjectMeta.Annotations["infra.stackrox.io/availability"] != "" {
		value, ok := v1.FlavorAvailability_value[template.ObjectMeta.Annotations["infra.stackrox.io/availability"]]
		if !ok {
			log.Printf("[WARN] Ignoring a workflow template with an unknown infra.stackrox.io/availability annotation: %v, %v\n",
				template.ObjectMeta.Name, template.ObjectMeta.Annotations["infra.stackrox.io/availability"])
			valid = false
		}
		availability = v1.FlavorAvailability(value)
	}
	if !valid {
		return nil
	}

	parameters := make(map[string]*v1.Parameter)
	for _, wfParameter := range template.Spec.Arguments.Parameters {
		parameter := &v1.Parameter{
			Name: wfParameter.Name,
		}
		if wfParameter.Description != nil {
			parameter.Description = wfParameter.Description.String()
		}
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

	flavor := &v1.Flavor{
		ID:           template.ObjectMeta.Name,
		Name:         template.ObjectMeta.Annotations["infra.stackrox.io/name"],
		Description:  template.ObjectMeta.Annotations["infra.stackrox.io/description"],
		Availability: availability,
		Parameters:   parameters,
		Artifacts:    make(map[string]*v1.FlavorArtifact),
	}

	return flavor
}
