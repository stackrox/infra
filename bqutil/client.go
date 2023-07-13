package bqutil

import (
	"context"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/pkg/errors"
	"github.com/stackrox/infra/config"
	"google.golang.org/api/option"
)

const (
	bigqueryInsertTimeout = 10 * time.Second
)

// Client is a BigQuery client that operates on `IngestionRecord`s.
type Client struct {
	creationInserter *bigquery.Inserter
	deletionInserter *bigquery.Inserter
}

type disabledClient struct{}

func (*disabledClient) InsertClusterCreationRecord(ctx context.Context, clusterID, flavor, actor string) error {
	return nil
}

func (*disabledClient) InsertClusterDeletionRecord(ctx context.Context, clusterID string) error {
	return nil
}

type clusterCreationRecord struct {
	ClusterID         string
	Flavor            string
	Actor             string
	CreationTimestamp time.Time
}

type clusterDeletionRecord struct {
	ClusterID         string
	DeletionTimestamp time.Time
}

// NewClient returns a new BigQuery client
func NewClient(cfg config.BigQueryConfig) (*Client, error) {
	if cfg.Project == "" || cfg.Dataset == "" || cfg.CreationTable == "" || cfg.DeletionTable == "" {
		return nil, errors.Errorf("malformed BigQuery config: all of project, dataset, table must be defined")
	}

	client, err := bigquery.NewClient(context.Background(), cfg.Project, option.WithCredentialsJSON())
	if err != nil {
		return nil, errors.Wrap(err, "creating BigQuery client")
	}

	creationInserter := client.Dataset(cfg.Dataset).Table(cfg.CreationTable).Inserter()
	deletionInserter := client.Dataset(cfg.Dataset).Table(cfg.DeletionTable).Inserter()
	return &Client{
		creationInserter: creationInserter,
		deletionInserter: deletionInserter,
	}, nil
}

// InsertClusterCreationRecord inserts a new cluster creation record into BigQuery.
func (c *Client) InsertClusterCreationRecord(ctx context.Context, clusterID, flavor, actor string) error {
	subCtx, cancel := context.WithTimeout(ctx, bigqueryInsertTimeout)
	defer cancel()

	clusterCreationRecord := &clusterCreationRecord{
		ClusterID:         clusterID,
		Flavor:            flavor,
		Actor:             actor,
		CreationTimestamp: time.Now(),
	}

	return c.creationInserter.Put(subCtx, clusterCreationRecord)
}

// InsertClusterDeletionRecord inserts a new cluster deletion record into BigQuery.
func (c *Client) InsertClusterDeletionRecord(ctx context.Context, clusterID string) error {
	subCtx, cancel := context.WithTimeout(ctx, bigqueryInsertTimeout)
	defer cancel()

	clusterDeletionRecord := &clusterDeletionRecord{
		ClusterID:         clusterID,
		DeletionTimestamp: time.Now(),
	}

	return c.creationInserter.Put(subCtx, clusterDeletionRecord)
}
