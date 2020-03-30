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

	"github.com/stackrox/infra/config"

	"github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	workflowv1 "github.com/argoproj/argo/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/argoproj/argo/workflow/util"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"github.com/stackrox/infra/calendar"
	"github.com/stackrox/infra/flavor"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/service/middleware"
	"github.com/stackrox/infra/signer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8sv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	// resumeExpiredWorkflowInterval is how often to periodically check for
	// expired workflows.
	resumeExpiredWorkflowInterval = 5 * time.Minute

	// calendarCheckInterval is how often to periodically check the calendar
	// for scheduled demos.
	calendarCheckInterval = 5 * time.Minute

	// slackCheckInterval is how often to periodically check for workflow
	// updates to send Slack messages.
	slackCheckInterval = 1 * time.Minute
)

type clusterImpl struct {
	clientWorkflows workflowv1.WorkflowInterface
	clientPods      k8sv1.PodInterface
	registry        *flavor.Registry
	signer          *signer.Signer
	eventSource     calendar.EventSource
	slackClient     *slack.Client
	slackChannel    string
}

var (
	_ middleware.APIService   = (*clusterImpl)(nil)
	_ v1.ClusterServiceServer = (*clusterImpl)(nil)
)

// NewClusterService creates a new ClusterService.
func NewClusterService(registry *flavor.Registry, signer *signer.Signer, eventSource calendar.EventSource, slackCfg config.SlackConfig) (middleware.APIService, error) {
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
		slackClient:     slack.New(slackCfg.Token),
		slackChannel:    slackCfg.Channel,
	}

	go impl.startSlackCheck()

	go impl.cleanupExpiredWorkflows()

	go impl.startCalendarCheck()
	return impl, nil
}

// clusterFromWorkflow converts an Argo workflow into a cluster.
func clusterFromWorkflow(workflow v1alpha1.Workflow) *v1.Cluster {
	cluster := &v1.Cluster{
		ID:          workflow.GetName(),
		Status:      workflowStatus(workflow.Status),
		Flavor:      GetFlavor(&workflow),
		Owner:       GetOwner(&workflow),
		Lifespan:    GetLifespan(&workflow),
		Description: GetDescription(&workflow),
	}

	cluster.CreatedOn, _ = ptypes.TimestampProto(workflow.Status.StartedAt.Time.UTC())

	if !workflow.Status.FinishedAt.Time.IsZero() {
		cluster.DestroyedOn, _ = ptypes.TimestampProto(workflow.Status.FinishedAt.Time.UTC())
	}

	return cluster
}

// Info implements ClusterService.Info.
func (s *clusterImpl) Info(ctx context.Context, clusterID *v1.ResourceByID) (*v1.Cluster, error) {
	workflow, err := s.clientWorkflows.Get(clusterID.Id, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return clusterFromWorkflow(*workflow), nil
}

// List implements ClusterService.List.
func (s *clusterImpl) List(ctx context.Context, clusterID *empty.Empty) (*v1.ClusterListResponse, error) {
	workflows, err := s.clientWorkflows.List(metav1.ListOptions{})
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
	workflow, err := s.clientWorkflows.Patch(req.Id, types.JSONPatchType, payloadBytes)
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
		owner = svcacct.GetEmail()
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
		annotationFlavorKey:      flav.ID,
		annotationLifespanKey:    fmt.Sprint(lifespan),
		annotationOwnerKey:       owner,
		annotationDescriptionKey: req.Description,
	})

	workflow.Spec.Arguments.Parameters = make([]v1alpha1.Parameter, 0, len(req.Parameters))
	for paramName, paramValue := range req.Parameters {
		workflow.Spec.Arguments.Parameters = append(workflow.Spec.Arguments.Parameters, v1alpha1.Parameter{
			Name:  paramName,
			Value: proto.String(paramValue),
		})
	}

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

