package cluster

import (
	"context"
	"io"
	"sort"
	"sync"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/golang/protobuf/ptypes"
	v1 "github.com/stackrox/infra/generated/api/v1"
	corev1 "k8s.io/api/core/v1"
)

func (s *clusterImpl) Logs(ctx context.Context, clusterID *v1.ResourceByID) (*v1.LogsResponse, error) {
	workflow, err := s.argoClient.GetMostRecentArgoWorkflowFromClusterID(clusterID.GetId())
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

func (s *clusterImpl) getLogs(ctx context.Context, node v1alpha1.NodeStatus) *v1.Log {
	var body []byte
	started, _ := ptypes.TimestampProto(node.StartedAt.UTC())
	log := &v1.Log{
		Name:    node.DisplayName,
		Body:    body,
		Started: started,
		Message: node.Message,
	}

	stream, err := s.k8sPodsClient.GetLogs(node.ID, &corev1.PodLogOptions{
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
