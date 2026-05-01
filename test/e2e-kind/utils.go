//go:build e2ekind
// +build e2ekind

package e2ekind

import (
	"bytes"
	"context"
	"io"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

// CreateTestNamespace creates a unique test namespace
func CreateTestNamespace() *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "kubecodriver-kind-e2e-",
		},
	}
	Expect(k8sClient.Create(ctx, ns)).To(Succeed())
	return ns
}

// DeleteTestNamespace deletes a test namespace
func DeleteTestNamespace(ns *corev1.Namespace) {
	Expect(k8sClient.Delete(ctx, ns)).To(Succeed())
}

// CreateTargetPod creates a target pod for profiling
func CreateTargetPod(namespace, name string, labels map[string]string) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:  "app",
				Image: "nginx:latest",
				Command: []string{
					"sh", "-c",
					"while true; do echo 'Running...'; sleep 10; done",
				},
			}},
		},
	}
	Expect(k8sClient.Create(ctx, pod)).To(Succeed())
	return pod
}

// WaitForPodRunning waits for a pod to reach Running phase
func WaitForPodRunning(pod *corev1.Pod) {
	Eventually(func() bool {
		updated := &corev1.Pod{}
		err := k8sClient.Get(ctx, client.ObjectKeyFromObject(pod), updated)
		if err != nil {
			return false
		}
		return updated.Status.Phase == corev1.PodRunning
	}, "60s", "2s").Should(BeTrue())
}

// CreateCoDriverJob creates a CoDriverJob resource
func CreateCoDriverJob(namespace, name string, spec v1alpha1.CoDriverJobSpec) *v1alpha1.CoDriverJob {
	pt := &v1alpha1.CoDriverJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: spec,
	}
	Expect(k8sClient.Create(ctx, pt)).To(Succeed())
	return pt
}

// CreateCoDriverTool creates a CoDriverTool resource
func CreateCoDriverTool(namespace, name string) *v1alpha1.CoDriverTool {
	allowPrivileged := true
	ptc := &v1alpha1.CoDriverTool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.CoDriverToolSpec{
			Name:  name,
			Image: "ghcr.io/codriverlabs/ce/kubecodriver-aperf:latest",
			SecurityContext: v1alpha1.SecuritySpec{
				AllowPrivileged: &allowPrivileged,
				Capabilities: &v1alpha1.Capabilities{
					Add: []string{"SYS_ADMIN", "SYS_PTRACE"},
				},
			},
		},
	}
	Expect(k8sClient.Create(ctx, ptc)).To(Succeed())
	return ptc
}

// GetPodLogs retrieves logs from a pod container
func GetPodLogs(pod *corev1.Pod, containerName string) (string, error) {
	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Container: containerName,
	})

	podLogs, err := req.Stream(context.Background())
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// CreatePVC creates a PersistentVolumeClaim
func CreatePVC(namespace, name string, size string) *corev1.PersistentVolumeClaim {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: parseQuantity(size),
				},
			},
		},
	}
	Expect(k8sClient.Create(ctx, pvc)).To(Succeed())
	return pvc
}

// parseQuantity is a helper to parse resource quantities
func parseQuantity(s string) resource.Quantity {
	q, err := resource.ParseQuantity(s)
	Expect(err).NotTo(HaveOccurred())
	return q
}

// WaitForCoDriverJobPhase waits for CoDriverJob to reach expected phase
func WaitForCoDriverJobPhase(pt *v1alpha1.CoDriverJob, expectedPhase string) {
	Eventually(func() string {
		updated := &v1alpha1.CoDriverJob{}
		err := k8sClient.Get(ctx, client.ObjectKeyFromObject(pt), updated)
		if err != nil {
			return ""
		}
		if updated.Status.Phase == nil {
			return ""
		}
		return *updated.Status.Phase
	}, "120s", "2s").Should(Equal(expectedPhase))
}

// GetCoDriverJob retrieves the latest version of a CoDriverJob
func GetCoDriverJob(pt *v1alpha1.CoDriverJob) *v1alpha1.CoDriverJob {
	updated := &v1alpha1.CoDriverJob{}
	Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(pt), updated)).To(Succeed())
	return updated
}
