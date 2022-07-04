package claimrule

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jeremywohl/flatten"
	"github.com/pkg/errors"
)

const (
	// In defines claim check if value is withing the slice.
	In string = "in"

	// Equal defines claim check if value is equal.
	Equal string = "eq"
)

// ClaimRule represents the configuration for checking access token claim.
type ClaimRule struct {
	Value interface{} `json:"value"`
	Key   string      `json:"key"`
	Op    string      `json:"op"`
}

// equalCheck checks exact claim key against token claims.
func (cr *ClaimRule) equalCheck(flatTokenClaims map[string]interface{}, key string) (bool, error) {
	tokenClaimValue, found := flatTokenClaims[key]
	if !found {
		return false, errors.Errorf("expected claim %q not found", key)
	}

	return cr.Value == tokenClaimValue, nil
}

// Checks expected claim against token claims. This function expects flattened
// JSON created from access token claims.
func (cr *ClaimRule) check(flatTokenClaims map[string]interface{}) (bool, error) {
	if cr.Op == Equal {
		return cr.equalCheck(flatTokenClaims, cr.Key)
	}

	if cr.Op == In {
		// Loop will exit when we check for not existing key in flatTokenClaims.
		// Or early exit if value is found in list.
		for i := 0; ; i++ {
			// Flattened lists in token are suffixed with index. i.e.
			// Token JSON: { test: [1, 2] }
			// Flattened Token JSON: { "test.0": 1, "test.1": 2 }
			key := fmt.Sprintf("%s.%d", cr.Key, i)

			isValid, err := cr.equalCheck(flatTokenClaims, key)
			if err != nil {
				return false, errors.Errorf("value provided within claim %q is not found", cr.Key)
			}

			if isValid {
				return isValid, nil
			}
		}
	}

	return false, errors.Errorf("unsupported rule %q for claim %q", cr.Op, cr.Key)
}

// ClaimRules represents the collection of claim ruels that should be validated
// against an access token claims.
type ClaimRules []ClaimRule

func decodeAccessToken(rawAccessToken string) (map[string]interface{}, error) {
	rawTokenParts := strings.Split(rawAccessToken, ".")
	if len(rawTokenParts) < 2 {
		return nil, errors.New("jws: invalid token received")
	}

	decoded, errDecode := base64.RawURLEncoding.DecodeString(rawTokenParts[1])
	if errDecode != nil {
		return nil, errDecode
	}

	tokenClaims := map[string]interface{}{}
	errDecode = json.NewDecoder(bytes.NewBuffer(decoded)).Decode(&tokenClaims)

	return tokenClaims, errDecode
}

// IsEmpty returns if claim rules are not defined.
func (cos *ClaimRules) IsEmpty() bool {
	return cos == nil || len(*cos) == 0
}

// Validate validates all defined claim rules for the access token.
func (cos *ClaimRules) Validate(rawAccessToken string) error {
	if cos.IsEmpty() {
		return nil
	}

	tokenClaims, errDecode := decodeAccessToken(rawAccessToken)
	if errDecode != nil {
		return errDecode
	}

	flatTokenClaims, errFlatten := flatten.Flatten(tokenClaims, "", flatten.DotStyle)
	if errFlatten != nil {
		return errFlatten
	}

	for _, expectedClaim := range *cos {
		isValid, errCheck := expectedClaim.check(flatTokenClaims)
		if errCheck != nil {
			return errCheck
		}

		if !isValid {
			return errors.Errorf("claim for key %q is not valid", expectedClaim.Key)
		}
	}

	return nil
}
