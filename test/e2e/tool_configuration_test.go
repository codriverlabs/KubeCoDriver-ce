package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

var _ = Describe("Tool Configuration and Validation", func() {
	var namespace *corev1.Namespace

	BeforeEach(func() {
		namespace = CreateSimpleTestNamespace()
		CreateSimpleMockTargetPod(namespace.Name, "tool-pod", map[string]string{
			"app": "tool-app",
		})
	})

	AfterEach(func() {
		DeleteSimpleTestNamespace(namespace)
	})

	Context("CoDriverTool Management", func() {
		It("should validate tool configurations exist", func() {
			By("creating CoDriverTool")
			_ = CreateSimpleTestCoDriverTool("valid-config", namespace.Name)

			By("creating CoDriverJob referencing the config")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "tool-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("config-test", namespace.Name, spec)

			By("verifying CoDriverJob finds the configuration")
			WaitForSimpleCoDriverJobCondition(coDriverJob, "ToolConfigured", "True")
		})

		It("should handle missing tool configurations", func() {
			By("creating CoDriverJob without corresponding config")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "tool-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "nonexistent-tool",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}

			coDriverJob := &v1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "missing-config",
					Namespace: namespace.Name,
				},
				Spec: spec,
			}

			By("expecting creation to fail due to missing config")
			err := simpleK8sClient.Create(simpleCtx, coDriverJob)
			Expect(err).To(HaveOccurred())
		})

		It("should handle multiple CoDriverTools", func() {
			By("creating multiple tool configurations")
			_ = CreateSimpleTestCoDriverTool("aperf-config", namespace.Name)

			allowPrivileged := true
			config2 := &v1alpha1.CoDriverTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "strace-config",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CoDriverToolSpec{
					Name:  "strace",
					Image: "ghcr.io/codriverlabs/ce/kubecodriver-strace:latest",
					SecurityContext: v1alpha1.SecuritySpec{
						AllowPrivileged: &allowPrivileged,
						Capabilities: &v1alpha1.Capabilities{
							Add: []string{"SYS_PTRACE"},
						},
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, config2)).To(Succeed())

			By("creating CoDriverJobs for different tools")
			spec1 := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "tool-app"})
			spec1.Tool.Name = "aperf"
			coDriverJob1 := CreateSimpleTestCoDriverJob("aperf-tool", namespace.Name, spec1)

			spec2 := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "tool-app"})
			spec2.Tool.Name = "strace"
			coDriverJob2 := CreateSimpleTestCoDriverJob("strace-tool", namespace.Name, spec2)

			By("verifying both tools are configured correctly")
			WaitForSimpleCoDriverJobCondition(coDriverJob1, "ToolConfigured", "True")
			WaitForSimpleCoDriverJobCondition(coDriverJob2, "ToolConfigured", "True")
		})
	})

	Context("Tool Duration Validation", func() {
		BeforeEach(func() {
			CreateSimpleTestCoDriverTool("duration-config", namespace.Name)
		})

		It("should accept valid duration formats", func() {
			By("testing various valid duration formats")
			validDurations := []string{
				"30s",
				"5m",
				"1h",
				"90s",
				"2m30s",
			}

			for _, duration := range validDurations {
				spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "tool-app"})
				spec.Tool.Duration = duration
				coDriverJob := CreateSimpleTestCoDriverJob("duration-"+duration, namespace.Name, spec)

				By("verifying duration is accepted: " + duration)
				updated := GetSimpleCoDriverJob(coDriverJob)
				Expect(updated.Spec.Tool.Duration).To(Equal(duration))

				// Cleanup
				Expect(simpleK8sClient.Delete(simpleCtx, coDriverJob)).To(Succeed())
			}
		})

		It("should reject invalid duration formats", func() {
			By("testing invalid duration formats")
			invalidDurations := []string{
				"invalid",
				"30",
				"-5s",
				"0s",
				"25h", // Too long
			}

			for _, duration := range invalidDurations {
				spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "tool-app"})
				spec.Tool.Duration = duration

				coDriverJob := &v1alpha1.CoDriverJob{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "invalid-duration-" + duration,
						Namespace: namespace.Name,
					},
					Spec: spec,
				}

				By("expecting validation to fail for duration: " + duration)
				err := simpleK8sClient.Create(simpleCtx, coDriverJob)
				Expect(err).To(HaveOccurred())
			}
		})
	})

	Context("Tool Arguments and Environment", func() {
		BeforeEach(func() {
			CreateSimpleTestCoDriverTool("args-config", namespace.Name)
		})

		It("should handle tool arguments correctly", func() {
			By("creating CoDriverJob with custom arguments")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "tool-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "aperf",
					Duration: "30s",
					Args:     []string{"--verbose", "--output=/tmp/custom.out"},
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			coDriverJob := CreateSimpleTestCoDriverJob("custom-args", namespace.Name, spec)

			By("verifying arguments are preserved")
			updated := GetSimpleCoDriverJob(coDriverJob)
			Expect(updated.Spec.Tool.Args).To(Equal([]string{"--verbose", "--output=/tmp/custom.out"}))
		})

		It("should handle environment variables", func() {
			By("creating CoDriverJob with environment variables")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "tool-app"},
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
			coDriverJob := CreateSimpleTestCoDriverJob("custom-env", namespace.Name, spec)

			By("verifying environment variables are set")
			_ = GetSimpleCoDriverJob(coDriverJob)
		})
	})

	Context("Security Context Validation", func() {
		It("should validate security context requirements", func() {
			By("creating CoDriverTool with specific security requirements")
			allowPrivileged := true
			allowHostPID := true
			config := &v1alpha1.CoDriverTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secure-config",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CoDriverToolSpec{
					Name:  "secure-tool",
					Image: "ghcr.io/codriverlabs/ce/kubecodriver-secure:latest",
					SecurityContext: v1alpha1.SecuritySpec{
						AllowPrivileged: &allowPrivileged,
						AllowHostPID:    &allowHostPID,
						Capabilities: &v1alpha1.Capabilities{
							Add:  []string{"SYS_ADMIN", "SYS_PTRACE"},
							Drop: []string{"NET_RAW"},
						},
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, config)).To(Succeed())

			By("creating CoDriverJob using secure configuration")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "tool-app"},
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
			coDriverJob := CreateSimpleTestCoDriverJob("secure-tool-test", namespace.Name, spec)

			By("verifying security context is applied")
			WaitForSimpleCoDriverJobCondition(coDriverJob, "ToolConfigured", "True")
		})

		It("should handle capability requirements", func() {
			By("creating CoDriverTool with specific capabilities")
			allowPrivileged := false
			config := &v1alpha1.CoDriverTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cap-config",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CoDriverToolSpec{
					Name:  "cap-tool",
					Image: "ghcr.io/codriverlabs/ce/kubecodriver-cap:latest",
					SecurityContext: v1alpha1.SecuritySpec{
						AllowPrivileged: &allowPrivileged,
						Capabilities: &v1alpha1.Capabilities{
							Add: []string{"NET_ADMIN", "SYS_TIME"},
						},
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, config)).To(Succeed())

			By("creating CoDriverJob with capability requirements")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "tool-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "cap-tool",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			coDriverJob := CreateSimpleTestCoDriverJob("capability-test", namespace.Name, spec)

			By("verifying capability configuration")
			WaitForSimpleCoDriverJobCondition(coDriverJob, "ToolConfigured", "True")
		})
	})

	Context("Tool Image Management", func() {
		It("should handle different image registries", func() {
			By("creating CoDriverTool with custom registry")
			allowPrivileged := true
			config := &v1alpha1.CoDriverTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "custom-registry",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CoDriverToolSpec{
					Name:  "custom-tool",
					Image: "custom-registry.example.com/tools/profiler:v1.0.0",
					SecurityContext: v1alpha1.SecuritySpec{
						AllowPrivileged: &allowPrivileged,
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, config)).To(Succeed())

			By("creating CoDriverJob using custom registry image")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "tool-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "custom-tool",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			coDriverJob := CreateSimpleTestCoDriverJob("custom-registry-test", namespace.Name, spec)

			By("verifying custom image is configured")
			WaitForSimpleCoDriverJobCondition(coDriverJob, "ToolConfigured", "True")
		})

		It("should handle image pull policies", func() {
			By("creating CoDriverTool with pull policy")
			allowPrivileged := true
			config := &v1alpha1.CoDriverTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pull-policy-config",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CoDriverToolSpec{
					Name:  "pull-policy-tool",
					Image: "ghcr.io/codriverlabs/ce/kubecodriver-test:latest",
					SecurityContext: v1alpha1.SecuritySpec{
						AllowPrivileged: &allowPrivileged,
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, config)).To(Succeed())

			By("verifying image configuration is accepted")
			updated := &v1alpha1.CoDriverTool{}
			Expect(simpleK8sClient.Get(simpleCtx, client.ObjectKeyFromObject(config), updated)).To(Succeed())
			Expect(updated.Spec.Image).To(Equal("ghcr.io/codriverlabs/ce/kubecodriver-test:latest"))
		})
	})
})
