// Package artifacts implements the infractl artifacts command.
package artifacts

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const examples = `# List the artifacts for cluster "example-s3maj".
$ infractl artifacts example-s3maj

# Download the artifacts for cluster "example-s3maj" into the "artifacts" directory.
$ infractl artifacts example-s3maj -d artifacts`

// Command defines the handler for infractl artifacts.
func Command() *cobra.Command {
	// $ infractl artifacts
	cmd := &cobra.Command{
		Use:     "artifacts CLUSTER",
		Short:   "Download cluster artifacts",
		Long:    "Download artifacts from a cluster",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(1), args),
		RunE:    common.WithGRPCHandler(run),
	}

	cmd.Flags().StringP("download-dir", "d", "", "artifact download directory")
	return cmd
}

func args(_ *cobra.Command, args []string) error {
	if args[0] == "" {
		return errors.New("no cluster ID given")
	}
	return nil
}

func run(ctx context.Context, conn *grpc.ClientConn, cmd *cobra.Command, args []string) (common.PrettyPrinter, error) {
	downloadDir, _ := cmd.Flags().GetString("download-dir")
	client := v1.NewClusterServiceClient(conn)

	return DownloadArtifacts(ctx, client, args[0], downloadDir)
}

// DownloadArtifacts grabs all artifacts
func DownloadArtifacts(ctx context.Context, client v1.ClusterServiceClient, id string, downloadDir string) (common.PrettyPrinter, error) {
	resp, err := client.Artifacts(ctx, &v1.ResourceByID{Id: id})
	if err != nil {
		return nil, err
	}

	// If no --download-dir flag was given, skip downloading the artifacts
	// altogether.
	if downloadDir == "" {
		return prettyClusterArtifacts(*resp), nil
	}

	for _, artifact := range resp.Artifacts {
		filename, err := download(downloadDir, *artifact)
		if err != nil {
			return nil, err
		}
		if strings.HasSuffix(filename, ".tgz") {
			unpackSingleArtifact(filename, downloadDir, *artifact)
		}
	}

	return prettyClusterArtifacts(*resp), nil
}

// download will save the given cluster artifact to disk inside the given
// directory.
func download(downloadDir string, artifact v1.Artifact) (filename string, err error) {
	// Create the download directory if it doesn't exist. All artifacts will be
	// downloaded into this directory.
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return "", err
	}

	// Download the (GCS signed) URL.
	resp, err := http.Get(artifact.URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() //nolint:errcheck

	var artifactName string
	if strings.Contains(resp.Header.Get("Content-Type"), "gzip") {
		artifactName = artifact.Name + ".tgz"
	} else {
		artifactName = artifact.Name
	}

	// Create a new, empty file.
	filename = filepath.Join(downloadDir, artifactName)
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer func() {
		// if the file fails to close, return that error only if there was no
		// original error.
		if ferr := file.Close(); ferr != nil && err == nil {
			err = ferr
		}
	}()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", err
	}

	if err = os.Chmod(filename, fs.FileMode(artifact.Mode)); err != nil {
		return "", err
	}

	return filename, nil
}

// Unpack single file .tgz's. Workflows that specify single file artifacts
// without indicating compression are tar'd and gzip'd. Note: errors are largely
// ignored here as the artifact is already saved.
//nolint:errcheck
func unpackSingleArtifact(tgzFilename string, downloadDir string, artifact v1.Artifact) {
	file, err := os.Open(tgzFilename)
	if err != nil {
		return
	}
	defer file.Close()

	gr, err := gzip.NewReader(file)
	if err != nil {
		return
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	hdr, err := tr.Next()
	if err == io.EOF {
		return // empty .tgz
	}
	if err != nil {
		return
	}

	if hdr.Typeflag != tar.TypeReg {
		// Not a tar with just a single file
		return
	}

	singleFilename := filepath.Join(downloadDir, artifact.Name)
	singleFile, err := os.Create(singleFilename)
	if err != nil {
		return
	}
	if _, err := io.Copy(singleFile, tr); err != nil {
		singleFile.Close()
		os.Remove(singleFilename)
		return
	}
	singleFile.Close()

	_, err = tr.Next()
	if err != nil && err == io.EOF {
		// The tar is a single file and now unpacked to the artifact name
		file.Close()
		os.Remove(tgzFilename)
		return
	}

	// The .tgz is more than just a single artifact
	os.Remove(singleFilename)
}
