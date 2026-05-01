package controller

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestIsContainerRunning(t *testing.T) {
	r := &CoDriverJobReconciler{}

	tests := []struct {
		name          string
		pod           *corev1.Pod
		containerName string
		want          bool
	}{
		{
			name: "container running",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					EphemeralContainers: []corev1.EphemeralContainer{
						{EphemeralContainerCommon: corev1.EphemeralContainerCommon{Name: "profiler"}},
					},
				},
				Status: corev1.PodStatus{
					EphemeralContainerStatuses: []corev1.ContainerStatus{
						{
							Name: "profiler",
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{},
							},
						},
					},
				},
			},
			containerName: "profiler",
			want:          true,
		},
		{
			name: "container not found",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					EphemeralContainerStatuses: []corev1.ContainerStatus{},
				},
			},
			containerName: "profiler",
			want:          false,
		},
		{
			name: "container terminated",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					EphemeralContainerStatuses: []corev1.ContainerStatus{
						{
							Name: "profiler",
							State: corev1.ContainerState{
								Terminated: &corev1.ContainerStateTerminated{
									ExitCode: 0,
								},
							},
						},
					},
				},
			},
			containerName: "profiler",
			want:          false,
		},
		{
			name: "container waiting",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					EphemeralContainerStatuses: []corev1.ContainerStatus{
						{
							Name: "profiler",
							State: corev1.ContainerState{
								Waiting: &corev1.ContainerStateWaiting{
									Reason: "ContainerCreating",
								},
							},
						},
					},
				},
			},
			containerName: "profiler",
			want:          false,
		},
		{
			name: "pod with no ephemeral containers",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					EphemeralContainerStatuses: nil,
				},
			},
			containerName: "profiler",
			want:          false,
		},
		{
			name: "multiple containers, target running",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					EphemeralContainers: []corev1.EphemeralContainer{
						{EphemeralContainerCommon: corev1.EphemeralContainerCommon{Name: "other"}},
						{EphemeralContainerCommon: corev1.EphemeralContainerCommon{Name: "profiler"}},
					},
				},
				Status: corev1.PodStatus{
					EphemeralContainerStatuses: []corev1.ContainerStatus{
						{
							Name: "other",
							State: corev1.ContainerState{
								Terminated: &corev1.ContainerStateTerminated{},
							},
						},
						{
							Name: "profiler",
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{},
							},
						},
					},
				},
			},
			containerName: "profiler",
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.isContainerRunning(*tt.pod, tt.containerName)
			if got != tt.want {
				t.Errorf("isContainerRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}
