package controller

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	kubecodriverv1alpha1 "github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

func TestGetToolConfig(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = kubecodriverv1alpha1.AddToScheme(scheme)

	tests := []struct {
		name        string
		toolName    string
		configs     []kubecodriverv1alpha1.CoDriverTool
		expectFound bool
		expectError bool
	}{
		{
			name:     "config found in kubecodriver-system",
			toolName: "perf",
			configs: []kubecodriverv1alpha1.CoDriverTool{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "perf-config",
						Namespace: "kubecodriver-system",
					},
					Spec: kubecodriverv1alpha1.CoDriverToolSpec{
						Image: "test:latest",
					},
				},
			},
			expectFound: true,
			expectError: false,
		},
		{
			name:     "config found in default",
			toolName: "strace",
			configs: []kubecodriverv1alpha1.CoDriverTool{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "strace-config",
						Namespace: "default",
					},
					Spec: kubecodriverv1alpha1.CoDriverToolSpec{
						Image: "test:latest",
					},
				},
			},
			expectFound: true,
			expectError: false,
		},
		{
			name:        "config not found",
			toolName:    "nonexistent",
			configs:     []kubecodriverv1alpha1.CoDriverTool{},
			expectFound: false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objects := []runtime.Object{}
			for i := range tt.configs {
				objects = append(objects, &tt.configs[i])
			}

			client := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objects...).
				Build()

			r := &CoDriverJobReconciler{
				Client: client,
				Scheme: scheme,
			}

			config, err := r.getToolConfig(context.Background(), tt.toolName)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.expectFound && config == nil {
				t.Error("expected config to be found, got nil")
			}

			if !tt.expectFound && config != nil {
				t.Error("expected config to be nil, got non-nil")
			}
		})
	}
}
