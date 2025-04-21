package metrics

import (
	"context"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"

	corev1 "k8s.io/api/core/v1"

	multidcpdbv1 "k8s.tochka.com/multidc-pdb-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

type PDBStateCollector struct {
	Clusters          map[string]*cluster.Cluster
	Ctx               context.Context
	MetricUpdateDelay time.Duration
}

func (p *PDBStateCollector) CalculatePDBMetrics() error {

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = multidcpdbv1.AddToScheme(scheme)

	for {

		for _, cl := range p.Clusters {

			cfg := (*cl).GetConfig() // get the rest.Config
			k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
			if err != nil {
				klog.Error(err)
				continue
			}

			namespaceList := &corev1.NamespaceList{}

			if err := k8sClient.List(p.Ctx, namespaceList, &client.ListOptions{}); err != nil {
				klog.Error(err)
				continue
			}

			multidcPDBList := &multidcpdbv1.MultidcPodDisruptionBudgetList{}

			for _, ns := range namespaceList.Items {
				if err := k8sClient.List(p.Ctx, multidcPDBList, &client.ListOptions{Namespace: ns.Name}); err != nil {
					klog.Error(err)
					continue
				}
				for _, mpdb := range multidcPDBList.Items {
					defaultLabels := prometheus.Labels{"multidcpdb": mpdb.Name, "namespace": mpdb.Namespace, "cluster": (*cl).GetConfig().ServerName}

					if minAvailable, err := strconv.Atoi(mpdb.Spec.MinAvailable); err == nil {
						MultidcpdbStatusDesiredAvailable.With(defaultLabels).Set(float64(minAvailable))
					}
					if maxUnavailable, err := strconv.Atoi(mpdb.Spec.MaxUnavailable); err == nil {
						MultidcpdbStatusPodDisruptionsAllowed.With(defaultLabels).Set(float64(maxUnavailable))
					}

				}

			}

		}
		time.Sleep(p.MetricUpdateDelay)
	}
}
