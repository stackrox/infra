package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/stackrox/infra/config"
	"github.com/stackrox/infra/flavor"
	"github.com/stackrox/infra/pkg/buildinfo"
	"github.com/stackrox/infra/server"
	"github.com/stackrox/infra/service"
	"github.com/stackrox/infra/service/cluster"
	"github.com/stackrox/infra/service/middleware"
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
		flagConfig       = flag.String("config", "infra.toml", "path to configuration file")
		flagFlavorConfig = flag.String("flavor-config", "flavors.yaml", "path to flavor configuration file")
		flagVersion      = flag.Bool("version", false, fmt.Sprintf("print the version %s and exit", buildinfo.Version()))
	)
	flag.Parse()

	// If the -version flag was given, print the version and exit.
	if *flagVersion {
		fmt.Println(buildinfo.Version())
		return nil
	}

	log.Printf("Starting infra server version %s", buildinfo.All().Version)

	cfg, err := config.Load(*flagConfig)
	if err != nil {
		return errors.Wrapf(err, "failed to load config file %q", *flagConfig)
	}

	registry, err := flavor.NewFromConfig(*flagFlavorConfig)
	if err != nil {
		return err
	}

	// Construct each individual service.
	services, err := middleware.Services(
		func() (middleware.APIService, error) {
			return service.NewFlavorService(registry)
		},
		service.NewUserService,
		service.NewVersionService,
		func() (middleware.APIService, error) {
			return cluster.NewClusterService(registry)
		},
	)
	if err != nil {
		return err
	}

	srv, err := server.New(*cfg, services...)
	if err != nil {
		return err
	}

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
