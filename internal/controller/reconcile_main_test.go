package controller

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	kubecodriverv1alpha1 "github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

func TestReconcile_CoDriverJobNotFoundReturnsEmpty(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, kubecodriverv1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	k8sClient := k8sfake.NewSimpleClientset()
	reconciler := NewCoDriverJobReconciler(fakeClient, scheme, k8sClient)

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent-tool",
			Namespace: "default",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
}

func TestReconcile_DeletionHandling(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, kubecodriverv1alpha1.AddToScheme(scheme))
	require.NoError(t, corev1.AddToScheme(scheme))

	coDriverJob := &kubecodriverv1alpha1.CoDriverJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-tool",
			Namespace:         "default",
			DeletionTimestamp: &metav1.Time{Time: time.Now()},
			Finalizers:        []string{"kubecodriver.codriverlabs.ai/codriverjob-cleanup"},
		},
		Spec: kubecodriverv1alpha1.CoDriverJobSpec{
			Tool: kubecodriverv1alpha1.ToolSpec{
				Name: "aperf",
			},
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(coDriverJob).
		Build()

	k8sClient := k8sfake.NewSimpleClientset()
	reconciler := NewCoDriverJobReconciler(fakeClient, scheme, k8sClient)

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      coDriverJob.Name,
			Namespace: coDriverJob.Namespace,
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
}
