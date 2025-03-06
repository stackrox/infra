package claimrule

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/jeremywohl/flatten/v2"
	"github.com/pkg/errors"
)

type operation string

var _ json.Unmarshaler = (*operation)(nil)

const (
	// In defines claim check if value is withing the slice.
	In operation = "in"

	// Equal defines claim check if value is equal.
	Equal operation = "eq"
)

// UnmarshalJSON does validation of supported operations for claim rule.
func (op *operation) UnmarshalJSON(b []byte) error {
	var strOp string
	err := json.Unmarshal(b, &strOp)
	if err != nil {
		return err
	}

	*op = operation(strOp)
	if *op != In && *op != Equal {
		return errors.Errorf("unsupported operation %q in claim rule", *op)
	}

	return nil
}

// ClaimRule represents the configuration for checking access token claims.
type ClaimRule struct {
	// Values is used to compare with retrieved value from token claims.
	Value interface{} `json:"value"`

	// Op represents defined operation for the claim rule that should be used
	// during checking of token claims. It can be "eq" or "in".
	// - "eq" is used to compare single value from token claims.
	// - "in" is used to look if defined value is in the list of values for
	//   defined token claim path.
	Op operation `json:"op"`

	// Path represent JSON path to specific key in the token claims. Nested
	// fields are separated by '.'. i.e. "top_level.field.sub_field".
	Path string `json:"path"`
}

// jsonPathExists checks if jsonPath exists in flat token claims.
func (cr *ClaimRule) jsonPathExists(flatTokenClaims map[string]interface{}, jsonPath string) bool {
	_, found := flatTokenClaims[jsonPath]

	return found
}

// equalCheck checks exact claim key against token claims.
func (cr *ClaimRule) equalCheck(flatTokenClaims map[string]interface{}, jsonPath string) error {
	tokenClaimValue, found := flatTokenClaims[jsonPath]
	if !found {
		return errors.Errorf("expected claim %q is not found", jsonPath)
	}

	pattern := fmt.Sprintf("^%s$", cr.Value)
	found, err := regexp.MatchString(pattern, tokenClaimValue.(string))
	if !found {
		return errors.Errorf("expected claim %q is not correct", jsonPath)
	}
	if err != nil {
		return errors.Wrapf(err, "error matching claim %s to expected value", tokenClaimValue)
	}

	return nil
}

// Checks expected claim against token claims. This function expects flattened
// JSON created from access token claims.
func (cr *ClaimRule) check(flatTokenClaims map[string]interface{}) error {
	if cr.Op == Equal {
		return cr.equalCheck(flatTokenClaims, cr.Path)
	}

	if cr.Op == In {
		// Loop will exit when we check for not existing key in flatTokenClaims.
		// Or early exit if value is found in list.
		for i := 0; ; i++ {
			// Flattened lists in token are suffixed with index. i.e.
			// Token JSON: { test: [1, 2] }
			// Flattened Token JSON: { "test.0": 1, "test.1": 2 }
			jsonPath := fmt.Sprintf("%s.%d", cr.Path, i)
			if !cr.jsonPathExists(flatTokenClaims, jsonPath) {
				return errors.Errorf("value %q is not found in claim rule path %q", cr.Value, cr.Path)
			}

			if errCheck := cr.equalCheck(flatTokenClaims, jsonPath); errCheck == nil {
				return nil
			}
		}
	}

	return errors.Errorf("unsupported operation %q for claim rule path %q", cr.Op, cr.Path)
}

// ClaimRules represents the collection of claim ruels that should be validated
// against an access token claims.
type ClaimRules []ClaimRule

func decodeAccessToken(rawAccessToken string) (map[string]interface{}, error) {
	rawTokenParts := strings.Split(rawAccessToken, ".")
	if len(rawTokenParts) < 2 {
		return nil, errors.New("invalid token received")
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
		if errCheck := expectedClaim.check(flatTokenClaims); errCheck != nil {
			return errCheck
		}
	}

	return nil
}
