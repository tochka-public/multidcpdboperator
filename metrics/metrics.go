package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var MultidcpdbStatusDesiredAvailable = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "multidcpdb_status_desired_healthy",
	Help: "Number of custom resources by status",
}, []string{"multidcpdb", "cluster", "namespace"})

var MultidcpdbStatusPodDisruptionsAllowed = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "MultidcpdbStatusPodDisruptionsAllowed",
	Help: "Unix creation timestamp",
}, []string{"multidcpdb", "cluster", "namespace"})

func init() {
	// Register metrics with Prometheus
	metrics.Registry.MustRegister(MultidcpdbStatusDesiredAvailable)
	metrics.Registry.MustRegister(MultidcpdbStatusPodDisruptionsAllowed)
}
