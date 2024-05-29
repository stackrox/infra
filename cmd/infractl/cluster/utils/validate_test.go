package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateParameterArgument(t *testing.T) {
	err := ValidateParameterArgument([]string{"machine-type", "e2-medium"})
	assert.NoError(t, err, "no error expected")

	err = ValidateParameterArgument([]string{"container-type", "container_d"})
	assert.NoError(t, err, "no error expected")

	err = ValidateParameterArgument([]string{"user-arns", "arn:aws:iam::393282794030:user/joey@stackrox.com"})
	assert.NoError(t, err, "no error expected")
}
