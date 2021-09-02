package common

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"

	"github.com/golang/protobuf/ptypes/empty"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"github.com/stackrox/infra/pkg/buildinfo"
	"google.golang.org/grpc"
)

func checkForVersionDiff(ctx context.Context, conn *grpc.ClientConn, cmd *cobra.Command) {
	if cmd.Use == "version" {
		return
	}

	clientVersion := buildinfo.All()
	serverVersion, _ := v1.NewVersionServiceClient(conn).GetVersion(ctx, &empty.Empty{})

	if clientVersion.Version != serverVersion.Version {
		fmt.Printf("---\ninfractl and infra-server versions are different.\n"+
			"%s -v- %s\n"+
			"You can use `infractl cli upgrade` to update.\n---\n",
			clientVersion.Version, serverVersion.Version)
	}
}
