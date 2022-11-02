// Package create implements the infractl create command.
package create

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
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
      derived from the tag of the last commit.`

	nameProvidedToOther = `NOTE: infractl no longer requires a NAME parameter when creating a cluster.
      If ommitted a name will be generated using your infra initials and a 
      short date.`

	mainImageProvidedToQaDemoInStackroxContext = `NOTE: infractl no longer requires a --arg main-image=<image> when creating 
      a qa-demo cluster in a stackrox repo context. An image will be choosen 
      to match the last commit. That commit should be pushed in order to ensure 
      that the image is built. By default opensource (quay.io/stackrox-io) 
      images are used. Pass --rhacs to get RedHat images.`
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
	for _, osArg := range os.Args {
		if strings.Contains(osArg, "qa-demo") {
			cmd.Flags().Bool("rhacs", false, "use RedHat branded images for qa-demo (the default is to use opensource images)")
		}
	}
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

	determineWorkingEnvironment()
	displayUserNotes(cmd, args, &req)

	if len(args) > 1 {
		req.Parameters["name"] = args[1]
	} else {
		name, err := determineName(ctx, conn, args[0])
		if err != nil {
			return nil, err
		}
		req.Parameters["name"] = name
	}

	assignDefaults(cmd, &req)

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

func determineWorkingEnvironment() {
	workingEnvironment.gitTopLevel = ""
	workingEnvironment.tag = ""

	topLevel := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := topLevel.Output()
	if err != nil {
		return
	}
	rootDir := string(out)
	rootDir = strings.TrimSpace(rootDir)
	workingEnvironment.gitTopLevel = rootDir

	makeTag := exec.Command("make", "--quiet", "tag")
	makeTag.Dir = rootDir
	out, err = makeTag.Output()
	if err != nil {
		return
	}
	tag := string(out)
	tag = strings.TrimSpace(tag)
	workingEnvironment.tag = tag
}

func determineName(ctx context.Context, conn *grpc.ClientConn, flavorID string) (string, error) {
	initials, err := getUserInitials(ctx, conn)
	if err != nil {
		return "", err
	}

	suffix := getNameForFlavor(flavorID)
	if suffix == "" {
		suffix = time.Now().Format("01-02")
	}

	unconflicted, err := avoidConflicts(ctx, conn, initials+"-"+suffix)
	if err != nil {
		return "", err
	}

	return unconflicted, nil
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

func getNameForFlavor(flavorID string) string {
	switch flavorID {
	case "qa-demo", "test-qa-demo":
		return getNameForQaDemoFlavor()
	}
	return ""
}

func getNameForQaDemoFlavor() string {
	if !strings.Contains(workingEnvironment.gitTopLevel, "stackrox/stackrox") {
		return ""
	}

	if workingEnvironment.tag == "" {
		return ""
	}

	name := strings.TrimSuffix(workingEnvironment.tag, "-dirty")
	name = strings.ReplaceAll(name, ".", "-")

	return name
}

func avoidConflicts(ctx context.Context, conn *grpc.ClientConn, nameSoFar string) (string, error) {
	req := v1.ClusterListRequest{
		All:     true,
		Expired: true,
	}

	resp, err := v1.NewClusterServiceClient(conn).List(ctx, &req)
	if err != nil {
		return "", err
	}

	for i := 1; i <= 11; i++ {
		potential := nameSoFar + "-" + strconv.Itoa(i)
		inUse := false
		for _, cluster := range resp.Clusters {
			if cluster.ID == potential {
				inUse = true
				break
			}
		}
		if !inUse {
			return potential, nil
		}
	}

	return "", errors.New("could not find a default name for this cluster")
}

func assignDefaults(cmd *cobra.Command, req *v1.CreateClusterRequest) {
	if !strings.Contains(req.ID, "qa-demo") {
		return
	}

	if req.Parameters["main-image"] != "" {
		return
	}

	if !strings.Contains(workingEnvironment.gitTopLevel, "stackrox/stackrox") {
		return
	}

	registry := openSourceRegistry
	if rhacsImages, _ := cmd.Flags().GetBool("rhacs"); rhacsImages {
		registry = rhacsRegistry
	}

	tag := strings.TrimSuffix(workingEnvironment.tag, "-dirty")
	req.Parameters["main-image"] = registry + "/main:" + tag
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

func displayUserNotes(cmd *cobra.Command, args []string, req *v1.CreateClusterRequest) {
	if len(args) >= 2 && args[1] != "" {
		if strings.Contains(args[0], "qa-demo") &&
			strings.Contains(workingEnvironment.gitTopLevel, "stackrox/stackrox") {
			cmd.Println(nameProvidedToQaDemoInStackroxContext)
		} else {
			cmd.Println(nameProvidedToOther)
		}
	}
	if len(args) >= 1 && strings.Contains(args[0], "qa-demo") &&
		strings.Contains(workingEnvironment.gitTopLevel, "stackrox/stackrox") &&
		req.Parameters["main-image"] != "" {
		cmd.Println(mainImageProvidedToQaDemoInStackroxContext)
	}
}
