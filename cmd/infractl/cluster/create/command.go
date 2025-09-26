// Package create implements the infractl create command.
package create

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/cluster/artifacts"
	"github.com/stackrox/infra/cmd/infractl/cluster/utils"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	examples = `# Create a "gke-default" cluster with default naming.
$ infractl create gke-default
ID: jb-10-21-1

# Create another "gke-default" cluster with an 8 hour lifespan.
$ infractl create gke-default --lifespan 8h
ID: jb-10-21-2

# Create a demo cluster with a name of your own choosing.
$ infractl create qa-demo my-demo-for-me
ID: my-demo-for-me`

	openSourceRegistry = "quay.io/stackrox-io"
	rhacsRegistry      = "quay.io/rhacs-eng"

	nameProvidedToQaDemoInStackroxContext = `NOTE: infractl no longer requires a NAME parameter when creating a cluster.
      qa-demo flavors created from a stackrox repo context will get a name
      derived from the tag of the last commit when the name is not specified.`

	nameProvidedToOther = `NOTE: infractl no longer requires a NAME parameter when creating a cluster.
      If ommitted a name will be generated using your infra user initials, a
      short date and a counter for uniqueness. e.g. jb-10-31-1`

	mainImageProvidedToQaDemoInStackroxContext = `NOTE: infractl no longer requires a --arg main-image=<image> when creating
      a qa-demo cluster in a stackrox repo context. An image will be choosen
      to match the last commit. That commit should be pushed in order to ensure
      that the image is built. By default opensource (quay.io/stackrox-io)
      images are used. Pass --rhacs to get Red Hat images.`
)

// Command defines the handler for infractl create.
func Command() *cobra.Command {
	// $ infractl create
	cmd := &cobra.Command{
		Use:     "create FLAVOR [NAME]",
		Short:   "Create a new cluster",
		Long:    "Creates a new cluster",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.RangeArgs(1, 2)),
		RunE:    common.WithGRPCHandler(run),
	}

	cmd.Flags().StringArray("arg", []string{}, "repeated key=value parameter pairs")
	cmd.Flags().String("description", "", "description for this cluster")
	cmd.Flags().Duration("lifespan", 3*time.Hour, "initial lifespan of the cluster")
	cmd.Flags().Bool("wait", false, "wait for cluster to be ready")
	cmd.Flags().Bool("no-slack", false, "skip sending Slack messages for lifecycle events")
	cmd.Flags().Bool("slack-me", false, "send slack messages directly and not to the #infra_notifications channel")
	cmd.Flags().StringP("download-dir", "d", "", "wait for readiness and download artifacts to this dir")
	cmd.Flags().Bool("rhacs", false, "use Red Hat branded images (only for qa-demo, defaults to open source images)")
	return cmd
}

var workingEnvironment struct {
	gitTopLevel string
	tag         string
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

	if err := utils.ValidateLifespan(lifespan); err != nil {
		return nil, err
	}

	client := v1.NewClusterServiceClient(conn)

	req := v1.CreateClusterRequest{
		ID:          args[0],
		Parameters:  make(map[string]string),
		Lifespan:    durationpb.New(lifespan),
		Description: description,
		NoSlack:     noSlack,
		SlackDM:     slackDM,
	}

	for _, arg := range params {
		parts := strings.SplitN(arg, "=", 2)
		if err := utils.ValidateParameterArgument(parts); err != nil {
			return nil, fmt.Errorf("bad parameter argument %q: %v", arg, err)
		}
		req.Parameters[parts[0]] = parts[1]
	}

	currentWorkingEnvironment := newCurrentWorkingEnvironment()
	displayUserNotes(cmd, args, &req, currentWorkingEnvironment)

	name, err := determineClusterName(ctx, conn, currentWorkingEnvironment, args)
	if err != nil {
		return nil, err
	}
	req.Parameters["name"] = name

	assignDefaults(cmd, &req, currentWorkingEnvironment)

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

	return prettyResourceByID{clusterID}, nil
}

func assignDefaults(cmd *cobra.Command, req *v1.CreateClusterRequest, cwe *currentWorkingEnvironment) {
	if !isQaDemoFlavor(req.GetID()) {
		return
	}

	if req.GetParameters()["main-image"] != "" {
		return
	}

	if !cwe.isInStackroxRepo() {
		return
	}

	registry := openSourceRegistry
	if rhacsImages, _ := cmd.Flags().GetBool("rhacs"); rhacsImages {
		registry = rhacsRegistry
	}

	req.Parameters["main-image"] = registry + "/main:" + getCleaned(cwe.tag)
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

func displayUserNotes(cmd *cobra.Command, args []string, req *v1.CreateClusterRequest, cwe *currentWorkingEnvironment) {
	if wasNameProvided(args) {
		if isQaDemoFlavor(args[0]) && cwe.isInStackroxRepo() {
			cmd.PrintErrln(nameProvidedToQaDemoInStackroxContext)
		} else {
			cmd.PrintErrln(nameProvidedToOther)
		}
	}
	if isQaDemoFlavor(args[0]) && cwe.isInStackroxRepo() && req.Parameters["main-image"] != "" {
		cmd.PrintErrln(mainImageProvidedToQaDemoInStackroxContext)
	}
}
