package common

import (
	"context"
	"regexp"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/cobra"
	v1 "github.com/stackrox/infra/generated/proto/api/v1"
	"github.com/stackrox/infra/pkg/buildinfo"
	"google.golang.org/grpc"
)

func getMajorMinorVersion(cmd *cobra.Command, version *v1.Version) string {
	versionStr := version.GetVersion()
	if regexp.MustCompile(`\d+\.\d+\.\d+`).MatchString(versionStr) {
		return strings.Join(strings.Split(versionStr, ".")[:2], ".")
	}
	return ""
}

func versionsDoNotMatch(cmd *cobra.Command, clientVersion *v1.Version, serverVersion *v1.Version) bool {
	return getMajorMinorVersion(cmd, clientVersion) != getMajorMinorVersion(cmd, serverVersion)
}

func areLocalVersions(clientVersion *v1.Version, serverVersion *v1.Version) bool {
	localVersionID := "local"
	return clientVersion.GetVersion() == localVersionID || serverVersion.GetVersion() == localVersionID
}

func checkForVersionDiff(ctx context.Context, conn *grpc.ClientConn, cmd *cobra.Command) {
	if cmd.Use == "version" || cmd.Use == "cli" {
		return
	}

	clientVersion := buildinfo.All()
	serverVersion, _ := v1.NewVersionServiceClient(conn).GetVersion(ctx, &empty.Empty{})

	if serverVersion != nil && !areLocalVersions(clientVersion, serverVersion) && versionsDoNotMatch(cmd, clientVersion, serverVersion) {
		cmd.PrintErrf("---\ninfractl and infra-server versions are different.\n"+
			"%s -v- %s\n"+
			"You can use `infractl cli upgrade` to update.\n---\n",
			clientVersion.Version, serverVersion.Version)
	}
}
