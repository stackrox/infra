// Package cluster provides an implementation for the Cluster gRPC service.
package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	workflowv1 "github.com/argoproj/argo/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/argoproj/argo/workflow/util"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"github.com/stackrox/infra/flavor"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/service/middleware"
	"github.com/stackrox/infra/signer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	resumeExpiredWorkflowInterval = 5 * time.Minute
)

type clusterImpl struct {
	argo     workflowv1.WorkflowInterface
	registry *flavor.Registry
	signer   *signer.Signer
}

var (
	_ middleware.APIService   = (*clusterImpl)(nil)
	_ v1.ClusterServiceServer = (*clusterImpl)(nil)
)

// NewClusterService creates a new ClusterService.
func NewClusterService(registry *flavor.Registry, signer *signer.Signer) (middleware.APIService, error) {
	argo, err := argoClient()
	if err != nil {
		return nil, err
	}

	impl := &clusterImpl{
		argo:     argo,
		registry: registry,
		signer:   signer,
	}

	go impl.cleanupExpiredWorkflows()
	return impl, nil
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

// Create implements ClusterService.Create.
func (s *clusterImpl) Create(ctx context.Context, req *v1.CreateClusterRequest) (*v1.ResourceByID, error) {
	flav, workflow, found := s.registry.Get(req.ID)
	if !found {
		return nil, status.Errorf(codes.NotFound, "flavor %q not found", req.ID)
	}

	if err := flavor.CheckParametersEquivalence(flav, req.Parameters); err != nil {
		return nil, err
	}

	var owner string
	if user, found := middleware.UserFromContext(ctx); found {
		owner = user.GetEmail()
	} else if svcacct, found := middleware.ServiceAccountFromContext(ctx); found {
		owner = svcacct.GetName()
	} else {
		return nil, errors.New("could not determine owner")
	}

	lifespan, _ := ptypes.Duration(req.Lifespan)
	if lifespan <= 0 {
		lifespan = 3 * time.Hour
	}
	if lifespan > 12*time.Hour {
		lifespan = 12 * time.Hour
	}

	workflow.SetAnnotations(map[string]string{
		annotationFlavorKey:   flav.ID,
		annotationLifespanKey: fmt.Sprint(lifespan),
		annotationOwnerKey:    owner,
	})

	workflow.Spec.Arguments.Parameters = make([]v1alpha1.Parameter, 0, len(req.Parameters))
	for paramName, paramValue := range req.Parameters {
		workflow.Spec.Arguments.Parameters = append(workflow.Spec.Arguments.Parameters, v1alpha1.Parameter{
			Name:  paramName,
			Value: proto.String(paramValue),
		})
	}

	created, err := s.argo.Create(&workflow)
	if err != nil {
		return nil, err
	}

	return &v1.ResourceByID{Id: created.Name}, nil
}

// Artifacts implements ClusterService.Artifacts.
func (s *clusterImpl) Artifacts(_ context.Context, clusterID *v1.ResourceByID) (*v1.ClusterArtifacts, error) {
	workflow, err := s.argo.Get(clusterID.Id, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	resp := v1.ClusterArtifacts{}

	for _, nodeStatus := range workflow.Status.Nodes {
		if nodeStatus.Outputs != nil {
			for _, artifact := range nodeStatus.Outputs.Artifacts {
				if artifact.S3 == nil {
					continue
				}

				url, err := s.signer.Generate(artifact.S3.Bucket, artifact.S3.Key)
				if err != nil {
					return nil, err
				}

				resp.Artifacts = append(resp.Artifacts, &v1.Artifact{
					Name: artifact.Name,
					URL:  url,
				})
			}
		}
	}

	return &resp, nil
}

// Access configures access for this service.
func (s *clusterImpl) Access() map[string]middleware.Access {
	return map[string]middleware.Access{
		"/v1.ClusterService/Info":      middleware.Authenticated,
		"/v1.ClusterService/List":      middleware.Authenticated,
		"/v1.ClusterService/Lifespan":  middleware.Authenticated,
		"/v1.ClusterService/Create":    middleware.Authenticated,
		"/v1.ClusterService/Artifacts": middleware.Authenticated,
	}
}

func (s *clusterImpl) Delete(ctx context.Context, req *v1.ResourceByID) (*empty.Empty, error) {
	// Set lifespan to zero.
	lifespanReq := &v1.LifespanRequest{
		Id:       req.Id,
		Lifespan: &duration.Duration{},
	}

	if _, err := s.Lifespan(ctx, lifespanReq); err != nil {
		return nil, err
	}

	// Resume workflow for this cluster.
	if err := util.ResumeWorkflow(s.argo, req.Id); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

// RegisterServiceServer registers this service with the given gRPC Server.
func (s *clusterImpl) RegisterServiceServer(server *grpc.Server) {
	v1.RegisterClusterServiceServer(server, s)
}

// RegisterServiceHandler registers this service with the given gRPC Gateway endpoint.
func (s *clusterImpl) RegisterServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return v1.RegisterClusterServiceHandler(ctx, mux, conn)
}

func (s *clusterImpl) cleanupExpiredWorkflows() {
	for ; ; time.Sleep(resumeExpiredWorkflowInterval) {
		workflowList, err := s.argo.List(metav1.ListOptions{})
		if err != nil {
			log.Printf("failed to list workflows: %v", err)
			continue
		}

		for idx, workflow := range workflowList.Items {
			if workflowStatus(workflow.Status) == v1.Status_FINISHED || !isWorkflowExpired(&workflowList.Items[idx]) {
				continue
			}

			log.Printf("Resuming workflow: %s\n", workflow.GetName())
			err := util.ResumeWorkflow(s.argo, workflow.GetName())
			if err != nil {
				log.Printf("failed to resume workflow %q: %v", workflow.GetName(), err)
			}
		}
	}
}

func isWorkflowExpired(workflow *v1alpha1.Workflow) bool {
	lifespan, err := ptypes.Duration(GetLifespan(workflow))
	if err != nil {
		log.Printf("Error while determining lifespan of workflow: %s\n", workflow.GetName())
		return false
	}

	workflowExpiryTime := workflow.Status.StartedAt.Time.Add(lifespan)
	return time.Now().After(workflowExpiryTime)
}
