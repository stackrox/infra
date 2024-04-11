package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateParameterArgument(t *testing.T) {
	err := ValidateParameterArgument([]string{"machine-type", "e2-medium"})
	assert.NoError(t, err, "no error expected")
}
