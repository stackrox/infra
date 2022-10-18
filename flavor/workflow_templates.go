// Package flavor is augmented with available argo.WorkflowTemplates
package flavor

import (
	"context"
	"log"

	argov3client "github.com/argoproj/argo-workflows/v3/cmd/argo/commands/client"
	workflowtemplatepkg "github.com/argoproj/argo-workflows/v3/pkg/apiclient/workflowtemplate"
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

func (r *Registry) appendWorkflowTemplates(results []v1.Flavor) []v1.Flavor {
	templates, err := r.argoWorkflowTemplatesClient.ListWorkflowTemplates(r.argoClientCtx, &workflowtemplatepkg.WorkflowTemplateListRequest{
		Namespace: r.workflowTemplateNamespace,
	})
	if err != nil {
		log.Printf("[ERROR] failed to list argo workflow templates: %v", err)
		return results
	}

	for _, template := range templates.Items {
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
			continue
		}
		flavor := &v1.Flavor{
			ID:           template.ObjectMeta.Name,
			Name:         template.ObjectMeta.Annotations["infra.stackrox.io/name"],
			Description:  template.ObjectMeta.Annotations["infra.stackrox.io/description"],
			Availability: availability,
			Parameters:   make(map[string]*v1.Parameter),
			Artifacts:    make(map[string]*v1.FlavorArtifact),
		}
		results = append(results, *flavor)
	}

	return results
}
