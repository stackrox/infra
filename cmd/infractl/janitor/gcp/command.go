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

const prefixThreshold = 3

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
$ infractl janitor gcp

# List only instances without matching clusters
$ infractl janitor --quiet`

// Command defines the handler for infractl janitor find.
func Command() *cobra.Command {
	// $ infractl janitor find
	cmd := &cobra.Command{
		Use:     "gcp",
		Short:   "Find orphaned GCP VMs",
		Long:    "Find orphaned GCP compute instances by matching them to running clusters",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(0)),
		RunE:    common.WithGRPCHandler(run),
	}

	return cmd
}

type ComputeInstance struct {
	Name         string
	Status       string
	Zone         string
	Labels       map[string]string
	OriginalName string
}

type CandidateMapping map[*ComputeInstance][]*v1.Cluster

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
		handleQaDemoClusters(i)
		i.Name = strings.ReplaceAll(i.Name, "-", "")
		if !uniqueMap[i.Name] {
			uniqueMap[i.Name] = true
			result = append(result, i)
		}
	}

	sortInstances(result)
	return result
}

// handleQaDemoClusters takes the name of the cluster from the labels instead of the
func handleQaDemoClusters(i *ComputeInstance) {
	if _, ok := i.Labels["name"]; ok {
		i.Name = strings.TrimSuffix(i.Labels["name"], "-prod")
	}
}

func sortInstances(instances []*ComputeInstance) {
	sort.Slice(instances, func(i, j int) bool {
		return instances[i].Name < instances[j].Name
	})
}

type clusterRegistry interface {
	lookup(*ComputeInstance) []string
}

type clusterRegistryImpl struct {
	clusters []*v1.Cluster
}

func findCandidateClustersForInstances(instances []*ComputeInstance, runningClusters []*v1.Cluster) map[*ComputeInstance][]*v1.Cluster {
	result := CandidateMapping{}
	registry := newClusterRegistry(runningClusters)
	for _, vm := range instances {
		result[vm] = registry.lookup(vm)
	}
	return result
}

func newClusterRegistry(clusters []*v1.Cluster) *clusterRegistryImpl {
	return &clusterRegistryImpl{clusters: clusters}
}

// lookup checks all known clusters for ownership of the VM.
func (c *clusterRegistryImpl) lookup(vm *ComputeInstance) []*v1.Cluster {
	out := []*v1.Cluster{}
	for _, cluster := range c.clusters {
		normalizedClusterID := strings.ReplaceAll(cluster.ID, "-", "")
		commonPrefix := findCommonPrefix([2]string{vm.Name, normalizedClusterID})
		if len(commonPrefix) >= prefixThreshold {
			out = append(out, cluster)
		}
	}
	return out
}

func findCommonPrefix(data [2]string) string {
	a, b := data[0], data[1]
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

func filterInstancesWithoutCandidates(clusters CandidateMapping) {
	for instance, candidates := range clusters {
		if len(candidates) > 0 {
			delete(clusters, instance)
		}
	}
}
