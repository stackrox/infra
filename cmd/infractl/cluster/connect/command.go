// Package connect implements the infractl connect command.
package connect

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/cluster/utils"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const examples = `# Connect to the cluster "example-s3maj".
$ infractl connect example-s3maj`

// Command defines the handler for infractl connect.
func Command() *cobra.Command {
	// $ infractl connect
	cmd := &cobra.Command{
		Use:     "connect CLUSTER",
		Short:   "Connect to cluster",
		Long:    "Connect local kubectl to the cluster",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(1), args),
		RunE:    common.WithGRPCHandler(run),
	}
	return cmd
}

func args(_ *cobra.Command, args []string) error {
	if args[0] == "" {
		return errors.New("no cluster ID given")
	}
	return utils.ValidateClusterName(args[0])
}

func run(ctx context.Context, conn *grpc.ClientConn, cmd *cobra.Command, args []string) (common.PrettyPrinter, error) {
	client := v1.NewClusterServiceClient(conn)
	clusterID := args[0]
	if err := validateClusterSupportConnect(ctx, client, clusterID); err != nil {
		return nil, err
	}

	if err := connectToCluster(ctx, client, clusterID); err != nil {
		return nil, err
	}

	return prettyCluster(v1.Cluster{ID: clusterID}), nil
}

func validateClusterSupportConnect(ctx context.Context, client v1.ClusterServiceClient, id string) error {
	// TODO: refactor this check into a field on the flavor
	resp, err := client.Info(ctx, &v1.ResourceByID{Id: id})
	if err != nil {
		return err
	}

	connect := resp.GetConnect()
	if connect == "" {
		return fmt.Errorf("Flavor %s does not support connect.", resp.GetFlavor())
	}
	return nil
}

func connectToCluster(ctx context.Context, client v1.ClusterServiceClient, id string) error {
	resp, err := client.Info(ctx, &v1.ResourceByID{Id: id})
	if err != nil {
		return err
	}

	connect := resp.GetConnect()
	connectWithoutSheBang := strings.ReplaceAll(connect, "#!/bin/sh", "")
	connectWithoutSheBangLineBreaks := strings.ReplaceAll(connectWithoutSheBang, "\n", "")

	commandSplitted := strings.Split(connectWithoutSheBangLineBreaks, " ")

	cmd := exec.Command(commandSplitted[0], commandSplitted[1:]...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	err = cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	outStr, errStr := stdoutBuf.String(), stderrBuf.String()
	fmt.Printf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)

	return nil
}
