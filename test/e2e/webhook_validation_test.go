package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

var _ = Describe("Webhook Validation", func() {
	var namespace *corev1.Namespace

	BeforeEach(func() {
		namespace = CreateSimpleTestNamespace()
		CreateSimpleTestCoDriverTool("webhook-config", namespace.Name)
		CreateSimpleMockTargetPod(namespace.Name, "webhook-pod", map[string]string{
			"app": "webhook-app",
		})
	})

	AfterEach(func() {
		DeleteSimpleTestNamespace(namespace)
	})

	Context("CoDriverJob Validation", func() {
		It("should validate required fields", func() {
			By("attempting to create CoDriverJob without required fields")
			coDriverJob := &v1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-required",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CoDriverJobSpec{
					// Missing required fields
				},
			}

			By("expecting validation webhook to reject")
			err := simpleK8sClient.Create(simpleCtx, coDriverJob)
			Expect(err).To(HaveOccurred())
		})

		It("should validate tool name format", func() {
			By("testing invalid tool names")
			invalidNames := []string{
				"",
				"UPPERCASE",
				"with spaces",
				"with-special-chars!",
				"toolname-with-very-long-name-that-exceeds-reasonable-limits",
			}

			for _, name := range invalidNames {
				spec := v1alpha1.CoDriverJobSpec{
					Targets: v1alpha1.TargetSpec{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "webhook-app"},
						},
					},
					Tool: v1alpha1.ToolSpec{
						Name:     name,
						Duration: "30s",
					},
					Output: v1alpha1.OutputSpec{
						Mode: "ephemeral",
					},
				}

				coDriverJob := &v1alpha1.CoDriverJob{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "invalid-name-" + name,
						Namespace: namespace.Name,
					},
					Spec: spec,
				}

				By("expecting validation to fail for name: " + name)
				err := simpleK8sClient.Create(simpleCtx, coDriverJob)
				Expect(err).To(HaveOccurred())
			}
		})

		It("should validate duration constraints", func() {
			By("testing duration limits")
			testCases := []struct {
				duration    string
				shouldFail  bool
				description string
			}{
				{"1s", true, "too short"},
				{"5s", false, "minimum valid"},
				{"30m", false, "reasonable duration"},
				{"2h", false, "maximum valid"},
				{"25h", true, "too long"},
				{"0s", true, "zero duration"},
				{"-5s", true, "negative duration"},
			}

			for _, tc := range testCases {
				spec := v1alpha1.CoDriverJobSpec{
					Targets: v1alpha1.TargetSpec{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "webhook-app"},
						},
					},
					Tool: v1alpha1.ToolSpec{
						Name:     "webhook-config",
						Duration: tc.duration,
					},
					Output: v1alpha1.OutputSpec{
						Mode: "ephemeral",
					},
				}

				coDriverJob := &v1alpha1.CoDriverJob{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "duration-" + tc.duration,
						Namespace: namespace.Name,
					},
					Spec: spec,
				}

				By("testing duration: " + tc.duration + " (" + tc.description + ")")
				err := simpleK8sClient.Create(simpleCtx, coDriverJob)
				if tc.shouldFail {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
					// Cleanup successful creation
					Expect(simpleK8sClient.Delete(simpleCtx, coDriverJob)).To(Succeed())
				}
			}
		})

		It("should validate label selector complexity", func() {
			By("testing complex label selectors")
			validSelector := &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":     "webhook-app",
					"version": "v1.0",
				},
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "tier",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"frontend", "backend"},
					},
				},
			}

			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: validSelector,
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "webhook-config",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			coDriverJob := CreateSimpleTestCoDriverJob("complex-selector", namespace.Name, spec)

			By("verifying complex selector is accepted")
			updated := GetSimpleCoDriverJob(coDriverJob)
			Expect(updated.Spec.Targets.LabelSelector.MatchLabels).To(HaveLen(2))
			Expect(updated.Spec.Targets.LabelSelector.MatchExpressions).To(HaveLen(1))
		})

		It("should validate output configuration consistency", func() {
			By("testing PVC mode without PVC spec")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "webhook-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "webhook-config",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "pvc",
					// Missing PVC spec
				},
			}

			coDriverJob := &v1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-pvc-config",
					Namespace: namespace.Name,
				},
				Spec: spec,
			}

			By("expecting validation to fail for incomplete PVC config")
			err := simpleK8sClient.Create(simpleCtx, coDriverJob)
			Expect(err).To(HaveOccurred())
		})

		It("should validate collector configuration", func() {
			By("testing collector mode with invalid endpoint")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "webhook-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "webhook-config",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "collector",
					Collector: &v1alpha1.CollectorSpec{
						Endpoint: "invalid-url-format",
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

			By("expecting validation to fail for invalid collector endpoint")
			err := simpleK8sClient.Create(simpleCtx, coDriverJob)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("CoDriverTool Validation", func() {
		It("should validate security context requirements", func() {
			By("testing invalid capability combinations")
			allowPrivileged := false
			invalidConfig := &v1alpha1.CoDriverTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-security",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CoDriverToolSpec{
					Name:  "invalid-tool",
					Image: "test:latest",
					SecurityContext: v1alpha1.SecuritySpec{
						AllowPrivileged: &allowPrivileged,
						Capabilities: &v1alpha1.Capabilities{
							Add: []string{"INVALID_CAPABILITY"},
						},
					},
				},
			}

			By("expecting validation to fail for invalid capabilities")
			err := simpleK8sClient.Create(simpleCtx, invalidConfig)
			Expect(err).To(HaveOccurred())
		})

		It("should validate image format", func() {
			By("testing invalid image formats")
			invalidImages := []string{
				"",
				"invalid-image-format",
				"registry.com/",
				"registry.com/:invalid-tag",
			}

			for _, image := range invalidImages {
				allowPrivileged := true
				config := &v1alpha1.CoDriverTool{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "invalid-image-" + image,
						Namespace: namespace.Name,
					},
					Spec: v1alpha1.CoDriverToolSpec{
						Name:  "test-tool",
						Image: image,
						SecurityContext: v1alpha1.SecuritySpec{
							AllowPrivileged: &allowPrivileged,
						},
					},
				}

				By("expecting validation to fail for image: " + image)
				err := simpleK8sClient.Create(simpleCtx, config)
				Expect(err).To(HaveOccurred())
			}
		})

		It("should validate allowed namespaces format", func() {
			By("testing invalid namespace names")
			allowPrivileged := true
			invalidConfig := &v1alpha1.CoDriverTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-namespaces",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CoDriverToolSpec{
					Name:  "test-tool",
					Image: "test:latest",
					SecurityContext: v1alpha1.SecuritySpec{
						AllowPrivileged: &allowPrivileged,
					},
					AllowedNamespaces: []string{
						"valid-namespace",
						"INVALID-NAMESPACE", // Uppercase not allowed
						"invalid namespace", // Spaces not allowed
					},
				},
			}

			By("expecting validation to fail for invalid namespace names")
			err := simpleK8sClient.Create(simpleCtx, invalidConfig)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Mutation Webhook", func() {
		It("should apply default values", func() {
			By("creating CoDriverJob without optional fields")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "webhook-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "webhook-config",
					Duration: "30s",
					// No Args, Env specified
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
					// No Path specified
				},
			}
			coDriverJob := CreateSimpleTestCoDriverJob("default-values", namespace.Name, spec)

			By("verifying default values are applied")
			_ = GetSimpleCoDriverJob(coDriverJob)
		})

		It("should normalize resource names", func() {
			By("creating CoDriverJob with name that needs normalization")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "webhook-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("Name-With-Caps", namespace.Name, spec)

			By("verifying name normalization")
			updated := GetSimpleCoDriverJob(coDriverJob)
			Expect(updated.Name).To(Equal("name-with-caps"))
		})

		It("should add required labels and annotations", func() {
			By("creating CoDriverJob")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "webhook-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("label-annotation-test", namespace.Name, spec)

			By("verifying required labels and annotations are added")
			updated := GetSimpleCoDriverJob(coDriverJob)
			Expect(updated.Labels).To(HaveKey("kubecodriver.codriverlabs.ai/managed-by"))
			Expect(updated.Annotations).To(HaveKey("kubecodriver.codriverlabs.ai/created-at"))
		})
	})

	Context("Webhook Error Handling", func() {
		It("should handle webhook timeout gracefully", func() {
			By("creating CoDriverJob that might trigger webhook timeout")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "webhook-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("timeout-test", namespace.Name, spec)

			By("verifying CoDriverJob is eventually created despite potential delays")
			Eventually(func() error {
				return simpleK8sClient.Get(simpleCtx, types.NamespacedName{Name: coDriverJob.Name, Namespace: coDriverJob.Namespace}, coDriverJob)
			}, "60s", "2s").Should(Succeed())
		})

		It("should provide clear validation error messages", func() {
			By("creating CoDriverJob with multiple validation errors")
			coDriverJob := &v1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "multi-error-test",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CoDriverJobSpec{
					Targets: v1alpha1.TargetSpec{
						// Missing label selector
					},
					Tool: v1alpha1.ToolSpec{
						Name:     "",        // Empty name
						Duration: "invalid", // Invalid duration
					},
					Output: v1alpha1.OutputSpec{
						Mode: "invalid-mode", // Invalid mode
					},
				},
			}

			By("expecting detailed validation error")
			err := simpleK8sClient.Create(simpleCtx, coDriverJob)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("validation"))
		})
	})
})
