package controller

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kubecodriverv1alpha1 "github.com/codriverlabs/KubeCoDriver/api/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:rbac:groups=kubecodriver.codriverlabs.ai,resources=codrivertools,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubecodriver.codriverlabs.ai,resources=codrivertools/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubecodriver.codriverlabs.ai,resources=codrivertools/finalizers,verbs=update

type CoDriverToolReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile is part of the main kubernetes reconciliation loop
func (r *CoDriverToolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the CoDriverTool instance
	var toolConfig kubecodriverv1alpha1.CoDriverTool
	if err := r.Get(ctx, req.NamespacedName, &toolConfig); err != nil {
		logger.Error(err, "unable to fetch CoDriverTool")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.V(1).Info("Reconciling CoDriverTool", "name", toolConfig.Name, "tool", toolConfig.Spec.Name)

	// Update status to indicate validation
	now := metav1.Now()
	toolConfig.Status.LastValidated = &now
	toolConfig.Status.Phase = stringPtr("Ready")

	// Add condition
	condition := kubecodriverv1alpha1.CoDriverToolCondition{
		Type:               "Ready",
		Status:             "True",
		LastTransitionTime: now,
		Reason:             "ConfigurationValid",
		Message:            "CoDriverTool is valid and ready for use",
	}

	// Update or add the condition
	toolConfig.Status.Conditions = updateCondition(toolConfig.Status.Conditions, condition)

	if err := r.Status().Update(ctx, &toolConfig); err != nil {
		logger.Error(err, "failed to update CoDriverTool status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
}

// Helper function to update conditions
func updateCondition(conditions []kubecodriverv1alpha1.CoDriverToolCondition, newCondition kubecodriverv1alpha1.CoDriverToolCondition) []kubecodriverv1alpha1.CoDriverToolCondition {
	for i, condition := range conditions {
		if condition.Type == newCondition.Type {
			conditions[i] = newCondition
			return conditions
		}
	}
	return append(conditions, newCondition)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// SetupWithManager sets up the controller with the Manager.
func (r *CoDriverToolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubecodriverv1alpha1.CoDriverTool{}).
		Complete(r)
}
