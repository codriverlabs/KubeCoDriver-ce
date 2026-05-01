package controller

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubecodriverv1alpha1 "github.com/codriverlabs/KubeCoDriver/api/v1alpha1"
)

func TestSetCondition_Comprehensive(t *testing.T) {
	tests := []struct {
		name            string
		initialStatus   kubecodriverv1alpha1.CoDriverJobStatus
		conditionType   string
		status          string
		reason          string
		message         string
		expectedCount   int
		expectedType    string
		expectedStatus  string
		expectedReason  string
		expectedMessage string
	}{
		{
			name:            "add new condition to empty status",
			initialStatus:   kubecodriverv1alpha1.CoDriverJobStatus{},
			conditionType:   "Ready",
			status:          "True",
			reason:          "PodFound",
			message:         "Target pod found and ready",
			expectedCount:   1,
			expectedType:    "Ready",
			expectedStatus:  "True",
			expectedReason:  "PodFound",
			expectedMessage: "Target pod found and ready",
		},
		{
			name: "update existing condition",
			initialStatus: kubecodriverv1alpha1.CoDriverJobStatus{
				Conditions: []kubecodriverv1alpha1.CoDriverJobCondition{
					{
						Type:    "Ready",
						Status:  "False",
						Reason:  "PodNotFound",
						Message: "Target pod not found",
					},
				},
			},
			conditionType:   "Ready",
			status:          "True",
			reason:          "PodFound",
			message:         "Target pod found and ready",
			expectedCount:   1,
			expectedType:    "Ready",
			expectedStatus:  "True",
			expectedReason:  "PodFound",
			expectedMessage: "Target pod found and ready",
		},
		{
			name: "add second condition",
			initialStatus: kubecodriverv1alpha1.CoDriverJobStatus{
				Conditions: []kubecodriverv1alpha1.CoDriverJobCondition{
					{
						Type:   "Ready",
						Status: "True",
						Reason: "PodFound",
					},
				},
			},
			conditionType:   "Progressing",
			status:          "True",
			reason:          "ContainerStarted",
			message:         "Ephemeral container started",
			expectedCount:   2,
			expectedType:    "Progressing",
			expectedStatus:  "True",
			expectedReason:  "ContainerStarted",
			expectedMessage: "Ephemeral container started",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coDriverJob := &kubecodriverv1alpha1.CoDriverJob{
				Status: tt.initialStatus,
			}

			reconciler := &CoDriverJobReconciler{}
			reconciler.setCondition(coDriverJob, tt.conditionType, tt.status, tt.reason, tt.message)

			assert.Len(t, coDriverJob.Status.Conditions, tt.expectedCount)

			// Find the condition we just set/updated
			var foundCondition *kubecodriverv1alpha1.CoDriverJobCondition
			for i := range coDriverJob.Status.Conditions {
				if coDriverJob.Status.Conditions[i].Type == tt.expectedType {
					foundCondition = &coDriverJob.Status.Conditions[i]
					break
				}
			}

			assert.NotNil(t, foundCondition, "Expected condition not found")
			assert.Equal(t, tt.expectedStatus, foundCondition.Status)
			assert.Equal(t, tt.expectedReason, foundCondition.Reason)
			assert.Equal(t, tt.expectedMessage, foundCondition.Message)
			assert.NotNil(t, foundCondition.LastTransitionTime)
		})
	}
}

func TestSetCondition_TimestampUpdate(t *testing.T) {
	coDriverJob := &kubecodriverv1alpha1.CoDriverJob{
		Status: kubecodriverv1alpha1.CoDriverJobStatus{
			Conditions: []kubecodriverv1alpha1.CoDriverJobCondition{
				{
					Type:               "Ready",
					Status:             "False",
					Reason:             "PodNotFound",
					Message:            "Target pod not found",
					LastTransitionTime: metav1.Time{Time: time.Now().Add(-1 * time.Hour)},
				},
			},
		},
	}

	oldTime := coDriverJob.Status.Conditions[0].LastTransitionTime

	reconciler := &CoDriverJobReconciler{}
	reconciler.setCondition(coDriverJob, "Ready", "True", "PodFound", "Target pod found")

	newTime := coDriverJob.Status.Conditions[0].LastTransitionTime
	assert.True(t, newTime.After(oldTime.Time), "LastTransitionTime should be updated")
}

func TestSetCondition_NoTimestampUpdateForSameStatus(t *testing.T) {
	originalTime := metav1.Time{Time: time.Now().Add(-1 * time.Hour)}
	coDriverJob := &kubecodriverv1alpha1.CoDriverJob{
		Status: kubecodriverv1alpha1.CoDriverJobStatus{
			Conditions: []kubecodriverv1alpha1.CoDriverJobCondition{
				{
					Type:               "Ready",
					Status:             "True",
					Reason:             "PodFound",
					Message:            "Target pod found",
					LastTransitionTime: originalTime,
				},
			},
		},
	}

	reconciler := &CoDriverJobReconciler{}
	reconciler.setCondition(coDriverJob, "Ready", "True", "PodFound", "Target pod found and ready")

	// Status didn't change, so timestamp should remain the same
	assert.Equal(t, originalTime, coDriverJob.Status.Conditions[0].LastTransitionTime)
	// But message should be updated
	assert.Equal(t, "Target pod found and ready", coDriverJob.Status.Conditions[0].Message)
}

func TestGetRequeueInterval_AllPhases(t *testing.T) {
	tests := []struct {
		name             string
		phase            *string
		expectedInterval time.Duration
	}{
		{
			name:             "nil phase",
			phase:            nil,
			expectedInterval: SetupTeardownInterval,
		},
		{
			name:             "running phase",
			phase:            &[]string{"Running"}[0],
			expectedInterval: ActiveRunningInterval,
		},
		{
			name:             "completed phase",
			phase:            &[]string{"Completed"}[0],
			expectedInterval: CompletedJobInterval,
		},
		{
			name:             "failed phase",
			phase:            &[]string{"Failed"}[0],
			expectedInterval: CompletedJobInterval, // Failed jobs use completed interval
		},
		{
			name:             "pending phase",
			phase:            &[]string{"Pending"}[0],
			expectedInterval: SetupTeardownInterval,
		},
		{
			name:             "unknown phase",
			phase:            &[]string{"Unknown"}[0],
			expectedInterval: SetupTeardownInterval,
		},
	}

	reconciler := &CoDriverJobReconciler{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coDriverJob := &kubecodriverv1alpha1.CoDriverJob{
				Status: kubecodriverv1alpha1.CoDriverJobStatus{
					Phase: tt.phase,
				},
			}

			interval := reconciler.getRequeueInterval(coDriverJob)
			assert.Equal(t, tt.expectedInterval, interval)
		})
	}
}

func TestGetRequeueInterval_EdgeCases(t *testing.T) {
	reconciler := &CoDriverJobReconciler{}

	// Test with empty CoDriverJob
	emptyTool := &kubecodriverv1alpha1.CoDriverJob{}
	interval := reconciler.getRequeueInterval(emptyTool)
	assert.Equal(t, SetupTeardownInterval, interval)

	// Test with empty string phase
	emptyPhase := ""
	toolWithEmptyPhase := &kubecodriverv1alpha1.CoDriverJob{
		Status: kubecodriverv1alpha1.CoDriverJobStatus{
			Phase: &emptyPhase,
		},
	}
	interval = reconciler.getRequeueInterval(toolWithEmptyPhase)
	assert.Equal(t, SetupTeardownInterval, interval)
}
