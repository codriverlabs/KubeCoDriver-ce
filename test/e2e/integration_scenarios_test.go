package e2e

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

var _ = Describe("Integration Scenarios", func() {
	var namespace *corev1.Namespace

	BeforeEach(func() {
		namespace = CreateSimpleTestNamespace()
		CreateSimpleTestCoDriverTool("integration-config", namespace.Name)
	})

	AfterEach(func() {
		DeleteSimpleTestNamespace(namespace)
	})

	Context("Real-World Application Scenarios", func() {
		It("should profile a web application deployment", func() {
			By("creating a web application deployment")
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "web-app",
					Namespace: namespace.Name,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(3),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "web-app",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app":  "web-app",
								"tier": "frontend",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "web",
									Image: "nginx:latest",
									Ports: []corev1.ContainerPort{
										{ContainerPort: 80},
									},
								},
							},
						},
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, deployment)).To(Succeed())

			By("waiting for deployment to be ready")
			Eventually(func() int32 {
				updated := &appsv1.Deployment{}
				simpleK8sClient.Get(simpleCtx, client.ObjectKeyFromObject(deployment), updated)
				return updated.Status.ReadyReplicas
			}, "60s", "2s").Should(Equal(int32(3)))

			By("creating CoDriverJob to profile the web application")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app":  "web-app",
							"tier": "frontend",
						},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "integration-config",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			coDriverJob := CreateSimpleTestCoDriverJob("web-app-profile", namespace.Name, spec)

			By("verifying CoDriverJob targets all deployment pods")
			Eventually(func() int {
				updated := GetSimpleCoDriverJob(coDriverJob)
				return len(updated.Status.ActivePods)
			}, "60s", "2s").Should(Equal(3))
		})

		It("should handle database profiling scenario", func() {
			By("creating a database StatefulSet")
			statefulSet := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "database",
					Namespace: namespace.Name,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: int32Ptr(1),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "database",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app":  "database",
								"tier": "backend",
								"type": "stateful",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "db",
									Image: "postgres:13",
									Env: []corev1.EnvVar{
										{Name: "POSTGRES_PASSWORD", Value: "testpass"},
									},
								},
							},
						},
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, statefulSet)).To(Succeed())

			By("creating CoDriverJob for database profiling")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app":  "database",
							"type": "stateful",
						},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "integration-config",
					Duration: "45s",
					Args:     []string{"--database-mode", "--io-trace"},
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			coDriverJob := CreateSimpleTestCoDriverJob("database-profile", namespace.Name, spec)

			By("verifying database profiling configuration")
			WaitForSimpleCoDriverJobPhase(coDriverJob, "Pending")
			updated := GetSimpleCoDriverJob(coDriverJob)
			Expect(updated.Spec.Tool.Args).To(ContainElement("--database-mode"))
		})

		It("should handle microservices profiling", func() {
			By("creating multiple microservice deployments")
			services := []string{"auth-service", "user-service", "order-service"}

			for _, service := range services {
				deployment := &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      service,
						Namespace: namespace.Name,
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: int32Ptr(2),
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": service,
							},
						},
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"app":         service,
									"tier":        "microservice",
									"environment": "test",
								},
							},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "service",
										Image: "nginx:latest", // Placeholder
									},
								},
							},
						},
					},
				}
				Expect(simpleK8sClient.Create(simpleCtx, deployment)).To(Succeed())
			}

			By("creating CoDriverJob to profile all microservices")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"tier":        "microservice",
							"environment": "test",
						},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "integration-config",
					Duration: "60s",
					Args:     []string{"--microservice-mode", "--trace-calls"},
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			coDriverJob := CreateSimpleTestCoDriverJob("microservices-profile", namespace.Name, spec)

			By("verifying all microservice pods are targeted")
			Eventually(func() int {
				updated := GetSimpleCoDriverJob(coDriverJob)
				return len(updated.Status.ActivePods)
			}, "90s", "3s").Should(Equal(6)) // 3 services × 2 replicas
		})
	})

	Context("Multi-Tool Coordination", func() {
		It("should coordinate multiple profiling tools", func() {
			By("creating target application")
			CreateSimpleMockTargetPod(namespace.Name, "multi-tool-app", map[string]string{
				"app": "multi-tool-app",
			})

			By("creating multiple CoDriverTools")
			configs := []struct {
				name string
				tool string
			}{
				{"perf-config", "perf"},
				{"strace-config", "strace"},
				{"tcpdump-config", "tcpdump"},
			}

			for _, config := range configs {
				allowPrivileged := true
				toolConfig := &v1alpha1.CoDriverTool{
					ObjectMeta: metav1.ObjectMeta{
						Name:      config.name,
						Namespace: namespace.Name,
					},
					Spec: v1alpha1.CoDriverToolSpec{
						Name:  config.tool,
						Image: fmt.Sprintf("ghcr.io/codriverlabs/ce/kubecodriver-%s:latest", config.tool),
						SecurityContext: v1alpha1.SecuritySpec{
							AllowPrivileged: &allowPrivileged,
							Capabilities: &v1alpha1.Capabilities{
								Add: []string{"SYS_ADMIN", "SYS_PTRACE"},
							},
						},
					},
				}
				Expect(simpleK8sClient.Create(simpleCtx, toolConfig)).To(Succeed())
			}

			By("creating coordinated CoDriverJobs with time offsets")
			for i, config := range configs {
				spec := v1alpha1.CoDriverJobSpec{
					Targets: v1alpha1.TargetSpec{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "multi-tool-app"},
						},
					},
					Tool: v1alpha1.ToolSpec{
						Name:     config.tool,
						Duration: "30s",
					},
					Output: v1alpha1.OutputSpec{
						Mode: "ephemeral",
					},
				}

				// Stagger the creation to avoid conflicts
				if i > 0 {
					time.Sleep(5 * time.Second)
				}

				coDriverJob := CreateSimpleTestCoDriverJob(fmt.Sprintf("multi-tool-%s", config.tool), namespace.Name, spec)

				By(fmt.Sprintf("verifying %s tool is configured", config.tool))
				WaitForSimpleCoDriverJobCondition(coDriverJob, "ToolConfigured", "True")
			}
		})

		It("should handle tool priority and scheduling", func() {
			By("creating target pod")
			CreateSimpleMockTargetPod(namespace.Name, "priority-app", map[string]string{
				"app": "priority-app",
			})

			By("creating high-priority CoDriverJob")
			highPrioritySpec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "priority-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "integration-config",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			highPriorityTool := CreateSimpleTestCoDriverJob("high-priority", namespace.Name, highPrioritySpec)

			By("creating low-priority CoDriverJob")
			lowPrioritySpec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "priority-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "integration-config",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			lowPriorityTool := CreateSimpleTestCoDriverJob("low-priority", namespace.Name, lowPrioritySpec)

			By("verifying priority handling")
			// First tool should succeed
			WaitForSimpleCoDriverJobPhase(highPriorityTool, "Pending")

			// Second tool should detect conflict
			Eventually(func() string {
				updated := GetSimpleCoDriverJob(lowPriorityTool)
				if updated.Status.Phase == nil {
					return ""
				}
				return *updated.Status.Phase
			}, "30s", "1s").Should(Equal("Failed"))
		})
	})

	Context("Performance and Scale Integration", func() {
		It("should handle large-scale deployments", func() {
			By("creating large deployment")
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "large-app",
					Namespace: namespace.Name,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(10),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "large-app",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app":   "large-app",
								"scale": "large",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "app",
									Image: "nginx:latest",
								},
							},
						},
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, deployment)).To(Succeed())

			By("creating CoDriverJob for large-scale profiling")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app":   "large-app",
							"scale": "large",
						},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "integration-config",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			coDriverJob := CreateSimpleTestCoDriverJob("large-scale-profile", namespace.Name, spec)

			By("verifying large-scale profiling performance")
			startTime := time.Now()
			Eventually(func() int {
				updated := GetSimpleCoDriverJob(coDriverJob)
				return len(updated.Status.ActivePods)
			}, "120s", "3s").Should(Equal(10))

			elapsed := time.Since(startTime)
			Expect(elapsed).To(BeNumerically("<", 90*time.Second))
		})

		It("should maintain performance under resource constraints", func() {
			By("creating resource-constrained pods")
			for i := 0; i < 5; i++ {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("constrained-pod-%d", i),
						Namespace: namespace.Name,
						Labels: map[string]string{
							"app":        "constrained-app",
							"constraint": "resource",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "app",
								Image: "nginx:latest",
								Resources: corev1.ResourceRequirements{
									Limits: corev1.ResourceList{
										"cpu":    resource.MustParse("100m"),
										"memory": resource.MustParse("128Mi"),
									},
									Requests: corev1.ResourceList{
										"cpu":    resource.MustParse("50m"),
										"memory": resource.MustParse("64Mi"),
									},
								},
							},
						},
					},
				}
				Expect(simpleK8sClient.Create(simpleCtx, pod)).To(Succeed())

				// Update status to running
				pod.Status.Phase = corev1.PodRunning
				pod.Status.ContainerStatuses = []corev1.ContainerStatus{
					{
						Name:  "app",
						Ready: true,
						State: corev1.ContainerState{
							Running: &corev1.ContainerStateRunning{},
						},
					},
				}
				Expect(simpleK8sClient.Status().Update(simpleCtx, pod)).To(Succeed())
			}

			By("creating CoDriverJob under resource constraints")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app":        "constrained-app",
							"constraint": "resource",
						},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "integration-config",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			coDriverJob := CreateSimpleTestCoDriverJob("constrained-profile", namespace.Name, spec)

			By("verifying performance under constraints")
			Eventually(func() int {
				updated := GetSimpleCoDriverJob(coDriverJob)
				return len(updated.Status.ActivePods)
			}, "60s", "2s").Should(Equal(5))
		})
	})

	Context("Error Recovery Integration", func() {
		It("should recover from pod restarts during profiling", func() {
			By("creating deployment that will be restarted")
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "restart-app",
					Namespace: namespace.Name,
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: int32Ptr(2),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "restart-app",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app":      "restart-app",
								"recovery": "test",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "app",
									Image: "nginx:latest",
								},
							},
						},
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, deployment)).To(Succeed())

			By("creating CoDriverJob")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app":      "restart-app",
							"recovery": "test",
						},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "integration-config",
					Duration: "60s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			coDriverJob := CreateSimpleTestCoDriverJob("recovery-test", namespace.Name, spec)

			By("waiting for initial profiling setup")
			WaitForSimpleCoDriverJobPhase(coDriverJob, "Pending")

			By("simulating pod restart by updating deployment")
			updated := &appsv1.Deployment{}
			Expect(simpleK8sClient.Get(simpleCtx, client.ObjectKeyFromObject(deployment), updated)).To(Succeed())
			updated.Spec.Template.Spec.Containers[0].Env = append(updated.Spec.Template.Spec.Containers[0].Env,
				corev1.EnvVar{Name: "RESTART_TRIGGER", Value: "true"},
			)
			Expect(simpleK8sClient.Update(simpleCtx, updated)).To(Succeed())

			By("verifying CoDriverJob handles pod restarts gracefully")
			Consistently(func() string {
				current := GetSimpleCoDriverJob(coDriverJob)
				if current.Status.Phase == nil {
					return ""
				}
				return *current.Status.Phase
			}, "30s", "2s").Should(Or(Equal("Pending"), Equal("Running")))
		})
	})
})

// Helper function for int32 pointer
func int32Ptr(i int32) *int32 {
	return &i
}
