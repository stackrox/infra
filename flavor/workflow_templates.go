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
		log.Printf("Found workflow template: %v\n", template)
		flavor := &v1.Flavor{
			ID:           template.ObjectMeta.Name,
			Name:         template.ObjectMeta.Annotations["infra.stackrox.io/name"],
			Description:  template.ObjectMeta.Annotations["infra.stackrox.io/description"],
			Availability: v1.Flavor_alpha,
			Parameters:   make(map[string]*v1.Parameter),
			Artifacts:    make(map[string]*v1.FlavorArtifact),
		}
		results = append(results, *flavor)
	}

	return results
}
