package common

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/buildinfo"
	"google.golang.org/grpc"
)

func checkForVersionDiff(ctx context.Context, conn *grpc.ClientConn, cmd *cobra.Command) {
	if cmd.Use == "version" || cmd.Use == "cli" {
		return
	}

	clientVersion := buildinfo.All()
	serverVersion, _ := v1.NewVersionServiceClient(conn).GetVersion(ctx, &empty.Empty{})

	if serverVersion != nil && clientVersion.Version != serverVersion.Version {
		cmd.Printf("---\ninfractl and infra-server versions are different.\n"+
			"%s -v- %s\n"+
			"You can use `infractl cli upgrade` to update.\n---\n",
			clientVersion.Version, serverVersion.Version)
	}
}
