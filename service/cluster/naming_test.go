package cluster

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleName(t *testing.T) {
	tests := []struct {
		title    string
		input    string
		expected string
	}{
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
			t.Parallel()
			actual := simpleName(test.input)
			assert.Equal(t, actual, test.expected)
		})
	}
}
