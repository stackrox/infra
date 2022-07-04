package claimrule

import (
	"encoding/json"
	"testing"

	"gopkg.in/square/go-jose.v2"

	"github.com/stretchr/testify/require"
)

type dataSet struct {
	tokenClaims map[string]interface{}
	rules       ClaimRules
	err         bool
}

func getRawToken(t *testing.T, jwt map[string]interface{}) string {
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: "HS256", Key: []byte("secret")}, (&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		t.Fatal(err)
	}

	payload, err := json.Marshal(jwt)
	if err != nil {
		t.Fatal(err)
	}

	jws, err := signer.Sign(payload)
	if err != nil {
		t.Fatal(err)
	}

	data, err := jws.CompactSerialize()
	if err != nil {
		t.Fatal(err)
	}

	return data
}

func getDataSets() map[string]dataSet {
	return map[string]dataSet{
		"nil-claim-rules": {
			tokenClaims: map[string]interface{}{
				"field": "val",
			},
			rules: nil,
			err:   false,
		},
		"empty-claim-rules": {
			tokenClaims: map[string]interface{}{
				"field": "val",
			},
			rules: ClaimRules{},
			err:   false,
		},
		"empty-token-claims": {
			tokenClaims: map[string]interface{}{},
			rules: ClaimRules{{
				Value: "val",
				Key:   "field",
				Op:    "eq",
			}},
			err: true,
		},
		"unsupported-rule": {
			tokenClaims: map[string]interface{}{
				"field": "val",
			},
			rules: ClaimRules{{
				Value: "val",
				Key:   "field",
				Op:    "no",
			}},
			err: true,
		},
		"eq-valid": {
			tokenClaims: map[string]interface{}{
				"field": "val",
			},
			rules: ClaimRules{{
				Value: "val",
				Key:   "field",
				Op:    "eq",
			}},
			err: false,
		},
		"eq-not-same-value": {
			tokenClaims: map[string]interface{}{
				"field": "not-val",
			},
			rules: ClaimRules{{
				Value: "val",
				Key:   "field",
				Op:    "eq",
			}},
			err: true,
		},
		"eq-no-filed": {
			tokenClaims: map[string]interface{}{
				"no-field": "val",
			},
			rules: ClaimRules{{
				Value: "val",
				Key:   "field",
				Op:    "eq",
			}},
			err: true,
		},
		"in-valid": {
			tokenClaims: map[string]interface{}{
				"field": []string{"val1", "val2"},
			},
			rules: ClaimRules{{
				Value: "val2",
				Key:   "field",
				Op:    "in",
			}},
			err: false,
		},
		"eq-nested-valid": {
			tokenClaims: map[string]interface{}{
				"field": map[string]interface{}{
					"nested": map[string]interface{}{
						"level": "val",
					},
				},
			},
			rules: ClaimRules{{
				Value: "val",
				Key:   "field.nested.level",
				Op:    "eq",
			}},
			err: false,
		},
		"in-nested-valid": {
			tokenClaims: map[string]interface{}{
				"field": map[string]interface{}{
					"nested": map[string]interface{}{
						"level": []string{"val1", "val2"},
					},
				},
			},
			rules: ClaimRules{{
				Value: "val2",
				Key:   "field.nested.level",
				Op:    "in",
			}},
			err: false,
		},
		"in-nested-not-same-value": {
			tokenClaims: map[string]interface{}{
				"field": map[string]interface{}{
					"nested": map[string]interface{}{
						"level": []string{"val1", "val3"},
					},
				},
			},
			rules: ClaimRules{{
				Value: "val2",
				Key:   "field.nested.level",
				Op:    "in",
			}},
			err: true,
		},
	}
}

func TestClaimRulesValidate(t *testing.T) {
	tests := getDataSets()

	for testName, testSet := range tests {
		t.Run(testName, func(t *testing.T) {
			rawToken := getRawToken(t, testSet.tokenClaims)
			err := testSet.rules.Validate(rawToken)

			if testSet.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClaimRulesValidateNotJWT(t *testing.T) {
	rules := ClaimRules{{
		Value: "val",
		Key:   "field",
		Op:    "eq",
	}}

	err := rules.Validate("bad-token")
	require.Error(t, err)

	err = rules.Validate("not.goo.token")
	require.Error(t, err)

	err = rules.Validate("(not).(base64).(token)")
	require.Error(t, err)

	err = rules.Validate("")
	require.Error(t, err)

	err = rules.Validate("..")
	require.Error(t, err)

	// Access token is not valid if claim rules are defined and token in not JWT.
	exampleToken := "dmFsaWQtYWNjZXNzLXRva2Vu"
	err = rules.Validate(exampleToken)
	require.Error(t, err)

	// Access token is valid when claim rules are not defined.
	rules = ClaimRules{}
	err = rules.Validate(exampleToken)
	require.NoError(t, err)

	rules = nil
	err = rules.Validate(exampleToken)
	require.NoError(t, err)
}
