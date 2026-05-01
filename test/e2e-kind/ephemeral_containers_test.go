//go:build e2ekind
// +build e2ekind

package e2ekind

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

var _ = Describe("Ephemeral Container Profiling", func() {
	var (
		testNs    *corev1.Namespace
		targetPod *corev1.Pod
		coDriverJob *v1alpha1.CoDriverJob
		ctx       = context.Background()
	)

	BeforeEach(func() {
		testNs = CreateTestNamespace()
		targetPod = CreateTargetPod(testNs.Name, "target-app", map[string]string{"app": "test"})
		WaitForPodRunning(targetPod)
	})

	AfterEach(func() {
		if coDriverJob != nil {
			k8sClient.Delete(ctx, coDriverJob)
		}
		DeleteTestNamespace(testNs)
	})

	Context("Container Creation", func() {
		It("should create ephemeral container in target pod", func() {
			By("creating CoDriverJob with ephemeral output mode")
			coDriverJob = CreateCoDriverJob(testNs.Name, "test-ephemeral", v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "10s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			})

			By("waiting for CoDriverJob to reach Running phase")
			Eventually(func() string {
				updated := &v1alpha1.CoDriverJob{}
				k8sClient.Get(ctx, client.ObjectKeyFromObject(coDriverJob), updated)
				if updated.Status.Phase != nil {
					return *updated.Status.Phase
				}
				return ""
			}, "60s", "2s").Should(Equal("Running"))

			By("verifying ephemeral container exists in target pod")
			Eventually(func() bool {
				updated := &corev1.Pod{}
				k8sClient.Get(ctx, client.ObjectKeyFromObject(targetPod), updated)
				return len(updated.Spec.EphemeralContainers) > 0
			}, "30s", "2s").Should(BeTrue())

			By("verifying ephemeral container is running")
			Eventually(func() bool {
				updated := &corev1.Pod{}
				k8sClient.Get(ctx, client.ObjectKeyFromObject(targetPod), updated)
				for _, status := range updated.Status.EphemeralContainerStatuses {
					if status.State.Running != nil {
						return true
					}
				}
				return false
			}, "30s", "2s").Should(BeTrue())
		})

		It("should handle ephemeral container lifecycle", func() {
			coDriverJob = CreateCoDriverJob(testNs.Name, "test-lifecycle", v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "5s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			})

			By("waiting for container to start")
			Eventually(func() bool {
				updated := &corev1.Pod{}
				k8sClient.Get(ctx, client.ObjectKeyFromObject(targetPod), updated)
				for _, status := range updated.Status.EphemeralContainerStatuses {
					if status.State.Running != nil {
						return true
					}
				}
				return false
			}, "30s", "2s").Should(BeTrue())

			By("waiting for container to complete")
			Eventually(func() bool {
				updated := &corev1.Pod{}
				k8sClient.Get(ctx, client.ObjectKeyFromObject(targetPod), updated)
				for _, status := range updated.Status.EphemeralContainerStatuses {
					if status.State.Terminated != nil {
						return true
					}
				}
				return false
			}, "30s", "2s").Should(BeTrue())

			By("verifying CoDriverJob reaches Completed phase")
			Eventually(func() string {
				updated := &v1alpha1.CoDriverJob{}
				k8sClient.Get(ctx, client.ObjectKeyFromObject(coDriverJob), updated)
				if updated.Status.Phase != nil {
					return *updated.Status.Phase
				}
				return ""
			}, "60s", "2s").Should(Equal("Completed"))
		})
	})

	Context("Tool Execution", func() {
		It("should execute profiling tool in ephemeral container", func() {
			coDriverJob = CreateCoDriverJob(testNs.Name, "test-execution", v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "10s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			})

			By("waiting for ephemeral container to run")
			time.Sleep(15 * time.Second)

			By("checking container logs for tool output")
			updated := &corev1.Pod{}
			k8sClient.Get(ctx, client.ObjectKeyFromObject(targetPod), updated)

			var containerName string
			for _, ec := range updated.Spec.EphemeralContainers {
				containerName = ec.Name
				break
			}
			Expect(containerName).NotTo(BeEmpty())

			logs, err := GetPodLogs(targetPod, containerName)
			Expect(err).NotTo(HaveOccurred())
			Expect(logs).NotTo(BeEmpty())
		})

		It("should handle tool timeout correctly", func() {
			coDriverJob = CreateCoDriverJob(testNs.Name, "test-timeout", v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "5s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			})

			By("verifying tool stops after duration")
			Eventually(func() bool {
				updated := &corev1.Pod{}
				k8sClient.Get(ctx, client.ObjectKeyFromObject(targetPod), updated)
				for _, status := range updated.Status.EphemeralContainerStatuses {
					if status.State.Terminated != nil {
						return true
					}
				}
				return false
			}, "20s", "2s").Should(BeTrue())
		})
	})

	Context("Data Collection", func() {
		It("should collect profiling data in ephemeral mode", func() {
			coDriverJob = CreateCoDriverJob(testNs.Name, "test-data", v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "10s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			})

			By("waiting for profiling to complete")
			Eventually(func() string {
				updated := &v1alpha1.CoDriverJob{}
				k8sClient.Get(ctx, client.ObjectKeyFromObject(coDriverJob), updated)
				if updated.Status.Phase != nil {
					return *updated.Status.Phase
				}
				return ""
			}, "60s", "2s").Should(Equal("Completed"))

			By("verifying data artifacts in status")
			updated := &v1alpha1.CoDriverJob{}
			k8sClient.Get(ctx, client.ObjectKeyFromObject(coDriverJob), updated)
			Expect(updated.Status.Artifacts).NotTo(BeEmpty())
		})
	})

	Context("Multiple Target Pods", func() {
		It("should create ephemeral containers in multiple pods", func() {
			By("creating additional target pods")
			targetPod2 := CreateTargetPod(testNs.Name, "target-app-2", map[string]string{"app": "test"})
			targetPod3 := CreateTargetPod(testNs.Name, "target-app-3", map[string]string{"app": "test"})
			WaitForPodRunning(targetPod2)
			WaitForPodRunning(targetPod3)

			By("creating CoDriverJob targeting all pods")
			coDriverJob = CreateCoDriverJob(testNs.Name, "test-multi", v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "10s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			})

			By("verifying ephemeral containers in all pods")
			for _, pod := range []*corev1.Pod{targetPod, targetPod2, targetPod3} {
				Eventually(func() bool {
					updated := &corev1.Pod{}
					k8sClient.Get(ctx, client.ObjectKeyFromObject(pod), updated)
					return len(updated.Spec.EphemeralContainers) > 0
				}, "30s", "2s").Should(BeTrue())
			}
		})
	})
})
