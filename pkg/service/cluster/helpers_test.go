package cluster

import (
	"regexp"
	"slices"
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

func TestParseVMOSList(t *testing.T) {
	tests := []struct {
		name    string
		vmOS    string
		want    []string
		wantErr string
	}{
		{name: "empty", vmOS: ""},
		{name: "whitespace", vmOS: "  "},
		{name: "single", vmOS: "rhel9", want: []string{"rhel9"}},
		{name: "rhel8", vmOS: "rhel8", want: []string{"rhel8"}},
		{name: "mixed with spaces", vmOS: "rhel8, rhel9, rhel10", want: []string{"rhel8", "rhel9", "rhel10"}},
		{name: "uppercase", vmOS: "RHEL9", want: []string{"rhel9"}},
		{name: "empty entry", vmOS: "rhel9,,rhel10", wantErr: "empty entry"},
		{name: "trailing comma", vmOS: "rhel9,", wantErr: "empty entry"},
		{name: "unsupported", vmOS: "rhel7", wantErr: "unsupported vm-os"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVMOSList(tt.vmOS)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got none", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("parseVMOSList(%q) = %q, want %q", tt.vmOS, got, tt.want)
			}
		})
	}
}

func TestValidateVirtWorkerNodeType(t *testing.T) {
	tests := []struct {
		name       string
		vmOS       string
		workerType string
		wantErr    bool
	}{
		{
			name:       "empty vm-os default e2",
			vmOS:       "",
			workerType: "e2-standard-8",
		},
		{
			name:       "whitespace vm-os",
			vmOS:       "  ",
			workerType: "e2-standard-8",
		},
		{
			name:       "rhel9 on n2-standard-8",
			vmOS:       "rhel9",
			workerType: "n2-standard-8",
		},
		{
			name:       "list on n2-standard-4",
			vmOS:       "rhel9,rhel10",
			workerType: "n2-standard-4",
		},
		{
			name:       "uppercase os on n2",
			vmOS:       "RHEL9",
			workerType: "n2-standard-8",
		},
		{
			name:       "rhel9 on c3",
			vmOS:       "rhel9",
			workerType: "c3-standard-8",
		},
		{
			name:       "rhel9 on n4d (AMD exception)",
			vmOS:       "rhel9",
			workerType: "n4d-standard-8",
		},
		{
			name:       "rhel9 on default e2",
			vmOS:       "rhel9",
			workerType: "e2-standard-8",
			wantErr:    true,
		},
		{
			name:       "rhel9 on n2d (AMD)",
			vmOS:       "rhel9",
			workerType: "n2d-standard-8",
			wantErr:    true,
		},
		{
			name:       "rhel9 on t2a (ARM)",
			vmOS:       "rhel9",
			workerType: "t2a-standard-8",
			wantErr:    true,
		},
		{
			name:       "rhel9 on m3 (memory-optimized)",
			vmOS:       "rhel9",
			workerType: "m3-ultramem-32",
			wantErr:    true,
		},
		{
			name:       "rhel9 on empty worker type",
			vmOS:       "rhel9",
			workerType: "",
			wantErr:    true,
		},
		{
			name:       "unsupported os",
			vmOS:       "debian",
			workerType: "n2-standard-8",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVirtWorkerNodeType(tt.vmOS, tt.workerType)
			if tt.wantErr && err == nil {
				t.Errorf("validateVirtWorkerNodeType(%q, %q) expected error but got none", tt.vmOS, tt.workerType)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateVirtWorkerNodeType(%q, %q) expected no error but got: %v", tt.vmOS, tt.workerType, err)
			}
		})
	}
}

func TestCheckAndEnrichParameters_VirtWorkerType(t *testing.T) {
	flavorParams := map[string]*v1.Parameter{
		"name": {Name: "name"},
		"vm-os": {
			Name:     "vm-os",
			Value:    "",
			Optional: true,
		},
		"worker-node-type": {
			Name:     "worker-node-type",
			Value:    "e2-standard-8",
			Optional: true,
		},
	}

	tests := []struct {
		name           string
		req            map[string]string
		wantErr        string
		wantWorkerType string
	}{
		{
			name:           "vm-os omitted keeps default e2",
			req:            map[string]string{"name": "abc"},
			wantWorkerType: "e2-standard-8",
		},
		{
			name: "rhel9 with n2-standard-8",
			req: map[string]string{
				"name":             "abc",
				"vm-os":            "rhel9",
				"worker-node-type": "n2-standard-8",
			},
			wantWorkerType: "n2-standard-8",
		},
		{
			name: "list with n2-standard-4",
			req: map[string]string{
				"name":             "abc",
				"vm-os":            "rhel9,rhel10",
				"worker-node-type": "n2-standard-4",
			},
			wantWorkerType: "n2-standard-4",
		},
		{
			name: "rhel9 with default e2",
			req: map[string]string{
				"name":  "abc",
				"vm-os": "rhel9",
			},
			wantErr: "vm-os requires a worker-node-type with nested kvm",
		},
		{
			name: "rhel9 with explicit e2",
			req: map[string]string{
				"name":             "abc",
				"vm-os":            "rhel9",
				"worker-node-type": "e2-standard-8",
			},
			wantErr: "vm-os requires a worker-node-type with nested kvm",
		},
		{
			name: "unsupported os",
			req: map[string]string{
				"name":             "abc",
				"vm-os":            "debian",
				"worker-node-type": "n2-standard-8",
			},
			wantErr: "unsupported vm-os",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkAndEnrichParameters(flavorParams, tt.req)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got none", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotType := workflowParameterValue(got, "worker-node-type"); gotType != tt.wantWorkerType {
				t.Errorf("worker-node-type = %q, want %q", gotType, tt.wantWorkerType)
			}
		})
	}
}
