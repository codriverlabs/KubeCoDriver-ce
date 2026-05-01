package controller

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestFindPVCVolumeName(t *testing.T) {
	r := &CoDriverJobReconciler{}

	tests := []struct {
		name      string
		pod       corev1.Pod
		claimName string
		want      string
	}{
		{
			name: "PVC found",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "data-volume",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "my-pvc",
								},
							},
						},
					},
				},
			},
			claimName: "my-pvc",
			want:      "data-volume",
		},
		{
			name: "PVC not found - returns default",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "data-volume",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "other-pvc",
								},
							},
						},
					},
				},
			},
			claimName: "my-pvc",
			want:      "profiling-storage",
		},
		{
			name: "multiple volumes, PVC in middle",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "config-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{},
							},
						},
						{
							Name: "data-volume",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "target-pvc",
								},
							},
						},
						{
							Name: "secret-volume",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{},
							},
						},
					},
				},
			},
			claimName: "target-pvc",
			want:      "data-volume",
		},
		{
			name: "pod with no volumes",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{},
				},
			},
			claimName: "my-pvc",
			want:      "profiling-storage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.findPVCVolumeName(tt.pod, tt.claimName)
			if got != tt.want {
				t.Errorf("findPVCVolumeName() = %v, want %v", got, tt.want)
			}
		})
	}
}
