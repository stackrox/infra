package common

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	v1 "github.com/stackrox/infra/generated/api/v1"
)

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
