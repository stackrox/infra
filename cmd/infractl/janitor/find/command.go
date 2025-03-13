package find

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const commonPrefixThreshold = 3

var (
	infraFlavorsOnGCP = []string{
		"gke-default",
		"demo",
		"openshift-4",
		"openshift-4-demo",
		"osd-on-gcp",
		"qa-demo",
		"openshift-4-perf-scale",
		"openshift-multi",
	}
	relevantStatusesForJanitor = []v1.Status{
		v1.Status_CREATING,
		v1.Status_READY,
		v1.Status_DESTROYING,
	}
)

const examples = `# List GCP compute instances and matching infra clusters.
$ infractl janitor find-gcp

# List only instances without matching clusters
$ infractl janitor find-gcp--quiet`

// Command defines the handler for infractl janitor find.
func Command() *cobra.Command {
	// $ infractl janitor find-gcp
	cmd := &cobra.Command{
		Use:     "find-gcp",
		Short:   "Find orphaned GCP VMs",
		Long:    "Find orphaned GCP compute instances by matching them to running clusters",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(0)),
		RunE:    common.WithGRPCHandler(run),
	}

	return cmd
}

// ComputeInstance represents the type for a GCP compute instance as returned by 'gcloud instances list --json'.
type ComputeInstance struct {
	Name         string
	Status       string
	Zone         string
	Labels       map[string]string
	OriginalName string
}

type candidateMapping map[*ComputeInstance][]*v1.Cluster

func run(ctx context.Context, conn *grpc.ClientConn, _ *cobra.Command, _ []string) (common.PrettyPrinter, error) {
	runningClusters, err := listGCPInfraClusters(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("error listing infra clusters on GCP flavors: %v", err)
	}

	instances, err := readInstances()
	if err != nil {
		return nil, fmt.Errorf("error reading instances: %v", err)
	}

	instances = FormatInstanceNames(instances)

	instancesWithCandidates := findCandidateClustersForInstances(instances, runningClusters)
	filterInstancesWithoutCandidates(instancesWithCandidates)

	return prettyJanitorFindResponse{
		instances: instancesWithCandidates,
	}, nil
}

func listGCPInfraClusters(ctx context.Context, conn *grpc.ClientConn) ([]*v1.Cluster, error) {
	req := v1.ClusterListRequest{
		All:             true,
		AllowedStatuses: relevantStatusesForJanitor,
		AllowedFlavors:  infraFlavorsOnGCP,
	}

	resp, err := v1.NewClusterServiceClient(conn).List(ctx, &req)
	if err != nil {
		return []*v1.Cluster{}, err
	}

	return resp.Clusters, nil
}

func readInstances() ([]*ComputeInstance, error) {
	instances := []*ComputeInstance{}
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&instances); err != nil {
		return nil, fmt.Errorf("error decoding instances: %v", err)
	}
	return instances, nil
}

// FormatInstanceNames removes GKE and OCP specific prefix and suffix from compute instance names
func FormatInstanceNames(instances []*ComputeInstance) []*ComputeInstance {
	result := []*ComputeInstance{}
	uniqueMap := make(map[string]bool)

	pattern := "^gke-|-default-pool.*|-master.*|-worker.*|-bootstrap.*"
	re := regexp.MustCompile(pattern)
	for _, i := range instances {
		i.OriginalName = i.Name
		i.Name = re.ReplaceAllString(i.Name, "")
		handleDemoClusters(i)
		i.Name = strings.ReplaceAll(i.Name, "-", "")
		if !uniqueMap[i.Name] {
			uniqueMap[i.Name] = true
			result = append(result, i)
		}
	}

	sortInstances(result)
	return result
}

// handleDemoClusters takes the name of the cluster from the labels instead of the instance itself.
func handleDemoClusters(i *ComputeInstance) {
	if _, ok := i.Labels["name"]; ok {
		i.Name = strings.TrimSuffix(i.Labels["name"], "-prod")
	}
}

func sortInstances(instances []*ComputeInstance) {
	sort.Slice(instances, func(i, j int) bool {
		return instances[i].Name < instances[j].Name
	})
}

func findCandidateClustersForInstances(instances []*ComputeInstance, runningClusters []*v1.Cluster) map[*ComputeInstance][]*v1.Cluster {
	result := candidateMapping{}
	for _, vm := range instances {
		result[vm] = listMatchingClustersForInstance(vm, runningClusters)
	}
	return result
}

// listMatchingClustersForInstance returns a list of clusters that have the same prefix as the instance.
func listMatchingClustersForInstance(vm *ComputeInstance, clusters []*v1.Cluster) []*v1.Cluster {
	out := []*v1.Cluster{}
	for _, cluster := range clusters {
		normalizedClusterID := strings.ReplaceAll(cluster.ID, "-", "")
		commonPrefix := findCommonPrefix(vm.Name, normalizedClusterID)
		if len(commonPrefix) >= commonPrefixThreshold {
			out = append(out, cluster)
		}
	}
	return out
}

func findCommonPrefix(a, b string) string {
	shorterStringLen := compareLen(a, b)
	i := 0
	for i < shorterStringLen && a[i] == b[i] {
		i++
	}
	return a[:i]
}

func compareLen(a, b string) int {
	if len(a) < len(b) {
		return len(a)
	}
	return len(b)
}

func filterInstancesWithoutCandidates(clusters candidateMapping) {
	for instance, candidates := range clusters {
		if len(candidates) > 0 {
			delete(clusters, instance)
		}
	}
}
