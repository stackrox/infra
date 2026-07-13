package cluster

import (
	"testing"
)

func TestEmailToLabelValue(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "simple email",
			email:    "user@example.com",
			expected: "user.at.example.com",
		},
		{
			name:     "email with plus",
			email:    "user+tag@example.com",
			expected: "user.plus.tag.at.example.com",
		},
		{
			name:     "email with dot",
			email:    "user.name@example.com",
			expected: "user.name.at.example.com",
		},
		{
			name:     "long email that exceeds 63 chars",
			email:    "very-long-email-address-that-exceeds-the-maximum-length@example.com",
			expected: "very-long-email-address-that-exceeds-the-maximum-length.at.exam",
		},
		{
			name:     "complex email",
			email:    "user.name+test@subdomain.example.com",
			expected: "user.name.plus.test.at.subdomain.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := emailToLabelValue(tt.email)
			if result != tt.expected {
				t.Errorf("emailToLabelValue(%q) = %q, want %q", tt.email, result, tt.expected)
			}
			if len(result) > 63 {
				t.Errorf("emailToLabelValue(%q) returned %q with length %d, exceeds 63 chars", tt.email, result, len(result))
			}
		})
	}
}
