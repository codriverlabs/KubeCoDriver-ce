package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

var _ = Describe("Advanced Reconciliation Scenarios", func() {
	var namespace *corev1.Namespace

	BeforeEach(func() {
		namespace = CreateSimpleTestNamespace()
		CreateSimpleTestCoDriverTool("aperf-config", namespace.Name)
	})

	AfterEach(func() {
		DeleteSimpleTestNamespace(namespace)
	})

	Context("Cross-Namespace Scenarios", func() {
		var secondNamespace *corev1.Namespace

		BeforeEach(func() {
			secondNamespace = CreateSimpleTestNamespace()
			CreateSimpleTestCoDriverTool("aperf-config", secondNamespace.Name)
		})

		AfterEach(func() {
			DeleteSimpleTestNamespace(secondNamespace)
		})

		It("should isolate CoDriverJobs between namespaces", func() {
			By("creating pods in both namespaces")
			CreateSimpleMockTargetPod(namespace.Name, "ns1-pod", map[string]string{
				"app": "cross-ns-app",
			})
			CreateSimpleMockTargetPod(secondNamespace.Name, "ns2-pod", map[string]string{
				"app": "cross-ns-app",
			})

			By("creating CoDriverJob in first namespace")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "cross-ns-app"})
			coDriverJob1 := CreateSimpleTestCoDriverJob("ns1-tool", namespace.Name, spec)

			By("creating CoDriverJob in second namespace")
			coDriverJob2 := CreateSimpleTestCoDriverJob("ns2-tool", secondNamespace.Name, spec)

			By("verifying each CoDriverJob only targets pods in its namespace")
			Eventually(func() int {
				updated := GetSimpleCoDriverJob(coDriverJob1)
				return len(updated.Status.ActivePods)
			}, "30s", "1s").Should(Equal(1))

			Eventually(func() int {
				updated := GetSimpleCoDriverJob(coDriverJob2)
				return len(updated.Status.ActivePods)
			}, "30s", "1s").Should(Equal(1))
		})
	})

	Context("Resource Ownership and Cleanup", func() {
		It("should handle orphaned resources gracefully", func() {
			By("creating target pod")
			targetPod := CreateSimpleMockTargetPod(namespace.Name, "orphan-pod", map[string]string{
				"app": "orphan-app",
			})

			By("creating CoDriverJob")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "orphan-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("orphan-test", namespace.Name, spec)

			By("waiting for reconciliation")
			WaitForSimpleCoDriverJobPhase(coDriverJob, "Pending")

			By("simulating pod deletion without CoDriverJob cleanup")
			Expect(simpleK8sClient.Delete(simpleCtx, targetPod)).To(Succeed())

			By("verifying CoDriverJob handles orphaned state")
			Eventually(func() int {
				updated := GetSimpleCoDriverJob(coDriverJob)
				return len(updated.Status.ActivePods)
			}, "30s", "1s").Should(Equal(0))

			By("verifying CoDriverJob phase transitions appropriately")
			Eventually(func() string {
				updated := GetSimpleCoDriverJob(coDriverJob)
				if updated.Status.Phase == nil {
					return ""
				}
				return *updated.Status.Phase
			}, "30s", "1s").Should(Equal("Failed"))
		})

		It("should clean up resources on CoDriverJob deletion", func() {
			By("creating target pod")
			CreateSimpleMockTargetPod(namespace.Name, "cleanup-pod", map[string]string{
				"app": "cleanup-app",
			})

			By("creating CoDriverJob")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "cleanup-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("cleanup-test", namespace.Name, spec)

			By("waiting for reconciliation")
			WaitForSimpleCoDriverJobPhase(coDriverJob, "Pending")

			By("storing initial resource state")
			initialPod := GetSimplePod(&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cleanup-pod",
					Namespace: namespace.Name,
				},
			})

			By("deleting CoDriverJob")
			Expect(simpleK8sClient.Delete(simpleCtx, coDriverJob)).To(Succeed())

			By("verifying CoDriverJob is deleted")
			Eventually(func() bool {
				updated := &v1alpha1.CoDriverJob{}
				err := simpleK8sClient.Get(simpleCtx, client.ObjectKeyFromObject(coDriverJob), updated)
				return err != nil
			}, "30s", "1s").Should(BeTrue())

			By("verifying target pod is not affected by CoDriverJob deletion")
			finalPod := GetSimplePod(initialPod)
			Expect(finalPod.Name).To(Equal(initialPod.Name))
		})
	})

	Context("Edge Cases and Boundary Conditions", func() {
		It("should handle empty label selectors", func() {
			By("creating pods with various labels")
			CreateSimpleMockTargetPod(namespace.Name, "any-pod-1", map[string]string{
				"app": "any-app",
			})
			CreateSimpleMockTargetPod(namespace.Name, "any-pod-2", map[string]string{
				"service": "any-service",
			})

			By("creating CoDriverJob with empty label selector")
			coDriverJob := &v1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "empty-selector",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CoDriverJobSpec{
					Targets: v1alpha1.TargetSpec{
						LabelSelector: &metav1.LabelSelector{}, // Empty selector
					},
					Tool: v1alpha1.ToolSpec{
						Name:     "aperf",
						Duration: "30s",
					},
					Output: v1alpha1.OutputSpec{
						Mode: "ephemeral",
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, coDriverJob)).To(Succeed())

			By("verifying all pods in namespace are selected")
			Eventually(func() int {
				updated := GetSimpleCoDriverJob(coDriverJob)
				return len(updated.Status.ActivePods)
			}, "30s", "1s").Should(Equal(2))
		})

		It("should handle rapid CoDriverJob creation and deletion", func() {
			By("creating target pod")
			CreateSimpleMockTargetPod(namespace.Name, "rapid-pod", map[string]string{
				"app": "rapid-app",
			})

			By("rapidly creating and deleting CoDriverJobs")
			for i := 0; i < 3; i++ {
				spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "rapid-app"})
				coDriverJob := CreateSimpleTestCoDriverJob(fmt.Sprintf("rapid-tool-%d", i), namespace.Name, spec)

				// Brief wait to allow some reconciliation
				Eventually(func() bool {
					updated := GetSimpleCoDriverJob(coDriverJob)
					return updated.Status.Phase != nil
				}, "10s", "1s").Should(BeTrue())

				// Delete immediately
				Expect(simpleK8sClient.Delete(simpleCtx, coDriverJob)).To(Succeed())

				// Verify deletion
				Eventually(func() bool {
					updated := &v1alpha1.CoDriverJob{}
					err := simpleK8sClient.Get(simpleCtx, client.ObjectKeyFromObject(coDriverJob), updated)
					return err != nil
				}, "15s", "1s").Should(BeTrue())
			}
		})

		It("should handle CoDriverJob updates during reconciliation", func() {
			By("creating target pod")
			CreateSimpleMockTargetPod(namespace.Name, "update-pod", map[string]string{
				"app": "update-app",
			})

			By("creating CoDriverJob")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "update-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("update-test", namespace.Name, spec)

			By("waiting for initial reconciliation")
			WaitForSimpleCoDriverJobPhase(coDriverJob, "Pending")

			By("updating CoDriverJob spec")
			updated := GetSimpleCoDriverJob(coDriverJob)
			updated.Spec.Tool.Duration = "60s"
			Expect(simpleK8sClient.Update(simpleCtx, updated)).To(Succeed())

			By("verifying reconciliation handles the update")
			Eventually(func() string {
				current := GetSimpleCoDriverJob(coDriverJob)
				return current.Spec.Tool.Duration
			}, "30s", "1s").Should(Equal("60s"))
		})
	})

	Context("Stress Testing Scenarios", func() {
		It("should handle multiple concurrent CoDriverJobs", func() {
			By("creating multiple target pods")
			for i := 0; i < 10; i++ {
				CreateSimpleMockTargetPod(namespace.Name, fmt.Sprintf("stress-pod-%d", i), map[string]string{
					"app":   "stress-app",
					"index": fmt.Sprintf("%d", i),
				})
			}

			By("creating multiple CoDriverJobs with different selectors")
			var coDriverJobs []*v1alpha1.CoDriverJob
			for i := 0; i < 5; i++ {
				spec := v1alpha1.CoDriverJobSpec{
					Targets: v1alpha1.TargetSpec{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app":   "stress-app",
								"index": fmt.Sprintf("%d", i*2), // Select every other pod
							},
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
				coDriverJob := CreateSimpleTestCoDriverJob(fmt.Sprintf("stress-tool-%d", i), namespace.Name, spec)
				coDriverJobs = append(coDriverJobs, coDriverJob)
			}

			By("verifying all CoDriverJobs reconcile successfully")
			for i, coDriverJob := range coDriverJobs {
				Eventually(func() string {
					updated := GetSimpleCoDriverJob(coDriverJob)
					if updated.Status.Phase == nil {
						return ""
					}
					return *updated.Status.Phase
				}, "45s", "1s").Should(Or(Equal("Pending"), Equal("Running")),
					fmt.Sprintf("CoDriverJob %d should reach Pending or Running phase", i))
			}
		})

		It("should maintain consistency under resource pressure", func() {
			By("creating many target pods")
			for i := 0; i < 20; i++ {
				CreateSimpleMockTargetPod(namespace.Name, fmt.Sprintf("pressure-pod-%d", i), map[string]string{
					"app":  "pressure-app",
					"tier": fmt.Sprintf("tier-%d", i%3), // 3 tiers
				})
			}

			By("creating CoDriverJob targeting all pods")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "pressure-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("pressure-test", namespace.Name, spec)

			By("verifying consistent reconciliation despite scale")
			Eventually(func() int {
				updated := GetSimpleCoDriverJob(coDriverJob)
				return len(updated.Status.ActivePods)
			}, "60s", "2s").Should(Equal(20))

			By("verifying status remains consistent")
			Consistently(func() int {
				updated := GetSimpleCoDriverJob(coDriverJob)
				return len(updated.Status.ActivePods)
			}, "10s", "1s").Should(Equal(20))
		})
	})

	Context("Reconciliation Idempotency", func() {
		It("should produce consistent results on repeated reconciliation", func() {
			By("creating target pod")
			CreateSimpleMockTargetPod(namespace.Name, "idempotent-pod", map[string]string{
				"app": "idempotent-app",
			})

			By("creating CoDriverJob")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "idempotent-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("idempotent-test", namespace.Name, spec)

			By("capturing initial state after reconciliation")
			WaitForSimpleCoDriverJobPhase(coDriverJob, "Pending")
			initialState := GetSimpleCoDriverJob(coDriverJob)

			By("triggering re-reconciliation by updating annotation")
			updated := GetSimpleCoDriverJob(coDriverJob)
			if updated.Annotations == nil {
				updated.Annotations = make(map[string]string)
			}
			updated.Annotations["test.reconcile"] = "trigger"
			Expect(simpleK8sClient.Update(simpleCtx, updated)).To(Succeed())

			By("verifying state remains consistent after re-reconciliation")
			Eventually(func() bool {
				current := GetSimpleCoDriverJob(coDriverJob)
				return len(current.Status.ActivePods) == len(initialState.Status.ActivePods) &&
					*current.Status.Phase == *initialState.Status.Phase
			}, "30s", "1s").Should(BeTrue())
		})
	})
})
