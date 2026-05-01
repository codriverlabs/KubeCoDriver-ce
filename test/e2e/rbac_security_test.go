package e2e

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

var _ = Describe("RBAC and Security", func() {
	var namespace *corev1.Namespace
	var restrictedNamespace *corev1.Namespace

	BeforeEach(func() {
		namespace = CreateSimpleTestNamespace()
		restrictedNamespace = CreateSimpleTestNamespace()

		// Add restricted label to second namespace
		restrictedNamespace.Labels = map[string]string{
			"kubecodriver.codriverlabs.ai/restricted": "true",
		}
		Expect(simpleK8sClient.Update(simpleCtx, restrictedNamespace)).To(Succeed())

		CreateSimpleTestCoDriverTool("rbac-config", namespace.Name)
		CreateSimpleMockTargetPod(namespace.Name, "rbac-pod", map[string]string{
			"app": "rbac-app",
		})
	})

	AfterEach(func() {
		DeleteSimpleTestNamespace(namespace)
		DeleteSimpleTestNamespace(restrictedNamespace)
	})

	Context("Namespace Access Control", func() {
		It("should enforce namespace restrictions", func() {
			By("creating CoDriverTool with namespace restrictions")
			allowPrivileged := true
			restrictedConfig := &v1alpha1.CoDriverTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "restricted-config",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CoDriverToolSpec{
					Name:  "restricted-tool",
					Image: "ghcr.io/codriverlabs/ce/kubecodriver-restricted:latest",
					SecurityContext: v1alpha1.SecuritySpec{
						AllowPrivileged: &allowPrivileged,
					},
					AllowedNamespaces: []string{namespace.Name}, // Only allow current namespace
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, restrictedConfig)).To(Succeed())

			By("creating target pod in restricted namespace")
			CreateSimpleMockTargetPod(restrictedNamespace.Name, "restricted-pod", map[string]string{
				"app": "restricted-app",
			})

			By("attempting to create CoDriverJob in restricted namespace")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "restricted-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "restricted-tool",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}

			coDriverJob := &v1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "restricted-test",
					Namespace: restrictedNamespace.Name,
				},
				Spec: spec,
			}

			By("expecting creation to fail due to namespace restrictions")
			err := simpleK8sClient.Create(simpleCtx, coDriverJob)
			Expect(err).To(HaveOccurred())
		})

		It("should allow access to permitted namespaces", func() {
			By("creating CoDriverTool allowing multiple namespaces")
			allowPrivileged := true
			multiNsConfig := &v1alpha1.CoDriverTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "multi-ns-config",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CoDriverToolSpec{
					Name:  "multi-ns-tool",
					Image: "ghcr.io/codriverlabs/ce/kubecodriver-multi:latest",
					SecurityContext: v1alpha1.SecuritySpec{
						AllowPrivileged: &allowPrivileged,
					},
					AllowedNamespaces: []string{namespace.Name, restrictedNamespace.Name},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, multiNsConfig)).To(Succeed())

			By("creating CoDriverJob in allowed namespace")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "rbac-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "multi-ns-tool",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			coDriverJob := CreateSimpleTestCoDriverJob("multi-ns-test", namespace.Name, spec)

			By("verifying CoDriverJob is accepted")
			WaitForSimpleCoDriverJobPhase(coDriverJob, "Pending")
		})

		It("should handle empty allowed namespaces (allow all)", func() {
			By("creating CoDriverTool with no namespace restrictions")
			allowPrivileged := true
			openConfig := &v1alpha1.CoDriverTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "open-config",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CoDriverToolSpec{
					Name:  "open-tool",
					Image: "ghcr.io/codriverlabs/ce/kubecodriver-open:latest",
					SecurityContext: v1alpha1.SecuritySpec{
						AllowPrivileged: &allowPrivileged,
					},
					// AllowedNamespaces is empty, should allow all
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, openConfig)).To(Succeed())

			By("creating CoDriverJob in any namespace")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "rbac-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "open-tool",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			coDriverJob := CreateSimpleTestCoDriverJob("open-test", namespace.Name, spec)

			By("verifying CoDriverJob is accepted")
			WaitForSimpleCoDriverJobPhase(coDriverJob, "Pending")
		})
	})

	Context("Security Context Validation", func() {
		It("should validate privileged mode requirements", func() {
			By("creating CoDriverTool requiring privileged mode")
			allowPrivileged := true
			privilegedConfig := &v1alpha1.CoDriverTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "privileged-config",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CoDriverToolSpec{
					Name:  "privileged-tool",
					Image: "ghcr.io/codriverlabs/ce/kubecodriver-privileged:latest",
					SecurityContext: v1alpha1.SecuritySpec{
						AllowPrivileged: &allowPrivileged,
						Capabilities: &v1alpha1.Capabilities{
							Add: []string{"SYS_ADMIN", "SYS_PTRACE"},
						},
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, privilegedConfig)).To(Succeed())

			By("creating CoDriverJob using privileged configuration")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "rbac-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "privileged-tool",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			coDriverJob := CreateSimpleTestCoDriverJob("privileged-test", namespace.Name, spec)

			By("verifying privileged CoDriverJob is configured")
			WaitForSimpleCoDriverJobCondition(coDriverJob, "ToolConfigured", "True")
		})

		It("should handle capability restrictions", func() {
			By("creating CoDriverTool with specific capabilities")
			allowPrivileged := false
			capConfig := &v1alpha1.CoDriverTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "capability-config",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CoDriverToolSpec{
					Name:  "capability-tool",
					Image: "ghcr.io/codriverlabs/ce/kubecodriver-cap:latest",
					SecurityContext: v1alpha1.SecuritySpec{
						AllowPrivileged: &allowPrivileged,
						Capabilities: &v1alpha1.Capabilities{
							Add:  []string{"NET_ADMIN", "SYS_TIME"},
							Drop: []string{"MKNOD", "AUDIT_WRITE"},
						},
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, capConfig)).To(Succeed())

			By("creating CoDriverJob with capability requirements")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "rbac-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "capability-tool",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			coDriverJob := CreateSimpleTestCoDriverJob("capability-test", namespace.Name, spec)

			By("verifying capability configuration is applied")
			WaitForSimpleCoDriverJobCondition(coDriverJob, "ToolConfigured", "True")
		})

		It("should enforce hostPID restrictions", func() {
			By("creating CoDriverTool with hostPID requirements")
			allowPrivileged := false
			allowHostPID := true
			hostPIDConfig := &v1alpha1.CoDriverTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hostpid-config",
					Namespace: namespace.Name,
				},
				Spec: v1alpha1.CoDriverToolSpec{
					Name:  "hostpid-tool",
					Image: "ghcr.io/codriverlabs/ce/kubecodriver-hostpid:latest",
					SecurityContext: v1alpha1.SecuritySpec{
						AllowPrivileged: &allowPrivileged,
						AllowHostPID:    &allowHostPID,
						Capabilities: &v1alpha1.Capabilities{
							Add: []string{"SYS_PTRACE"},
						},
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, hostPIDConfig)).To(Succeed())

			By("creating CoDriverJob requiring hostPID access")
			spec := v1alpha1.CoDriverJobSpec{
				Targets: v1alpha1.TargetSpec{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "rbac-app"},
					},
				},
				Tool: v1alpha1.ToolSpec{
					Name:     "hostpid-tool",
					Duration: "30s",
				},
				Output: v1alpha1.OutputSpec{
					Mode: "ephemeral",
				},
			}
			coDriverJob := CreateSimpleTestCoDriverJob("hostpid-test", namespace.Name, spec)

			By("verifying hostPID configuration is accepted")
			WaitForSimpleCoDriverJobCondition(coDriverJob, "ToolConfigured", "True")
		})
	})

	Context("Service Account and RBAC", func() {
		It("should handle custom service accounts", func() {
			By("creating custom service account")
			serviceAccount := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kubecodriver-custom-sa",
					Namespace: namespace.Name,
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, serviceAccount)).To(Succeed())

			By("creating role for service account")
			role := &rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kubecodriver-custom-role",
					Namespace: namespace.Name,
				},
				Rules: []rbacv1.PolicyRule{
					{
						APIGroups: []string{""},
						Resources: []string{"pods"},
						Verbs:     []string{"get", "list", "watch"},
					},
					{
						APIGroups: []string{""},
						Resources: []string{"pods/ephemeralcontainers"},
						Verbs:     []string{"create", "update", "patch"},
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, role)).To(Succeed())

			By("creating role binding")
			roleBinding := &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kubecodriver-custom-binding",
					Namespace: namespace.Name,
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      "ServiceAccount",
						Name:      "kubecodriver-custom-sa",
						Namespace: namespace.Name,
					},
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "Role",
					Name:     "kubecodriver-custom-role",
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, roleBinding)).To(Succeed())

			By("verifying RBAC resources are created")
			Eventually(func() error {
				return simpleK8sClient.Get(simpleCtx, client.ObjectKeyFromObject(serviceAccount), serviceAccount)
			}, "30s", "1s").Should(Succeed())
		})

		It("should validate required permissions", func() {
			By("creating CoDriverJob that requires specific permissions")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "rbac-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("permission-test", namespace.Name, spec)

			By("verifying CoDriverJob can access required resources")
			WaitForSimpleCoDriverJobPhase(coDriverJob, "Pending")

			By("verifying no permission errors in status")
			updated := GetSimpleCoDriverJob(coDriverJob)
			if updated.Status.LastError != nil {
				Expect(*updated.Status.LastError).NotTo(ContainSubstring("forbidden"))
				Expect(*updated.Status.LastError).NotTo(ContainSubstring("unauthorized"))
			}
		})
	})

	Context("Resource Quotas and Limits", func() {
		It("should respect resource quotas", func() {
			By("creating resource quota")
			quota := &corev1.ResourceQuota{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kubecodriver-quota",
					Namespace: namespace.Name,
				},
				Spec: corev1.ResourceQuotaSpec{
					Hard: corev1.ResourceList{
						"requests.cpu":    resource.MustParse("1"),
						"requests.memory": resource.MustParse("1Gi"),
						"limits.cpu":      resource.MustParse("2"),
						"limits.memory":   resource.MustParse("2Gi"),
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, quota)).To(Succeed())

			By("creating CoDriverJob within quota limits")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "rbac-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("quota-test", namespace.Name, spec)

			By("verifying CoDriverJob respects quotas")
			WaitForSimpleCoDriverJobPhase(coDriverJob, "Pending")
		})

		It("should handle limit ranges", func() {
			By("creating limit range")
			limitRange := &corev1.LimitRange{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kubecodriver-limits",
					Namespace: namespace.Name,
				},
				Spec: corev1.LimitRangeSpec{
					Limits: []corev1.LimitRangeItem{
						{
							Type: corev1.LimitTypeContainer,
							Default: corev1.ResourceList{
								"cpu":    resource.MustParse("100m"),
								"memory": resource.MustParse("128Mi"),
							},
							DefaultRequest: corev1.ResourceList{
								"cpu":    resource.MustParse("50m"),
								"memory": resource.MustParse("64Mi"),
							},
						},
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, limitRange)).To(Succeed())

			By("creating CoDriverJob with limit range constraints")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "rbac-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("limit-range-test", namespace.Name, spec)

			By("verifying CoDriverJob works within limit ranges")
			WaitForSimpleCoDriverJobPhase(coDriverJob, "Pending")
		})
	})

	Context("Network Policies", func() {
		It("should handle network policy restrictions", func() {
			By("creating network policy")
			networkPolicy := &networkingv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kubecodriver-netpol",
					Namespace: namespace.Name,
				},
				Spec: networkingv1.NetworkPolicySpec{
					PodSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "rbac-app"},
					},
					PolicyTypes: []networkingv1.PolicyType{
						networkingv1.PolicyTypeIngress,
						networkingv1.PolicyTypeEgress,
					},
					Ingress: []networkingv1.NetworkPolicyIngressRule{
						{
							From: []networkingv1.NetworkPolicyPeer{
								{
									PodSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{"role": "profiler"},
									},
								},
							},
						},
					},
				},
			}
			Expect(simpleK8sClient.Create(simpleCtx, networkPolicy)).To(Succeed())

			By("creating CoDriverJob with network policy constraints")
			spec := CreateSimpleBasicCoDriverJobSpec(map[string]string{"app": "rbac-app"})
			coDriverJob := CreateSimpleTestCoDriverJob("netpol-test", namespace.Name, spec)

			By("verifying CoDriverJob handles network policies")
			WaitForSimpleCoDriverJobPhase(coDriverJob, "Pending")
		})
	})
})
