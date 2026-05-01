package controller

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubecodriverv1alpha1 "github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

func TestBuildCoDriverJobEnvVars_OutputModes(t *testing.T) {
	tests := []struct {
		name        string
		outputMode  string
		pvcSpec     *kubecodriverv1alpha1.PVCSpec
		checkEnvVar func(*testing.T, []corev1.EnvVar)
	}{
		{
			name:       "ephemeral mode",
			outputMode: "ephemeral",
			checkEnvVar: func(t *testing.T, envVars []corev1.EnvVar) {
				for _, env := range envVars {
					if env.Name == "OUTPUT_MODE" && env.Value != "ephemeral" {
						t.Errorf("expected OUTPUT_MODE=ephemeral, got %s", env.Value)
					}
				}
			},
		},
		{
			name:       "PVC mode",
			outputMode: "pvc",
			pvcSpec: &kubecodriverv1alpha1.PVCSpec{
				ClaimName: "test-pvc",
			},
			checkEnvVar: func(t *testing.T, envVars []corev1.EnvVar) {
				for _, env := range envVars {
					if env.Name == "OUTPUT_MODE" && env.Value != "pvc" {
						t.Errorf("expected OUTPUT_MODE=pvc, got %s", env.Value)
					}
				}
			},
		},
		{
			name:       "collector mode - basic env vars only",
			outputMode: "collector",
			checkEnvVar: func(t *testing.T, envVars []corev1.EnvVar) {
				// buildCoDriverJobEnvVars only sets OUTPUT_MODE
				// COLLECTOR_ENDPOINT and COLLECTOR_TOKEN are added in createEphemeralContainerForPod
				for _, env := range envVars {
					if env.Name == "OUTPUT_MODE" && env.Value != "collector" {
						t.Errorf("expected OUTPUT_MODE=collector, got %s", env.Value)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &CoDriverJobReconciler{}

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
						Name:     "perf",
						Duration: "30s",
					},
					Output: kubecodriverv1alpha1.OutputSpec{
						Mode: tt.outputMode,
						PVC:  tt.pvcSpec,
					},
				},
			}

			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
					Labels:    map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "app"}},
				},
			}

			envVars := r.buildCoDriverJobEnvVars(coDriverJob, pod)

			if tt.checkEnvVar != nil {
				tt.checkEnvVar(t, envVars)
			}
		})
	}
}

func TestBuildSecurityContext_AllCases(t *testing.T) {
	r := &CoDriverJobReconciler{}
	trueVal := true

	tests := []struct {
		name     string
		security kubecodriverv1alpha1.SecuritySpec
		check    func(*testing.T, *corev1.SecurityContext)
	}{
		{
			name:     "empty security spec - returns empty context",
			security: kubecodriverv1alpha1.SecuritySpec{},
			check: func(t *testing.T, sc *corev1.SecurityContext) {
				// Function always returns non-nil SecurityContext
				if sc == nil {
					t.Error("expected non-nil security context")
				}
				// But it should be empty
				if sc.Privileged != nil {
					t.Error("expected no privileged setting")
				}
				if sc.Capabilities != nil {
					t.Error("expected no capabilities")
				}
			},
		},
		{
			name: "privileged mode",
			security: kubecodriverv1alpha1.SecuritySpec{
				AllowPrivileged: &trueVal,
			},
			check: func(t *testing.T, sc *corev1.SecurityContext) {
				if sc == nil || sc.Privileged == nil || !*sc.Privileged {
					t.Error("expected privileged=true")
				}
			},
		},
		{
			name: "capabilities add and drop",
			security: kubecodriverv1alpha1.SecuritySpec{
				Capabilities: &kubecodriverv1alpha1.Capabilities{
					Add:  []string{"SYS_ADMIN", "SYS_PTRACE"},
					Drop: []string{"ALL"},
				},
			},
			check: func(t *testing.T, sc *corev1.SecurityContext) {
				if sc == nil || sc.Capabilities == nil {
					t.Fatal("expected capabilities to be set")
				}
				if len(sc.Capabilities.Add) != 2 {
					t.Errorf("expected 2 added capabilities, got %d", len(sc.Capabilities.Add))
				}
				if len(sc.Capabilities.Drop) != 1 {
					t.Errorf("expected 1 dropped capability, got %d", len(sc.Capabilities.Drop))
				}
			},
		},
		{
			name: "privileged and capabilities combined",
			security: kubecodriverv1alpha1.SecuritySpec{
				AllowPrivileged: &trueVal,
				Capabilities: &kubecodriverv1alpha1.Capabilities{
					Add: []string{"NET_ADMIN"},
				},
			},
			check: func(t *testing.T, sc *corev1.SecurityContext) {
				if sc == nil {
					t.Fatal("expected non-nil security context")
				}
				if sc.Privileged == nil || !*sc.Privileged {
					t.Error("expected privileged=true")
				}
				if sc.Capabilities == nil || len(sc.Capabilities.Add) != 1 {
					t.Error("expected 1 added capability")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := r.buildSecurityContext(tt.security)
			tt.check(t, sc)
		})
	}
}
