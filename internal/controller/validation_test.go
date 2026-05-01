package controller

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	kubecodriverv1alpha1 "github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

func TestValidateNamespaceAccess_Comprehensive(t *testing.T) {
	tests := []struct {
		name              string
		jobNamespace      string
		allowedNamespaces []string
		expectError       bool
		expectedErrorMsg  string
	}{
		{
			name:              "no restrictions - allow all",
			jobNamespace:      "any-namespace",
			allowedNamespaces: nil,
			expectError:       false,
		},
		{
			name:              "empty restrictions - allow all",
			jobNamespace:      "any-namespace",
			allowedNamespaces: []string{},
			expectError:       false,
		},
		{
			name:              "allowed namespace - exact match",
			jobNamespace:      "production",
			allowedNamespaces: []string{"production", "staging"},
			expectError:       false,
		},
		{
			name:              "disallowed namespace",
			jobNamespace:      "development",
			allowedNamespaces: []string{"production", "staging"},
			expectError:       true,
			expectedErrorMsg:  "namespace 'development' is not allowed",
		},
		{
			name:              "single allowed namespace - match",
			jobNamespace:      "production",
			allowedNamespaces: []string{"production"},
			expectError:       false,
		},
		{
			name:              "single allowed namespace - no match",
			jobNamespace:      "staging",
			allowedNamespaces: []string{"production"},
			expectError:       true,
			expectedErrorMsg:  "namespace 'staging' is not allowed",
		},
		{
			name:              "case sensitive namespace check",
			jobNamespace:      "Production",
			allowedNamespaces: []string{"production"},
			expectError:       true,
			expectedErrorMsg:  "namespace 'Production' is not allowed",
		},
	}

	reconciler := &CoDriverJobReconciler{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &kubecodriverv1alpha1.CoDriverJob{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tt.jobNamespace,
				},
			}

			toolConfig := &kubecodriverv1alpha1.CoDriverTool{
				Spec: kubecodriverv1alpha1.CoDriverToolSpec{
					AllowedNamespaces: tt.allowedNamespaces,
				},
			}

			err := reconciler.validateNamespaceAccess(job, toolConfig)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetTokenDuration_EdgeCases(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, kubecodriverv1alpha1.AddToScheme(scheme))

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	reconciler := &CoDriverJobReconciler{Client: fakeClient}

	tests := []struct {
		name               string
		collectionDuration time.Duration
		expectedMinimum    time.Duration // We'll check it's at least this much
	}{
		{
			name:               "zero duration",
			collectionDuration: 0,
			expectedMinimum:    10 * time.Minute,
		},
		{
			name:               "negative duration",
			collectionDuration: -5 * time.Minute,
			expectedMinimum:    10 * time.Minute,
		},
		{
			name:               "very small duration",
			collectionDuration: 1 * time.Second,
			expectedMinimum:    10 * time.Minute,
		},
		{
			name:               "exactly minimum duration",
			collectionDuration: 10 * time.Minute,
			expectedMinimum:    10 * time.Minute,
		},
		{
			name:               "above minimum duration",
			collectionDuration: 15 * time.Minute,
			expectedMinimum:    15 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration := reconciler.getTokenDuration(context.Background(), tt.collectionDuration)
			assert.GreaterOrEqual(t, duration, tt.expectedMinimum)
		})
	}
}

func TestGetToolConfig_SearchOrder(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, kubecodriverv1alpha1.AddToScheme(scheme))

	tests := []struct {
		name           string
		toolName       string
		configs        []kubecodriverv1alpha1.CoDriverTool
		expectedConfig *kubecodriverv1alpha1.CoDriverTool
		expectError    bool
	}{
		{
			name:     "config found in kubecodriver-system",
			toolName: "aperf",
			configs: []kubecodriverv1alpha1.CoDriverTool{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "aperf-config",
						Namespace: "kubecodriver-system",
					},
					Spec: kubecodriverv1alpha1.CoDriverToolSpec{
						Image: "kubecodriver-system/aperf:latest",
					},
				},
			},
			expectedConfig: &kubecodriverv1alpha1.CoDriverTool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aperf-config",
					Namespace: "kubecodriver-system",
				},
				Spec: kubecodriverv1alpha1.CoDriverToolSpec{
					Image: "kubecodriver-system/aperf:latest",
				},
			},
			expectError: false,
		},
		{
			name:           "config not found",
			toolName:       "nonexistent",
			configs:        []kubecodriverv1alpha1.CoDriverTool{},
			expectedConfig: nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objs := make([]client.Object, len(tt.configs))
			for i := range tt.configs {
				objs[i] = &tt.configs[i]
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(objs...).
				Build()

			reconciler := &CoDriverJobReconciler{Client: fakeClient}

			config, err := reconciler.getToolConfig(context.Background(), tt.toolName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				if assert.NotNil(t, config) {
					assert.Equal(t, tt.expectedConfig.Name, config.Name)
					assert.Equal(t, tt.expectedConfig.Namespace, config.Namespace)
					assert.Equal(t, tt.expectedConfig.Spec.Image, config.Spec.Image)
				}
			}
		})
	}
}

func TestGetToolConfig_ErrorHandling(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, kubecodriverv1alpha1.AddToScheme(scheme))

	// Test with empty tool name
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	reconciler := &CoDriverJobReconciler{Client: fakeClient}

	config, err := reconciler.getToolConfig(context.Background(), "")
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "not found")
}
