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

# If infractl cannot determine your OS and you can specify linux or darwin
$ infractl cli upgrade --os linux

# If infractl cannot determine your architecture and you can specify amd64 or arm64
$ infractl cli upgrade --arch amd64
`

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

	cmd.Flags().String("os", "", "Optionally choose an OS: darwin (macOS) or linux")
	cmd.Flags().String("arch", "", "Optionally choose and arch: amd64 (Intel-based) or arm64 (Apple Silicon)")

	return cmd
}

func run(ctx context.Context, conn *grpc.ClientConn, cmd *cobra.Command, _ []string) (common.PrettyPrinter, error) {
	argOS, _ := cmd.Flags().GetString("os")
	argArch, _ := cmd.Flags().GetString("arch")
	OS, arch, err := guessOSAndArchIfNotSet(argOS, argArch)
	if err != nil {
		return nil, err
	}

	reader, err := v1.NewCliServiceClient(conn).Upgrade(ctx, &v1.CliUpgradeRequest{Os: OS, Arch: arch})
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

func guessOSAndArchIfNotSet(os, arch string) (string, string, error) {
	if os != "" && arch != "" {
		return os, arch, nil
	}

	uname, err := exec.Command("uname", "-sm").Output()
	if err != nil {
		return "", "", errors.Wrap(err, "Cannot run uname -sm to determine OS")
	}

	if os == "" {
		switch {
		case strings.Contains(string(uname), "Darwin"):
			os = "darwin"
		case strings.Contains(string(uname), "Linux"):
			os = "linux"
		default:
			return "", "", errors.Errorf("uname returned invalid OS: %s; must be Darwin or Linux", string(uname))
		}
	}

	if arch == "" {
		switch {
		case strings.Contains(string(uname), "x86_64"):
			arch = "amd64"
		case strings.Contains(string(uname), "arm64"):
			arch = "arm64"
		default:
			return "", "", errors.Errorf("uname returned invalid arch: %s; must be x86_64 or arm64", string(uname))
		}
	}

	return os, arch, nil
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
