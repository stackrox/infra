package cluster

import (
	"regexp"
	"strings"
	"testing"

	v1 "github.com/stackrox/infra/generated/api/v1"
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
		{
			name:     "email starting with plus",
			email:    "+tag@example.com",
			expected: "plus.tag.at.example.com",
		},
		{
			name:     "long email truncating on dash",
			email:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-@example.com",
			expected: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		{
			name:     "long email truncating on dot",
			email:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.@example.com",
			expected: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
	}

	// Kubernetes label value regex
	labelValueRegex := regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9._-]*[A-Za-z0-9])?$`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := emailToLabelValue(tt.email)
			if result != tt.expected {
				t.Errorf("emailToLabelValue(%q) = %q, want %q", tt.email, result, tt.expected)
			}
			if len(result) > 63 {
				t.Errorf("emailToLabelValue(%q) returned %q with length %d, exceeds 63 chars", tt.email, result, len(result))
			}
			if result != "" && !labelValueRegex.MatchString(result) {
				t.Errorf("emailToLabelValue(%q) returned %q, not a valid K8s label value", tt.email, result)
			}
		})
	}
}

func TestBuildLabelSelector_FilterDeletedWorkflows(t *testing.T) {
	tests := []struct {
		name            string
		request         *v1.ClusterListRequest
		email           string
		expectedClauses []string // Expected clauses in the selector (order-independent)
	}{
		{
			name: "default - exclude deleted",
			request: &v1.ClusterListRequest{
				All:     false,
				Expired: false,
			},
			email:           "",
			expectedClauses: []string{"infra.stackrox.com/deleted!=true"},
		},
		{
			name: "expired flag - include deleted",
			request: &v1.ClusterListRequest{
				All:     false,
				Expired: true,
			},
			email:           "",
			expectedClauses: []string{}, // No deleted filter when expired=true
		},
		{
			name: "with owner filter",
			request: &v1.ClusterListRequest{
				All:     false,
				Expired: false,
			},
			email: "test@example.com",
			expectedClauses: []string{
				"infra.stackrox.com/deleted!=true",
				"infra.stackrox.com/owner=test.at.example.com",
			},
		},
		{
			name: "with flavor filter",
			request: &v1.ClusterListRequest{
				All:            false,
				Expired:        false,
				AllowedFlavors: []string{"gke-default", "eks-default"},
			},
			email: "",
			expectedClauses: []string{
				"infra.stackrox.com/deleted!=true",
				"infra.stackrox.com/flavor in (",
			},
		},
		{
			name: "with owner and flavor filters",
			request: &v1.ClusterListRequest{
				All:            false,
				Expired:        false,
				AllowedFlavors: []string{"gke-default"},
			},
			email: "user@example.com",
			expectedClauses: []string{
				"infra.stackrox.com/deleted!=true",
				"infra.stackrox.com/owner=user.at.example.com",
				"infra.stackrox.com/flavor in (gke-default)",
			},
		},
		{
			name: "all flag - exclude deleted unless expired",
			request: &v1.ClusterListRequest{
				All:     true,
				Expired: false,
			},
			email:           "user@example.com",
			expectedClauses: []string{"infra.stackrox.com/deleted!=true"},
		},
		{
			name: "expired and all flags - include deleted",
			request: &v1.ClusterListRequest{
				All:     true,
				Expired: true,
			},
			email:           "user@example.com",
			expectedClauses: []string{}, // No filters
		},
		{
			name: "expired with flavor filter - include deleted",
			request: &v1.ClusterListRequest{
				All:            false,
				Expired:        true,
				AllowedFlavors: []string{"gke-default"},
			},
			email: "",
			expectedClauses: []string{
				"infra.stackrox.com/flavor in (gke-default)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector, err := buildLabelSelector(tt.request, tt.email)
			if err != nil {
				t.Errorf("buildLabelSelector() returned error: %v", err)
				return
			}

			result := selector.String()

			// For empty expected clauses, result should be empty
			if len(tt.expectedClauses) == 0 {
				if result != "" {
					t.Errorf("expected empty selector, got %q", result)
				}
				return
			}

			// Check that all expected clauses are present (order-independent)
			for _, expectedClause := range tt.expectedClauses {
				if !strings.Contains(result, expectedClause) {
					t.Errorf("expected selector to contain %q, got %q", expectedClause, result)
				}
			}
		})
	}
}
