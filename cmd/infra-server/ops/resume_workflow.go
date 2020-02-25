package ops

import (
	"fmt"
	"time"

	"github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo/workflow/util"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
	"github.com/stackrox/infra/pkg/argoClient"
	"github.com/stackrox/infra/utils/workflowUtils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ResumeWorkflows(workflowNames ...string) error {
	argo := argoClient.NewArgoClient()
	var workflows []*v1alpha1.Workflow
	if len(workflowNames) == 0 {
		workflowList, err := argo.List(metav1.ListOptions{})
		if err != nil {
			fmt.Println(errors.Wrap(err, "Error while listing all workflows"))
			return err
		}

		for idx, _ := range workflowList.Items {
			workflows = append(workflows, &workflowList.Items[idx])
		}
	} else {
		for _, name := range workflowNames {
			workflow, err := argo.Get(name, metav1.GetOptions{})
			if err != nil {
				fmt.Println("Error: %v, while fetching workflow: %s", err, name)
				continue
			}

			workflows = append(workflows, workflow)
		}
	}

	for _, workflow := range workflows {
		if !isWorkflowExpired(workflow) {
			continue
		}

		fmt.Println("Resuming workflow: ", workflow.GetName())
		err := util.ResumeWorkflow(argo, workflow.GetName())
		if err != nil {
			fmt.Println("Error: %v, while resuming workflow: %s", err, workflow.GetName())
		}
	}

	return nil
}

func isWorkflowExpired(workflow *v1alpha1.Workflow) bool {
	lifespan, err := ptypes.Duration(workflowUtils.GetLifespan(workflow))
	if err != nil {
		fmt.Println("Error while determining lifespan of workflow: %v", workflow)
		return false
	}

	workflowExpiryTime := workflow.Status.StartedAt.Time.Add(lifespan)
	if time.Now().After(workflowExpiryTime) {
		return true
	}

	return false
}
