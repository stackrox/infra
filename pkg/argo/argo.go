package argo

import (
	"context"

	argov3client "github.com/argoproj/argo-workflows/v3/cmd/argo/commands/client"
	"github.com/argoproj/argo-workflows/v3/pkg/apiclient"
	workflowpkg "github.com/argoproj/argo-workflows/v3/pkg/apiclient/workflow"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/stackrox/infra/pkg/logging"
	"github.com/stackrox/infra/pkg/service/cluster/metadata"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

// ArgoInterface represents a type that can interact with the Argo Workflows API.
type ArgoInterface interface {
	CreateWorkflow(workflow *v1alpha1.Workflow) (*v1alpha1.Workflow, error)
	GetMostRecentArgoWorkflowFromClusterID(clusterID string) (*v1alpha1.Workflow, error)
	ListWorkflows() (*v1alpha1.WorkflowList, error)
	ResumeWorkflow(workflowName string) error
}

var (
	log = logging.CreateProductionLogger()

	_ ArgoInterface = (*argoClientImpl)(nil)
)

type argoClientImpl struct {
	client            apiclient.Client
	workflowsClient   workflowpkg.WorkflowServiceClient
	ctx               context.Context
	workflowNamespace string
}

// NewArgoClient creates a new argo client configured to talk to the workflowNamespace.
func NewArgoClient(ctx context.Context, workflowNamespace string) (*argoClientImpl, error) {
	ctx, client, err := argov3client.NewAPIClient(ctx)
	if err != nil {
		return nil, err
	}
	workflowsClient := client.NewWorkflowServiceClient()

	return &argoClientImpl{
		client:            client,
		workflowsClient:   workflowsClient,
		ctx:               ctx,
		workflowNamespace: workflowNamespace,
	}, nil
}

func (a *argoClientImpl) CreateWorkflow(workflow *v1alpha1.Workflow) (*v1alpha1.Workflow, error) {
	createdWorkflow, err := a.workflowsClient.CreateWorkflow(a.ctx, &workflowpkg.WorkflowCreateRequest{
		Workflow:  workflow,
		Namespace: a.workflowNamespace,
	})
	if err != nil {
		log.Log(logging.WARN, "creating argo workflow for a new cluster failed", "error", err)
		return nil, err
	}
	return createdWorkflow, nil
}

func (a *argoClientImpl) GetMostRecentArgoWorkflowFromClusterID(clusterID string) (*v1alpha1.Workflow, error) {
	workflowList, err := a.workflowsClient.ListWorkflows(a.ctx, &workflowpkg.WorkflowListRequest{
		Namespace:   a.workflowNamespace,
		ListOptions: &metav1.ListOptions{LabelSelector: buildLabelSelector(metadata.LabelClusterID, clusterID)},
	})
	if err != nil {
		log.Log(logging.ERROR, "failed to list workflows", "error", err)
		return nil, err
	}

	if len(workflowList.Items) == 0 {
		log.Log(logging.INFO, "could not find an argo workflow to match infra cluster by label", "cluster-id", clusterID)
		return &v1alpha1.Workflow{}, status.Errorf(
			codes.NotFound,
			"could not find a workflow for the requested infra cluster ID '%s'", clusterID,
		)
	}
	return &workflowList.Items[0], nil
}

func buildLabelSelector(label, value string) string {
	labelSelector := labels.NewSelector()
	req, _ := labels.NewRequirement(label, selection.Equals, []string{value})
	labelSelector = labelSelector.Add(*req)
	return labelSelector.String()
}

func (a *argoClientImpl) ListWorkflows() (*v1alpha1.WorkflowList, error) {
	return a.workflowsClient.ListWorkflows(a.ctx, &workflowpkg.WorkflowListRequest{
		Namespace: a.workflowNamespace,
	})
}

func (a *argoClientImpl) ResumeWorkflow(workflowName string) error {
	log.Log(logging.INFO, "resuming argo workflow", "workflow-name", workflowName)
	_, err := a.workflowsClient.ResumeWorkflow(a.ctx, &workflowpkg.WorkflowResumeRequest{
		Name:      workflowName,
		Namespace: a.workflowNamespace,
	})
	return err
}
