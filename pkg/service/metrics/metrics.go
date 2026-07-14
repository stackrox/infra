// Package metrics exposes custom metrics for the infra server.
package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// FlavorsUsedCounter is a Prometheus metric, counting the number of clusters created per flavor
	FlavorsUsedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "infra",
			Name:      "flavors_used",
			Help:      "Number of clusters created by flavor",
		},
		[]string{"flavor"},
	)

	// ArtifactCacheHitsCounter tracks successful cache lookups for GCS artifacts
	ArtifactCacheHitsCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "infra",
			Name:      "artifact_cache_hits_total",
			Help:      "Total number of artifact cache hits",
		},
	)

	// ArtifactCacheMissesCounter tracks cache misses requiring GCS API calls
	ArtifactCacheMissesCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "infra",
			Name:      "artifact_cache_misses_total",
			Help:      "Total number of artifact cache misses",
		},
	)

	// ArtifactCacheSizeGauge reports current number of entries in the artifact cache
	ArtifactCacheSizeGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "infra",
			Name:      "artifact_cache_size",
			Help:      "Current number of entries in the artifact cache",
		},
	)
)

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(FlavorsUsedCounter)
	prometheus.MustRegister(ArtifactCacheHitsCounter)
	prometheus.MustRegister(ArtifactCacheMissesCounter)
	prometheus.MustRegister(ArtifactCacheSizeGauge)
}
