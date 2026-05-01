package controller

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubecodriverv1alpha1 "github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

func TestBuildCoDriverJobEnvVars(t *testing.T) {
	r := &CoDriverJobReconciler{}

	tests := []struct {
		name        string
		coDriverJob *kubecodriverv1alpha1.CoDriverJob
		targetPod   corev1.Pod
		wantEnvs    map[string]string
	}{
		{
			name: "basic configuration with app label",
			coDriverJob: &kubecodriverv1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-profile",
					Namespace: "default",
				},
				Spec: kubecodriverv1alpha1.CoDriverJobSpec{
					Targets: kubecodriverv1alpha1.TargetSpec{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "nginx",
							},
						},
					},
					Tool: kubecodriverv1alpha1.ToolSpec{
						Name:     "aperf",
						Duration: "30s",
					},
					Output: kubecodriverv1alpha1.OutputSpec{
						Mode: "ephemeral",
					},
				},
			},
			targetPod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nginx-pod",
					Namespace: "default",
					Labels: map[string]string{
						"app": "nginx",
					},
				},
			},
			wantEnvs: map[string]string{
				"PROFILER_TOOL":       "aperf",
				"PROFILER_DURATION":   "30s",
				"TARGET_POD_NAME":     "nginx-pod",
				"TARGET_NAMESPACE":    "default",
				"POD_MATCHING_LABELS": "app-nginx",
				"OUTPUT_MODE":         "ephemeral",
			},
		},
		{
			name: "environment label",
			coDriverJob: &kubecodriverv1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "prod-profile",
					Namespace: "production",
				},
				Spec: kubecodriverv1alpha1.CoDriverJobSpec{
					Targets: kubecodriverv1alpha1.TargetSpec{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"env": "production",
							},
						},
					},
					Tool: kubecodriverv1alpha1.ToolSpec{
						Name:     "perf",
						Duration: "60s",
					},
					Output: kubecodriverv1alpha1.OutputSpec{
						Mode: "pvc",
					},
				},
			},
			targetPod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "api-pod",
					Namespace: "production",
					Labels: map[string]string{
						"env": "production",
					},
				},
			},
			wantEnvs: map[string]string{
				"PROFILER_TOOL":       "perf",
				"PROFILER_DURATION":   "60s",
				"TARGET_POD_NAME":     "api-pod",
				"TARGET_NAMESPACE":    "production",
				"POD_MATCHING_LABELS": "env-production",
				"OUTPUT_MODE":         "pvc",
			},
		},
		{
			name: "no matching labels - defaults to unknown",
			coDriverJob: &kubecodriverv1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-profile",
					Namespace: "default",
				},
				Spec: kubecodriverv1alpha1.CoDriverJobSpec{
					Targets: kubecodriverv1alpha1.TargetSpec{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "nginx",
							},
						},
					},
					Tool: kubecodriverv1alpha1.ToolSpec{
						Name:     "aperf",
						Duration: "30s",
					},
					Output: kubecodriverv1alpha1.OutputSpec{
						Mode: "ephemeral",
					},
				},
			},
			targetPod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "other-pod",
					Namespace: "default",
					Labels: map[string]string{
						"app": "apache",
					},
				},
			},
			wantEnvs: map[string]string{
				"PROFILER_TOOL":       "aperf",
				"PROFILER_DURATION":   "30s",
				"TARGET_POD_NAME":     "other-pod",
				"TARGET_NAMESPACE":    "default",
				"POD_MATCHING_LABELS": "unknown",
				"OUTPUT_MODE":         "ephemeral",
			},
		},
		{
			name: "custom tier label",
			coDriverJob: &kubecodriverv1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "backend-profile",
					Namespace: "default",
				},
				Spec: kubecodriverv1alpha1.CoDriverJobSpec{
					Targets: kubecodriverv1alpha1.TargetSpec{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"tier": "backend",
							},
						},
					},
					Tool: kubecodriverv1alpha1.ToolSpec{
						Name:     "aperf",
						Duration: "45s",
					},
					Output: kubecodriverv1alpha1.OutputSpec{
						Mode: "collector",
					},
				},
			},
			targetPod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "backend-pod",
					Namespace: "default",
					Labels: map[string]string{
						"tier": "backend",
					},
				},
			},
			wantEnvs: map[string]string{
				"PROFILER_TOOL":       "aperf",
				"PROFILER_DURATION":   "45s",
				"TARGET_POD_NAME":     "backend-pod",
				"TARGET_NAMESPACE":    "default",
				"POD_MATCHING_LABELS": "tier-backend",
				"OUTPUT_MODE":         "collector",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVars := r.buildCoDriverJobEnvVars(tt.coDriverJob, tt.targetPod)

			// Convert to map for easier comparison
			gotEnvs := make(map[string]string)
			for _, env := range envVars {
				gotEnvs[env.Name] = env.Value
			}

			// Check all expected env vars
			for key, wantValue := range tt.wantEnvs {
				gotValue, exists := gotEnvs[key]
				if !exists {
					t.Errorf("buildCoDriverJobEnvVars() missing env var %v", key)
					continue
				}
				if gotValue != wantValue {
					t.Errorf("buildCoDriverJobEnvVars() env var %v = %v, want %v", key, gotValue, wantValue)
				}
			}

			// Verify POD_MATCHING_LABELS is always present
			if _, exists := gotEnvs["POD_MATCHING_LABELS"]; !exists {
				t.Error("buildCoDriverJobEnvVars() missing POD_MATCHING_LABELS env var")
			}
		})
	}
}

