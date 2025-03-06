// Package list implements the infractl list command.
package list

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const examples = `# List your clusters.
$ infractl list

# List your clusters, including ones that have expired.
$ infractl list --expired

# List everyone's clusters.
$ infractl list --all

# List clusters whose name matches a prefix.
$ infractl list --prefix=<match>

# List only clusters with specified flavor(s).
$ infractl list --flavor=<flavor> [--flavor=<another>]

$ List only clusters with specified status(es).
$ infractl list --status=<status> [--status=<another>]

# List only the names of clusters
$ infractl list --quiet`

// Command defines the handler for infractl list.
func Command() *cobra.Command {
	// $ infractl list
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List clusters",
		Long:    "List the available clusters",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(0)),
		RunE:    common.WithGRPCHandler(run),
	}

	cmd.Flags().Bool("all", false, "include clusters not owned by you")
	cmd.Flags().Bool("expired", false, "include expired clusters")
	cmd.Flags().BoolP("quiet", "q", false, "only output cluster names")
	cmd.Flags().String("prefix", "", "only include clusters whose names matches this prefix")
	cmd.Flags().StringSlice("flavor", []string{}, "only include clusters with matching flavor(s)")
	cmd.Flags().StringSlice("status", []string{}, "only include clusters with matching status(es)")
	return cmd
}

func run(ctx context.Context, conn *grpc.ClientConn, cmd *cobra.Command, _ []string) (common.PrettyPrinter, error) {
	includeAll := common.MustBool(cmd.Flags(), "all")
	includeExpired := common.MustBool(cmd.Flags(), "expired")
	quietMode := common.MustBool(cmd.Flags(), "quiet")
	prefix, _ := cmd.Flags().GetString("prefix")
	allowedFlavors, _ := cmd.Flags().GetStringSlice("flavor")
	allowedStatuses, _ := cmd.Flags().GetStringSlice("status")

	protoAllowedStatuses := make([]v1.Status, len(allowedStatuses))
	for i, s := range allowedStatuses {
		value, ok := v1.Status_value[s]
		if !ok {
			return nil, fmt.Errorf("unknown cluster status: '%s'", s)
		}
		protoAllowedStatuses[i] = v1.Status(value)
	}

	req := v1.ClusterListRequest{
		All:             includeAll,
		Expired:         includeExpired,
		Prefix:          prefix,
		AllowedFlavors:  allowedFlavors,
		AllowedStatuses: protoAllowedStatuses,
	}

	resp, err := v1.NewClusterServiceClient(conn).List(ctx, &req)
	if err != nil {
		return nil, err
	}

	return prettyClusterListResponse{
		ClusterListResponse: *resp,
		QuietMode:           quietMode,
	}, nil
}
