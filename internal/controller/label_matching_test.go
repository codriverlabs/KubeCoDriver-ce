package controller

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestExtractMatchingLabels(t *testing.T) {
	r := &CoDriverJobReconciler{}

	tests := []struct {
		name      string
		selector  *metav1.LabelSelector
		podLabels map[string]string
		want      string
	}{
		{
			name: "single matching label",
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			podLabels: map[string]string{
				"app": "nginx",
			},
			want: "app-nginx",
		},
		{
			name: "multiple matching labels - returns first",
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
					"env": "prod",
				},
			},
			podLabels: map[string]string{
				"app": "nginx",
				"env": "prod",
			},
			want: "app-nginx", // or "env-prod" - depends on map iteration
		},
		{
			name: "no matching labels",
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			podLabels: map[string]string{
				"app": "apache",
			},
			want: "unknown",
		},
		{
			name:     "nil selector",
			selector: nil,
			podLabels: map[string]string{
				"app": "nginx",
			},
			want: "unknown",
		},
		{
			name: "nil matchLabels",
			selector: &metav1.LabelSelector{
				MatchLabels: nil,
			},
			podLabels: map[string]string{
				"app": "nginx",
			},
			want: "unknown",
		},
		{
			name: "empty matchLabels",
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{},
			},
			podLabels: map[string]string{
				"app": "nginx",
			},
			want: "unknown",
		},
		{
			name: "empty pod labels",
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			podLabels: map[string]string{},
			want:      "unknown",
		},
		{
			name: "environment label",
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"env": "production",
				},
			},
			podLabels: map[string]string{
				"env": "production",
			},
			want: "env-production",
		},
		{
			name: "tier label",
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"tier": "backend",
				},
			},
			podLabels: map[string]string{
				"tier": "backend",
			},
			want: "tier-backend",
		},
		{
			name: "custom label with special characters",
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"component": "api-v2",
				},
			},
			podLabels: map[string]string{
				"component": "api-v2",
			},
			want: "component-api-v2",
		},
		{
			name: "pod has extra labels",
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			podLabels: map[string]string{
				"app":     "nginx",
				"version": "1.0",
				"env":     "prod",
			},
			want: "app-nginx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.extractMatchingLabels(tt.selector, tt.podLabels)

			// For cases with multiple matching labels, we accept any of them
			if tt.name == "multiple matching labels - returns first" {
				if got != "app-nginx" && got != "env-prod" {
					t.Errorf("extractMatchingLabels() = %v, want app-nginx or env-prod", got)
				}
				return
			}

			if got != tt.want {
				t.Errorf("extractMatchingLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractMatchingLabels_POSIXCompliance(t *testing.T) {
	r := &CoDriverJobReconciler{}

	tests := []struct {
		name      string
		selector  *metav1.LabelSelector
		podLabels map[string]string
		wantRegex string
	}{
		{
			name: "uses hyphen not equals",
			selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			podLabels: map[string]string{
				"app": "nginx",
			},
			wantRegex: "^[a-zA-Z0-9-]+$", // Only alphanumeric and hyphen
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.extractMatchingLabels(tt.selector, tt.podLabels)

			// Verify no equals sign
			for _, char := range got {
				if char == '=' {
					t.Errorf("extractMatchingLabels() contains '=' which is not POSIX-compliant: %v", got)
				}
			}

			// Verify contains hyphen as separator
			if got != "unknown" && !contains(got, '-') {
				t.Errorf("extractMatchingLabels() should use hyphen as separator: %v", got)
			}
		})
	}
}

func contains(s string, char rune) bool {
	for _, c := range s {
		if c == char {
			return true
		}
	}
	return false
}
