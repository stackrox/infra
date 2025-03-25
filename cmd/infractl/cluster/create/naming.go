package create

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/stackrox/infra/cmd/infractl/cluster/utils"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const (
	maxClusterNameLength = 28
	uniqueAttempts       = 100
)

func determineClusterName(ctx context.Context, conn *grpc.ClientConn, cwe *currentWorkingEnvironment, args []string) (string, error) {
	if len(args) > 1 {
		name := args[1]
		err := utils.ValidateClusterName(name)
		if err != nil {
			return "", err
		}
		return name, nil
	}
	name, err := buildUnconflictedDefaultName(ctx, conn, cwe, args[0])
	if err != nil {
		return "", err
	}
	name = shortenName(name)
	return name, nil
}

func buildUnconflictedDefaultName(ctx context.Context, conn *grpc.ClientConn, cwe *currentWorkingEnvironment, flavorID string) (string, error) {
	initials, err := getUserInitials(ctx, conn)
	if err != nil {
		return "", err
	}

	suffix := getNameForFlavor(cwe, flavorID)
	uniqueClusterName, err := getUniqueClusterName(ctx, conn, initials+"-"+suffix)
	if err != nil {
		return "", err
	}

	return uniqueClusterName, nil
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
			initials += strings.ToLower(part[:1])
		}
		if len(initials) < 2 {
			return "", errors.Errorf("Cannot determine a default name for %s", name)
		}
		if len(initials) > 4 {
			initials = initials[:4]
		}
		return initials, nil
	default:
		return "", errors.New("anonymous authentication is not possible in this context")
	}
}

func getNameForFlavor(cwe *currentWorkingEnvironment, flavorID string) string {
	var name string
	if isQaDemoFlavor(flavorID) {
		name = getCurrentTag(cwe)
	}
	if name == "" {
		name = time.Now().Format("01-02")
	}
	return name
}

func getCurrentTag(cwe *currentWorkingEnvironment) string {
	if !cwe.isInStackroxRepo() {
		return ""
	}
	if !cwe.isTagged() {
		return ""
	}

	return getHyphened(getCleaned(cwe.tag))
}

func getUniqueClusterName(ctx context.Context, conn *grpc.ClientConn, prefix string) (string, error) {
	req := v1.ClusterListRequest{
		All:     true,
		Expired: true,
		Prefix:  prefix,
	}

	resp, err := v1.NewClusterServiceClient(conn).List(ctx, &req)
	if err != nil {
		return "", err
	}

	for i := 0; i <= uniqueAttempts; i++ {
		currentName := prefix + "-" + strconv.Itoa(i+1)
		clustersWithPrefix := resp.GetClusters()
		if len(clustersWithPrefix) == 0 {
			return currentName, nil
		}
		var inUse bool
		for _, cluster := range clustersWithPrefix {
			if cluster.ID == currentName {
				inUse = true
				break
			}
		}
		if !inUse {
			return currentName, nil
		}
	}

	return "", fmt.Errorf("could not find a unique cluster name for the prefix '%s'", prefix)
}

// Ensure that the generated name is a maximum of 28 characters long and does not end with hyphen.
func shortenName(name string) string {
	if len(name) > maxClusterNameLength {
		name = name[:28]
		name = strings.TrimSuffix(name, "-")
	}
	return name
}
