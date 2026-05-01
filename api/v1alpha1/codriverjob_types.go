/*
Copyright 2025.

*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CoDriverJob condition types
const (
	CoDriverJobConditionReady      = "Ready"
	CoDriverJobConditionRunning    = "Running"
	CoDriverJobConditionCompleted  = "Completed"
	CoDriverJobConditionFailed     = "Failed"
	CoDriverJobConditionConflicted = "Conflicted"
)

// CoDriverJob condition reasons
const (
	ReasonConflictDetected = "ConflictDetected"
	ReasonRunning          = "Running"
	ReasonCompleted        = "Completed"
	ReasonFailed           = "Failed"
	ReasonTargetsSelected  = "TargetsSelected"
)

// CoDriverJobSpec defines the desired state of CoDriverJob
type CoDriverJobSpec struct {
	Targets                 TargetSpec         `json:"targets"`
	Tool                    ToolSpec           `json:"tool"`
	Output                  OutputSpec         `json:"output"`
	Budgets                 *BudgetSpec        `json:"budgets,omitempty"`
	FailurePolicy           *FailurePolicySpec `json:"failurePolicy,omitempty"`
	Schedule                *string            `json:"schedule,omitempty"`
	TTLSecondsAfterFinished *int32             `json:"ttlSecondsAfterFinished,omitempty"`
}

// ToolSpec defines the tool configuration (renamed from ProfilerSpec)
type ToolSpec struct {
	Name string `json:"name"`
	// Args provides additional arguments that will be appended to defaultArgs from CoDriverTool
	// Users cannot override administrator-defined defaultArgs for security
	Args             []string `json:"args,omitempty"`
	Duration         string   `json:"duration"`
	Warmup           *string  `json:"warmup,omitempty"`
	ResolutionPreset *string  `json:"resolutionPreset,omitempty"`
	MaxCPUPercent    *int32   `json:"maxCPUPercent,omitempty"`
}

// CoDriverJobStatus defines the observed state of CoDriverJob
type CoDriverJobStatus struct {
	Phase         *string                `json:"phase,omitempty"`
	SelectedPods  *int32                 `json:"selectedPods,omitempty"`
	CompletedPods *int32                 `json:"completedPods,omitempty"`
	BytesWritten  *string                `json:"bytesWritten,omitempty"`
	Artifacts     []string               `json:"artifacts,omitempty"`
	LastError     *string                `json:"lastError,omitempty"`
	StartedAt     *metav1.Time           `json:"startedAt,omitempty"`
	FinishedAt    *metav1.Time           `json:"finishedAt,omitempty"`
	Conditions    []CoDriverJobCondition `json:"conditions,omitempty"`
	ActivePods    map[string]string      `json:"activePods,omitempty"` // podName -> containerName
}

// CoDriverJobCondition represents a condition of a CoDriverJob
type CoDriverJobCondition struct {
	Type               string      `json:"type"`
	Status             string      `json:"status"`
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	Reason             string      `json:"reason,omitempty"`
	Message            string      `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CoDriverJob is the Schema for the codriverjobs API
type CoDriverJob struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of CoDriverJob
	// +required
	Spec CoDriverJobSpec `json:"spec"`

	// status defines the observed state of CoDriverJob
	// +optional
	Status CoDriverJobStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// CoDriverJobList contains a list of CoDriverJob
type CoDriverJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CoDriverJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CoDriverJob{}, &CoDriverJobList{})
}
