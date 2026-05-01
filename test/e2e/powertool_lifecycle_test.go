package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

var _ = Describe("CoDriverJob Lifecycle", func() {
	var namespace *corev1.Namespace

	BeforeEach(func() {
		namespace = CreateSimpleTestNamespace()
		CreateSimpleMockTargetPod(namespace.Name, "target-pod", map[string]string{
			"app": "test-app",
			"env": "testing",
		})
		CreateSimpleTestCoDriverTool("aperf-config", namespace.Name)
	})

	AfterEach(func() {
		DeleteSimpleTestNamespace(namespace)
	})

	Context("CoDriverJob Creation", func() {
		It("should create CoDriverJob with valid spec", func() {
			By("creating a CoDriverJob with basic configuration")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "test-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("test-codriverjob", namespace.Name, spec)

			By("verifying CoDriverJob is created successfully")
			Expect(coDriverJob.Name).To(Equal("test-codriverjob"))
			Expect(coDriverJob.Namespace).To(Equal(namespace.Name))
			Expect(coDriverJob.Spec.Tool.Name).To(Equal("aperf"))

			By("waiting for CoDriverJob to be processed")
			WaitForSimpleCoDriverJobPhase(coDriverJob, "Pending")

			By("verifying status conditions are set")
			updated := GetSimpleCoDriverJob(coDriverJob)
			Expect(updated.Status.Conditions).NotTo(BeEmpty())
		})

		It("should handle CoDriverJob with multiple target pods", func() {
			By("creating additional target pods")
			CreateSimpleMockTargetPod(namespace.Name, "target-pod-2", map[string]string{
				"app": "test-app",
				"env": "testing",
			})
			CreateSimpleMockTargetPod(namespace.Name, "target-pod-3", map[string]string{
				"app": "test-app",
				"env": "production",
			})

			By("creating CoDriverJob targeting multiple pods")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "test-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("multi-target", namespace.Name, spec)

			By("verifying CoDriverJob processes multiple targets")
			WaitForSimpleCoDriverJobPhase(coDriverJob, "Pending")

			updated := GetSimpleCoDriverJob(coDriverJob)
			Expect(updated.Status.ActivePods).NotTo(BeEmpty())
		})

		It("should set appropriate finalizers", func() {
			By("creating a CoDriverJob")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "test-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("finalizer-test", namespace.Name, spec)

			By("verifying finalizer is set")
			updated := GetSimpleCoDriverJob(coDriverJob)
			Expect(updated.Finalizers).To(ContainElement("codriverjob.kubecodriver.codriverlabs.ai/finalizer"))
		})
	})

	Context("CoDriverJob Validation", func() {
		It("should reject CoDriverJob with invalid tool name", func() {
			By("attempting to create CoDriverJob with nonexistent tool")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name: "nonexistent-tool",
				},
			}

			coDriverJob := &v1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-tool",
					Namespace: namespace.Name,
				},
				Spec: spec,
			}

			By("expecting creation to fail")
			err := simpleK8sClient.Create(simpleCtx, coDriverJob)
			Expect(err).To(HaveOccurred())
		})

		It("should reject CoDriverJob with invalid duration", func() {
			By("attempting to create CoDriverJob with invalid duration")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "test-app"})
			spec.Tool.Duration = "invalid-duration"

			coDriverJob := &v1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-duration",
					Namespace: namespace.Name,
				},
				Spec: spec,
			}

			By("expecting creation to fail")
			err := simpleK8sClient.Create(simpleCtx, coDriverJob)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("CoDriverJob Status Updates", func() {
		It("should update status phase correctly", func() {
			By("creating a CoDriverJob")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "test-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("status-test", namespace.Name, spec)

			By("verifying initial status")
			Eventually(func() string {
				updated := GetSimpleCoDriverJob(coDriverJob)
				if updated.Status.Phase == nil {
					return ""
				}
				return *updated.Status.Phase
			}).Should(Equal("Pending"))

			By("verifying status conditions are populated")
			updated := GetSimpleCoDriverJob(coDriverJob)
			Expect(updated.Status.Conditions).NotTo(BeEmpty())
		})

		It("should track target pods in status", func() {
			By("creating a CoDriverJob")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "test-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("target-tracking", namespace.Name, spec)

			By("verifying target pods are tracked")
			Eventually(func() map[string]string {
				updated := GetSimpleCoDriverJob(coDriverJob)
				return updated.Status.ActivePods
			}).Should(Not(BeEmpty()))
		})
	})

	Context("CoDriverJob Deletion", func() {
		It("should handle deletion gracefully", func() {
			By("creating a CoDriverJob")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "test-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("deletion-test", namespace.Name, spec)

			By("waiting for CoDriverJob to be processed")
			WaitForSimpleCoDriverJobPhase(coDriverJob, "Pending")

			By("deleting the CoDriverJob")
			Expect(simpleK8sClient.Delete(simpleCtx, coDriverJob)).To(Succeed())

			By("verifying CoDriverJob is deleted")
			Eventually(func() bool {
				updated := &v1alpha1.CoDriverJob{}
				err := simpleK8sClient.Get(simpleCtx, client.ObjectKeyFromObject(coDriverJob), updated)
				return err != nil
			}).Should(BeTrue())
		})
	})
})
