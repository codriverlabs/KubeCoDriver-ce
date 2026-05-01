//go:build e2ekind
// +build e2ekind

package e2ekind

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

var (
	k8sClient client.Client
	clientset *kubernetes.Clientset
	scheme    *runtime.Scheme
	ctx       = context.Background()
)

func TestE2EKind(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TOE Kind E2E Suite")
}

var _ = BeforeSuite(func() {
	By("connecting to Kind cluster")
	config, err := ctrl.GetConfig()
	Expect(err).NotTo(HaveOccurred())
	Expect(config).NotTo(BeNil())

	By("initializing scheme")
	scheme = runtime.NewScheme()
	Expect(corev1.AddToScheme(scheme)).To(Succeed())
	Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())

	By("creating Kubernetes client")
	k8sClient, err = client.New(config, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	By("creating Kubernetes clientset")
	clientset, err = kubernetes.NewForConfig(config)
	Expect(err).NotTo(HaveOccurred())
	Expect(clientset).NotTo(BeNil())

	By("verifying cluster connectivity")
	_, err = clientset.Discovery().ServerVersion()
	Expect(err).NotTo(HaveOccurred())

	By("verifying TOE CRDs are installed")
	Eventually(func() error {
		ptList := &v1alpha1.CoDriverJobList{}
		return k8sClient.List(ctx, ptList)
	}, "30s", "2s").Should(Succeed())

	fmt.Fprintf(GinkgoWriter, "✅ Kind E2E suite initialized successfully\n")
})

var _ = AfterSuite(func() {
	By("cleaning up test resources")
	// Cleanup is handled by teardown-cluster.sh
	fmt.Fprintf(GinkgoWriter, "✅ Kind E2E suite completed\n")
})
