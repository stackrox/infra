package cluster

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleName(t *testing.T) {
	type tableTest struct {
		title    string
		input    string
		expected string
	}

	tests := []tableTest{
		{
			title: "empty string",
		},
		{
			title: "all invalid",
			input: "@#$%.",
		},
		{
			title:    "typical title",
			input:    "Meeting with Acme, inc.",
			expected: "meeting-with-acme-inc",
		},
		{
			title:    "excited title",
			input:    "!!!Meeting with Acme,,, inc.!!!",
			expected: "meeting-with-acme-inc",
		},
	}

	for index, test := range tests {
		name := fmt.Sprintf("%d %s", index+1, test.title)
		t.Run(name, func(t *testing.T) {
			func(current tableTest) {
				t.Parallel()
				actual := simpleName(current.input)
				assert.Equal(t, actual, current.expected)
			}(test)
		})
	}
}
