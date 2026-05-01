package controller

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestExtractMatchingLabels_EdgeCases(t *testing.T) {
	r := &CoDriverJobReconciler{}

	tests := []struct {
		name      string
		selector  *metav1.LabelSelector
		podLabels map[string]string
		expected  string
	}{
		{
			name:      "nil selector",
			selector:  nil,
			podLabels: map[string]string{"app": "test"},
			expected:  "unknown",
		},
		{
			name: "nil match labels",
			selector: &metav1.LabelSelector{
				MatchLabels: nil,
			},
			podLabels: map[string]string{"app": "test"},
			expected:  "unknown",
		},
		{
			name: "empty match labels",
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{},
			},
			podLabels: map[string]string{"app": "test"},
			expected:  "unknown",
		},
		{
			name: "no matching labels",
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "nginx"},
			},
			podLabels: map[string]string{"env": "prod"},
			expected:  "unknown",
		},
		{
			name: "single match",
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "nginx"},
			},
			podLabels: map[string]string{"app": "nginx", "env": "prod"},
			expected:  "app-nginx",
		},
		{
			name: "multiple matches - returns first",
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "nginx", "env": "prod"},
			},
			podLabels: map[string]string{"app": "nginx", "env": "prod"},
			expected:  "app-nginx", // or "env-prod" depending on map iteration
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.extractMatchingLabels(tt.selector, tt.podLabels)
			// For multiple matches, just verify it's not "unknown"
			if tt.name == "multiple matches - returns first" {
				if got == "unknown" {
					t.Errorf("expected a matching label, got %s", got)
				}
			} else if got != tt.expected {
				t.Errorf("extractMatchingLabels() = %v, want %v", got, tt.expected)
			}
		})
	}
}
