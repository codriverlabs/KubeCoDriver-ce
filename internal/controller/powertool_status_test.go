package controller

import (
	"testing"
	"time"

	kubecodriverv1alpha1 "github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

func TestSetCondition(t *testing.T) {
	reconciler := &CoDriverJobReconciler{}
	coDriverJob := &kubecodriverv1alpha1.CoDriverJob{
		Status: kubecodriverv1alpha1.CoDriverJobStatus{},
	}

	// Test adding new condition
	reconciler.setCondition(coDriverJob, kubecodriverv1alpha1.CoDriverJobConditionReady, "True", kubecodriverv1alpha1.ReasonTargetsSelected, "Test message")

	if len(coDriverJob.Status.Conditions) != 1 {
		t.Errorf("Expected 1 condition, got %d", len(coDriverJob.Status.Conditions))
	}

	condition := coDriverJob.Status.Conditions[0]
	if condition.Type != kubecodriverv1alpha1.CoDriverJobConditionReady {
		t.Errorf("Expected condition type %s, got %s", kubecodriverv1alpha1.CoDriverJobConditionReady, condition.Type)
	}
	if condition.Status != "True" {
		t.Errorf("Expected condition status True, got %s", condition.Status)
	}
	if condition.Reason != kubecodriverv1alpha1.ReasonTargetsSelected {
		t.Errorf("Expected reason %s, got %s", kubecodriverv1alpha1.ReasonTargetsSelected, condition.Reason)
	}

	// Test updating existing condition
	time.Sleep(time.Millisecond) // Ensure different timestamp
	reconciler.setCondition(coDriverJob, kubecodriverv1alpha1.CoDriverJobConditionReady, "False", kubecodriverv1alpha1.ReasonFailed, "Updated message")

	if len(coDriverJob.Status.Conditions) != 1 {
		t.Errorf("Expected 1 condition after update, got %d", len(coDriverJob.Status.Conditions))
	}

	updatedCondition := coDriverJob.Status.Conditions[0]
	if updatedCondition.Status != "False" {
		t.Errorf("Expected updated condition status False, got %s", updatedCondition.Status)
	}
	if updatedCondition.Message != "Updated message" {
		t.Errorf("Expected updated message 'Updated message', got %s", updatedCondition.Message)
	}
}

func TestGetRequeueInterval(t *testing.T) {
	reconciler := &CoDriverJobReconciler{}

	tests := []struct {
		name     string
		phase    *string
		expected time.Duration
	}{
		{
			name:     "nil phase",
			phase:    nil,
			expected: SetupTeardownInterval,
		},
		{
			name:     "running phase",
			phase:    stringPtrTest("Running"),
			expected: ActiveRunningInterval,
		},
		{
			name:     "completed phase",
			phase:    stringPtrTest("Completed"),
			expected: CompletedJobInterval,
		},
		{
			name:     "failed phase",
			phase:    stringPtrTest("Failed"),
			expected: CompletedJobInterval,
		},
		{
			name:     "unknown phase",
			phase:    stringPtrTest("Unknown"),
			expected: SetupTeardownInterval,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coDriverJob := &kubecodriverv1alpha1.CoDriverJob{
				Status: kubecodriverv1alpha1.CoDriverJobStatus{
					Phase: tt.phase,
				},
			}

			result := reconciler.getRequeueInterval(coDriverJob)
			if result != tt.expected {
				t.Errorf("Expected interval %v, got %v", tt.expected, result)
			}
		})
	}
}

func stringPtrTest(s string) *string {
	return &s
}
