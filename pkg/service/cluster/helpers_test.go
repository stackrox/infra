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
		expectError     bool
	}{
		{
			name: "empty email with All=false - should error",
			request: &v1.ClusterListRequest{
				All:     false,
				Expired: false,
			},
			email:       "",
			expectError: true,
		},
		{
			name: "empty email with All=false and Expired=true - should error",
			request: &v1.ClusterListRequest{
				All:     false,
				Expired: true,
			},
			email:       "",
			expectError: true,
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
			name: "empty email with flavor filter - should error",
			request: &v1.ClusterListRequest{
				All:            false,
				Expired:        false,
				AllowedFlavors: []string{"gke-default", "eks-default"},
			},
			email:       "",
			expectError: true,
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
			name: "empty email with expired and flavor filter - should error",
			request: &v1.ClusterListRequest{
				All:            false,
				Expired:        true,
				AllowedFlavors: []string{"gke-default"},
			},
			email:       "",
			expectError: true,
		},
		{
			name: "All=true with empty email - should succeed",
			request: &v1.ClusterListRequest{
				All:     true,
				Expired: false,
			},
			email:           "",
			expectedClauses: []string{"infra.stackrox.com/deleted!=true"},
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector, err := buildLabelSelector(tt.request, tt.email)

			if tt.expectError {
				if err == nil {
					t.Errorf("buildLabelSelector() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("buildLabelSelector() returned unexpected error: %v", err)
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

func TestValidateClusterID(t *testing.T) {
	tests := []struct {
		name        string
		clusterID   string
		expectError bool
	}{
		{
			name:        "valid simple ID",
			clusterID:   "my-cluster",
			expectError: false,
		},
		{
			name:        "valid ID with dots and underscores",
			clusterID:   "cluster-1.test_env",
			expectError: false,
		},
		{
			name:        "valid single character",
			clusterID:   "a",
			expectError: false,
		},
		{
			name:        "valid two characters",
			clusterID:   "ab",
			expectError: false,
		},
		{
			name:        "empty cluster ID",
			clusterID:   "",
			expectError: true,
		},
		{
			name:        "cluster ID with spaces",
			clusterID:   "my cluster",
			expectError: true,
		},
		{
			name:        "cluster ID too long (64 chars)",
			clusterID:   "a234567890123456789012345678901234567890123456789012345678901234",
			expectError: true,
		},
		{
			name:        "cluster ID exactly 63 chars",
			clusterID:   "a23456789012345678901234567890123456789012345678901234567890123",
			expectError: false,
		},
		{
			name:        "cluster ID starting with dash",
			clusterID:   "-cluster",
			expectError: true,
		},
		{
			name:        "cluster ID ending with dash",
			clusterID:   "cluster-",
			expectError: true,
		},
		{
			name:        "cluster ID starting with dot",
			clusterID:   ".cluster",
			expectError: true,
		},
		{
			name:        "cluster ID ending with dot",
			clusterID:   "cluster.",
			expectError: true,
		},
		{
			name:        "cluster ID with special characters",
			clusterID:   "cluster@test",
			expectError: true,
		},
		{
			name:        "cluster ID with slash",
			clusterID:   "cluster/test",
			expectError: true,
		},
		{
			name:        "cluster ID starting with underscore",
			clusterID:   "_cluster",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateClusterID(tt.clusterID)
			if tt.expectError && err == nil {
				t.Errorf("validateClusterID(%q) expected error but got none", tt.clusterID)
			}
			if !tt.expectError && err != nil {
				t.Errorf("validateClusterID(%q) expected no error but got: %v", tt.clusterID, err)
			}
		})
	}
}
