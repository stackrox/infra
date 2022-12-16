package e2e

import v1 "github.com/stackrox/infra/generated/api/v1"

// StatusResponse helper maps to the JSON response for infractl status operations.
type StatusResponse struct {
	Status v1.InfraStatus
}

// WhoamiResponse helper maps to the JSON response for infractl whoami operations.
type WhoamiResponse struct {
	Principal v1.WhoamiResponse_ServiceAccount
}
