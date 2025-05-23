// Package cluster provides an implementation for the Cluster gRPC service.
package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	argov3client "github.com/argoproj/argo-workflows/v3/cmd/argo/commands/client"
	"github.com/argoproj/argo-workflows/v3/pkg/apiclient"
	workflowpkg "github.com/argoproj/argo-workflows/v3/pkg/apiclient/workflow"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	workflowv1 "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/bqutil"
	"github.com/stackrox/infra/pkg/flavor"
	"github.com/stackrox/infra/pkg/kube"
	"github.com/stackrox/infra/pkg/logging"
	"github.com/stackrox/infra/pkg/service/metrics"
	"github.com/stackrox/infra/pkg/service/middleware"
	"github.com/stackrox/infra/pkg/signer"
	"github.com/stackrox/infra/pkg/slack"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	k8sv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var (
	// resumeExpiredClusterInterval is how often to periodically check for
	// expired workflows.
	resumeExpiredClusterInterval = 1 * time.Minute
)

const (
	// slackCheckInterval is how often to periodically check for workflow
	// updates to send Slack messages.
	slackCheckInterval = 1 * time.Minute

	// when is a cluster considered near expiration
	nearExpiry = 30 * time.Minute

	// warn when loops take too long
	loopDurationWarning = 5 * time.Second

	// default permissions for downloaded artifacts, this corresponds to -rw-r--r--
	artifactDefaultMode = 0o644

	artifactTagURL     = "url"
	artifactTagConnect = "connect"

	artifactTagInternal = "internal"
)

type clusterImpl struct {
	k8sWorkflowsClient  workflowv1.WorkflowInterface
	k8sPodsClient       k8sv1.PodInterface
	registry            *flavor.Registry
	signer              *signer.Signer
	slackClient         slack.Slacker
	argoClient          apiclient.Client
	argoWorkflowsClient workflowpkg.WorkflowServiceClient
	argoClientCtx       context.Context
	workflowNamespace   string
	bqClient            bqutil.BigQueryClient
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

	ctx, argoClient, err := argov3client.NewAPIClient(context.Background())
	if err != nil {
		return nil, err
	}
	argoWorkflowsClient := argoClient.NewWorkflowServiceClient()

	if os.Getenv("TEST_MODE") == "true" {
		log.Log(logging.INFO, "server is running in test mode")
		resumeExpiredClusterInterval = 5 * time.Second
	}

	impl := &clusterImpl{
		k8sWorkflowsClient:  k8sWorkflowsClient,
		k8sPodsClient:       k8sPodsClient,
		registry:            registry,
		signer:              signer,
		slackClient:         slackClient,
		argoClient:          argoClient,
		argoWorkflowsClient: argoWorkflowsClient,
		argoClientCtx:       ctx,
		workflowNamespace:   workflowNamespace,
		bqClient:            bqClient,
	}

	go impl.startSlackCheck()
	go impl.cleanupExpiredClusters()

	return impl, nil
}

// Info implements ClusterService.Info.
func (s *clusterImpl) Info(_ context.Context, clusterID *v1.ResourceByID) (*v1.Cluster, error) {
	workflow, err := s.getMostRecentArgoWorkflowFromClusterID(clusterID.GetId())
	if err != nil {
		return nil, err
	}

	metacluster, err := s.metaClusterFromWorkflow(*workflow)
	if err != nil {
		log.Log(logging.ERROR, "failed to convert argo workflow to infra meta-cluster", "workflow-name", workflow.GetName(), "error", err)
		return nil, err
	}

	return &metacluster.Cluster, nil
}

