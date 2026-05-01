package auth

import (
	"context"
	"errors"
	"testing"

	authv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestNewK8sTokenValidator(t *testing.T) {
	client := fake.NewSimpleClientset()
	audience := "test-audience"

	validator := NewK8sTokenValidator(client, audience)

	if validator == nil {
		t.Fatal("NewK8sTokenValidator returned nil")
	}

	if validator.audience != audience {
		t.Errorf("expected audience %v, got %v", audience, validator.audience)
	}
}

func TestValidateToken_Success(t *testing.T) {
	client := fake.NewSimpleClientset()

	// Configure fake client to return successful authentication
	client.PrependReactor("create", "tokenreviews", func(action k8stesting.Action) (bool, runtime.Object, error) {
		tr := action.(k8stesting.CreateAction).GetObject().(*authv1.TokenReview)
		tr.Status = authv1.TokenReviewStatus{
			Authenticated: true,
			User: authv1.UserInfo{
				Username: "system:serviceaccount:kubecodriver-system:kubecodriver-sdk-collector",
				UID:      "test-uid",
			},
		}
		return true, tr, nil
	})

	validator := NewK8sTokenValidator(client, "kubecodriver-sdk-collector")
	userInfo, err := validator.ValidateToken(context.Background(), "valid-token")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if userInfo == nil {
		t.Fatal("expected userInfo, got nil")
	}

	if userInfo.Username != "system:serviceaccount:kubecodriver-system:kubecodriver-sdk-collector" {
		t.Errorf("expected username 'system:serviceaccount:kubecodriver-system:kubecodriver-sdk-collector', got %v", userInfo.Username)
	}
}

func TestValidateToken_NotAuthenticated(t *testing.T) {
	client := fake.NewSimpleClientset()

	client.PrependReactor("create", "tokenreviews", func(action k8stesting.Action) (bool, runtime.Object, error) {
		tr := action.(k8stesting.CreateAction).GetObject().(*authv1.TokenReview)
		tr.Status = authv1.TokenReviewStatus{
			Authenticated: false,
			Error:         "token expired",
		}
		return true, tr, nil
	})

	validator := NewK8sTokenValidator(client, "kubecodriver-sdk-collector")
	userInfo, err := validator.ValidateToken(context.Background(), "expired-token")

	if err == nil {
		t.Error("expected error, got nil")
	}

	if userInfo != nil {
		t.Errorf("expected nil userInfo, got %v", userInfo)
	}
}

func TestValidateToken_K8sAPIError(t *testing.T) {
	client := fake.NewSimpleClientset()

	client.PrependReactor("create", "tokenreviews", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("API server unavailable")
	})

	validator := NewK8sTokenValidator(client, "kubecodriver-sdk-collector")
	userInfo, err := validator.ValidateToken(context.Background(), "any-token")

	if err == nil {
		t.Error("expected error, got nil")
	}

	if userInfo != nil {
		t.Errorf("expected nil userInfo, got %v", userInfo)
	}
}

func TestValidateToken_WithAudience(t *testing.T) {
	client := fake.NewSimpleClientset()
	expectedAudience := "custom-audience"

	client.PrependReactor("create", "tokenreviews", func(action k8stesting.Action) (bool, runtime.Object, error) {
		tr := action.(k8stesting.CreateAction).GetObject().(*authv1.TokenReview)

		// Verify audience is set correctly
		if len(tr.Spec.Audiences) != 1 || tr.Spec.Audiences[0] != expectedAudience {
			t.Errorf("expected audience [%v], got %v", expectedAudience, tr.Spec.Audiences)
		}

		tr.Status = authv1.TokenReviewStatus{
			Authenticated: true,
			User: authv1.UserInfo{
				Username: "test-user",
			},
		}
		return true, tr, nil
	})

	validator := NewK8sTokenValidator(client, expectedAudience)
	_, err := validator.ValidateToken(context.Background(), "token")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
