package controller

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubecodriverv1alpha1 "github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

func TestValidateNamespaceAccess(t *testing.T) {
	r := &CoDriverJobReconciler{}

	tests := []struct {
		name        string
		coDriverJob *kubecodriverv1alpha1.CoDriverJob
		toolConfig  *kubecodriverv1alpha1.CoDriverTool
		wantErr     bool
	}{
		{
			name: "no namespace restrictions - allow all",
			coDriverJob: &kubecodriverv1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "any-namespace",
				},
			},
			toolConfig: &kubecodriverv1alpha1.CoDriverTool{
				Spec: kubecodriverv1alpha1.CoDriverToolSpec{
					AllowedNamespaces: []string{},
				},
			},
			wantErr: false,
		},
		{
			name: "allowed namespace",
			coDriverJob: &kubecodriverv1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "production",
				},
			},
			toolConfig: &kubecodriverv1alpha1.CoDriverTool{
				Spec: kubecodriverv1alpha1.CoDriverToolSpec{
					AllowedNamespaces: []string{"production", "staging"},
				},
			},
			wantErr: false,
		},
		{
			name: "disallowed namespace",
			coDriverJob: &kubecodriverv1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "development",
				},
			},
			toolConfig: &kubecodriverv1alpha1.CoDriverTool{
				Spec: kubecodriverv1alpha1.CoDriverToolSpec{
					AllowedNamespaces: []string{"production", "staging"},
				},
			},
			wantErr: true,
		},
		{
			name: "nil allowed namespaces - allow all",
			coDriverJob: &kubecodriverv1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "any-namespace",
				},
			},
			toolConfig: &kubecodriverv1alpha1.CoDriverTool{
				Spec: kubecodriverv1alpha1.CoDriverToolSpec{
					AllowedNamespaces: nil,
				},
			},
			wantErr: false,
		},
		{
			name: "single allowed namespace - match",
			coDriverJob: &kubecodriverv1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "production",
				},
			},
			toolConfig: &kubecodriverv1alpha1.CoDriverTool{
				Spec: kubecodriverv1alpha1.CoDriverToolSpec{
					AllowedNamespaces: []string{"production"},
				},
			},
			wantErr: false,
		},
		{
			name: "single allowed namespace - no match",
			coDriverJob: &kubecodriverv1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "staging",
				},
			},
			toolConfig: &kubecodriverv1alpha1.CoDriverTool{
				Spec: kubecodriverv1alpha1.CoDriverToolSpec{
					AllowedNamespaces: []string{"production"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := r.validateNamespaceAccess(tt.coDriverJob, tt.toolConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateNamespaceAccess() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
