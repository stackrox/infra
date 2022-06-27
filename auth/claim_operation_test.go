package auth

import (
	"github.com/stackrox/infra/config"
	"github.com/stretchr/testify/require"
	"testing"
)

type dataSet struct {
	flattenedJSON map[string]interface{}
	operation     ClaimOperation
	err           bool
	res           bool
}

func getDataSets() []dataSet {
	return []dataSet{
		{
			flattenedJSON: map[string]interface{}{
				"field": "val",
			},
			operation: ClaimOperation{
				config.ClaimOperation{
					Value: "val",
					Key:   "field",
					Op:    "eq",
				},
			},
			res: true,
			err: false,
		},
		{
			flattenedJSON: map[string]interface{}{
				"field": "no-val",
			},
			operation: ClaimOperation{
				config.ClaimOperation{
					Value: "val",
					Key:   "field",
					Op:    "eq",
				},
			},
			res: false,
			err: false,
		},
		{
			flattenedJSON: map[string]interface{}{
				"no-field": "val",
			},
			operation: ClaimOperation{
				config.ClaimOperation{
					Value: "val",
					Key:   "field",
					Op:    "eq",
				},
			},
			res: false,
			err: true,
		},
		{
			flattenedJSON: map[string]interface{}{
				"field.0": "val1",
				"field.1": "val2",
			},
			operation: ClaimOperation{
				config.ClaimOperation{
					Value: "val2",
					Key:   "field",
					Op:    "in",
				},
			},
			res: true,
			err: false,
		},
		{
			flattenedJSON: map[string]interface{}{
				"field.nested.level": "val",
			},
			operation: ClaimOperation{
				config.ClaimOperation{
					Value: "val",
					Key:   "field.nested.level",
					Op:    "eq",
				},
			},
			res: true,
			err: false,
		},
		{
			flattenedJSON: map[string]interface{}{
				"field.nested.level.0": "val1",
				"field.nested.level.1": "val2",
			},
			operation: ClaimOperation{
				config.ClaimOperation{
					Value: "val2",
					Key:   "field.nested.level",
					Op:    "in",
				},
			},
			res: true,
			err: false,
		},
	}
}

func TestMarshal(t *testing.T) {
	tests := getDataSets()

	for _, testSet := range tests {
		res, err := testSet.operation.Check(testSet.flattenedJSON)

		require.Equal(t, testSet.res, res)
		if testSet.err {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