// List implements ClusterService.List.
func (s *clusterImpl) List(ctx context.Context, request *v1.ClusterListRequest) (*v1.ClusterListResponse, error) {
	workflowList, err := s.argoWorkflowsClient.ListWorkflows(s.argoClientCtx, &workflowpkg.WorkflowListRequest{
		Namespace: s.workflowNamespace,
	})
	if err != nil {
		return nil, err
	}

	// Obtain the email of the current principal.
	var email string
	if user, found := middleware.UserFromContext(ctx); found {
		email = user.Email
	} else if svcacct, found := middleware.ServiceAccountFromContext(ctx); found {
		email = svcacct.Email
	}

	clusters := make([]*v1.Cluster, 0, len(workflowList.Items))

	// Loop over all of the workflows, and keep only the ones that match our
	// request criteria.
	for _, workflow := range workflowList.Items {
		// This cluster is expired, and we did not request to include expired
		// clusters.
		if !request.Expired && isWorkflowExpired(workflow) {
			continue
		}

		// TODO(perf): move this to a listOption for the WorkflowListRequest to do the selection on K8s side
		// This cluster is not ours, and we did not request to include all clusters.
		if !request.All && GetOwner(&workflow) != email {
			continue
		}

		if request.Prefix != "" && !strings.HasPrefix(getClusterIDFromWorkflow(&workflow), request.Prefix) {
			continue
		}

		// If a filter for allowed flavors is active and the cluster is not one of the allowed, skip.
		if len(request.AllowedFlavors) > 0 && !isClusterOneOfAllowedFlavors(&workflow, request.AllowedFlavors) {
			continue
		}

		// If a status filter exists and the cluster is not one of the allowed, skip.
		if len(request.AllowedStatuses) > 0 && !isClusterOneOfAllowedStatuses(&workflow, request.AllowedStatuses) {
			continue
		}

		metacluster, err := s.metaClusterFromWorkflow(workflow)
		if err != nil {
			log.Log(logging.ERROR, "failed to convert argo workflow to infra meta-cluster", "workflow-name", workflow.GetName(), "error", err)
			continue
		}

		// This cluster wasn't rejected, so we'll keep it for the response.
		clusters = append(clusters, &metacluster.Cluster)
	}

	resp := &v1.ClusterListResponse{
		Clusters: clusters,
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
func (s *clusterImpl) Lifespan(ctx context.Context, req *v1.LifespanRequest) (*duration.Duration, error) {
	owner, err := middleware.GetOwnerFromContext(ctx)
	if err != nil {
		return nil, err
	}
	log.AuditLog(logging.INFO, "cluster-lifespan", "received a lifespan update request for infra cluster",
		"actor", owner,
		"cluster-id", req.GetId(),
		"lifespan-update-method", req.GetMethod().String(),
		"lifespan", req.GetLifespan().String(),
	)

	workflow, err := s.getMostRecentArgoWorkflowFromClusterID(req.GetId())
	if err != nil {
		return nil, err
	}

	return s.lifespan(ctx, req, workflow)
}

func (s *clusterImpl) lifespan(ctx context.Context, req *v1.LifespanRequest, workflow *v1alpha1.Workflow) (*duration.Duration, error) {
	log.Log(logging.INFO, "will apply a lifespan update to argo workflow",
		"workflow-name", workflow.GetName(),
		"lifespan-update-method", req.GetMethod().String(),
		"lifespan", req.GetLifespan().String(),
	)

	lifespanRequest := req.Lifespan.AsDuration()
	lifespanCurrent := time.Duration(0)
	lifespanUpdated := time.Duration(0)

	// If we're applying a relative lifespan (by adding or subtracting), we
	// need to know the current lifespan. Get the named workflow to obtain said
	// current lifespan.
	if req.Method != v1.LifespanRequest_REPLACE {
		lifespanCurrent = GetLifespan(workflow).AsDuration()
	}

	// Compute the updated lifespan using the requested method.
	switch req.Method {
	case v1.LifespanRequest_REPLACE:
		lifespanUpdated = lifespanRequest
	case v1.LifespanRequest_ADD:
		lifespanUpdated = lifespanCurrent + lifespanRequest
	case v1.LifespanRequest_SUBTRACT:
		lifespanUpdated = lifespanCurrent - lifespanRequest
	}

	// Sanity check that our updated lifespan doesn't go negative.
	if lifespanUpdated <= 0 {
		lifespanUpdated = 0
	}

	// Construct our replacement patch
	payloadBytes, err := formatAnnotationPatch(annotationLifespanKey, fmt.Sprint(lifespanUpdated))
	if err != nil {
		return nil, err
	}

	// Submit the patch.
	_, err = s.k8sWorkflowsClient.Patch(ctx, workflow.GetName(), types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
	if err != nil {
		log.Log(logging.ERROR, "error occurred updating the argo workflow", "workflow-name", workflow.GetName(), "error", err)
		return nil, err
	}

	// Return the remaining lifespan.
	remaining := time.Until(workflow.CreationTimestamp.Add(lifespanUpdated))
	return durationpb.New(remaining), nil
}

// Create implements ClusterService.Create.
func (s *clusterImpl) Create(ctx context.Context, req *v1.CreateClusterRequest) (*v1.ResourceByID, error) {
	owner, err := middleware.GetOwnerFromContext(ctx)
	if err != nil {
		return nil, err
	}

	log.AuditLog(logging.INFO, "cluster-create", "received a create request for flavor",
		"actor", owner,
		"flavor-id", req.GetID(),
	)
	return s.create(req, owner, "")
}

func (s *clusterImpl) create(req *v1.CreateClusterRequest, owner, eventID string) (*v1.ResourceByID, error) {
	flav, workflow, found := s.registry.Get(req.ID)
	if !found {
		return nil, status.Errorf(codes.NotFound, "flavor %q not found", req.ID)
	}

	// Combine any hardcoded or default workflow parameters with the user
	// provided parameters. Or return an error if the user provided
	// insufficient or superfluous parameters.
	workflowParams, err := checkAndEnrichParameters(flav.Parameters, req.Parameters)
	if err != nil {
		return nil, err
	}
	workflow.Spec.Arguments.Parameters = workflowParams

	// Use the user supplied name as the root of Argo workflow name and the Infra cluster Id.
	clusterID, ok := req.Parameters["name"]
	if ok {
		workflow.GenerateName = clusterID + "-"
	} else {
		return nil, fmt.Errorf("parameter 'name' was not provided")
	}

	// Make sure there is no running argo workflow for infra cluster with the same ID
	existingWorkflow, _ := s.getMostRecentArgoWorkflowFromClusterID(clusterID)
	if existingWorkflow != nil {
		switch workflowStatus(existingWorkflow.Status) {
		case v1.Status_FAILED, v1.Status_FINISHED:
			// It is ok to reuse a cluster ID from a failed or finished workflow.
			log.Log(logging.INFO, "a completed argo workflow exists",
				"workflow-name", existingWorkflow.GetName(),
				"cluster-id", clusterID,
				"workflow-phase", existingWorkflow.Status.Phase,
			)

		default:
			log.Log(logging.WARN, "infra cluster create failed due to an existing busy argo workflow",
				"workflow-name", existingWorkflow.GetName(),
				"cluster-id", clusterID,
				"workflow-phase", existingWorkflow.Status.Phase,
			)
			return nil, status.Errorf(
				codes.AlreadyExists,
				"An infra cluster ID %q already exists in state %s.",
				clusterID, workflowStatus(existingWorkflow.Status).String(),
			)
		}
	}

	// Determine the lifespan for this cluster. Apply some sanity/bounds
	// checking on provided lifespans.
	lifespan := req.Lifespan.AsDuration()
	if lifespan <= 0 {
		lifespan = 3 * time.Hour
	}

	var slackStatus slack.Status
	if req.NoSlack {
		slackStatus = slack.StatusSkip
	}

	slackDM := "no"
	if req.SlackDM {
		slackDM = "yes"
	}

	// Set workflow metadata annotations.
	workflow.SetAnnotations(map[string]string{
		annotationDescriptionKey: req.Description,
		annotationEventKey:       eventID,
		annotationFlavorKey:      flav.ID,
		annotationLifespanKey:    fmt.Sprint(lifespan),
		annotationOwnerKey:       owner,
		annotationSlackKey:       string(slackStatus),
		annotationSlackDMKey:     slackDM,
	})

	workflow.SetLabels(map[string]string{
		labelClusterID: clusterID,
	})

	log.Log(logging.INFO, "will create an infra cluster",
		"flavor-id", flav.GetID(),
		"cluster-id", clusterID,
		"cluster-owner", owner,
	)

	created, err := s.argoWorkflowsClient.CreateWorkflow(s.argoClientCtx, &workflowpkg.WorkflowCreateRequest{
		Workflow:  &workflow,
		Namespace: s.workflowNamespace,
	})
	if err != nil {
		log.Log(logging.WARN, "creating argo workflow for a new cluster failed", "error", err)
		return nil, err
	}

	log.Log(logging.INFO, "created an argo workflow for a new infra cluster",
		"workflow-name", created.GetName(),
		"cluster-id", clusterID,
	)

	metrics.FlavorsUsedCounter.WithLabelValues(flav.ID).Inc()

	err = s.bqClient.InsertClusterCreationRecord(context.Background(), clusterID, created.GetName(), flav.GetID(), owner)
	if err != nil {
		log.Log(logging.WARN, "failed to record cluster creation", "cluster-id", clusterID, "error", err)
	}

	return &v1.ResourceByID{Id: clusterID}, nil
}

// Artifacts implements ClusterService.Artifacts.
func (s *clusterImpl) Artifacts(_ context.Context, clusterID *v1.ResourceByID) (*v1.ClusterArtifacts, error) {
	workflow, err := s.getMostRecentArgoWorkflowFromClusterID(clusterID.GetId())
	if err != nil {
		return nil, err
	}

	flavorMetadata := make(map[string]*v1.FlavorArtifact)
	flavorName := GetFlavor(workflow)
	flavor, _, found := s.registry.Get(flavorName)
	if found && flavor.Artifacts != nil {
		flavorMetadata = flavor.Artifacts
	}

	resp := v1.ClusterArtifacts{}

	for _, nodeStatus := range workflow.Status.Nodes {
		if nodeStatus.Outputs != nil {
			for _, artifact := range nodeStatus.Outputs.Artifacts {
				if artifact.GCS == nil {
					continue
				}

				var description string

				meta, found := flavorMetadata[artifact.Name]
				if found {
					if _, isInternal := meta.Tags[artifactTagInternal]; isInternal {
						continue
					}

					description = meta.Description
				}

				bucket, key := handleArtifactMigration(*workflow, artifact)
				if bucket == "" || key == "" {
					continue
				}

				url, err := s.signer.Generate(bucket, key)
				if err != nil {
					return nil, err
				}

				var mode int32 = artifactDefaultMode
				if artifact.Mode != nil {
					mode = *artifact.Mode
				}

				resp.Artifacts = append(resp.Artifacts, &v1.Artifact{
					Name:        artifact.Name,
					Description: description,
					URL:         url,
					Mode:        mode,
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
		"/v1.ClusterService/Delete":    middleware.Authenticated,
		"/v1.ClusterService/Logs":      middleware.Authenticated,
	}
}

func (s *clusterImpl) Delete(ctx context.Context, req *v1.ResourceByID) (*empty.Empty, error) {
	owner, err := middleware.GetOwnerFromContext(ctx)
	if err != nil {
		return nil, err
	}
	log.AuditLog(logging.INFO, "cluster-delete", "received a delete request for infra cluster",
		"actor", owner,
		"cluster-id", req.GetId(),
	)

	workflow, err := s.getMostRecentArgoWorkflowFromClusterID(req.GetId())
	if err != nil {
		return &empty.Empty{}, err
	}

	// Set lifespan to zero so the workflow is examined in cleanupExpiredClusters().
	lifespanReq := &v1.LifespanRequest{
		Id:       req.Id,
		Lifespan: &duration.Duration{},
		Method:   v1.LifespanRequest_REPLACE,
	}

	if _, err := s.lifespan(ctx, lifespanReq, workflow); err != nil {
		log.Log(logging.ERROR, "failed to set lifespan to 0 for argo workflow",
			"workflow-name", workflow.GetName(),
			"error", err,
		)
		return nil, err
	}

	log.Log(logging.INFO, "resuming argo workflow", "workflow-name", workflow.GetName())

	// Resume the workflow so that it may move to the destroy phase without
	// waiting for cleanupExpiredClusters() to kick in.
	_, err = s.argoWorkflowsClient.ResumeWorkflow(s.argoClientCtx, &workflowpkg.WorkflowResumeRequest{
		Name:      workflow.GetName(),
		Namespace: s.workflowNamespace,
	})
	if err != nil {
		log.Log(logging.WARN, "failed to resume workflow, this is OK if the workflow is not waiting",
			"cluster-id", req.GetId(),
			"workflow-name", workflow.GetName(),
			"error", err,
		)
	}

	err = s.bqClient.InsertClusterDeletionRecord(context.Background(), req.GetId(), workflow.GetName())
	if err != nil {
		log.Log(logging.WARN, "failed to record cluster deletion", "cluster-id", req.GetId(), "error", err)
	}

	return &empty.Empty{}, nil
}

func (s *clusterImpl) Logs(ctx context.Context, clusterID *v1.ResourceByID) (*v1.LogsResponse, error) {
	workflow, err := s.getMostRecentArgoWorkflowFromClusterID(clusterID.GetId())
	if err != nil {
		return nil, err
	}

	var podNodes []v1alpha1.NodeStatus
	for _, node := range workflow.Status.Nodes {
		if node.Type == v1alpha1.NodeTypePod {
			podNodes = append(podNodes, node)
		}
	}

	// Fetch logs for each individual pod.
	var wg sync.WaitGroup
	logChan := make(chan *v1.Log)
	for _, node := range podNodes {
		wg.Add(1)
		go func(node v1alpha1.NodeStatus) {
			defer wg.Done()

			logChan <- s.getLogs(ctx, node)
		}(node)
	}

	// Close the channel when all goroutines are done.
	go func() {
		wg.Wait()
		close(logChan)
	}()

	// Consume all logs from the channel.
	logs := make([]*v1.Log, 0, len(podNodes))
	for log := range logChan {
		logs = append(logs, log)
	}

	// Sort the logs by when they started.
	sort.SliceStable(logs, func(i, j int) bool {
		return logs[i].Started.GetSeconds() < logs[j].Started.GetSeconds()
	})

	return &v1.LogsResponse{Logs: logs}, nil
}

// RegisterServiceServer registers this service with the given gRPC Server.
func (s *clusterImpl) RegisterServiceServer(server *grpc.Server) {
	v1.RegisterClusterServiceServer(server, s)
}

// RegisterServiceHandler registers this service with the given gRPC Gateway endpoint.
func (s *clusterImpl) RegisterServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return v1.RegisterClusterServiceHandler(ctx, mux, conn)
}

func (s *clusterImpl) getMostRecentArgoWorkflowFromClusterID(clusterID string) (*v1alpha1.Workflow, error) {
	listOpts := &metav1.ListOptions{}
	labelSelector := labels.NewSelector()
	req, _ := labels.NewRequirement(labelClusterID, selection.Equals, []string{clusterID})
	labelSelector = labelSelector.Add(*req)
	listOpts.LabelSelector = labelSelector.String()

	workflowList, err := s.argoWorkflowsClient.ListWorkflows(s.argoClientCtx, &workflowpkg.WorkflowListRequest{
		Namespace:   s.workflowNamespace,
		ListOptions: listOpts,
	})
	if err != nil {
		log.Log(logging.ERROR, "failed to list workflows", "error", err)
		return nil, err
	}
	if len(workflowList.Items) >= 1 {
		// Current behaviour - the cluster ID exists as a workflow label
		return &workflowList.Items[0], nil
	}

	log.Log(logging.INFO, "could not find an argo workflow to match infra cluster by label", "cluster-id", clusterID)

	// TODO: is this path ever executed?
	// Prior behaviour - Try to find using the cluster ID mapped to the workflow name
	return s.argoWorkflowsClient.GetWorkflow(s.argoClientCtx, &workflowpkg.WorkflowGetRequest{
		Name:      clusterID,
		Namespace: s.workflowNamespace,
	})
}

func (s *clusterImpl) cleanupExpiredClusters() {
	for ; ; time.Sleep(resumeExpiredClusterInterval) {
		start := time.Now()

		workflowList, err := s.argoWorkflowsClient.ListWorkflows(s.argoClientCtx, &workflowpkg.WorkflowListRequest{
			Namespace: s.workflowNamespace,
		})
		if err != nil {
			log.Log(logging.ERROR, "failed to list workflows", "error", err)
			continue
		}

		for _, workflow := range workflowList.Items {
			if workflowStatus(workflow.Status) != v1.Status_READY {
				continue
			}

			if !isWorkflowExpired(workflow) {
				continue
			}

			log.Log(logging.INFO, "resuming an argo workflow that has expired", "workflow-name", workflow.GetName())
			_, err = s.argoWorkflowsClient.ResumeWorkflow(s.argoClientCtx, &workflowpkg.WorkflowResumeRequest{
				Name:      workflow.GetName(),
				Namespace: s.workflowNamespace,
			})
			if err != nil {
				log.Log(logging.WARN, "failed to resume argo workflow", "workflow-name", workflow.GetName(), "error", err)
			}

			clusterID := strings.TrimSuffix(workflow.GenerateName, "-")
			err = s.bqClient.InsertClusterDeletionRecord(context.Background(), clusterID, workflow.GetName())
			if err != nil {
				log.Log(logging.WARN, "failed to record cluster deletion", "workflow-name", workflow.GetName(), "error", err)
			}
		}

		if time.Since(start) > loopDurationWarning {
			// TODO: why are we logging this?
			log.Log(logging.WARN, fmt.Sprintf("expire loop took %s", time.Since(start).String()))
		}
	}
}

func (s *clusterImpl) getLogs(ctx context.Context, node v1alpha1.NodeStatus) *v1.Log {
	var body []byte
	started := timestamppb.New(node.StartedAt.UTC())
	log := &v1.Log{
		Name:    node.DisplayName,
		Body:    body,
		Started: started,
		Message: node.Message,
	}

	podName := determinePodName(node)
	stream, err := s.k8sPodsClient.GetLogs(podName, &corev1.PodLogOptions{
		Container:  "main",
		Follow:     false,
		Timestamps: true,
	}).Stream(ctx)
	if err != nil {
		log.Body = []byte(err.Error())
		return log
	}

	logBody, err := io.ReadAll(stream)
	if err != nil {
		log.Body = []byte(err.Error())
		return log
	}
	log.Body = logBody

	return log
}

func determinePodName(node v1alpha1.NodeStatus) string {
	parts := strings.Split(node.ID, "-")
	baseName := strings.Join(parts[:len(parts)-1], "-")
	randomNumber := parts[len(parts)-1]
	return fmt.Sprintf("%s-%s-%s", baseName, node.TemplateName, randomNumber)
}

func (s *clusterImpl) startSlackCheck() {
	for ; ; time.Sleep(slackCheckInterval) {
		start := time.Now()

		workflowList, err := s.argoWorkflowsClient.ListWorkflows(s.argoClientCtx, &workflowpkg.WorkflowListRequest{
			Namespace: s.workflowNamespace,
		})
		if err != nil {
			log.Log(logging.ERROR, "failed to list workflows", "error", err)
			continue
		}

		for _, workflow := range workflowList.Items {
			s.slackCheckWorkflow(workflow)
		}

		if time.Since(start) > loopDurationWarning {
			// TODO: why are we logging this?
			log.Log(logging.WARN, fmt.Sprintf("slack loop took %s", time.Since(start).String()))
		}
	}
}

func (s *clusterImpl) slackCheckWorkflow(workflow v1alpha1.Workflow) {
	if slack.IsSlackComplete(slack.Status(GetSlack(&workflow))) {
		return
	}

	metacluster, err := s.metaClusterFromWorkflow(workflow)
	if err != nil {
		log.Log(logging.ERROR, "failed to convert workflow to meta-cluster", "workflow-name", workflow.Name, "error", err)
		return
	}

	// Generate a Slack message for our current cluster state.
	failureDetails := workflowFailureDetails(workflow.Status).Error()
	data := slackTemplateContext(s.slackClient, metacluster, failureDetails)
	newSlackStatus, message := slack.FormatSlackMessage(metacluster.Status, metacluster.NearingExpiry, metacluster.Slack, data)

	// Only bother to send a message if there is one to send.
	if message != nil {
		sent := false
		user, found := s.slackClient.LookupUser(metacluster.Owner)
		if found && metacluster.SlackDM {
			if err := s.slackClient.PostMessageToUser(user, message...); err != nil {
				log.Log(logging.ERROR, "failed to send Slack message directly to user", "user-email", user.Profile.Email, "error", err)
			} else {
				sent = true
			}
		}
		if !sent {
			if err := s.slackClient.PostMessage(message...); err != nil {
				log.Log(logging.ERROR, "failed to send Slack message", "error", err)
				return
			}
		}

		if metacluster.Status == v1.Status_FAILED {
			clusterID := getClusterIDFromWorkflow(&workflow)
			err = s.bqClient.InsertClusterDeletionRecord(context.Background(), clusterID, workflow.GetName())
			if err != nil {
				log.Log(logging.WARN, "failed to record cluster deletion", "cluster-id", clusterID, "error", err)
			}
		}
	}

	// Only bother to update workflow annotation if our phase has
	// transitioned.
	if newSlackStatus != metacluster.Slack {
		// Construct our replacement patch
		payloadBytes, err := formatAnnotationPatch(annotationSlackKey, string(newSlackStatus))
		if err != nil {
			log.Log(logging.ERROR, "failed to format Slack annotation patch", "error", err)
			return
		}

		// Submit the patch.
		_, err = s.k8sWorkflowsClient.Patch(context.Background(), workflow.GetName(), types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
		if err != nil {
			log.Log(logging.ERROR, "failed to patch Slack annotation",
				"cluster-id", metacluster.ID,
				"workflow-name", workflow.GetName(),
				"error", err,
			)
			return
		}
	}
}

func slackTemplateContext(client slack.Slacker, cluster *metaCluster, failureDetails string) slack.TemplateData {
	createdOn := cluster.CreatedOn.AsTime()
	lifespan := cluster.Lifespan.AsDuration()
	remaining := time.Until(createdOn.Add(lifespan))

	data := slack.TemplateData{
		Description:    cluster.Description,
		Flavor:         cluster.Flavor,
		ID:             cluster.ID,
		OwnerEmail:     cluster.Owner,
		Remaining:      common.FormatExpiration(remaining),
		Scheduled:      cluster.EventID != "",
		URL:            cluster.URL,
		FailureDetails: failureDetails,
	}

	if user, found := client.LookupUser(cluster.Owner); found {
		data.OwnerID = user.ID
	}

	return data
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
