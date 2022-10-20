package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	FlavorsUsedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "infra",
			Name:      "flavors_used",
			Help:      "Kubernetes cluster flavors being created",
		},
		[]string{"flavor"},
	)
)

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(FlavorsUsedCounter)
}