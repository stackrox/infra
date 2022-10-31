// Package create implements the infractl create command.
package create

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/cluster/artifacts"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const examples = `# Create a new "gke-default" cluster.
$ infractl create gke-default you-20200921-1 --arg nodes=3

# Create another "gke-default" cluster with a 30 minute lifespan.
$ infractl create gke-default you-20200921-2 --lifespan 30m

# Create a demo cluster with a name based on your environment.
# Uses your infra user initials - make tag if present OR MM-DD-N if not.
$ infractl create qa-demo`

// Command defines the handler for infractl create.
func Command() *cobra.Command {
	// $ infractl create
	cmd := &cobra.Command{
		Use:     "create FLAVOR [NAME|<defaults to initials + tag or short date>]",
		Short:   "Create a new cluster",
		Long:    "Creates a new cluster",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.RangeArgs(1, 2), args),
		RunE:    common.WithGRPCHandler(run),
	}

	cmd.Flags().StringArray("arg", []string{}, "repeated key=value parameter pairs")
	cmd.Flags().String("description", "", "description for this cluster")
	cmd.Flags().Duration("lifespan", 3*time.Hour, "initial lifespan of the cluster")
	cmd.Flags().Bool("wait", false, "wait for cluster to be ready")
	cmd.Flags().Bool("no-slack", false, "skip sending Slack messages for lifecycle events")
	cmd.Flags().Bool("slack-me", false, "send slack messages directly and not to the #infra_notifications channel")
	cmd.Flags().StringP("download-dir", "d", "", "wait for readiness and download artifacts to this dir")
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
	noSlack, _ := cmd.Flags().GetBool("no-slack")
	slackDM, _ := cmd.Flags().GetBool("slack-me")
	downloadDir, _ := cmd.Flags().GetString("download-dir")
	if downloadDir != "" {
		wait = true
	}
	client := v1.NewClusterServiceClient(conn)

	req := v1.CreateClusterRequest{
		ID:          args[0],
		Parameters:  make(map[string]string),
		Lifespan:    ptypes.DurationProto(lifespan),
		Description: description,
		NoSlack:     noSlack,
		SlackDM:     slackDM,
	}

	for _, arg := range params {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 || parts[1] == "" {
			return nil, fmt.Errorf("bad parameter argument %q", arg)
		}
		req.Parameters[parts[0]] = parts[1]
	}
	if len(args) > 1 {
		req.Parameters["name"] = args[1]
	} else {
		name, err := determineAName(ctx, conn)
		if err != nil {
			return nil, err
		}
		req.Parameters["name"] = name
	}

	clusterID, err := client.Create(ctx, &req)
	if err != nil {
		return nil, err
	}

	if wait {
		if err := waitForCluster(client, clusterID); err != nil {
			return nil, err
		}
		if downloadDir != "" {
			return artifacts.DownloadArtifacts(context.Background(), client, req.Parameters["name"], downloadDir)
		}
	}

	return prettyResourceByID(*clusterID), nil
}

func determineAName(ctx context.Context, conn *grpc.ClientConn) (string, error) {
	initials, err := getUserInitials(ctx, conn)
	if err != nil {
		return "", err
	}
	return initials, nil
}

func getUserInitials(ctx context.Context, conn *grpc.ClientConn) (string, error) {
	resp, err := v1.NewUserServiceClient(conn).Whoami(ctx, &empty.Empty{})
	if err != nil {
		return "", err
	}
	switch resp := resp.Principal.(type) {
	case *v1.WhoamiResponse_User:
		return "", errors.New("authenticating as a user is not possible in this context")
	case *v1.WhoamiResponse_ServiceAccount:
		initials := ""
		name := resp.ServiceAccount.GetName()
		for _, part := range regexp.MustCompile(`[\s-_\.]+`).Split(name, -1) {
			initials += strings.ToLower(part)[0:1]
		}
		if len(initials) < 2 {
			return "", errors.Errorf("Cannot determine a default name for %s", name)
		}
		if len(initials) > 4 {
			initials = initials[0:4]
		}
		return initials, nil
	case nil:
		return "", errors.New("anonymous authentication is not possible in this context")
	}

	panic("unexpected")
}

func waitForCluster(client v1.ClusterServiceClient, clusterID *v1.ResourceByID) error {
	const timeoutSleep = 30 * time.Second
	const timeoutAPI = 15 * time.Second

	fmt.Fprintf(os.Stderr, "...creating %s\n", clusterID.Id)
	for {
		time.Sleep(timeoutSleep)
		ctx, cancel := context.WithTimeout(context.Background(), timeoutAPI)

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
}
