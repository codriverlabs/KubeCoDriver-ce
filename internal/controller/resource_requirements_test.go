package controller

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	kubecodriverv1alpha1 "github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

func TestBuildResourceRequirements(t *testing.T) {
	reconciler := &CoDriverJobReconciler{}

	tests := []struct {
		name       string
		resources  *kubecodriverv1alpha1.ResourceSpec
		wantReqs   corev1.ResourceList
		wantLimits corev1.ResourceList
	}{
		{
			name:       "nil resources",
			resources:  nil,
			wantReqs:   nil,
			wantLimits: nil,
		},
		{
			name: "requests only",
			resources: &kubecodriverv1alpha1.ResourceSpec{
				Requests: &kubecodriverv1alpha1.ResourceList{
					CPU:    strPtr("100m"),
					Memory: strPtr("64Mi"),
				},
			},
			wantReqs: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("64Mi"),
			},
			wantLimits: nil,
		},
		{
			name: "limits only",
			resources: &kubecodriverv1alpha1.ResourceSpec{
				Limits: &kubecodriverv1alpha1.ResourceList{
					CPU:    strPtr("1000m"),
					Memory: strPtr("512Mi"),
				},
			},
			wantReqs: nil,
			wantLimits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1000m"),
				corev1.ResourceMemory: resource.MustParse("512Mi"),
			},
		},
		{
			name: "both requests and limits",
			resources: &kubecodriverv1alpha1.ResourceSpec{
				Requests: &kubecodriverv1alpha1.ResourceList{
					CPU:    strPtr("100m"),
					Memory: strPtr("64Mi"),
				},
				Limits: &kubecodriverv1alpha1.ResourceList{
					CPU:    strPtr("1000m"),
					Memory: strPtr("512Mi"),
				},
			},
			wantReqs: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("64Mi"),
			},
			wantLimits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1000m"),
				corev1.ResourceMemory: resource.MustParse("512Mi"),
			},
		},
		{
			name: "cpu only",
			resources: &kubecodriverv1alpha1.ResourceSpec{
				Requests: &kubecodriverv1alpha1.ResourceList{
					CPU: strPtr("200m"),
				},
				Limits: &kubecodriverv1alpha1.ResourceList{
					CPU: strPtr("500m"),
				},
			},
			wantReqs: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("200m"),
			},
			wantLimits: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("500m"),
			},
		},
		{
			name: "memory only",
			resources: &kubecodriverv1alpha1.ResourceSpec{
				Requests: &kubecodriverv1alpha1.ResourceList{
					Memory: strPtr("128Mi"),
				},
				Limits: &kubecodriverv1alpha1.ResourceList{
					Memory: strPtr("256Mi"),
				},
			},
			wantReqs: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("128Mi"),
			},
			wantLimits: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toolConfig := &kubecodriverv1alpha1.CoDriverTool{
				Spec: kubecodriverv1alpha1.CoDriverToolSpec{
					Resources: tt.resources,
				},
			}

			got := reconciler.buildResourceRequirements(toolConfig)

			if !resourceListEqual(got.Requests, tt.wantReqs) {
				t.Errorf("buildResourceRequirements() requests = %v, want %v", got.Requests, tt.wantReqs)
			}

			if !resourceListEqual(got.Limits, tt.wantLimits) {
				t.Errorf("buildResourceRequirements() limits = %v, want %v", got.Limits, tt.wantLimits)
			}
		})
	}
}

func resourceListEqual(a, b corev1.ResourceList) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || !v.Equal(bv) {
			return false
		}
	}
	return true
}

func strPtr(s string) *string {
	return &s
}
