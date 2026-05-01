package controller

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	kubecodriverv1alpha1 "github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

func TestBuildSecurityContext(t *testing.T) {
	r := &CoDriverJobReconciler{}

	tests := []struct {
		name         string
		securitySpec kubecodriverv1alpha1.SecuritySpec
		wantCheck    func(*corev1.SecurityContext) error
	}{
		{
			name: "privileged mode",
			securitySpec: kubecodriverv1alpha1.SecuritySpec{
				AllowPrivileged: boolPtr(true),
			},
			wantCheck: func(sc *corev1.SecurityContext) error {
				if sc.Privileged == nil || !*sc.Privileged {
					t.Error("expected Privileged=true")
				}
				return nil
			},
		},
		{
			name: "add capabilities",
			securitySpec: kubecodriverv1alpha1.SecuritySpec{
				Capabilities: &kubecodriverv1alpha1.Capabilities{
					Add: []string{"SYS_ADMIN", "NET_ADMIN"},
				},
			},
			wantCheck: func(sc *corev1.SecurityContext) error {
				if sc.Capabilities == nil {
					t.Error("expected Capabilities to be set")
					return nil
				}
				if len(sc.Capabilities.Add) != 2 {
					t.Errorf("expected 2 capabilities, got %d", len(sc.Capabilities.Add))
				}
				return nil
			},
		},
		{
			name: "drop capabilities",
			securitySpec: kubecodriverv1alpha1.SecuritySpec{
				Capabilities: &kubecodriverv1alpha1.Capabilities{
					Drop: []string{"ALL", "CHOWN"},
				},
			},
			wantCheck: func(sc *corev1.SecurityContext) error {
				if sc.Capabilities == nil {
					t.Error("expected Capabilities to be set")
					return nil
				}
				if len(sc.Capabilities.Drop) != 2 {
					t.Errorf("expected 2 drop capabilities, got %d", len(sc.Capabilities.Drop))
				}
				return nil
			},
		},
		{
			name: "both add and drop capabilities",
			securitySpec: kubecodriverv1alpha1.SecuritySpec{
				Capabilities: &kubecodriverv1alpha1.Capabilities{
					Add:  []string{"SYS_ADMIN"},
					Drop: []string{"ALL"},
				},
			},
			wantCheck: func(sc *corev1.SecurityContext) error {
				if sc.Capabilities == nil {
					t.Error("expected Capabilities to be set")
					return nil
				}
				if len(sc.Capabilities.Add) != 1 {
					t.Errorf("expected 1 add capability, got %d", len(sc.Capabilities.Add))
				}
				if len(sc.Capabilities.Drop) != 1 {
					t.Errorf("expected 1 drop capability, got %d", len(sc.Capabilities.Drop))
				}
				return nil
			},
		},
		{
			name:         "empty security spec",
			securitySpec: kubecodriverv1alpha1.SecuritySpec{},
			wantCheck: func(sc *corev1.SecurityContext) error {
				if sc.Privileged != nil {
					t.Error("expected Privileged to be nil")
				}
				if sc.Capabilities != nil {
					t.Error("expected Capabilities to be nil")
				}
				return nil
			},
		},
		{
			name: "nil capabilities",
			securitySpec: kubecodriverv1alpha1.SecuritySpec{
				Capabilities: nil,
			},
			wantCheck: func(sc *corev1.SecurityContext) error {
				if sc.Capabilities != nil {
					t.Error("expected Capabilities to be nil")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.buildSecurityContext(tt.securitySpec)
			if got == nil {
				t.Fatal("buildSecurityContext() returned nil")
			}
			tt.wantCheck(got)
		})
	}
}
