// Package cluster provides an implementation for the Cluster gRPC service.
package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strings"
	"sync"
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
	"github.com/stackrox/infra/calendar"
	"github.com/stackrox/infra/cmd/infractl/common"
	"github.com/stackrox/infra/flavor"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/service/middleware"
	"github.com/stackrox/infra/signer"
	"github.com/stackrox/infra/slack"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8sv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	// resumeExpiredClusterInterval is how often to periodically check for
	// expired workflows.
	resumeExpiredClusterInterval = 5 * time.Minute

	// calendarCheckInterval is how often to periodically check the calendar
	// for scheduled demos.
	calendarCheckInterval = 5 * time.Minute

	// slackCheckInterval is how often to periodically check for workflow
	// updates to send Slack messages.
	slackCheckInterval = 1 * time.Minute

	workflowDeleteCheckInterval = 5 * time.Minute

	artifactTagURL = "url"

	artifactTagInternal = "internal"
)

type clusterImpl struct {
	clientWorkflows workflowv1.WorkflowInterface
	clientPods      k8sv1.PodInterface
	registry        *flavor.Registry
	signer          *signer.Signer
	eventSource     calendar.EventSource
	slackClient     slack.Slacker
}

var (
	_ middleware.APIService   = (*clusterImpl)(nil)
	_ v1.ClusterServiceServer = (*clusterImpl)(nil)
)

// NewClusterService creates a new ClusterService.
func NewClusterService(registry *flavor.Registry, signer *signer.Signer, eventSource calendar.EventSource, slackClient slack.Slacker) (middleware.APIService, error) {
	clientWorkflows, err := workflowClient()
	if err != nil {
		return nil, err
	}

	clientPods, err := podsClient()
	if err != nil {
		return nil, err
	}

	impl := &clusterImpl{
		clientWorkflows: clientWorkflows,
		clientPods:      clientPods,
		registry:        registry,
		signer:          signer,
		eventSource:     eventSource,
		slackClient:     slackClient,
	}

	go impl.startSlackCheck()
	go impl.cleanupExpiredClusters()
	go impl.startCalendarCheck()
	go impl.startDeleteWorkflowCheck()

	return impl, nil
}

