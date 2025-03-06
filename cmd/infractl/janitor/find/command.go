package find

// import (
// 	"context"
// 	"log"
// 	"regexp"

// 	compute "cloud.google.com/go/compute/apiv1"
// 	protobuf "cloud.google.com/go/compute/apiv1/computepb"
// 	"github.com/spf13/cobra"
// 	"github.com/stackrox/infra/cmd/infractl/common"
// 	v1 "github.com/stackrox/infra/generated/api/v1"
// 	"google.golang.org/api/iterator"
// 	"google.golang.org/grpc"
// )

// const examples = `# List your clusters.
// $ infractl list

// # List your clusters, including ones that have expired.
// $ infractl list --expired

// # List everyone's clusters.
// $ infractl list --all

// # List clusters whose name matches a prefix.
// $ infractl list --prefix=<match>

// # List only the names of clusters
// $ infractl list --quiet`

// // Command defines the handler for infractl list.
// func Command() *cobra.Command {
// 	// $ infractl list
// 	cmd := &cobra.Command{
// 		Use:     "find",
// 		Short:   "Find orphaned VMs",
// 		Long:    "Find orphaned VMs",
// 		Example: examples,
// 		Args:    common.ArgsWithHelp(cobra.ExactArgs(0)),
// 		RunE:    common.WithGRPCHandler(run),
// 	}

// 	cmd.Flags().Bool("all", false, "include clusters not owned by you")
// 	cmd.Flags().Bool("expired", false, "include expired clusters")
// 	cmd.Flags().BoolP("quiet", "q", false, "only output cluster names")
// 	cmd.Flags().String("prefix", "", "only include clusters whose names matches this prefix")
// 	return cmd
// }

// type ComputeInstance struct {
// 	Name   string
// 	Status string
// 	Zone   string
// 	Labels map[string]string
// }

// func createClient(ctx context.Context) *compute.InstancesClient {
// 	client, err := compute.NewInstancesRESTClient(ctx)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	return client
// }

// func fetchInstances(ctx context.Context, project string) []*ComputeInstance {
// 	c := createClient(ctx)
// 	defer c.Close()

// 	computeInstances := []*ComputeInstance{}
// 	req := &protobuf.AggregatedListInstancesRequest{
// 		Project: project,
// 	}

// 	it := c.AggregatedList(ctx, req)
// 	for {
// 		resp, err := it.Next()
// 		if err == iterator.Done {
// 			break
// 		}
// 		if err != nil {
// 			log.Fatalf("failed to list instances: %v", err)
// 		}

// 		for _, i := range resp.Value.GetInstances() {
// 			computeInstances = append(computeInstances, &ComputeInstance{
// 				Name:   i.GetName(),
// 				Status: i.GetStatus(),
// 				Zone:   i.GetZone(),
// 				Labels: i.GetLabels(),
// 			})
// 		}
// 	}
// 	return computeInstances
// }

// // formatInstanceNames removes GKE and OCP specific prefix and suffix from compute instance names
// func formatInstanceNames(instances []*ComputeInstance) []*ComputeInstance {
// 	pattern := "gke-|-default-pool.*|-master.*|-worker.*"
// 	re := regexp.MustCompile(pattern)
// 	for _, i := range instances {
// 		i.Name = re.ReplaceAllString(i.Name, "")
// 	}

// 	return instances
// }

// func listGCPInfraClusters(ctx context.Context, conn *grpc.ClientConn) ([]*v1.Cluster, error) {
// 	req := v1.ClusterListRequest{
// 		All: true,
// 	}

// 	resp, err := v1.NewClusterServiceClient(conn).List(ctx, &req)
// 	if err != nil {
// 		return []*v1.Cluster{}, err
// 	}

// 	return resp.Clusters, nil
// }

// func identifyRelevantInfraClusters(clusters []*v1.Cluster) []*v1.Cluster {
// 	validStatuses := map[v1.Status]bool{
// 		v1.Status_CREATING:   true,
// 		v1.Status_READY:      true,
// 		v1.Status_DESTROYING: true,
// 	}

// 	validFlavors := map[string]bool{
// 		"gke-default":            true,
// 		"demo":                   true,
// 		"openshift-4":            true,
// 		"openshift-4-demo":       true,
// 		"osd-on-gcp":             true,
// 		"qa-demo":                true,
// 		"openshift-4-perf-scale": true,
// 		"openshift-multi":        true,
// 	}

// 	interestingClusters := []*v1.Cluster{}
// 	for _, c := range clusters {
// 		if validStatuses[c.Status] && validFlavors[c.Flavor] {
// 			interestingClusters = append(interestingClusters, c)
// 		}
// 	}
// 	return interestingClusters
// }

// func run(ctx context.Context, conn *grpc.ClientConn, _ *cobra.Command, _ []string) (common.PrettyPrinter, error) {

// 	_, _ = listGCPInfraClusters(ctx, conn)
// 	return nil, nil
// }
