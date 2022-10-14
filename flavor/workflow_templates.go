// This augments the infra.Flavors registry with available argo.WorkflowTemplates
package flavor

import (
	"github.com/stackrox/infra/pkg/kube"
)

func (r *Registry) initWorkflowTemplatesClient() error {
	workflowTemplateNamespace := "default"

	k8sWorkflowTemplatesClient, err := kube.GetK8sWorkflowTemplatesClient(workflowTemplateNamespace)
	if err != nil {
		return err
	}

	r.k8sWorkflowTemplatesClient = k8sWorkflowTemplatesClient

	return nil
}
