package common

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	v1 "github.com/stackrox/infra/generated/api/v1"
)

const (
	flagName                        = "wait-max-errors"
	defaultMaxConsecutiveWaitErrors = 10
)

// AddMaxWaitErrorsFlag adds a flag definition to cmd.
func AddMaxWaitErrorsFlag(cmd *cobra.Command) {
	cmd.Flags().Int(flagName, defaultMaxConsecutiveWaitErrors, "maximum number of consecutive errors before giving up waiting")
}

// GetMaxWaitErrorsFlagValue gets effective value of the flag after arguments are parsed.
func GetMaxWaitErrorsFlagValue(cmd *cobra.Command) int {
	maxWaitErrors, err := cmd.Flags().GetInt(flagName)
	if err != nil {
		panic(err)
	}
	return maxWaitErrors
}

// WaitForCluster waits for a created cluster to be in a ready state.
func WaitForCluster(client v1.ClusterServiceClient, clusterID *v1.ResourceByID, maxWaitErrors int) error {
	const timeoutSleep = 30 * time.Second

	nErrors := 0

	fmt.Fprintf(os.Stderr, "...waiting for %s\n", clusterID.Id)
	for {
		ctx, cancel := ContextWithTimeout()
		cluster, err := client.Info(ctx, clusterID)
		cancel()

		if err != nil {
			fmt.Fprintf(os.Stderr, "...error %s\n", err)
			nErrors++
			if nErrors >= maxWaitErrors {
				return errors.New("too many errors while waiting")
			}
		} else {
			nErrors = 0
			switch cluster.Status {
			case v1.Status_CREATING:
				fmt.Fprintln(os.Stderr, "...creating")
			case v1.Status_READY:
				fmt.Fprintln(os.Stderr, "...ready")
				return nil
			default:
				fmt.Fprintln(os.Stderr, "...failed")
				return errors.New("cluster failed provisioning")
			}
		}

		time.Sleep(timeoutSleep)
	}
}
