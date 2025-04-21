package main

import (
	"context"
	"flag"
	"os"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/cluster"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"

	// GetConfigWithContext
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	sigswebhook "sigs.k8s.io/controller-runtime/pkg/webhook"

	multidccrd "k8s.tochka.com/multidc-pdb-operator/api/v1"
	"k8s.tochka.com/multidc-pdb-operator/controllers"
	"k8s.tochka.com/multidc-pdb-operator/metrics"
	"k8s.tochka.com/multidc-pdb-operator/webhook"
	//+kubebuilder:scaffold:imports
)

var (
	scheme       = runtime.NewScheme()
	setupLog     = ctrl.Log.WithName("setup")
	peerClusters = map[string]*cluster.Cluster{}
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(multidccrd.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var certDir string
	var metricsAddr string
	var probeAddr string
	var customMetricsCollectDelay string
	flag.StringVar(&certDir, "cert-dir", "/pki", "Cert dir")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&customMetricsCollectDelay, "custom-metrics-collect-delay", "30s", "The delay for collecting metrics about CRD values. Default is 30s")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	customDelayAsDuration, err := time.ParseDuration(customMetricsCollectDelay)
	if err != nil {
		setupLog.Error(err, "unable to set metrics collect delay as duration. Using default 30s period")
		customDelayAsDuration, _ = time.ParseDuration("30s")
	}

	// caching only what is needed
	fieldSelector := fields.Everything()
	labelSelector := labels.Everything()
	selectorsByObject := cache.SelectorsByObject{
		&corev1.Pod{}: {
			Field: fieldSelector,
			Label: labelSelector,
		},
		&appsv1.ReplicaSet{}: {
			Field: fieldSelector,
			Label: labelSelector,
		},
		&appsv1.StatefulSet{}: {
			Field: fieldSelector,
			Label: labelSelector,
		},
		&appsv1.Deployment{}: {
			Field: fieldSelector,
			Label: labelSelector,
		},
		&multidccrd.MultidcPodDisruptionBudget{}: {
			Field: fieldSelector,
			Label: labelSelector,
		},
	}
	kubeMainContext := os.Getenv("KUBECONTEXT")
	if len(kubeMainContext) == 0 {
		setupLog.Error(nil, "unable to get env KUBECONTEXT")
		os.Exit(1)
	}

	kubeContextPeers := os.Getenv("KUBECONTEXT_PEERS")
	contexts := []string{kubeMainContext}
	if len(kubeContextPeers) > 0 {
		contexts = append(contexts, strings.Split(kubeContextPeers, ",")...)
	}

	ctx := context.Background()
	for _, kubeContext := range contexts {
		ctxConfig, err := config.GetConfigWithContext(kubeContext)
		if err != nil {
			setupLog.Error(err, "unable to get config", "context", kubeContext)
			os.Exit(1)
		}
		peerCluster, err := cluster.New(ctxConfig)
		if err != nil {
			setupLog.Error(err, "unable to create cluster", "config", ctxConfig)
			os.Exit(1)
		}
		go func() {
			if err := peerCluster.Start(ctx); err != nil {
				setupLog.Error(err, "problem running cluster", "context", kubeContext)
				os.Exit(1)
			}
		}()
		peerClusters[kubeContext] = &peerCluster
		go func() {
			pdbStateCollector := metrics.PDBStateCollector{
				Clusters:          peerClusters,
				Ctx:               ctx,
				MetricUpdateDelay: customDelayAsDuration,
			}
			if err := pdbStateCollector.CalculatePDBMetrics(); err != nil {
				setupLog.Error(err, "problem in metric producing", "context", kubeContext)
				os.Exit(1)
			}
		}()
	}
	restConfig, _ := config.GetConfigWithContext(kubeMainContext)
	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		Scheme:                 scheme,
		NewCache:               cache.BuilderWithOptions(cache.Options{SelectorsByObject: selectorsByObject}),
		CertDir:                certDir,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager", "context", kubeMainContext)
		os.Exit(1)
	}

	//go webhook.RunServe(webhook.CmdWebhook, []string{})
	hookServer := mgr.GetWebhookServer()
	setupLog.Info("registering webhooks to the webhook server")
	hookServer.Register("/validate-v1-pod", &sigswebhook.Admission{
		Handler: &webhook.PodValidator{Client: mgr.GetClient(), Clusters: peerClusters}},
	)

	if err = (&controllers.MultidcPodDisruptionBudgetReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "MultidcPodDisruptionBudget")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}
	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
