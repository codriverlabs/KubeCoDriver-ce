/*
Copyright 2025.

*/

package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kubecodriverv1alpha1 "github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

var _ = Describe("CoDriverJob Controller Integration", func() {
	Context("When reconciling a CoDriverJob resource with comprehensive scenarios", func() {
		const resourceName = "integration-test-codriverjob"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}

		BeforeEach(func() {
			By("creating the custom resource for the Kind CoDriverJob")
			coDriverJob := &kubecodriverv1alpha1.CoDriverJob{}
			err := k8sClient.Get(ctx, typeNamespacedName, coDriverJob)
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
									"app": "integration-test",
								},
							},
						},
						Tool: kubecodriverv1alpha1.ToolSpec{
							Name:     "nonexistent-tool",
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
			if err == nil {
				By("Cleanup the specific resource instance CoDriverJob")
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &CoDriverJobReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})

			// Should fail due to missing CoDriverTool, which is expected behavior
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("CoDriverTool not found for tool: nonexistent-tool"))
		})
	})
})
