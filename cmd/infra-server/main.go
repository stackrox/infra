package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/stackrox/infra/auth"
	"github.com/stackrox/infra/cmd/infra-server/ops"
	"github.com/stackrox/infra/config"
	"github.com/stackrox/infra/flavor"
	"github.com/stackrox/infra/pkg/buildinfo"
	"github.com/stackrox/infra/server"
	"github.com/stackrox/infra/service"
	"github.com/stackrox/infra/service/cluster"
	"github.com/stackrox/infra/service/middleware"
	"github.com/stackrox/infra/signer"
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
		fmt.Println(buildinfo.Version())
		return nil
	}

	log.Printf("Starting infra server version %s", buildinfo.All().Version)

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

	auth0ConfigFile := filepath.Join(*flagConfigDir, "auth0.yaml")
	auth0PublicKeyPEMFile := filepath.Join(*flagConfigDir, "auth0.pem")
	auth0, err := auth.NewFromConfig(auth0ConfigFile, auth0PublicKeyPEMFile)
	if err != nil {
		return errors.Wrapf(err, "failed to load auth0 config file %q", auth0ConfigFile)
	}

	signer, err := signer.NewFromEnv()
	if err != nil {
		return errors.Wrapf(err, "failed to load GCS signing credentials")
	}

	// Construct each individual service.
	services, err := middleware.Services(
		func() (middleware.APIService, error) {
			return service.NewFlavorService(registry)
		},
		func() (middleware.APIService, error) {
			return service.NewUserService(auth0.GenerateServiceAccountToken)
		},
		service.NewVersionService,
		func() (middleware.APIService, error) {
			return cluster.NewClusterService(registry, signer)
		},
	)
	if err != nil {
		return err
	}

	srv := server.New(*cfg, *auth0, services...)
	errCh, err := srv.RunServer()
	if err != nil {
		return err
	}

	startCronJobs()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return errors.Wrap(err, "server error received")
	case <-sigCh:
		return errors.New("signal caught")
	}
}

func startCronJobs() {
	go func() {
		for {
			err := ops.ResumeWorkflows()
			if err != nil {
				fmt.Println("Error while resuming all workflows", err)
			}

			time.Sleep(5 * time.Minute)
		}
	}()
}
