package auth

import (
	"context"
	"fmt"
	"time"

	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type K8sTokenManager struct {
	client    kubernetes.Interface
	namespace string
	audience  string
}

func NewK8sTokenManager(client kubernetes.Interface, namespace, audience string) *K8sTokenManager {
	return &K8sTokenManager{
		client:    client,
		namespace: namespace,
		audience:  audience,
	}
}

func (tm *K8sTokenManager) GenerateToken(ctx context.Context, jobID string, duration time.Duration) (string, error) {
	// Create TokenRequest for the collector's ServiceAccount
	treq := &authv1.TokenRequest{
		Spec: authv1.TokenRequestSpec{
			Audiences:         []string{tm.audience},
			ExpirationSeconds: ptr(int64(duration.Seconds())),
			// Note: No BoundObjectRef - Kubernetes doesn't support binding to custom resources
			// Token lifecycle is managed by expiration time instead
		},
	}

	// Request token from collector's ServiceAccount in kubecodriver-system
	result, err := tm.client.CoreV1().ServiceAccounts("kubecodriver-system").CreateToken(
		ctx,
		"kubecodriver-collector", // Collector's ServiceAccount (with kustomize prefix)
		treq,
		metav1.CreateOptions{},
	)
	if err != nil {
		return "", fmt.Errorf("failed to create token: %w", err)
	}

	return result.Status.Token, nil
}

func ptr(i int64) *int64 {
	return &i
}
