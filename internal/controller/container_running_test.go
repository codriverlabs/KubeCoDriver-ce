package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsContainerRunning_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		pod           corev1.Pod
		containerName string
		expected      bool
	}{
		{
			name: "empty container name",
			pod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
			},
			containerName: "",
			expected:      false,
		},
		{
			name: "nil ephemeral container statuses",
			pod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Status: corev1.PodStatus{
					EphemeralContainerStatuses: nil,
				},
			},
			containerName: "test",
			expected:      false,
		},
		{
			name: "container with nil state",
			pod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Status: corev1.PodStatus{
					EphemeralContainerStatuses: []corev1.ContainerStatus{
						{
							Name:  "test",
							State: corev1.ContainerState{},
						},
					},
				},
			},
			containerName: "test",
			expected:      false,
		},
		{
			name: "container running with start time",
			pod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Spec: corev1.PodSpec{
					EphemeralContainers: []corev1.EphemeralContainer{
						{
							EphemeralContainerCommon: corev1.EphemeralContainerCommon{
								Name: "test",
							},
						},
					},
				},
				Status: corev1.PodStatus{
					EphemeralContainerStatuses: []corev1.ContainerStatus{
						{
							Name: "test",
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{
									StartedAt: metav1.Now(),
								},
							},
						},
					},
				},
			},
			containerName: "test",
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &CoDriverJobReconciler{}
			result := r.isContainerRunning(tt.pod, tt.containerName)
			assert.Equal(t, tt.expected, result)
		})
	}
}
