package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	SyncTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitops_sync_total",
			Help: "Total number of sync operations, partitioned by status",
		},
		[]string{"status"},
	)
	SyncDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "gitops_sync_duration_seconds",
			Help:    "Duration of sync operations in seconds",
			Buckets: prometheus.LinearBuckets(0, 10, 10),
		},
	)
	ResourceManaged = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitops_resource_managed_total",
			Help: "Total number of resources managed by operation and kind",
		},
		[]string{"operation", "kind"},
	)

	LastSyncTimestamp = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "gitops_last_sync_timestamp",
			Help: "Timestamp of the last successful sync operation",
		},
	)
)

func Register() {

}
