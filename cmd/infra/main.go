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
	"github.com/stackrox/infra/pkg/buildinfo"
	"github.com/stackrox/infra/server"
	"github.com/stackrox/infra/service"
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
		flagConfig  = flag.String("config", "infra.toml", "path to configuration file")
		flagVersion = flag.Bool("version", false, fmt.Sprintf("print the version %s and exit", buildinfo.Version()))
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

	services := []service.APIService{
		service.NewUserService(),
		service.NewVersionService(),
	}

	// Start the gRPC server.
	grpcShutdown, grpcErr, err := server.RunGRPCServer(services, cfg)
	if err != nil {
		log.Fatalf("Error %v.\n", err)
	}
	defer grpcShutdown()

	// Start the HTTP/gRPC gateway server.
	httpShutdown, httpErr, err := server.RunHTTPServer(services, cfg)
	if err != nil {
		log.Fatalf("Error %v.\n", err)
	}
	defer httpShutdown()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-grpcErr:
		return errors.Wrap(err, "grpc error received")
	case err := <-httpErr:
		return errors.Wrap(err, "http error received")
	case <-sigint:
		return errors.New("signal caught")
	}
}
