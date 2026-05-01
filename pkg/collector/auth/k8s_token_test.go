package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	authv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestNewK8sTokenManager(t *testing.T) {
	client := fake.NewSimpleClientset()
	namespace := "test-namespace"
	audience := "test-audience"

	manager := NewK8sTokenManager(client, namespace, audience)

	if manager == nil {
		t.Fatal("NewK8sTokenManager returned nil")
	}

	if manager.namespace != namespace {
		t.Errorf("expected namespace %v, got %v", namespace, manager.namespace)
	}

	if manager.audience != audience {
		t.Errorf("expected audience %v, got %v", audience, manager.audience)
	}
}

func TestGenerateToken_Success(t *testing.T) {
	// Create service account
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kubecodriver-collector",
			Namespace: "kubecodriver-system",
		},
	}

	client := fake.NewSimpleClientset(sa)

	client.PrependReactor("create", "serviceaccounts", func(action k8stesting.Action) (bool, runtime.Object, error) {
		// This is for CreateToken subresource
		if action.GetSubresource() == "token" {
			return true, &authv1.TokenRequest{
				Status: authv1.TokenRequestStatus{
					Token: "generated-token-12345",
				},
			}, nil
		}
		return false, nil, nil
	})

	manager := NewK8sTokenManager(client, "kubecodriver-system", "kubecodriver-sdk-collector")
	token, err := manager.GenerateToken(context.Background(), "test-job", 10*time.Minute)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if token != "generated-token-12345" {
		t.Errorf("expected token 'generated-token-12345', got %v", token)
	}
}

func TestGenerateToken_Duration(t *testing.T) {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kubecodriver-collector",
			Namespace: "kubecodriver-system",
		},
	}

	client := fake.NewSimpleClientset(sa)

	var capturedDuration int64
	client.PrependReactor("create", "serviceaccounts", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() == "token" {
			createAction := action.(k8stesting.CreateAction)
			tr := createAction.GetObject().(*authv1.TokenRequest)
			if tr.Spec.ExpirationSeconds != nil {
				capturedDuration = *tr.Spec.ExpirationSeconds
			}
			return true, &authv1.TokenRequest{
				Status: authv1.TokenRequestStatus{
					Token: "test-token",
				},
			}, nil
		}
		return false, nil, nil
	})

	manager := NewK8sTokenManager(client, "kubecodriver-system", "kubecodriver-sdk-collector")
	duration := 15 * time.Minute
	_, err := manager.GenerateToken(context.Background(), "test-job", duration)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedSeconds := int64(duration.Seconds())
	if capturedDuration != expectedSeconds {
		t.Errorf("expected duration %v seconds, got %v", expectedSeconds, capturedDuration)
	}
}

func TestGenerateToken_ServiceAccountNotFound(t *testing.T) {
	// Don't create service account
	client := fake.NewSimpleClientset()

	client.PrependReactor("create", "serviceaccounts", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() == "token" {
			return true, nil, errors.New("serviceaccount not found")
		}
		return false, nil, nil
	})

	manager := NewK8sTokenManager(client, "kubecodriver-system", "kubecodriver-sdk-collector")
	token, err := manager.GenerateToken(context.Background(), "test-job", 10*time.Minute)

	if err == nil {
		t.Error("expected error, got nil")
	}

	if token != "" {
		t.Errorf("expected empty token, got %v", token)
	}
}

func TestGenerateToken_K8sAPIError(t *testing.T) {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kubecodriver-collector",
			Namespace: "kubecodriver-system",
		},
	}

	client := fake.NewSimpleClientset(sa)

	client.PrependReactor("create", "serviceaccounts", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.GetSubresource() == "token" {
			return true, nil, errors.New("API server error")
		}
		return false, nil, nil
	})

	manager := NewK8sTokenManager(client, "kubecodriver-system", "kubecodriver-sdk-collector")
	token, err := manager.GenerateToken(context.Background(), "test-job", 10*time.Minute)

	if err == nil {
		t.Error("expected error, got nil")
	}

	if token != "" {
		t.Errorf("expected empty token, got %v", token)
	}
}
