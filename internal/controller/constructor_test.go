package controller

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestNewCoDriverJobReconciler(t *testing.T) {
	scheme := runtime.NewScheme()
	var mockClient client.Client
	var mockK8sClient kubernetes.Interface

	r := NewCoDriverJobReconciler(mockClient, scheme, mockK8sClient)

	if r == nil {
		t.Fatal("expected non-nil reconciler")
	}

	if r.Client != mockClient {
		t.Error("Client not set correctly")
	}

	if r.Scheme != scheme {
		t.Error("Scheme not set correctly")
	}

	if r.K8sClient != mockK8sClient {
		t.Error("K8sClient not set correctly")
	}
}
