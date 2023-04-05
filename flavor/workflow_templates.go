// Package flavor is augmented with available argo.WorkflowTemplates
package flavor

import (
	"context"
	"log"
	"strings"
	"time"

	argov3client "github.com/argoproj/argo-workflows/v3/cmd/argo/commands/client"
	workflowtemplatepkg "github.com/argoproj/argo-workflows/v3/pkg/apiclient/workflowtemplate"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	v1 "github.com/stackrox/infra/generated/proto/api/v1"
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

func (r *Registry) addWorkflowTemplates(results []v1.Flavor) []v1.Flavor {
	templates, err := r.argoWorkflowTemplatesClient.ListWorkflowTemplates(r.argoClientCtx, &workflowtemplatepkg.WorkflowTemplateListRequest{
		Namespace: r.workflowTemplateNamespace,
	})
	if err != nil {
		log.Printf("[ERROR] failed to list argo workflow templates: %v", err)
		return results
	}

	for i := range templates.Items {
		flavor := workflowTemplate2Flavor(&templates.Items[i])
		if flavor != nil {
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
			log.Printf("[WARN] Failed to get an argo workflow template: %s, %v", id, err)
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

	return workflowTemplate2Flavor(template), workflow
}

func workflowTemplate2Flavor(template *v1alpha1.WorkflowTemplate) *v1.Flavor {
	valid := true
	if template.ObjectMeta.Annotations["infra.stackrox.io/description"] == "" {
		log.Printf("[WARN] Ignoring a workflow template without infra.stackrox.io/description annotation: %s", template.ObjectMeta.Name)
		valid = false
	}
	availability := v1.Flavor_alpha
	if template.ObjectMeta.Annotations["infra.stackrox.io/availability"] != "" {
		value, ok := v1.FlavorAvailability_value[template.ObjectMeta.Annotations["infra.stackrox.io/availability"]]
		if !ok {
			log.Printf("[WARN] Ignoring a workflow template with an unknown infra.stackrox.io/availability annotation: %s, %s",
				template.ObjectMeta.Name, template.ObjectMeta.Annotations["infra.stackrox.io/availability"])
			valid = false
		}
		availability = v1.FlavorAvailability(value)
	}
	for _, wfParameter := range template.Spec.Arguments.Parameters {
		if wfParameter.Description == nil {
			log.Printf("[WARN] Ignoring a workflow template with a parameter (%s) that has no description: %s", wfParameter.Name, template.ObjectMeta.Name)
			valid = false
		}
	}
	if !valid {
		return nil
	}

	flavor := &v1.Flavor{
		ID:           template.ObjectMeta.Name,
		Name:         template.ObjectMeta.Name,
		Description:  template.ObjectMeta.Annotations["infra.stackrox.io/description"],
		Availability: availability,
		Parameters:   getParametersFromWorkflowTemplate(template),
		Artifacts:    make(map[string]*v1.FlavorArtifact),
	}

	return flavor
}

func getParametersFromWorkflowTemplate(template *v1alpha1.WorkflowTemplate) map[string]*v1.Parameter {
	parameters := make(map[string]*v1.Parameter)

	for idx, wfParameter := range template.Spec.Arguments.Parameters {
		if wfParameter.Description == nil {
			log.Printf("[WARN] Ignoring a workflow template with a parameter (%s) that has no description: %s", wfParameter.Name, template.ObjectMeta.Name)
			continue
		}
		parameter := &v1.Parameter{
			Name:  wfParameter.Name,
			Order: int32(idx + 1),
		}
		parameter.Description = wfParameter.Description.String()
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
