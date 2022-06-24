package auth

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/stackrox/infra/config"
)

const (
	// OpIn defines claim check if value is withing the slice
	OpIn string = "in"

	// OpEqual defines claim check if value is exactly equal
	OpEqual string = "eq"
)

// ClaimOperation represents the configuration for checking access token claims.
type ClaimOperation struct {
	config.ClaimOperation
}

// equalCheck checks exact claim key against token claims
func (co *ClaimOperation) equalCheck(flatTokenClaims map[string]interface{}, key string) (bool, error) {
	tokenClaimValue, found := flatTokenClaims[key]
	if !found {
		return false, errors.Errorf("Expected claim %q not found", key)
	}

	return co.Value == tokenClaimValue, nil
}

// Check checks expected claim against token claims
func (co *ClaimOperation) Check(flatTokenClaims map[string]interface{}) (bool, error) {
	if co.Op == OpEqual {
		return co.equalCheck(flatTokenClaims, co.Key)
	}

	if co.Op == OpIn {
		// Loop will exit when we check for not existing key in flatTokenClaims.
		// Or early exit if value is found in list.
		for i := 0; ; i++ {
			// Flattened lists in token are suffixed with index. i.e.
			// Token JSON: { test: [1, 2] }
			// Flattened Token JSON: { "test.0": 1, "test.1": 2 }
			key := fmt.Sprintf("%s.%d", co.Key, i)

			isValid, err := co.equalCheck(flatTokenClaims, key)
			if err != nil {
				return false, errors.Errorf("Value provided within claim %q is not found", co.Key)
			}

			if isValid {
				return isValid, nil
			}
		}
	}

	return false, errors.Errorf("Unsupported operation %q for claim %q", co.Op, co.Key)
}