func TestBuildCoDriverJobEnvVars_WithPVCPath(t *testing.T) {
	r := &CoDriverJobReconciler{}

	pvcPath := "/custom/path"
	coDriverJob := &kubecodriverv1alpha1.CoDriverJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-profile",
			Namespace: "default",
		},
		Spec: kubecodriverv1alpha1.CoDriverJobSpec{
			Targets: kubecodriverv1alpha1.TargetSpec{
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "test",
					},
				},
			},
			Tool: kubecodriverv1alpha1.ToolSpec{
				Name:     "aperf",
				Duration: "30s",
			},
			Output: kubecodriverv1alpha1.OutputSpec{
				Mode: "pvc",
				PVC: &kubecodriverv1alpha1.PVCSpec{
					ClaimName: "test-pvc",
					Path:      &pvcPath,
				},
			},
		},
	}

	targetPod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Labels: map[string]string{
				"app": "test",
			},
		},
	}

	envVars := r.buildCoDriverJobEnvVars(coDriverJob, targetPod)

	// Convert to map
	gotEnvs := make(map[string]string)
	for _, env := range envVars {
		gotEnvs[env.Name] = env.Value
	}

	// Verify PVC_PATH is set
	if gotEnvs["PVC_PATH"] != pvcPath {
		t.Errorf("buildCoDriverJobEnvVars() PVC_PATH = %v, want %v", gotEnvs["PVC_PATH"], pvcPath)
	}
}

func TestBuildCoDriverJobEnvVars_ToolArgs(t *testing.T) {
	r := &CoDriverJobReconciler{}

	tests := []struct {
		name     string
		args     []string
		wantEnvs map[string]string
	}{
		{
			name: "valid string args",
			args: []string{"--frequency", "99", "--duration", "30"},
			wantEnvs: map[string]string{
				"TOOL_ARGS":  "--frequency 99 --duration 30",
				"TOOL_ARG_0": "--frequency",
				"TOOL_ARG_1": "99",
				"TOOL_ARG_2": "--duration",
				"TOOL_ARG_3": "30",
			},
		},
		{
			name: "single arg",
			args: []string{"suspend"},
			wantEnvs: map[string]string{
				"TOOL_ARGS":  "suspend",
				"TOOL_ARG_0": "suspend",
			},
		},
		{
			name:     "empty args",
			args:     []string{},
			wantEnvs: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coDriverJob := &kubecodriverv1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-profile",
					Namespace: "default",
				},
				Spec: kubecodriverv1alpha1.CoDriverJobSpec{
					Targets: kubecodriverv1alpha1.TargetSpec{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "test"},
						},
					},
					Tool: kubecodriverv1alpha1.ToolSpec{
						Name:     "aperf",
						Duration: "30s",
						Args:     tt.args,
					},
					Output: kubecodriverv1alpha1.OutputSpec{
						Mode: "ephemeral",
					},
				},
			}

			targetPod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
					Labels:    map[string]string{"app": "test"},
				},
			}

			envVars := r.buildCoDriverJobEnvVars(coDriverJob, targetPod)

			gotEnvs := make(map[string]string)
			for _, env := range envVars {
				gotEnvs[env.Name] = env.Value
			}

			for key, wantValue := range tt.wantEnvs {
				gotValue, exists := gotEnvs[key]
				if !exists {
					t.Errorf("missing env var %v", key)
					continue
				}
				if gotValue != wantValue {
					t.Errorf("env var %v = %v, want %v", key, gotValue, wantValue)
				}
			}
		})
	}
}

func TestBuildCoDriverJobEnvVars_NilArgs(t *testing.T) {
	r := &CoDriverJobReconciler{}

	coDriverJob := &kubecodriverv1alpha1.CoDriverJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-profile",
			Namespace: "default",
		},
		Spec: kubecodriverv1alpha1.CoDriverJobSpec{
			Targets: kubecodriverv1alpha1.TargetSpec{
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
			},
			Tool: kubecodriverv1alpha1.ToolSpec{
				Name:     "aperf",
				Duration: "30s",
				Args:     nil,
			},
			Output: kubecodriverv1alpha1.OutputSpec{
				Mode: "ephemeral",
			},
		},
	}

	targetPod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Labels:    map[string]string{"app": "test"},
		},
	}

	// Should not panic with invalid JSON
	envVars := r.buildCoDriverJobEnvVars(coDriverJob, targetPod)

	// Verify basic env vars are still set
	gotEnvs := make(map[string]string)
	for _, env := range envVars {
		gotEnvs[env.Name] = env.Value
	}

	if gotEnvs["PROFILER_TOOL"] != "aperf" {
		t.Error("basic env vars should still be set with invalid JSON")
	}
}
