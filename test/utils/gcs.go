package utils

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/iterator"
)

const bucketName = "infra-e2e-upload-test"

func CheckGCSObjectExists(ctx context.Context, clusterID string) (bool, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return false, err
	}
	defer client.Close()

	query := &storage.Query{Prefix: clusterID}
	it := client.Bucket(bucketName).Objects(ctx, query)
	_, err = it.Next()
	if err == iterator.Done {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("error finding file for prefix (%s): %v", clusterID, err)
	}

	return true, nil
}

func CheckGCSObjectEventuallyDeleted(ctx context.Context, t *testing.T, clusterID string) {
	tick := 1 * time.Second
	conditionMet := func() bool {
		exists, err := CheckGCSObjectExists(ctx, clusterID)
		if err != nil {
			log.Printf("error when looking for object: %v", err)
			return false
		}

		return !exists
	}

	assert.Eventually(t, conditionMet, defaultTimeout, tick)
}
