package find_test

import (
	"testing"

	"github.com/stackrox/infra/cmd/infractl/janitor/find"
	"github.com/stretchr/testify/assert"
)

func TestFormatInstanceNames(t *testing.T) {
	instance := &find.ComputeInstance{Name: "gke-pr-03-10-work-gke-default-pool-53807d4f-x0tb"}
	expected := "pr0310workgke"

	assert.Equal(t, expected, find.FormatInstanceNames([]*find.ComputeInstance{instance})[0].Name, "they should match")
}
