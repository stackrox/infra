package e2e

type StatusResponse struct {
	Status struct {
		MaintenanceActive bool
		Maintainer        string
	}
}

type WhoamiResponse struct {
	Principal struct {
		ServiceAccount struct {
			Email string
		}
	}
}
