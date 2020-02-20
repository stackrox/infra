package ops

import (
	"fmt"
	"github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo/workflow/util"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
	"github.com/stackrox/infra/pkg/argoClient"
	"github.com/stackrox/infra/service/cluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func ResumeWorkflows(clusterIds ...string) error {
	argo := argoClient.Singleton()
	if len(clusterIds) == 0 {
		workflows, err := argo.List(metav1.ListOptions{})
		if err != nil {
			fmt.Println(errors.Wrap(err, "Error while listing all workflows"))
			return err
		}

		for _, workflow := range workflows.Items {
			// Only resume suspended workflows.
			if workflow.Status.Phase != v1alpha1.NodeRunning {
				continue
			}

			nodeSuspended := true
			for _, node := range workflow.Status.Nodes {
				if node.Type != v1alpha1.NodeTypeSuspend || node.Phase != v1alpha1.NodeRunning {
					nodeSuspended = false
					break
				}
			}

			if !nodeSuspended {
				continue
			}

			// Check expiry time.
			lifespan, err := ptypes.Duration(cluster.GetLifespan(&workflow))
			if err != nil {
				continue
			}

			workflowExpiryTime := workflow.Status.StartedAt.Time.Add(lifespan)
			if time.Now().Before(workflowExpiryTime) {
				continue
			}

			clusterIds = append(clusterIds, workflow.GetName())
		}
	}

	for _, clusterId := range clusterIds {
		fmt.Println("Resuming workflow: ", clusterId)
		err := util.ResumeWorkflow(argo, clusterId)
		if err != nil {
			fmt.Println("Error: %v, while resuming workflow: %s", err, clusterId)
		}
	}

	return nil
}