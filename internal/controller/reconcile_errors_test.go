package controller

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kubecodriverv1alpha1 "github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

func TestReconcile_CoDriverJobNotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = kubecodriverv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	r := &CoDriverJobReconciler{
		Client: client,
		Scheme: scheme,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(context.Background(), req)

	if err != nil {
		t.Errorf("expected no error for nonexistent CoDriverJob, got %v", err)
	}

	if result.Requeue {
		t.Error("expected no requeue for nonexistent CoDriverJob")
	}
}

func TestReconcile_ToolConfigNotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = kubecodriverv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	coDriverJob := &kubecodriverv1alpha1.CoDriverJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-tool",
			Namespace: "default",
		},
		Spec: kubecodriverv1alpha1.CoDriverJobSpec{
			Targets: kubecodriverv1alpha1.TargetSpec{
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
			},
			Tool: kubecodriverv1alpha1.ToolSpec{
				Name:     "nonexistent-tool",
				Duration: "30s",
			},
			Output: kubecodriverv1alpha1.OutputSpec{
				Mode: "ephemeral",
			},
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(coDriverJob).
		Build()

	r := &CoDriverJobReconciler{
		Client: client,
		Scheme: scheme,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-tool",
			Namespace: "default",
		},
	}

	_, err := r.Reconcile(context.Background(), req)

	if err == nil {
		t.Error("expected error for nonexistent ToolConfig")
	}
}

func TestReconcile_NoMatchingPods(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = kubecodriverv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	coDriverJob := &kubecodriverv1alpha1.CoDriverJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-tool",
			Namespace: "default",
		},
		Spec: kubecodriverv1alpha1.CoDriverJobSpec{
			Targets: kubecodriverv1alpha1.TargetSpec{
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "nonexistent"},
				},
			},
			Tool: kubecodriverv1alpha1.ToolSpec{
				Name:     "perf",
				Duration: "30s",
			},
			Output: kubecodriverv1alpha1.OutputSpec{
				Mode: "ephemeral",
			},
		},
	}

	toolConfig := &kubecodriverv1alpha1.CoDriverTool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "perf-config", // Must match {toolName}-config pattern
			Namespace: "kubecodriver-system",
		},
		Spec: kubecodriverv1alpha1.CoDriverToolSpec{
			Image:           "test-image:latest",
			SecurityContext: kubecodriverv1alpha1.SecuritySpec{},
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(coDriverJob, toolConfig).
		WithStatusSubresource(coDriverJob).
		Build()

	r := &CoDriverJobReconciler{
		Client: client,
		Scheme: scheme,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-tool",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(context.Background(), req)

	if err != nil {
		t.Errorf("expected no error for no matching pods, got %v", err)
	}

	// Should requeue to check again later
	if !result.Requeue && result.RequeueAfter == 0 {
		t.Error("expected requeue for no matching pods")
	}
}
