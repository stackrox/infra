package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
	"github.com/stackrox/infra/auth"
	"github.com/stackrox/infra/bqutil"
	"github.com/stackrox/infra/config"
	"github.com/stackrox/infra/flavor"
	"github.com/stackrox/infra/pkg/buildinfo"
	"github.com/stackrox/infra/pkg/logging"
	"github.com/stackrox/infra/server"
	"github.com/stackrox/infra/service"
	"github.com/stackrox/infra/service/cluster"
	"github.com/stackrox/infra/service/middleware"
	"github.com/stackrox/infra/signer"
	"github.com/stackrox/infra/slack"
)

// main is the entry point of the infra server.
func main() {
	if err := mainCmd(); err != nil {
		log.Fatalf("infra: %v", err)
	}
}

// mainCmd composes all the components together and can return an error for
// convenience.
func mainCmd() error {
	var (
		flagConfigDir = flag.String("config-dir", "configuration", "path to configuration dir")
		flagVersion   = flag.Bool("version", false, fmt.Sprintf("print the version %s and exit", buildinfo.Version()))
	)
	flag.Parse()

	// If the -version flag was given, print the version and exit.
	if *flagVersion {
		fmt.Println(buildinfo.Version()) //nolint:forbidigo
		return nil
	}

	log := logging.CreateProductionLogger()
	log.Infow("starting infra server", "version", buildinfo.All().Version)

	serverConfigFile := filepath.Join(*flagConfigDir, "infra.yaml")
	cfg, err := config.Load(serverConfigFile)
	if err != nil {
		return errors.Wrapf(err, "failed to load server config file %q", serverConfigFile)
	}

	flavorConfigFile := filepath.Join(*flagConfigDir, "flavors.yaml")
	registry, err := flavor.NewFromConfig(flavorConfigFile)
	if err != nil {
		return errors.Wrapf(err, "failed to load flavor config file %q", flavorConfigFile)
	}

	oidcConfigFile := filepath.Join(*flagConfigDir, "oidc.yaml")
	oidc, err := auth.NewFromConfig(oidcConfigFile)
	if err != nil {
		return errors.Wrapf(err, "failed to load oidc config file %q", oidcConfigFile)
	}

	signer, err := signer.NewFromEnv()
	if err != nil {
		return errors.Wrapf(err, "failed to load GCS signing credentials")
	}

	slackClient, err := slack.New(cfg.Slack)
	if err != nil {
		return errors.Wrapf(err, "failed to create Slack client")
	}

	bqClient, err := bqutil.NewClient(*cfg.BigQuery)
	if err != nil {
		return errors.Wrapf(err, "failed to create bqClient")
	}

	// Construct each individual service.
	services, err := middleware.Services(
		func() (middleware.APIService, error) {
			return service.NewFlavorService(registry)
		},
		func() (middleware.APIService, error) {
			return service.NewUserService(oidc.GenerateServiceAccountToken)
		},
		service.NewCliService,
		service.NewStatusService,
		service.NewVersionService,
		func() (middleware.APIService, error) {
			return cluster.NewClusterService(registry, signer, slackClient, bqClient)
		},
	)
	if err != nil {
		return err
	}

	srv := server.New(*cfg, *oidc, services...)
	errCh, err := srv.RunServer()
	if err != nil {
		return err
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return errors.Wrap(err, "server error received")
	case <-sigCh:
		return errors.New("signal caught")
	}
}
