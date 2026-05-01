package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestGetTargetContainer_SingleContainer(t *testing.T) {
	r := &CoDriverJobReconciler{}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "main"},
			},
		},
	}

	// No container specified - should return first
	container := r.getTargetContainer(pod, nil)
	assert.NotNil(t, container)
	assert.Equal(t, "main", container.Name)
}

func TestGetTargetContainer_MultiContainer_Specified(t *testing.T) {
	r := &CoDriverJobReconciler{}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "sidecar"},
				{Name: "main-app"},
			},
		},
	}

	targetName := "main-app"
	container := r.getTargetContainer(pod, &targetName)
	assert.NotNil(t, container)
	assert.Equal(t, "main-app", container.Name)
}

func TestGetTargetContainer_MultiContainer_NotSpecified(t *testing.T) {
	r := &CoDriverJobReconciler{}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "sidecar"},
				{Name: "main-app"},
			},
		},
	}

	// No container specified - should return first
	container := r.getTargetContainer(pod, nil)
	assert.NotNil(t, container)
	assert.Equal(t, "sidecar", container.Name)
}

func TestGetTargetContainer_NotFound_FallbackToFirst(t *testing.T) {
	r := &CoDriverJobReconciler{}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "sidecar"},
				{Name: "main-app"},
			},
		},
	}

	targetName := "nonexistent"
	container := r.getTargetContainer(pod, &targetName)
	assert.NotNil(t, container)
	assert.Equal(t, "sidecar", container.Name) // Falls back to first
}

func TestGetTargetContainer_EmptyPod(t *testing.T) {
	r := &CoDriverJobReconciler{}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{},
		},
	}

	container := r.getTargetContainer(pod, nil)
	assert.Nil(t, container)
}

func TestGetTargetContainer_EmptyString(t *testing.T) {
	r := &CoDriverJobReconciler{}

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "main"},
			},
		},
	}

	emptyString := ""
	container := r.getTargetContainer(pod, &emptyString)
	assert.NotNil(t, container)
	assert.Equal(t, "main", container.Name) // Empty string treated as nil
}

func TestSecurityContextInheritance_TargetContainer(t *testing.T) {
	user1 := int64(2000)
	user2 := int64(1001)

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "sidecar",
					SecurityContext: &corev1.SecurityContext{
						RunAsUser: &user1,
					},
				},
				{
					Name: "main-app",
					SecurityContext: &corev1.SecurityContext{
						RunAsUser: &user2,
					},
				},
			},
		},
	}

	r := &CoDriverJobReconciler{}
	targetName := "main-app"
	container := r.getTargetContainer(pod, &targetName)

	assert.NotNil(t, container)
	assert.NotNil(t, container.SecurityContext)
	assert.NotNil(t, container.SecurityContext.RunAsUser)
	assert.Equal(t, int64(1001), *container.SecurityContext.RunAsUser)
}

func TestSecurityContextInheritance_FirstContainer(t *testing.T) {
	user1 := int64(2000)
	user2 := int64(1001)

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "sidecar",
					SecurityContext: &corev1.SecurityContext{
						RunAsUser: &user1,
					},
				},
				{
					Name: "main-app",
					SecurityContext: &corev1.SecurityContext{
						RunAsUser: &user2,
					},
				},
			},
		},
	}

	r := &CoDriverJobReconciler{}
	// No target specified - should use first container
	container := r.getTargetContainer(pod, nil)

	assert.NotNil(t, container)
	assert.Equal(t, "sidecar", container.Name)
	assert.NotNil(t, container.SecurityContext)
	assert.NotNil(t, container.SecurityContext.RunAsUser)
	assert.Equal(t, int64(2000), *container.SecurityContext.RunAsUser)
}
