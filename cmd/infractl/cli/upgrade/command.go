// Package upgrade implements the infractl cli upgrade command.
package upgrade

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
	v1 "github.com/stackrox/infra/generated/api/v1"
	"google.golang.org/grpc"
)

const examples = `# Upgrade infractl in place
$ infractl cli upgrade

# If infractl cannot determine your OS you can specify linux or darwin
$ infractl cli upgrade --os linux`

// Command defines the handler for infractl cli upgrade.
func Command() *cobra.Command {
	// $ infractl cli upgrade
	cmd := &cobra.Command{
		Use:     "upgrade",
		Short:   "Upgrade infractl",
		Long:    "Downloads and installs in-place the latest infractl",
		Example: examples,
		Args:    common.ArgsWithHelp(cobra.ExactArgs(0)),
		RunE:    common.WithGRPCHandler(run),
	}

	cmd.Flags().String("os", "", "Optionally choose an OS: darwin or linux")

	return cmd
}

func run(ctx context.Context, conn *grpc.ClientConn, cmd *cobra.Command, _ []string) (common.PrettyPrinter, error) {
	argOS, _ := cmd.Flags().GetString("os")
	OS, err := guessOSIfNotSet(argOS)
	if err != nil {
		return nil, err
	}

	reader, err := v1.NewCliServiceClient(conn).Upgrade(ctx, &v1.CliUpgradeRequest{Os: OS})
	if err != nil {
		return nil, err
	}
	bytes, err := recvBytes(reader)
	if err != nil {
		return nil, err
	}

	tempFilename, err := writeToTempFileAndTest(bytes)
	if err != nil {
		return nil, err
	}

	infractlFilename, err := moveIntoPlace(tempFilename)
	if err != nil {
		return nil, err
	}

	return prettyCliUpgrade{infractlFilename}, nil
}

func guessOSIfNotSet(os string) (string, error) {
	if os != "" {
		return os, nil
	}

	uname, err := exec.Command("uname", "-a").Output()
	if err != nil {
		return "", errors.Wrap(err, "Cannot run uname -a to determine OS")
	}

	if strings.Contains(string(uname), "Darwin") {
		os = "darwin"
	} else if strings.Contains(string(uname), "Linux") {
		os = "linux"
	} else {
		return "", errors.New("Cannot determine OS from: " + string(uname))
	}

	return os, nil
}

func recvBytes(reader v1.CliService_UpgradeClient) ([]byte, error) {
	var bytes []byte
	for {
		chunk, err := reader.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "Error reading from infra-server")
		}
		bytes = append(bytes, chunk.FileChunk...)
	}
	return bytes, nil
}

func writeToTempFileAndTest(bytes []byte) (string, error) {
	tempFile, err := os.CreateTemp("", "infractl-download-")
	if err != nil {
		return "", err
	}

	_, err = tempFile.Write(bytes)
	if err != nil {
		return "", errors.Wrap(err, "Cannot write to a temp file")
	}
	err = tempFile.Close()
	if err != nil {
		return "", err
	}

	err = os.Chmod(tempFile.Name(), 0755)
	if err != nil {
		return "", err
	}

	err = exec.Command(tempFile.Name()).Run()
	if err != nil {
		return "", errors.Wrap(err, "Cannot run the downloaded infractl")
	}

	return tempFile.Name(), nil
}

func moveIntoPlace(tempFilename string) (string, error) {
	infractlFilename, err := filepath.Abs(os.Args[0])
	if err != nil {
		return "", errors.Wrap(err, "Cannot determine infractl path")
	}

	err = os.Rename(tempFilename, infractlFilename)
	if err != nil {
		return "", errors.Wrap(err, "Cannot move the download into place")
	}

	return infractlFilename, nil
}
