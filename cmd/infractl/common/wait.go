package common

import (
	"errors"
	"fmt"
	"os"
	"time"

	v1 "github.com/stackrox/infra/generated/api/v1"
)

const DefaultMaxConsecutiveWaitErrors = 10

func WaitForCluster(client v1.ClusterServiceClient, clusterID *v1.ResourceByID, maxConsecutiveErrors int) error {
	const timeoutSleep = 30 * time.Second

	nErrors := 0

	fmt.Fprintf(os.Stderr, "...waiting for %s\n", clusterID.Id)
	for {
		ctx, cancel := ContextWithTimeout()
		cluster, err := client.Info(ctx, clusterID)
		cancel()

		if err != nil {
			fmt.Fprintf(os.Stderr, "...error %s\n", err)
			nErrors += 1
			if nErrors >= maxConsecutiveErrors {
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
