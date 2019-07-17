package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/pkg/errors"
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
		flagVersion = flag.Bool("version", false, fmt.Sprintf("print the version %s and exit", buildinfo.Version()))
	)
	flag.Parse()

	// If the -version flag was given, print the version and exit.
	if *flagVersion {
		fmt.Println(buildinfo.Version())
		return nil
	}

	log.Printf("Starting infra server version %s", buildinfo.All().Version)
	services := []service.APIService{
		service.NewVersionService(),
	}

	// Start the gRPC server.
	grpcShutdown, grpcErr, err := server.RunGRPCServer(services, "localhost:9001")
	if err != nil {
		log.Fatalf("Error %v.\n", err)
	}
	defer grpcShutdown()

	// Start the HTTP/gRPC gateway server.
	httpShutdown, httpErr, err := server.RunHTTPServer(services, "localhost:8080", "localhost:9001")
	if err != nil {
		log.Fatalf("Error %v.\n", err)
	}
	defer httpShutdown()

	sigint := make(chan os.Signal)
	signal.Notify(sigint, os.Interrupt, os.Kill)

	select {
	case err := <-grpcErr:
		return errors.Wrap(err, "grpc error received")
	case err := <-httpErr:
		return errors.Wrap(err, "http error received")
	case <-sigint:
		return errors.New("sigint caught")
	}
}
