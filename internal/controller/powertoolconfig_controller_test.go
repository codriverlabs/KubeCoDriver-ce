package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	kubecodriverv1alpha1 "github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

func TestCoDriverToolReconciler_Reconcile(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = kubecodriverv1alpha1.AddToScheme(scheme)

	config := &kubecodriverv1alpha1.CoDriverTool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "default",
		},
		Spec: kubecodriverv1alpha1.CoDriverToolSpec{
			Name:  "test-tool",
			Image: "test-image:latest",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(config).
		WithStatusSubresource(config).
		Build()

	reconciler := &CoDriverToolReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-config",
			Namespace: "default",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify status was updated
	var updated kubecodriverv1alpha1.CoDriverTool
	err = fakeClient.Get(context.Background(), req.NamespacedName, &updated)
	assert.NoError(t, err)
	assert.NotNil(t, updated.Status.LastValidated)
	assert.NotNil(t, updated.Status.Phase)
	assert.Equal(t, "Ready", *updated.Status.Phase)
	assert.Len(t, updated.Status.Conditions, 1)
	assert.Equal(t, "Ready", updated.Status.Conditions[0].Type)
}

func TestCoDriverToolReconciler_NotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = kubecodriverv1alpha1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	reconciler := &CoDriverToolReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent",
			Namespace: "default",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
}

func TestUpdateCondition(t *testing.T) {
	now := metav1.Now()

	tests := []struct {
		name       string
		existing   []kubecodriverv1alpha1.CoDriverToolCondition
		new        kubecodriverv1alpha1.CoDriverToolCondition
		expectLen  int
		expectType string
	}{
		{
			name:     "add new condition",
			existing: []kubecodriverv1alpha1.CoDriverToolCondition{},
			new: kubecodriverv1alpha1.CoDriverToolCondition{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: now,
			},
			expectLen:  1,
			expectType: "Ready",
		},
		{
			name: "update existing condition",
			existing: []kubecodriverv1alpha1.CoDriverToolCondition{
				{
					Type:               "Ready",
					Status:             "False",
					LastTransitionTime: now,
				},
			},
			new: kubecodriverv1alpha1.CoDriverToolCondition{
				Type:               "Ready",
				Status:             "True",
				LastTransitionTime: now,
			},
			expectLen:  1,
			expectType: "Ready",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := updateCondition(tt.existing, tt.new)
			assert.Len(t, result, tt.expectLen)
			assert.Equal(t, tt.expectType, result[0].Type)
		})
	}
}

func TestStringPtr(t *testing.T) {
	s := "test"
	ptr := stringPtr(s)
	assert.NotNil(t, ptr)
	assert.Equal(t, "test", *ptr)
}
