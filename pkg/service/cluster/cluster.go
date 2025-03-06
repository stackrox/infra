// Package cluster provides an implementation for the Cluster gRPC service.
package cluster

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	workflowv1 "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/argo"
	"github.com/stackrox/infra/pkg/bqutil"
	"github.com/stackrox/infra/pkg/flavor"
	"github.com/stackrox/infra/pkg/kube"
	"github.com/stackrox/infra/pkg/logging"
	"github.com/stackrox/infra/pkg/service/middleware"
	"github.com/stackrox/infra/pkg/signer"
	"github.com/stackrox/infra/pkg/slack"
	"google.golang.org/grpc"
	k8sv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var (
	// resumeExpiredClusterInterval is how often to periodically check for
	// expired workflows.
	// This var is overriden in TEST_MODE.
	resumeExpiredClusterInterval = 1 * time.Minute
)

const (
	// slackCheckInterval is how often to periodically check for workflow
	// updates to send Slack messages.
	slackCheckInterval = 1 * time.Minute

	// default permissions for downloaded artifacts, this corresponds to -rw-r--r--
	artifactDefaultMode = 0o644

	artifactTagURL     = "url"
	artifactTagConnect = "connect"

	artifactTagInternal = "internal"
)

type clusterImpl struct {
	argoClient         argo.ArgoInterface
	bqClient           bqutil.BigQueryClient
	k8sWorkflowsClient workflowv1.WorkflowInterface
	k8sPodsClient      k8sv1.PodInterface
	registry           *flavor.Registry
	signer             *signer.Signer
	slackClient        slack.Slacker
	workflowNamespace  string
}

var (
	log = logging.CreateProductionLogger()

	_ middleware.APIService   = (*clusterImpl)(nil)
	_ v1.ClusterServiceServer = (*clusterImpl)(nil)
)

// NewClusterService creates a new ClusterService.
func NewClusterService(registry *flavor.Registry, signer *signer.Signer, slackClient slack.Slacker, bqClient bqutil.BigQueryClient) (middleware.APIService, error) {
	workflowNamespace := "default"

	k8sWorkflowsClient, err := kube.GetK8sWorkflowsClient(workflowNamespace)
	if err != nil {
		return nil, err
	}

	k8sPodsClient, err := kube.GetK8sPodsClient(workflowNamespace)
	if err != nil {
		return nil, err
	}

	argoClient, err := argo.NewArgoClient(context.Background(), workflowNamespace)
	if err != nil {
		return nil, err
	}

	if os.Getenv("TEST_MODE") == "true" {
		log.Log(logging.INFO, "server is running in test mode")
		resumeExpiredClusterInterval = 5 * time.Second
	}

	impl := &clusterImpl{
		argoClient:         argoClient,
		bqClient:           bqClient,
		k8sWorkflowsClient: k8sWorkflowsClient,
		k8sPodsClient:      k8sPodsClient,
		registry:           registry,
		signer:             signer,
		slackClient:        slackClient,
		workflowNamespace:  workflowNamespace,
	}

	go impl.startSlackCheck()
	go impl.cleanupExpiredClusters()

	return impl, nil
}

// Access configures access for this service.
func (s *clusterImpl) Access() map[string]middleware.Access {
	return map[string]middleware.Access{
		"/v1.ClusterService/Info":      middleware.Authenticated,
		"/v1.ClusterService/List":      middleware.Authenticated,
		"/v1.ClusterService/Lifespan":  middleware.Authenticated,
		"/v1.ClusterService/Create":    middleware.Authenticated,
		"/v1.ClusterService/Artifacts": middleware.Authenticated,
		"/v1.ClusterService/Delete":    middleware.Authenticated,
		"/v1.ClusterService/Logs":      middleware.Authenticated,
	}
}

// RegisterServiceServer registers this service with the given gRPC Server.
func (s *clusterImpl) RegisterServiceServer(server *grpc.Server) {
	v1.RegisterClusterServiceServer(server, s)
}

// RegisterServiceHandler registers this service with the given gRPC Gateway endpoint.
func (s *clusterImpl) RegisterServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return v1.RegisterClusterServiceHandler(ctx, mux, conn)
}

func checkAndEnrichParameters(flavorParams map[string]*v1.Parameter, requestParams map[string]string) ([]v1alpha1.Parameter, error) {
	allParams := make([]v1alpha1.Parameter, 0, len(flavorParams))

	for flavorParamName, flavorParam := range flavorParams {
		requestValue, found := requestParams[flavorParamName]
		var value string

		switch {
		case flavorParam.Internal:
			// Extra sanity check to reject any internal parameters from the
			// user.
			if found {
				return nil, fmt.Errorf("rejecting an internal parameter: %q", flavorParamName)
			}

			// Parameter is internally hardcoded.
			value = flavorParam.Value

		case flavorParam.Optional:
			// Parameter is optional, so fall back to a default if the user
			// hasn't provided a replacement value.
			if !found {
				// use default value.
				value = flavorParam.Value
			} else {
				// Use user-provided value.
				value = requestValue
			}

		default:
			// Parameter is required. The user must provide a value.
			if !found {
				return nil, fmt.Errorf("parameter %q was not provided", flavorParamName)
			}
			value = requestValue
		}

		anyString := v1alpha1.ParseAnyString(value)

		allParams = append(allParams, v1alpha1.Parameter{
			Name:  flavorParamName,
			Value: &anyString,
		})
	}

	for requestParamName := range requestParams {
		flavorParam, found := flavorParams[requestParamName]
		if !found || flavorParam.Internal {
			return nil, fmt.Errorf("passed parameter %q is not defined for this flavor", requestParamName)
		}
	}

	return allParams, nil
}
