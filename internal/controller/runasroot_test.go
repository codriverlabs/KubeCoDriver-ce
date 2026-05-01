package controller

import (
	"testing"

	kubecodriverv1alpha1 "github.com/codriverlabs/KubeCoDriver/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

func TestRunAsRootSecurityContext(t *testing.T) {
	tests := []struct {
		name                string
		toolConfigSecurity  kubecodriverv1alpha1.SecuritySpec
		podSecurityContext  *corev1.PodSecurityContext
		containerSecContext *corev1.SecurityContext
		expectedUser        *int64
		expectedGroup       *int64
		expectedNonRoot     *bool
	}{
		{
			name: "runAsRoot enabled with container group",
			toolConfigSecurity: kubecodriverv1alpha1.SecuritySpec{
				RunAsRoot: ptr.To(true),
			},
			containerSecContext: &corev1.SecurityContext{
				RunAsUser:  ptr.To(int64(1001)),
				RunAsGroup: ptr.To(int64(1001)),
			},
			expectedUser:    ptr.To(int64(0)),
			expectedGroup:   ptr.To(int64(1001)),
			expectedNonRoot: ptr.To(false),
		},
		{
			name: "runAsRoot enabled with pod group",
			toolConfigSecurity: kubecodriverv1alpha1.SecuritySpec{
				RunAsRoot: ptr.To(true),
			},
			podSecurityContext: &corev1.PodSecurityContext{
				RunAsUser:  ptr.To(int64(2000)),
				RunAsGroup: ptr.To(int64(2000)),
			},
			expectedUser:    ptr.To(int64(0)),
			expectedGroup:   ptr.To(int64(2000)),
			expectedNonRoot: ptr.To(false),
		},
		{
			name: "runAsRoot enabled, container group overrides pod group",
			toolConfigSecurity: kubecodriverv1alpha1.SecuritySpec{
				RunAsRoot: ptr.To(true),
			},
			podSecurityContext: &corev1.PodSecurityContext{
				RunAsGroup: ptr.To(int64(2000)),
			},
			containerSecContext: &corev1.SecurityContext{
				RunAsGroup: ptr.To(int64(1001)),
			},
			expectedUser:    ptr.To(int64(0)),
			expectedGroup:   ptr.To(int64(1001)),
			expectedNonRoot: ptr.To(false),
		},
		{
			name: "runAsRoot disabled, normal inheritance",
			toolConfigSecurity: kubecodriverv1alpha1.SecuritySpec{
				RunAsRoot: ptr.To(false),
			},
			containerSecContext: &corev1.SecurityContext{
				RunAsUser:  ptr.To(int64(1001)),
				RunAsGroup: ptr.To(int64(1001)),
			},
			expectedUser:  ptr.To(int64(1001)),
			expectedGroup: ptr.To(int64(1001)),
		},
		{
			name:               "runAsRoot not set, normal inheritance",
			toolConfigSecurity: kubecodriverv1alpha1.SecuritySpec{},
			podSecurityContext: &corev1.PodSecurityContext{
				RunAsUser:    ptr.To(int64(2000)),
				RunAsGroup:   ptr.To(int64(2000)),
				RunAsNonRoot: ptr.To(true),
			},
			expectedUser:    ptr.To(int64(2000)),
			expectedGroup:   ptr.To(int64(2000)),
			expectedNonRoot: ptr.To(true),
		},
		{
			name: "runAsRoot enabled overrides pod runAsNonRoot: true",
			toolConfigSecurity: kubecodriverv1alpha1.SecuritySpec{
				RunAsRoot: ptr.To(true),
			},
			podSecurityContext: &corev1.PodSecurityContext{
				RunAsUser:    ptr.To(int64(1001)),
				RunAsGroup:   ptr.To(int64(1001)),
				RunAsNonRoot: ptr.To(true),
			},
			expectedUser:    ptr.To(int64(0)),
			expectedGroup:   ptr.To(int64(1001)),
			expectedNonRoot: ptr.To(false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "main"},
					},
					SecurityContext: tt.podSecurityContext,
				},
			}

			if tt.containerSecContext != nil {
				pod.Spec.Containers[0].SecurityContext = tt.containerSecContext
			}

			toolConfig := &kubecodriverv1alpha1.CoDriverTool{
				Spec: kubecodriverv1alpha1.CoDriverToolSpec{
					SecurityContext: tt.toolConfigSecurity,
				},
			}

			r := &CoDriverJobReconciler{}
			targetContainer := r.getTargetContainer(pod, nil)

			securityContext := r.buildSecurityContext(toolConfig.Spec.SecurityContext)

			runAsRoot := toolConfig.Spec.SecurityContext.RunAsRoot != nil &&
				*toolConfig.Spec.SecurityContext.RunAsRoot

			if runAsRoot {
				rootUser := int64(0)
				securityContext.RunAsUser = &rootUser
				runAsNonRootFalse := false
				securityContext.RunAsNonRoot = &runAsNonRootFalse

				if targetContainer != nil && targetContainer.SecurityContext != nil && targetContainer.SecurityContext.RunAsGroup != nil {
					securityContext.RunAsGroup = targetContainer.SecurityContext.RunAsGroup
				} else if pod.Spec.SecurityContext != nil && pod.Spec.SecurityContext.RunAsGroup != nil {
					securityContext.RunAsGroup = pod.Spec.SecurityContext.RunAsGroup
				}
			} else {
				if pod.Spec.SecurityContext != nil {
					if pod.Spec.SecurityContext.RunAsUser != nil {
						securityContext.RunAsUser = pod.Spec.SecurityContext.RunAsUser
					}
					if pod.Spec.SecurityContext.RunAsGroup != nil {
						securityContext.RunAsGroup = pod.Spec.SecurityContext.RunAsGroup
					}
					if pod.Spec.SecurityContext.RunAsNonRoot != nil {
						securityContext.RunAsNonRoot = pod.Spec.SecurityContext.RunAsNonRoot
					}
				}

				if targetContainer != nil && targetContainer.SecurityContext != nil {
					if targetContainer.SecurityContext.RunAsUser != nil {
						securityContext.RunAsUser = targetContainer.SecurityContext.RunAsUser
					}
					if targetContainer.SecurityContext.RunAsGroup != nil {
						securityContext.RunAsGroup = targetContainer.SecurityContext.RunAsGroup
					}
					if targetContainer.SecurityContext.RunAsNonRoot != nil {
						securityContext.RunAsNonRoot = targetContainer.SecurityContext.RunAsNonRoot
					}
				}
			}

			if !int64PtrEqual(securityContext.RunAsUser, tt.expectedUser) {
				t.Errorf("Expected runAsUser %v, got %v", tt.expectedUser, securityContext.RunAsUser)
			}
			if !int64PtrEqual(securityContext.RunAsGroup, tt.expectedGroup) {
				t.Errorf("Expected runAsGroup %v, got %v", tt.expectedGroup, securityContext.RunAsGroup)
			}
			if !boolPtrEqual(securityContext.RunAsNonRoot, tt.expectedNonRoot) {
				t.Errorf("Expected runAsNonRoot %v, got %v", tt.expectedNonRoot, securityContext.RunAsNonRoot)
			}
		})
	}
}

func int64PtrEqual(a, b *int64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func boolPtrEqual(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
