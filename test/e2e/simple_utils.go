package e2e

import (
	"context"
	"fmt"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

var (
	simpleK8sClient client.Client
	simpleClientset *kubernetes.Clientset
	simpleCtx       = context.Background()
)

// InitializeSimpleClients initializes clients for simple E2E tests
func InitializeSimpleClients() error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}

	scheme := runtime.NewScheme()
	err = v1alpha1.AddToScheme(scheme)
	if err != nil {
		return err
	}

	simpleK8sClient, err = client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return err
	}

	simpleClientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	return nil
}

// CreateSimpleTestNamespace creates a unique test namespace
func CreateSimpleTestNamespace() *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "kubecodriver-simple-e2e-",
		},
	}
	Expect(simpleK8sClient.Create(simpleCtx, ns)).To(Succeed())
	return ns
}

// DeleteSimpleTestNamespace deletes a test namespace
func DeleteSimpleTestNamespace(ns *corev1.Namespace) {
	Expect(simpleK8sClient.Delete(simpleCtx, ns)).To(Succeed())
}

// CreateSimpleMockTargetPod creates a mock pod for testing
func CreateSimpleMockTargetPod(namespace, name string, labels map[string]string) *corev1.Pod {
	if labels == nil {
		labels = map[string]string{
			"app": "test-app",
			"env": "testing",
		}
	}

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
			}},
		},
	}
	Expect(simpleK8sClient.Create(simpleCtx, pod)).To(Succeed())

	// Update status to simulate running pod
	pod.Status = corev1.PodStatus{
		Phase: corev1.PodRunning,
		ContainerStatuses: []corev1.ContainerStatus{{
			Name:  "app",
			Ready: true,
			State: corev1.ContainerState{
				Running: &corev1.ContainerStateRunning{},
			},
		}},
	}
	Expect(simpleK8sClient.Status().Update(simpleCtx, pod)).To(Succeed())
	return pod
}

// CreateSimpleTestCoDriverJob creates a CoDriverJob for testing
func CreateSimpleTestCoDriverJob(name, namespace string, spec v1alpha1.CoDriverJobSpec) *v1alpha1.CoDriverJob {
	coDriverJob := &v1alpha1.CoDriverJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: spec,
	}
	Expect(simpleK8sClient.Create(simpleCtx, coDriverJob)).To(Succeed())
	return coDriverJob
}

// CreateSimpleTestCoDriverTool creates a CoDriverTool for testing
func CreateSimpleTestCoDriverTool(name, namespace string) *v1alpha1.CoDriverTool {
	allowPrivileged := true
	coDriverTool := &v1alpha1.CoDriverTool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.CoDriverToolSpec{
			Name:  "aperf",
			Image: "ghcr.io/codriverlabs/ce/kubecodriver-aperf:latest",
			SecurityContext: v1alpha1.SecuritySpec{
				AllowPrivileged: &allowPrivileged,
				Capabilities: &v1alpha1.Capabilities{
					Add: []string{"SYS_ADMIN", "SYS_PTRACE"},
				},
			},
		},
	}
	Expect(simpleK8sClient.Create(simpleCtx, coDriverTool)).To(Succeed())
	return coDriverTool
}

// WaitForSimpleCoDriverJobPhase waits for CoDriverJob to reach expected phase
func WaitForSimpleCoDriverJobPhase(coDriverJob *v1alpha1.CoDriverJob, expectedPhase string) {
	Eventually(func() string {
		updated := &v1alpha1.CoDriverJob{}
		err := simpleK8sClient.Get(simpleCtx, client.ObjectKeyFromObject(coDriverJob), updated)
		if err != nil {
			return ""
		}
		if updated.Status.Phase == nil {
			return ""
		}
		return *updated.Status.Phase
	}, "30s", "1s").Should(Equal(expectedPhase))
}

// WaitForSimpleCoDriverJobCondition waits for CoDriverJob to have expected condition
func WaitForSimpleCoDriverJobCondition(coDriverJob *v1alpha1.CoDriverJob, conditionType string, status string) {
	Eventually(func() bool {
		updated := &v1alpha1.CoDriverJob{}
		err := simpleK8sClient.Get(simpleCtx, client.ObjectKeyFromObject(coDriverJob), updated)
		if err != nil {
			return false
		}

		for _, condition := range updated.Status.Conditions {
			if condition.Type == conditionType && condition.Status == status {
				return true
			}
		}
		return false
	}, "30s", "1s").Should(BeTrue())
}

// GetSimpleCoDriverJob retrieves the latest version of a CoDriverJob
func GetSimpleCoDriverJob(coDriverJob *v1alpha1.CoDriverJob) *v1alpha1.CoDriverJob {
	updated := &v1alpha1.CoDriverJob{}
	Expect(simpleK8sClient.Get(simpleCtx, client.ObjectKeyFromObject(coDriverJob), updated)).To(Succeed())
	return updated
}

// GetSimplePod retrieves the latest version of a Pod
func GetSimplePod(pod *corev1.Pod) *corev1.Pod {
	updated := &corev1.Pod{}
	Expect(simpleK8sClient.Get(simpleCtx, client.ObjectKeyFromObject(pod), updated)).To(Succeed())
	return updated
}

// CreateSimpleBasicCoDriverJobSpec creates a basic CoDriverJob spec for testing
func CreateSimpleBasicCoDriverJobSpec(targetLabels map[string]string) v1alpha1.CoDriverJobSpec {
	return v1alpha1.CoDriverJobSpec{
		Targets: v1alpha1.TargetSpec{
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: targetLabels,
			},
		},
		Tool: v1alpha1.ToolSpec{
			Name:     "aperf",
			Duration: "30s",
		},
		Output: v1alpha1.OutputSpec{
			Mode: "ephemeral",
		},
	}
}

// LogSimpleCoDriverJobStatus logs the current status of a CoDriverJob for debugging
func LogSimpleCoDriverJobStatus(coDriverJob *v1alpha1.CoDriverJob) {
	updated := GetSimpleCoDriverJob(coDriverJob)
	fmt.Printf("CoDriverJob %s/%s Status:\n", updated.Namespace, updated.Name)
	if updated.Status.Phase != nil {
		fmt.Printf("  Phase: %s\n", *updated.Status.Phase)
	} else {
		fmt.Printf("  Phase: <nil>\n")
	}
	fmt.Printf("  Conditions:\n")
	for _, condition := range updated.Status.Conditions {
		fmt.Printf("    %s: %s - %s\n", condition.Type, condition.Status, condition.Message)
	}
}
