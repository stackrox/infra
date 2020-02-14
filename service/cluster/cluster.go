// Package cluster provides an implementation for the Cluster gRPC service.
package cluster

import (
	"context"

	"github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	workflowv1 "github.com/argoproj/argo/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	// "github.com/stackrox/infra/argo"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/service/middleware"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type clusterImpl struct {
	argo workflowv1.WorkflowInterface
}

var (
	_ middleware.APIService   = (*clusterImpl)(nil)
	_ v1.ClusterServiceServer = (*clusterImpl)(nil)
)

// NewClusterService creates a new ClusterService.
func NewClusterService() (middleware.APIService, error) {
	client, err := argoClient()
	if err != nil {
		return nil, err
	}

	return &clusterImpl{
		argo: client,
	}, nil
}

// clusterFromWorkflow converts an Argo workflow into a cluster.
func clusterFromWorkflow(workflow v1alpha1.Workflow) *v1.Cluster {
	cluster := &v1.Cluster{
		ID:       workflow.GetName(),
		Status:   workflowStatus(workflow.Status),
		Flavor:   GetFlavor(&workflow),
		Owner:    GetOwner(&workflow),
		Lifespan: GetLifespan(&workflow),
	}

	cluster.CreatedOn, _ = ptypes.TimestampProto(workflow.Status.StartedAt.Time.UTC())

	if !workflow.Status.FinishedAt.Time.IsZero() {
		cluster.DestroyedOn, _ = ptypes.TimestampProto(workflow.Status.FinishedAt.Time.UTC())
	}

	return cluster
}

// Info implements ClusterService.Info.
func (s *clusterImpl) Info(ctx context.Context, clusterID *v1.ResourceByID) (*v1.Cluster, error) {
	workflow, err := s.argo.Get(clusterID.Id, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return clusterFromWorkflow(*workflow), nil
}

// List implements ClusterService.List.
func (s *clusterImpl) List(ctx context.Context, clusterID *empty.Empty) (*v1.ClusterListResponse, error) {
	workflows, err := s.argo.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	resp := &v1.ClusterListResponse{
		Clusters: make([]*v1.Cluster, len(workflows.Items)),
	}
	for index, workflow := range workflows.Items {
		resp.Clusters[index] = clusterFromWorkflow(workflow)
	}

	return resp, nil
}

// AllowAnonymous declares that this service can be called anonymously.
func (s *clusterImpl) AllowAnonymous() bool {
	return false
}

// RegisterServiceServer registers this service with the given gRPC Server.
func (s *clusterImpl) RegisterServiceServer(server *grpc.Server) {
	v1.RegisterClusterServiceServer(server, s)
}

// RegisterServiceHandler registers this service with the given gRPC Gateway endpoint.
func (s *clusterImpl) RegisterServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return v1.RegisterClusterServiceHandler(ctx, mux, conn)
}
