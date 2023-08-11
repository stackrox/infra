package bqutil

import (
	"context"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/pkg/errors"
	"github.com/stackrox/infra/config"
	"github.com/stackrox/infra/pkg/logging"
	"google.golang.org/api/option"
)

const (
	bigqueryInsertTimeout = 10 * time.Second
)

// BigQueryClient is a BigQuery client that operates on `IngestionRecord`s.
type BigQueryClient interface {
	InsertClusterCreationRecord(ctx context.Context, clusterID, workflowName, flavor, actor string) error
	InsertClusterDeletionRecord(ctx context.Context, clusterID, workflowName string) error
}

var (
	log = logging.CreateProductionLogger()

	_ BigQueryClient = (*enabledClient)(nil)
	_ BigQueryClient = (*disabledClient)(nil)
)

type enabledClient struct {
	environment      string
	creationInserter *bigquery.Inserter
	deletionInserter *bigquery.Inserter
}

type disabledClient struct{}

func (*disabledClient) InsertClusterCreationRecord(_ context.Context, _, _, _, _ string) error {
	return nil
}

func (*disabledClient) InsertClusterDeletionRecord(_ context.Context, _, _ string) error {
	return nil
}

type clusterCreationRecord struct {
	Environment       string
	ClusterID         string
	WorkflowName      string
	Flavor            string
	Actor             string
	CreationTimestamp time.Time
}

type clusterDeletionRecord struct {
	Environment       string
	ClusterID         string
	WorkflowName      string
	DeletionTimestamp time.Time
}

// NewClient returns a new BigQuery client
func NewClient(cfg *config.BigQueryConfig) (BigQueryClient, error) {
	if os.Getenv("TEST_MODE") == "true" {
		log.Log(logging.INFO, "disabling BigQuery integration because we are in TEST_MODE")
		return &disabledClient{}, nil
	}

	// If the config was missing a BigQuery configuration, disable the integration
	// altogether.
	if cfg == nil {
		log.Log(logging.INFO, "disabling BigQuery integration due to missing configuration")
		return &disabledClient{}, nil
	}

	if cfg.CredentialsFile == "" || cfg.Environment == "" || cfg.Project == "" || cfg.Dataset == "" || cfg.CreationTable == "" || cfg.DeletionTable == "" {
		return nil, errors.Errorf("malformed BigQuery config: all of credentialsFile, environment, project, dataset, tables must be defined")
	}

	client, err := bigquery.NewClient(context.Background(), cfg.Project, option.WithCredentialsFile(cfg.CredentialsFile))
	if err != nil {
		return nil, errors.Wrap(err, "creating BigQuery client")
	}

	creationInserter := client.Dataset(cfg.Dataset).Table(cfg.CreationTable).Inserter()
	deletionInserter := client.Dataset(cfg.Dataset).Table(cfg.DeletionTable).Inserter()
	bigQueryClient := &enabledClient{
		environment:      cfg.Environment,
		creationInserter: creationInserter,
		deletionInserter: deletionInserter,
	}

	log.Log(logging.INFO, "enabled BigQuery integration")

	return bigQueryClient, nil
}

// InsertClusterCreationRecord inserts a new cluster creation record into BigQuery.
func (c *enabledClient) InsertClusterCreationRecord(ctx context.Context, clusterID, workflowName, flavor, actor string) error {
	subCtx, cancel := context.WithTimeout(ctx, bigqueryInsertTimeout)
	defer cancel()

	clusterCreationRecord := &clusterCreationRecord{
		Environment:       c.environment,
		ClusterID:         clusterID,
		WorkflowName:      workflowName,
		Flavor:            flavor,
		Actor:             actor,
		CreationTimestamp: time.Now(),
	}

	return c.creationInserter.Put(subCtx, clusterCreationRecord)
}

// InsertClusterDeletionRecord inserts a new cluster deletion record into BigQuery.
func (c *enabledClient) InsertClusterDeletionRecord(ctx context.Context, clusterID, workflowName string) error {
	subCtx, cancel := context.WithTimeout(ctx, bigqueryInsertTimeout)
	defer cancel()

	clusterDeletionRecord := &clusterDeletionRecord{
		Environment:       c.environment,
		ClusterID:         clusterID,
		WorkflowName:      workflowName,
		DeletionTimestamp: time.Now(),
	}

	return c.deletionInserter.Put(subCtx, clusterDeletionRecord)
}
