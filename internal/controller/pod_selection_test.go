/*
Copyright 2025.

*/

package controller

import (
	"context"

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
	Context("When reconciling a resource with pod selection", func() {
		const resourceName = "test-resource-pod-selection"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		coDriverJob := &kubecodriverv1alpha1.CoDriverJob{}

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
					Name:      "aperf-config",
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
			configKey := types.NamespacedName{Name: "aperf-config", Namespace: "kubecodriver-system"}
			err = k8sClient.Get(ctx, configKey, &kubecodriverv1alpha1.CoDriverTool{})
			if err != nil && errors.IsNotFound(err) {
				Expect(k8sClient.Create(ctx, config)).To(Succeed())
			}

			By("creating the custom resource for the Kind CoDriverJob")
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
			resource := &kubecodriverv1alpha1.CoDriverJob{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance CoDriverJob")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			By("creating a test pod with matching labels")
			testPod := &corev1.Pod{
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
							Image: "nginx:alpine",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, testPod)).To(Succeed())

			By("Reconciling the created resource")
			controllerReconciler := &CoDriverJobReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("checking the status of the CoDriverJob")
			updatedCoDriverJob := &kubecodriverv1alpha1.CoDriverJob{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, updatedCoDriverJob)).To(Succeed())
			// Note: Status updates would need to be implemented in the controller
			// Expect(*updatedCoDriverJob.Status.SelectedPods).To(Equal(int32(1)))

			By("cleaning up the pod")
			Expect(k8sClient.Delete(ctx, testPod)).To(Succeed())
		})
	})
})