// Info implements ClusterService.Info.
func (s *clusterImpl) Info(ctx context.Context, clusterID *v1.ResourceByID) (*v1.Cluster, error) {
	workflow, err := s.clientWorkflows.Get(clusterID.Id, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	metacluster, err := s.metaClusterFromWorkflow(*workflow)
	if err != nil {
		log.Printf("failed to convert workflow to meta-cluster: %v", err)
		return nil, err
	}

	return &metacluster.Cluster, nil
}

// List implements ClusterService.List.
func (s *clusterImpl) List(ctx context.Context, request *v1.ClusterListRequest) (*v1.ClusterListResponse, error) {
	workflowList, err := s.clientWorkflows.List(metav1.ListOptions{})
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
		metacluster, err := s.metaClusterFromWorkflow(workflow)
		if err != nil {
			log.Printf("failed to convert workflow to meta-cluster: %v", err)
			continue
		}

		// This cluster is expired, and we did not request to include expired
		// clusters.
		if !request.Expired && metacluster.Expired {
			continue
		}

		// This cluster is not ours, and we did not request to include all
		// clusters.
		if !request.All && metacluster.Owner != email {
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
	workflow, err := s.clientWorkflows.Patch(req.Id, types.JSONPatchType, payloadBytes)
	if err != nil {
		return nil, err
	}

	return GetLifespan(workflow), nil
}

// Create implements ClusterService.Create.
func (s *clusterImpl) Create(ctx context.Context, req *v1.CreateClusterRequest) (*v1.ResourceByID, error) {
	// Determine the owner for this cluster, which is derived from information
	// about the current authenticated user stored in the request context.
	var owner string
	if user, found := middleware.UserFromContext(ctx); found {
		owner = user.GetEmail()
	} else if svcacct, found := middleware.ServiceAccountFromContext(ctx); found {
		owner = svcacct.GetEmail()
	} else {
		return nil, errors.New("could not determine owner")
	}

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

	// Determine the lifespan for this cluster. Apply some sanity/bounds
	// checking on provided lifespans.
	lifespan, _ := ptypes.Duration(req.Lifespan)
	if lifespan <= 0 {
		lifespan = 3 * time.Hour
	}
	if lifespan > 12*time.Hour {
		lifespan = 12 * time.Hour
	}

	var slackStatus slack.Status
	if req.NoSlack {
		slackStatus = slack.StatusSkip
	}

	// Set workflow metadata annotations.
	workflow.SetAnnotations(map[string]string{
		annotationDescriptionKey: req.Description,
		annotationEventKey:       eventID,
		annotationFlavorKey:      flav.ID,
		annotationLifespanKey:    fmt.Sprint(lifespan),
		annotationOwnerKey:       owner,
		annotationSlackKey:       string(slackStatus),
	})

	created, err := s.clientWorkflows.Create(&workflow)
	if err != nil {
		return nil, err
	}

	return &v1.ResourceByID{Id: created.Name}, nil
}

// Artifacts implements ClusterService.Artifacts.
func (s *clusterImpl) Artifacts(_ context.Context, clusterID *v1.ResourceByID) (*v1.ClusterArtifacts, error) {
	workflow, err := s.clientWorkflows.Get(clusterID.Id, metav1.GetOptions{})
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
				if artifact.S3 == nil {
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

				url, err := s.signer.Generate(artifact.S3.Bucket, artifact.S3.Key)
				if err != nil {
					return nil, err
				}

				resp.Artifacts = append(resp.Artifacts, &v1.Artifact{
					Name:        artifact.Name,
					Description: description,
					URL:         url,
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
	// Set lifespan to zero.
	lifespanReq := &v1.LifespanRequest{
		Id:       req.Id,
		Lifespan: &duration.Duration{},
	}

	if _, err := s.Lifespan(ctx, lifespanReq); err != nil {
		return nil, err
	}

	// Resume workflow for this cluster.
	if err := util.ResumeWorkflow(s.clientWorkflows, req.Id); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *clusterImpl) Logs(_ context.Context, clusterID *v1.ResourceByID) (*v1.LogsResponse, error) {
	workflow, err := s.clientWorkflows.Get(clusterID.Id, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var podNodes []v1alpha1.NodeStatus
	for _, node := range workflow.Status.Nodes {
		if node.Type == v1alpha1.NodeTypePod && node.Phase != v1alpha1.NodeError {
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

			if log, err := s.getLogs(node); err == nil {
				logChan <- log
			}
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

func (s *clusterImpl) cleanupExpiredClusters() {
	for ; ; time.Sleep(resumeExpiredClusterInterval) {
		workflowList, err := s.clientWorkflows.List(metav1.ListOptions{})
		if err != nil {
			log.Printf("[ERROR] failed to list workflows: %v", err)
			continue
		}

		for _, workflow := range workflowList.Items {
			metacluster, err := s.metaClusterFromWorkflow(workflow)
			if err != nil {
				log.Printf("failed to convert workflow to meta-cluster: %v", err)
				continue
			}

			if metacluster.Status != v1.Status_READY {
				continue
			}

			if !metacluster.Expired {
				continue
			}

			log.Printf("resuming workflow %q", metacluster.ID)
			if err := util.ResumeWorkflow(s.clientWorkflows, metacluster.ID); err != nil {
				log.Printf("[ERROR] failed to resume workflow %q: %v", metacluster.ID, err)
			}
		}
	}
}

func (s *clusterImpl) startCalendarCheck() {
	for ; ; time.Sleep(calendarCheckInterval) {
		// Retrieve upcoming calendar events.
		events, err := s.eventSource.Events()
		if err != nil {
			log.Printf("[ERROR] failed to list calendar events: %v", err)
			continue
		}

		// If there are no events scheduled, then there's nothing to do here.
		if len(events) == 0 {
			continue
		}

		// List out all of the current workflows.
		workflowList, err := s.clientWorkflows.List(metav1.ListOptions{})
		if err != nil {
			log.Printf("[ERROR] failed to list workflows: %v", err)
			continue
		}

		// Build a lookup of current workflow IDs that were launched from
		// calendar events.
		existingWorkflowEventIDs := make(map[string]struct{})
		for _, workflow := range workflowList.Items {
			metacluster, err := s.metaClusterFromWorkflow(workflow)
			if err != nil {
				log.Printf("failed to convert workflow to meta-cluster: %v", err)
				continue
			}

			if metacluster.EventID != "" {
				existingWorkflowEventIDs[metacluster.EventID] = struct{}{}
			}
		}

		for _, event := range events {
			// If there is already a workflow with this event ID
			if _, found := existingWorkflowEventIDs[event.ID]; found {
				log.Printf("[DEBUG] skipping scheduled demo for %q", event.Title)
				continue
			}

			id, err := s.createFromEvent(event)
			if err != nil {
				log.Printf("[ERROR] failed to launch scheduled demo for %q: %v", event.Title, err)
				continue
			} else {
				log.Printf("Launched scheduled demo for %q: %s", event.Title, id.Id)
			}
		}
	}
}

func (s *clusterImpl) createFromEvent(event calendar.Event) (*v1.ResourceByID, error) {
	// Lookup the default flavor.
	defaultFlavorID := s.registry.Default()

	// Set lifespan to range from right now, until 1 hour after the event ends.
	lifespan := time.Until(event.End.Add(time.Hour))

	// Build cluster creation request.
	req := &v1.CreateClusterRequest{
		ID:       defaultFlavorID,
		Lifespan: ptypes.DurationProto(lifespan),
		Parameters: map[string]string{
			"name": simpleName(event.Title),
		},
		Description: event.Title,
	}

	return s.create(req, event.Email, event.ID)
}

func (s *clusterImpl) getLogs(node v1alpha1.NodeStatus) (*v1.Log, error) {
	stream, err := s.clientPods.GetLogs(node.ID, &corev1.PodLogOptions{
		Container:  "main",
		Follow:     false,
		Timestamps: true,
	}).Stream()
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(stream)
	if err != nil {
		return nil, err
	}

	started, _ := ptypes.TimestampProto(node.StartedAt.UTC())
	return &v1.Log{
		Name:    node.DisplayName,
		Body:    body,
		Started: started,
	}, nil
}

func (s *clusterImpl) startSlackCheck() {
	for ; ; time.Sleep(slackCheckInterval) {
		workflowList, err := s.clientWorkflows.List(metav1.ListOptions{})
		if err != nil {
			log.Printf("[ERROR] failed to list workflows: %v", err)
			continue
		}

		for _, workflow := range workflowList.Items {
			metacluster, err := s.metaClusterFromWorkflow(workflow)
			if err != nil {
				log.Printf("failed to convert workflow to meta-cluster: %v", err)
				continue
			}

			// Generate a Slack message for our current cluster state.
			data := slackTemplateContext(s.slackClient, metacluster)
			newSlackStatus, message := slack.FormatSlackMessage(metacluster.Status, metacluster.Slack, data)

			// Only bother to send a message if there is one to send.
			if message != nil {
				if err := s.slackClient.PostMessage(message...); err != nil {
					log.Printf("failed to send Slack message: %v", err)
					continue
				}
			}

			// Only bother to update workflow annotation if our phase has
			// transitioned.
			if newSlackStatus != metacluster.Slack {
				// Construct our replacement patch
				payloadBytes, err := formatAnnotationPatch(annotationSlackKey, string(newSlackStatus))
				if err != nil {
					log.Printf("failed to format Slack annotation patch: %v", err)
					continue
				}

				// Submit the patch.
				_, err = s.clientWorkflows.Patch(metacluster.Cluster.ID, types.JSONPatchType, payloadBytes)
				if err != nil {
					log.Printf("failed to patch Slack annotation for cluster %s: %v", metacluster.Cluster.ID, err)
					continue
				}
			}
		}
	}
}

func (s *clusterImpl) startDeleteWorkflowCheck() {
	// workflowMaxAgeGracePeriod is the time duration after a workflow finishes
	// before it becomes eligible for deletion. This is so that cluster
	// logs/artifacts can still be retrieved from dead clusters for forensic
	// purposes.
	//
	// E.g. a workflow must have finished more than 24 hours ago before it is
	// considered for deletion.
	const workflowMaxAgeGracePeriod = 24 * time.Hour

	for ; ; time.Sleep(workflowDeleteCheckInterval) {
		workflowList, err := s.clientWorkflows.List(metav1.ListOptions{})
		if err != nil {
			log.Printf("[ERROR] failed to list workflows: %v", err)
			continue
		}

		for _, workflow := range workflowList.Items {
			metacluster, err := s.metaClusterFromWorkflow(workflow)
			if err != nil {
				log.Printf("failed to convert workflow to meta-cluster: %v", err)
				continue
			}

			// If this workflow is not finished or failed (aka still running)
			// reject it.
			status := metacluster.Status
			if status != v1.Status_FAILED && status != v1.Status_FINISHED {
				continue
			}

			// If this workflow isn't old enough, reject it.
			cutoff := time.Now().Add(-workflowMaxAgeGracePeriod)
			if time.Unix(metacluster.DestroyedOn.Seconds, 0).After(cutoff) {
				continue
			}

			// Perform the actual deletion via the Kube API.
			var gracePeriod int64 = 0
			if err := s.clientWorkflows.Delete(metacluster.ID, &metav1.DeleteOptions{
				GracePeriodSeconds: &gracePeriod,
			}); err != nil {
				log.Printf("[ERROR] failed to delete workflow %q: %v", metacluster.ID, err)
			} else {
				log.Printf("[INFO] deleted workflow %q", metacluster.ID)
			}
		}
	}
}

func slackTemplateContext(client slack.Slacker, cluster *metaCluster) slack.TemplateData {
	createdOn, _ := ptypes.Timestamp(cluster.CreatedOn)
	lifespan, _ := ptypes.Duration(cluster.Lifespan)
	remaining := time.Until(createdOn.Add(lifespan))

	data := slack.TemplateData{
		Description: cluster.Description,
		Flavor:      cluster.Flavor,
		ID:          cluster.ID,
		OwnerEmail:  cluster.Owner,
		Remaining:   common.FormatExpiration(remaining),
		Scheduled:   cluster.EventID != "",
		URL:         cluster.URL,
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
				return nil, fmt.Errorf("parameter %q was not requested", flavorParamName)
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

		allParams = append(allParams, v1alpha1.Parameter{
			Name:  flavorParamName,
			Value: proto.String(value),
		})
	}

	for requestParamName := range requestParams {
		flavorParam, found := flavorParams[requestParamName]
		if !found || flavorParam.Internal {
			return nil, fmt.Errorf("parameter %q was not requested", requestParamName)
		}
	}

	return allParams, nil
}
