// Package create implements the infractl cluster create command.
package create

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const examples = `# Create a new "gke-default" cluster.
$ infractl cluster create gke-default --arg name=test --arg nodes=3

# Create a new "gke-default" cluster with a 30 minute lifespan.
$ infractl cluster create gke-default --lifespan 30m --arg name=test --arg nodes=3`

// Command defines the handler for infractl cluster create.
func Command() *cobra.Command {
	// $ infractl cluster create
	cmd := &cobra.Command{
		Use:     "create FLAVOR",
		Short:   "Create a new cluster",
		Long:    "Creates a new cluster",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(1), args),
		RunE:    common.WithGRPCHandler(run),
	}

	cmd.Flags().StringArray("arg", []string{}, "repeated key=value parameter pairs")
	cmd.Flags().String("description", "", "description for this cluster")
	cmd.Flags().Duration("lifespan", 3*time.Hour, "initial lifespan of the cluster")
	cmd.Flags().Bool("wait", false, "wait for cluster to be ready")
	return cmd
}

func args(_ *cobra.Command, args []string) error {
	if args[0] == "" {
		return errors.New("no flavor ID given")
	}
	return nil
}

func run(ctx context.Context, conn *grpc.ClientConn, cmd *cobra.Command, args []string) (common.PrettyPrinter, error) {
	params, _ := cmd.Flags().GetStringArray("arg")
	description, _ := cmd.Flags().GetString("description")
	lifespan, _ := cmd.Flags().GetDuration("lifespan")
	wait, _ := cmd.Flags().GetBool("wait")
	client := v1.NewClusterServiceClient(conn)

	req := v1.CreateClusterRequest{
		ID:          args[0],
		Parameters:  make(map[string]string),
		Lifespan:    ptypes.DurationProto(lifespan),
		Description: description,
	}

	for _, arg := range params {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 || parts[1] == "" {
			return nil, fmt.Errorf("bad parameter argument %q", arg)
		}
		req.Parameters[parts[0]] = parts[1]
	}

	clusterID, err := client.Create(ctx, &req)
	if err != nil {
		return nil, err
	}

	if wait {
		fmt.Fprintf(os.Stderr, "...creating %s\n", clusterID.Id)
		err := func() error {
			for {
				time.Sleep(30 * time.Second)
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

				cluster, err := client.Info(ctx, clusterID)
				cancel()
				if err != nil {
					fmt.Fprintln(os.Stderr, "...error")
					continue
				}

				switch cluster.Status {
				case v1.Status_CREATING:
					fmt.Fprintln(os.Stderr, "...creating")
					continue
				case v1.Status_READY:
					fmt.Fprintln(os.Stderr, "...ready")
					return nil
				default:
					fmt.Fprintln(os.Stderr, "...failed")
					return errors.New("failed to provision cluster")
				}
			}
		}()
		if err != nil {
			return nil, err
		}
	}

	return prettyResourceByID(*clusterID), nil
}
