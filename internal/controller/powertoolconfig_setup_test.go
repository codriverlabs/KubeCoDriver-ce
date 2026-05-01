package controller

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("CoDriverToolReconciler SetupWithManager", func() {
	It("should set up the reconciler with manager", func() {
		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: k8sClient.Scheme(),
		})
		Expect(err).NotTo(HaveOccurred())

		r := &CoDriverToolReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}

		err = r.SetupWithManager(mgr)
		Expect(err).NotTo(HaveOccurred())
	})
})
