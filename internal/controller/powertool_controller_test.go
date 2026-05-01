/*
Copyright 2025.

*/

package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kubecodriverv1alpha1 "github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

var _ = Describe("CoDriverJob Controller", func() {
	Context("When reconciling a CoDriverJob resource", func() {
		const resourceName = "test-codriverjob"
		const configName = "aperf-config"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}

		configNamespacedName := types.NamespacedName{
			Name:      configName,
			Namespace: "kubecodriver-system",
		}

		BeforeEach(func() {
			By("creating the kubecodriver-system namespace")
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "kubecodriver-system",
				},
			}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: "kubecodriver-system"}, &corev1.Namespace{})
			if err != nil && errors.IsNotFound(err) {
				Expect(k8sClient.Create(ctx, namespace)).To(Succeed())
			}

			By("creating the CoDriverTool")
			config := &kubecodriverv1alpha1.CoDriverTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      configName,
					Namespace: "kubecodriver-system",
				},
				Spec: kubecodriverv1alpha1.CoDriverToolSpec{
					Name:  "aperf",
					Image: "test-registry/aperf:latest",
					SecurityContext: kubecodriverv1alpha1.SecuritySpec{
						AllowPrivileged: boolPtr(true),
					},
				},
			}
			err = k8sClient.Get(ctx, configNamespacedName, &kubecodriverv1alpha1.CoDriverTool{})
			if err != nil && errors.IsNotFound(err) {
				Expect(k8sClient.Create(ctx, config)).To(Succeed())
			}

			By("creating the CoDriverJob resource")
			coDriverJob := &kubecodriverv1alpha1.CoDriverJob{}
			err = k8sClient.Get(ctx, typeNamespacedName, coDriverJob)
			if err != nil && errors.IsNotFound(err) {
				resource := &kubecodriverv1alpha1.CoDriverJob{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: kubecodriverv1alpha1.CoDriverJobSpec{
						Targets: kubecodriverv1alpha1.TargetSpec{
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app": "test-app",
								},
							},
						},
						Tool: kubecodriverv1alpha1.ToolSpec{
							Name:     "aperf",
							Duration: "30s",
						},
						Output: kubecodriverv1alpha1.OutputSpec{
							Mode: "ephemeral",
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			By("cleaning up the CoDriverJob resource")
			resource := &kubecodriverv1alpha1.CoDriverJob{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}

			By("cleaning up the CoDriverTool")
			config := &kubecodriverv1alpha1.CoDriverTool{}
			err = k8sClient.Get(ctx, configNamespacedName, config)
			if err == nil {
				Expect(k8sClient.Delete(ctx, config)).To(Succeed())
			}
		})

		It("should successfully reconcile and initialize status", func() {
			By("reconciling the created resource")
			controllerReconciler := &CoDriverJobReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("checking that status is initialized")
			coDriverJob := &kubecodriverv1alpha1.CoDriverJob{}
			err = k8sClient.Get(ctx, typeNamespacedName, coDriverJob)
			Expect(err).NotTo(HaveOccurred())

			// Verify status initialization
			Expect(coDriverJob.Status.Phase).NotTo(BeNil())
			Expect(*coDriverJob.Status.Phase).To(Equal("Pending"))
			Expect(coDriverJob.Status.StartedAt).NotTo(BeNil())
			Expect(coDriverJob.Status.SelectedPods).NotTo(BeNil())
			Expect(*coDriverJob.Status.SelectedPods).To(Equal(int32(0))) // No matching pods

			// Verify conditions
			Expect(coDriverJob.Status.Conditions).NotTo(BeEmpty())
			readyCondition := findCondition(coDriverJob.Status.Conditions, kubecodriverv1alpha1.CoDriverJobConditionReady)
			Expect(readyCondition).NotTo(BeNil())
			Expect(readyCondition.Status).To(Equal("False"))
			Expect(readyCondition.Reason).To(Equal(kubecodriverv1alpha1.ReasonTargetsSelected))
		})

		It("should handle missing CoDriverTool gracefully", func() {
			By("deleting the CoDriverTool")
			config := &kubecodriverv1alpha1.CoDriverTool{}
			err := k8sClient.Get(ctx, configNamespacedName, config)
			Expect(err).NotTo(HaveOccurred())
			Expect(k8sClient.Delete(ctx, config)).To(Succeed())

			By("reconciling without config")
			controllerReconciler := &CoDriverJobReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("CoDriverTool not found"))
		})

		It("should detect conflicts with other CoDriverJobs", func() {
			By("creating a target pod")
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
					Labels: map[string]string{
						"app": "test-app",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "nginx:latest",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, pod)).To(Succeed())

			By("creating a second CoDriverJob targeting the same pod")
			conflictingCoDriverJob := &kubecodriverv1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "conflicting-codriverjob",
					Namespace: "default",
				},
				Spec: kubecodriverv1alpha1.CoDriverJobSpec{
					Targets: kubecodriverv1alpha1.TargetSpec{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"app": "test-app",
							},
						},
					},
					Tool: kubecodriverv1alpha1.ToolSpec{
						Name:     "aperf",
						Duration: "30s",
					},
					Output: kubecodriverv1alpha1.OutputSpec{
						Mode: "ephemeral",
					},
				},
			}
			Expect(k8sClient.Create(ctx, conflictingCoDriverJob)).To(Succeed())

			By("reconciling both CoDriverJobs")
			controllerReconciler := &CoDriverJobReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			// First CoDriverJob should succeed
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Second CoDriverJob should detect conflict
			conflictingNamespacedName := types.NamespacedName{
				Name:      "conflicting-codriverjob",
				Namespace: "default",
			}
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: conflictingNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("checking conflict detection")
			conflictingResource := &kubecodriverv1alpha1.CoDriverJob{}
			err = k8sClient.Get(ctx, conflictingNamespacedName, conflictingResource)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() string {
				err := k8sClient.Get(ctx, conflictingNamespacedName, conflictingResource)
				if err != nil {
					return ""
				}
				if conflictingResource.Status.Phase != nil {
					return *conflictingResource.Status.Phase
				}
				return ""
			}, time.Second*5, time.Millisecond*100).Should(Equal("Conflicted"))

			// Cleanup
			Expect(k8sClient.Delete(ctx, conflictingCoDriverJob)).To(Succeed())
			Expect(k8sClient.Delete(ctx, pod)).To(Succeed())
		})
	})
})

// Helper functions
func findCondition(conditions []kubecodriverv1alpha1.CoDriverJobCondition, conditionType string) *kubecodriverv1alpha1.CoDriverJobCondition {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return &condition
		}
	}
	return nil
}

func boolPtr(b bool) *bool {
	return &b
}
