package claimrule

import (
	"encoding/json"
	"testing"

	"gopkg.in/square/go-jose.v2"

	"github.com/stretchr/testify/assert"
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
				Path:  "field",
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
				Path:  "field",
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
				Path:  "field",
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
				Path:  "field",
				Op:    "eq",
			}},
			err: true,
		},
		"eq-no-field": {
			tokenClaims: map[string]interface{}{
				"no-field": "val",
			},
			rules: ClaimRules{{
				Value: "val",
				Path:  "field",
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
				Path:  "field",
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
				Path:  "field.nested.level",
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
				Path:  "field.nested.level",
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
				Path:  "field.nested.level",
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
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClaimRulesValidateNotJWT(t *testing.T) {
	rules := ClaimRules{{
		Value: "val",
		Path:  "field",
		Op:    "eq",
	}}

	err := rules.Validate("bad-token")
	assert.Error(t, err)

	err = rules.Validate("not.goo.token")
	assert.Error(t, err)

	err = rules.Validate("(not).(base64).(token)")
	assert.Error(t, err)

	err = rules.Validate("")
	assert.Error(t, err)

	err = rules.Validate("..")
	assert.Error(t, err)

	// Access token is not valid if claim rules are defined and token in not JWT.
	exampleToken := "dmFsaWQtYWNjZXNzLXRva2Vu"
	err = rules.Validate(exampleToken)
	assert.Error(t, err)

	// Access token is valid when claim rules are not defined.
	rules = ClaimRules{}
	err = rules.Validate(exampleToken)
	assert.NoError(t, err)

	rules = nil
	err = rules.Validate(exampleToken)
	assert.NoError(t, err)
}

func TestOperationUnmarshalJSON(t *testing.T) {
	type operationUnmarshalTestSet struct {
		rawData []byte
		error   bool
	}

	testCases := map[string]operationUnmarshalTestSet{
		"fail-unsupported": {
			rawData: []byte("\"abc\""),
			error:   true,
		},
		"fail-number": {
			rawData: []byte("1"),
			error:   true,
		},
		"in": {
			rawData: []byte("\"in\""),
			error:   false,
		},
		"eq": {
			rawData: []byte("\"eq\""),
			error:   false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			var op operation
			err := json.Unmarshal(testCase.rawData, &op)

			assert.Equal(t, testCase.error, err != nil)
		})
	}
}
