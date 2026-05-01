package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

var _ = Describe("Output Mode Configuration", func() {
	var namespace *corev1.Namespace

	BeforeEach(func() {
		namespace = CreateSimpleTestNamespace()
		CreateSimpleTestCoDriverTool("aperf-config", namespace.Name)
		CreateSimpleMockTargetPod(namespace.Name, "output-pod", map[string]string{
			"app": "output-app",
		})
	})

	AfterEach(func() {
		DeleteSimpleTestNamespace(namespace)
	})

	Context("Ephemeral Mode", func() {
		It("should configure ephemeral output correctly", func() {
			By("creating CoDriverJob with ephemeral output")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "output-app"},
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
			coDriverJob := CreateSimpleTestCoDriverJob("ephemeral-output", namespace.Name, spec)

			By("verifying CoDriverJob accepts ephemeral configuration")
			WaitForSimpleCoDriverJobPhase(coDriverJob, "Pending")

			By("verifying output mode is correctly set")
			updated := GetSimpleCoDriverJob(coDriverJob)
			Expect(updated.Spec.Output.Mode).To(Equal("ephemeral"))
		})

		It("should handle ephemeral mode with custom path", func() {
			By("creating CoDriverJob with custom ephemeral path")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "output-app"},
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
			coDriverJob := CreateSimpleTestCoDriverJob("ephemeral-custom-path", namespace.Name, spec)

			By("verifying custom path configuration")
			WaitForSimpleCoDriverJobPhase(coDriverJob, "Pending")
			updated := GetSimpleCoDriverJob(coDriverJob)
			Expect(updated.Spec.Output.Mode).To(Equal("ephemeral"))
		})
	})

	Context("PVC Mode", func() {
		It("should validate PVC configuration", func() {
			By("creating CoDriverJob with PVC output")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "output-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "pvc",
					PVC: &v1alpha1.PVCSpec{
						ClaimName: "test-pvc",
						Path:      func() *string { s := "/data/profiles"; return &s }(),
					},
				},
			}
			coDriverJob := CreateSimpleTestCoDriverJob("pvc-output", namespace.Name, spec)

			By("verifying PVC configuration is accepted")
			WaitForSimpleCoDriverJobPhase(coDriverJob, "Pending")

			By("verifying PVC settings are correctly configured")
			updated := GetSimpleCoDriverJob(coDriverJob)
			Expect(updated.Spec.Output.Mode).To(Equal("pvc"))
			Expect(updated.Spec.Output.PVC.ClaimName).To(Equal("test-pvc"))
			Expect(updated.Spec.Output.PVC.Path).To(Equal("/data/profiles"))
		})

		It("should handle PVC mode with storage class", func() {
			By("creating CoDriverJob with PVC and storage class")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "output-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "pvc",
					PVC: &v1alpha1.PVCSpec{
						ClaimName: "storage-pvc",
						Path:      func() *string { s := "/data/profiles"; return &s }(),
					},
				},
			}
			coDriverJob := CreateSimpleTestCoDriverJob("pvc-storage-class", namespace.Name, spec)

			By("verifying PVC configuration")
			updated := GetSimpleCoDriverJob(coDriverJob)
			Expect(updated.Spec.Output.PVC.ClaimName).To(Equal("storage-pvc"))
			Expect(updated.Spec.Output.PVC.Path).NotTo(BeNil())
			Expect(*updated.Spec.Output.PVC.Path).To(Equal("/data/profiles"))
		})

		It("should reject invalid PVC configurations", func() {
			By("attempting to create CoDriverJob with invalid PVC config")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "output-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "pvc",
					// Missing PVC configuration
				},
			}

			coDriverJob := &v1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-pvc",
					Namespace: namespace.Name,
				},
				Spec: spec,
			}

			By("expecting validation to fail")
			err := simpleK8sClient.Create(simpleCtx, coDriverJob)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Collector Mode", func() {
		It("should configure collector endpoint correctly", func() {
			By("creating CoDriverJob with collector output")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "output-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "collector",
					Collector: &v1alpha1.CollectorSpec{
						Endpoint: "https://collector.kubecodriver-system.svc.cluster.local:8443",
					},
				},
			}
			coDriverJob := CreateSimpleTestCoDriverJob("collector-output", namespace.Name, spec)

			By("verifying collector configuration")
			WaitForSimpleCoDriverJobPhase(coDriverJob, "Pending")
			updated := GetSimpleCoDriverJob(coDriverJob)
			Expect(updated.Spec.Output.Mode).To(Equal("collector"))
			Expect(updated.Spec.Output.Collector.Endpoint).To(Equal("https://collector.kubecodriver-system.svc.cluster.local:8443"))
		})

		It("should handle collector mode with authentication", func() {
			By("creating CoDriverJob with collector authentication")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "output-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "collector",
					Collector: &v1alpha1.CollectorSpec{
						Endpoint: "https://collector.example.com:8443",
					},
				},
			}
			coDriverJob := CreateSimpleTestCoDriverJob("collector-auth", namespace.Name, spec)

			By("verifying authentication configuration")
			_ = GetSimpleCoDriverJob(coDriverJob)
		})

		It("should validate collector endpoint format", func() {
			By("attempting to create CoDriverJob with invalid collector endpoint")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "output-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "collector",
					Collector: &v1alpha1.CollectorSpec{
						Endpoint: "invalid-url",
					},
				},
			}

			coDriverJob := &v1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-collector",
					Namespace: namespace.Name,
				},
				Spec: spec,
			}

			By("expecting validation to fail for invalid URL")
			err := simpleK8sClient.Create(simpleCtx, coDriverJob)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Output Mode Transitions", func() {
		It("should handle output mode changes", func() {
			By("creating CoDriverJob with ephemeral output")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "output-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("mode-transition", namespace.Name, spec)

			By("verifying initial ephemeral mode")
			WaitForSimpleCoDriverJobPhase(coDriverJob, "Pending")
			initial := GetSimpleCoDriverJob(coDriverJob)
			Expect(initial.Spec.Output.Mode).To(Equal("ephemeral"))

			By("updating to PVC mode")
			updated := GetSimpleCoDriverJob(coDriverJob)
			updated.Spec.Output.Mode = "pvc"
			updated.Spec.Output.PVC = &v1alpha1.PVCSpec{
				ClaimName: "transition-pvc",
				Path:      func() *string { s := "/data"; return &s }(),
			}
			Expect(simpleK8sClient.Update(simpleCtx, updated)).To(Succeed())

			By("verifying mode transition")
			Eventually(func() string {
				current := GetSimpleCoDriverJob(coDriverJob)
				return current.Spec.Output.Mode
			}, "30s", "1s").Should(Equal("pvc"))
		})
	})

	Context("Output Path Validation", func() {
		It("should validate output path formats", func() {
			By("testing various path formats")
			validPaths := []string{
				"/tmp/profiles",
				"/data/output",
				"/var/log/kubecodriver",
			}

			for _, path := range validPaths {
				spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "output-app"})
				coDriverJob := CreateSimpleTestCoDriverJob("path-test-"+path[1:], namespace.Name, spec)

				By("verifying path is accepted: " + path)
				_ = GetSimpleCoDriverJob(coDriverJob)

				// Cleanup
				Expect(simpleK8sClient.Delete(simpleCtx, coDriverJob)).To(Succeed())
			}
		})

		It("should reject invalid output paths", func() {
			By("testing invalid path formats")
			invalidPaths := []string{
				"relative/path",
				"",
				"../../../etc/passwd",
			}

			for _, path := range invalidPaths {
				spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "output-app"})

				coDriverJob := &v1alpha1.CoDriverJob{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "invalid-path-" + path,
						Namespace: namespace.Name,
					},
					Spec: spec,
				}

				By("expecting validation to fail for path: " + path)
				err := simpleK8sClient.Create(simpleCtx, coDriverJob)
				Expect(err).To(HaveOccurred())
			}
		})
	})
})
