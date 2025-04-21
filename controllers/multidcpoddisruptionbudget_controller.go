package controllers

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/log"

	//"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	//"sigs.k8s.io/controller-runtime/pkg/log"

	multidccrd "k8s.tochka.com/multidc-pdb-operator/api/v1"
)

// var (
// MultidcPodDisruptionIndexKey = ".spec.selector"
//apiGVStr                     = multidccrd.GroupVersion.String()
// )

// MultidcPodDisruptionBudgetReconciler reconciles a MultidcPodDisruptionBudget object
type MultidcPodDisruptionBudgetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps,resources=replicasets;deployments;statefulsets,verbs=get;list;watch
//+kubebuilder:rbac:groups=k8s.tochka.com,resources=multidcpoddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k8s.tochka.com,resources=multidcpoddisruptionbudgets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k8s.tochka.com,resources=multidcpoddisruptionbudgets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MultidcPodDisruptionBudget object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *MultidcPodDisruptionBudgetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	rlog := log.FromContext(ctx)
	rlog.Info(fmt.Sprintf("Reconcile req: %+v", req))
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MultidcPodDisruptionBudgetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&multidccrd.MultidcPodDisruptionBudget{}).
		Complete(r)
}
