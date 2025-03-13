package mock

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/stackrox/infra/cmd/infractl/common"
)

// PrepareCommand adds common flags and default args to a cobra.Command for test simulation.
func PrepareCommand(cmd *cobra.Command, asJSON bool, args ...string) *bytes.Buffer {
	common.AddCommonFlags(cmd)

	defaultArgs := []string{"--endpoint=localhost:8443", "--insecure"}
	args = append(args, defaultArgs...)
	if asJSON {
		args = append(args, "--json")
	}

	cmd.SetArgs(args)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	return buf
}

func PrepareCommandStdIn(cmd *cobra.Command, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	os.Stdin = file
	return nil
}

// RetrieveCommandOutput stringifies the contents of a buffer to read a command's output.
func RetrieveCommandOutput(buf *bytes.Buffer) (string, error) {
	data, err := io.ReadAll(buf)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// RetrieveCommandOutputJSON parses the contents of a buffer to a map.
func RetrieveCommandOutputJSON(buf *bytes.Buffer, outJSON interface{}) error {
	data, err := io.ReadAll(buf)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &outJSON)
	if err != nil {
		return err
	}
	return nil
}
