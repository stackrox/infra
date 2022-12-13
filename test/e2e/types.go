package e2e

// StatusResponse helper maps to the JSON response for infractl status operations.
type StatusResponse struct {
	Status struct {
		MaintenanceActive bool
		Maintainer        string
	}
}

// WhoamiResponse helper maps to the JSON response for infractl whoami operations.
type WhoamiResponse struct {
	Principal struct {
		ServiceAccount struct {
			Email string
		}
	}
}
