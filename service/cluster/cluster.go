// Package cluster provides an implementation for the Cluster gRPC service.
package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	workflowv1 "github.com/argoproj/argo/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	// "github.com/stackrox/infra/argo"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/service/middleware"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

// formatAnnotationPatch generates a raw patch for updating the given annotation.
func formatAnnotationPatch(annotationKey string, annotationValue string) ([]byte, error) {
	// The annotation key needs to be escaped, since it may contain '/'
	// characters, which already have meaning in the path spec. See
	// https://tools.ietf.org/html/rfc6901#section-3 for more details.
	//
	// Because the characters '~' (%x7E) and '/' (%x2F) have special
	// meanings in JSON Pointer, '~' needs to be encoded as '~0' and '/'
	// needs to be encoded as '~1' when these characters appear in a
	// reference token.
	annotationKey = strings.ReplaceAll(annotationKey, "~", "~0")
	annotationKey = strings.ReplaceAll(annotationKey, "/", "~1")
	path := "/metadata/annotations/" + annotationKey

	//  patch specifies a patch operation for a string.
	payload := []struct {
		Op    string `json:"op"`
		Path  string `json:"path"`
		Value string `json:"value"`
	}{{
		Op:    "replace",
		Path:  path,
		Value: annotationValue,
	}}

	return json.Marshal(payload)
}

// Lifespan implements ClusterService.Lifespan.
func (s *clusterImpl) Lifespan(_ context.Context, req *v1.LifespanRequest) (*duration.Duration, error) {
	lifespan, _ := ptypes.Duration(req.Lifespan)
	// Sanity check that our lifespan doesn't go negative.
	if lifespan <= 0 {
		lifespan = 0
	}

	// Construct our replacement patch
	payloadBytes, err := formatAnnotationPatch(annotationLifespanKey, fmt.Sprint(lifespan))
	if err != nil {
		return nil, err
	}

	// Submit the patch.
	workflow, err := s.argo.Patch(req.Id, types.JSONPatchType, payloadBytes)
	if err != nil {
		return nil, err
	}

	return GetLifespan(workflow), nil
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
