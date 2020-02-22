// Package artifacts implements the infractl cluster artifacts command.
package artifacts

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const examples = `# List the artifacts for cluster "example-s3maj".
$ infractl cluster artifacts example-s3maj

# Download the artifacts for cluster "example-s3maj" into the "artifacts" directory.
$ infractl cluster artifacts example-s3maj --download-dir=artifacts
`

// Command defines the handler for infractl cluster artifacts.
func Command() *cobra.Command {
	// $ infractl cluster artifacts
	cmd := &cobra.Command{
		Use:     "artifacts <cluster id>",
		Short:   "Download cluster artifacts",
		Long:    "Download artifacts from a cluster",
		Example: examples,
		RunE:    common.WithGRPCHandler(artifacts),
	}

	cmd.Flags().String("download-dir", "", "artifact download directory")
	return cmd
}

func artifacts(ctx context.Context, conn *grpc.ClientConn, cmd *cobra.Command, args []string) (common.PrettyPrinter, error) {
	if len(args) != 1 {
		return nil, errors.New("invalid arguments")
	}

	downloadDir, _ := cmd.Flags().GetString("download-dir")

	resp, err := v1.NewClusterServiceClient(conn).Artifacts(ctx, &v1.ResourceByID{Id: args[0]})
	if err != nil {
		return nil, err
	}

	// If no --download-dir flag was given, skip downloading the artifacts
	// altogether.
	if downloadDir == "" {
		return clusterArtifacts(*resp), nil
	}

	for _, artifact := range resp.Artifacts {
		if err := download(downloadDir, *artifact); err != nil {
			return nil, err
		}
	}

	return clusterArtifacts(*resp), nil
}

// download will save the given cluster artifact to disk inside the given
// directory.
func download(downloadDir string, artifact v1.Artifact) (err error) {
	// Create the download directory if it doesn't exist. All artifacts will be
	// downloaded into this directory.
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return err
	}

	// Create a new, empty file.
	filename := filepath.Join(downloadDir, artifact.Name)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		// if the file fails to close, return that error only if there was no
		// original error.
		if ferr := file.Close(); ferr != nil && err == nil {
			err = ferr
		}
	}()

	// Download the (GCS signed) URL.
	resp, err := http.Get(artifact.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close() // nolint:errcheck

	// Archive is gzipped, so we need to strip that away.
	gr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gr.Close() // nolint:errcheck

	// Archive is a normal tar archive.
	tr := tar.NewReader(gr)

	// We're expecting 1 and only 1 file in the archive, so read just the
	// first entry.
	if _, err := tr.Next(); err != nil {
		if err == io.EOF {
			return fmt.Errorf("unexpected EOF reading artifact %q", artifact.Name)
		}
		return err
	}

	// Copy the entirety of the archive artifact to its final destination on
	// disk.
	if _, err := io.Copy(file, tr); err != nil {
		return err
	}

	return nil
}
