package controller

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubecodriverv1alpha1 "github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

func TestHandleDeletion(t *testing.T) {
	r := &CoDriverJobReconciler{}
	ctx := context.Background()

	tests := []struct {
		name        string
		coDriverJob *kubecodriverv1alpha1.CoDriverJob
		wantErr     bool
	}{
		{
			name: "no finalizer - returns nil",
			coDriverJob: &kubecodriverv1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-tool",
					Namespace:  "default",
					Finalizers: []string{},
				},
			},
			wantErr: false,
		},
		{
			name: "with finalizer - cleanup executed",
			coDriverJob: &kubecodriverv1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-tool",
					Namespace:  "default",
					Finalizers: []string{"kubecodriver.codriverlabs.ai/finalizer"},
				},
			},
			wantErr: false,
		},
		{
			name: "with other finalizer - returns nil",
			coDriverJob: &kubecodriverv1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-tool",
					Namespace:  "default",
					Finalizers: []string{"other.io/finalizer"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := r.handleDeletion(ctx, tt.coDriverJob)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleDeletion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
