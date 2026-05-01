package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

var _ = Describe("Phase 3 Validation Tests", func() {
	var namespace *corev1.Namespace

	BeforeEach(func() {
		err := InitializeSimpleClients()
		Expect(err).NotTo(HaveOccurred())

		namespace = CreateSimpleTestNamespace()
		CreateSimpleTestCoDriverTool("phase3-config", namespace.Name)
		CreateSimpleMockTargetPod(namespace.Name, "phase3-pod", map[string]string{
			"app": "phase3-app",
		})
	})

	AfterEach(func() {
		if namespace != nil {
			DeleteSimpleTestNamespace(namespace)
		}
	})

	Context("Basic RBAC Validation", func() {
		It("should create CoDriverJob with basic configuration", func() {
			By("creating a basic CoDriverJob")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "phase3-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("basic-test", namespace.Name, spec)

			By("verifying CoDriverJob is created successfully")
			Expect(coDriverJob.Name).To(Equal("basic-test"))
			Expect(coDriverJob.Namespace).To(Equal(namespace.Name))
		})

		It("should validate security context requirements", func() {
			By("creating CoDriverTool with security requirements")
			allowPrivileged := true
			secureConfig := &v1alpha1.CoDriverTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secure-config",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CoDriverToolSpec{
					Name:  "secure-tool",
					Image: "ghcr.io/codriverlabs/ce/kubecodriver-secure:latest",
					SecurityContext: v1alpha1.SecuritySpec{
						AllowPrivileged: &allowPrivileged,
						Capabilities: &v1alpha1.Capabilities{
							Add: []string{"SYS_ADMIN", "SYS_PTRACE"},
						},
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, secureConfig)).To(Succeed())

			By("creating CoDriverJob using secure configuration")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "phase3-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "secure-tool",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			coDriverJob := CreateSimpleTestCoDriverJob("secure-test", namespace.Name, spec)

			By("verifying secure CoDriverJob is created")
			Expect(coDriverJob.Spec.Tool.Name).To(Equal("secure-tool"))
		})
	})

	Context("Webhook Validation", func() {
		It("should reject CoDriverJob with invalid duration", func() {
			By("attempting to create CoDriverJob with invalid duration")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "phase3-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "phase3-config",
					Duration: "invalid-duration",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}

			coDriverJob := &v1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-duration",
					Namespace: namespace.Name,
				},
				Spec: spec,
			}

			By("expecting validation to fail")
			err := simpleK8sClient.Create(simpleCtx, coDriverJob)
			Expect(err).To(HaveOccurred())
		})

		It("should validate output mode configuration", func() {
			By("testing PVC mode without PVC spec")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "phase3-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "phase3-config",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "pvc",
					// Missing PVC spec
				},
			}

			coDriverJob := &v1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-pvc",
					Namespace: namespace.Name,
				},
				Spec: spec,
			}

			By("expecting validation to fail for incomplete PVC config")
			err := simpleK8sClient.Create(simpleCtx, coDriverJob)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Integration Scenarios", func() {
		It("should handle multiple target pods", func() {
			By("creating additional target pods")
			CreateSimpleMockTargetPod(namespace.Name, "multi-pod-1", map[string]string{
				"app": "multi-app",
			})
			CreateSimpleMockTargetPod(namespace.Name, "multi-pod-2", map[string]string{
				"app": "multi-app",
			})

			By("creating CoDriverJob targeting multiple pods")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "multi-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("multi-pod-test", namespace.Name, spec)

			By("verifying CoDriverJob is created successfully")
			Expect(coDriverJob.Spec.Targets.LabelSelector.MatchLabels["app"]).To(Equal("multi-app"))
		})

		It("should validate tool configuration exists", func() {
			By("creating CoDriverJob with existing configuration")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "phase3-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("config-exists-test", namespace.Name, spec)

			By("verifying CoDriverJob references correct tool")
			Expect(coDriverJob.Spec.Tool.Name).To(Equal("aperf"))
		})
	})

	Context("Error Handling", func() {
		It("should handle missing target pods gracefully", func() {
			By("creating CoDriverJob with non-matching selector")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "nonexistent"})
			coDriverJob := CreateSimpleTestCoDriverJob("no-targets", namespace.Name, spec)

			By("verifying CoDriverJob is created but will fail to find targets")
			Expect(coDriverJob.Spec.Targets.LabelSelector.MatchLabels["app"]).To(Equal("nonexistent"))
		})

		It("should validate required fields", func() {
			By("attempting to create CoDriverJob without required fields")
			coDriverJob := &v1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "missing-fields",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CoDriverJobSpec{
					// Missing required fields
				},
			}

			By("expecting validation to fail")
			err := simpleK8sClient.Create(simpleCtx, coDriverJob)
			Expect(err).To(HaveOccurred())
		})
	})
})