func (s *clusterImpl) cleanupExpiredWorkflows() {
	for ; ; time.Sleep(resumeExpiredWorkflowInterval) {
		workflowList, err := s.clientWorkflows.List(metav1.ListOptions{})
		if err != nil {
			log.Printf("[ERROR] failed to list workflows: %v", err)
			continue
		}

		for idx, workflow := range workflowList.Items {
			if workflowStatus(workflow.Status) != v1.Status_READY {
				continue
			}

			if expired, err := isWorkflowExpired(&workflowList.Items[idx]); err != nil {
				log.Printf("[ERROR] failed to determine expiration of workflow %q: %v", workflow.GetName(), err)
			} else if !expired {
				continue
			}

			log.Printf("resuming workflow %q", workflow.GetName())
			err := util.ResumeWorkflow(s.clientWorkflows, workflow.GetName())
			if err != nil {
				log.Printf("[ERROR] failed to resume workflow %q: %v", workflow.GetName(), err)
			}
		}
	}
}

func (s *clusterImpl) startCalendarCheck() {
	for ; ; time.Sleep(calendarCheckInterval) {
		workflowList, err := s.clientWorkflows.List(metav1.ListOptions{})
		if err != nil {
			log.Printf("[ERROR] failed to list workflows: %v", err)
			continue
		}

		events, err := s.eventSource.Events()
		if err != nil {
			log.Printf("[ERROR] failed to list calendar events: %v", err)
			continue
		}

		existingWorkflowEventIDs := make(map[string]struct{})
		for _, workflow := range workflowList.Items {
			workflow := workflow
			if eventID := GetEventID(&workflow); eventID != "" {
				existingWorkflowEventIDs[eventID] = struct{}{}
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
	flav, workflow := s.registry.Default()

	// Set lifespan to range from right now, until 1 hour after the event ends.
	lifespan := time.Until(event.End.Add(time.Hour))

	// Set some workflow metadata.
	workflow.SetAnnotations(map[string]string{
		annotationEventKey:       event.ID,
		annotationFlavorKey:      flav.ID,
		annotationLifespanKey:    fmt.Sprint(lifespan),
		annotationOwnerKey:       event.Email,
		annotationDescriptionKey: event.Title,
	})

	// Set the only parameter.
	workflow.Spec.Arguments.Parameters = []v1alpha1.Parameter{
		{
			Name:  "name",
			Value: proto.String(simpleName(event.Title)),
		},
	}

	// Launch the demo!
	created, err := s.clientWorkflows.Create(&workflow)
	if err != nil {
		return nil, err
	}

	return &v1.ResourceByID{Id: created.Name}, nil
}

func isWorkflowExpired(workflow *v1alpha1.Workflow) (bool, error) {
	lifespan, err := ptypes.Duration(GetLifespan(workflow))
	if err != nil {
		return false, err
	}

	workflowExpiryTime := workflow.Status.StartedAt.Time.Add(lifespan)
	return time.Now().After(workflowExpiryTime), nil
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
			workflow := workflow
			cluster := clusterFromWorkflow(workflow)
			wfStatus := workflowStatus(workflow.Status)
			slackStatus := slackStatus(GetSlack(&workflow))

			// Generate a Slack message for our current cluster state.
			newSlackStatus, message := formatSlackMessage(cluster, wfStatus, slackStatus)

			// Only bother to send a message if there is one to send.
			if message != nil {
				if _, _, err := s.slackClient.PostMessage(s.slackChannel, message...); err != nil {
					log.Printf("failed to send Slack message: %v", err)
					continue
				}
			}

			// Only bother to update workflow annotation if our phase has
			// transitioned.
			if newSlackStatus != slackStatus {
				// Construct our replacement patch
				payloadBytes, err := formatAnnotationPatch(annotationSlackKey, string(newSlackStatus))
				if err != nil {
					log.Printf("failed to format Slack annotation patch: %v", err)
					continue
				}

				// Submit the patch.
				_, err = s.clientWorkflows.Patch(cluster.ID, types.JSONPatchType, payloadBytes)
				if err != nil {
					log.Printf("failed to patch Slack annotation for cluster %s: %v", cluster.ID, err)
					continue
				}
			}
		}
	}
}
