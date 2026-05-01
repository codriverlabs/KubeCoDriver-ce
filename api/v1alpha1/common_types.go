/*
Copyright 2025.

*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Common types shared across CoDriverJob, CoDriverTool, and related resources

// TargetSpec defines the target for tool execution
type TargetSpec struct {
	NamespaceSelector *NamespaceSelector    `json:"namespaceSelector,omitempty"`
	LabelSelector     *metav1.LabelSelector `json:"labelSelector"`
	Container         *string               `json:"container,omitempty"`
}

// NamespaceSelector defines the namespace selection criteria
type NamespaceSelector struct {
	MatchNames []string `json:"matchNames,omitempty"`
	MatchRegex *string  `json:"matchRegex,omitempty"`
}

// OutputSpec defines the output configuration
type OutputSpec struct {
	Mode            string         `json:"mode"`
	PVC             *PVCSpec       `json:"pvc,omitempty"`
	Collector       *CollectorSpec `json:"collector,omitempty"`
	RollingInterval *string        `json:"rollingInterval,omitempty"`
	Compress        *string        `json:"compress,omitempty"`
	RetentionDays   *int32         `json:"retentionDays,omitempty"`
}

// PVCSpec defines the PVC output configuration
type PVCSpec struct {
	ClaimName string  `json:"claimName"`
	Path      *string `json:"path,omitempty"`
}

// CollectorSpec defines the collector output configuration
type CollectorSpec struct {
	Endpoint string `json:"endpoint"`
}

// BudgetSpec defines the budget configuration
type BudgetSpec struct {
	MaxConcurrentPods    *int32 `json:"maxConcurrentPods,omitempty"`
	PerPodOverheadTarget *int32 `json:"perPodOverheadTarget,omitempty"`
}

// SecuritySpec defines the security configuration
type SecuritySpec struct {
	AllowPrivileged *bool         `json:"allowPrivileged,omitempty"`
	AllowHostPID    *bool         `json:"allowHostPID,omitempty"`
	RunAsRoot       *bool         `json:"runAsRoot,omitempty"`
	Capabilities    *Capabilities `json:"capabilities,omitempty"`
}

// Capabilities defines the container capabilities
type Capabilities struct {
	Add  []string `json:"add,omitempty"`
	Drop []string `json:"drop,omitempty"`
}

// FailurePolicySpec defines the failure policy
type FailurePolicySpec struct {
	OnError *string      `json:"onError,omitempty"`
	Backoff *BackoffSpec `json:"backoff,omitempty"`
}

// BackoffSpec defines the backoff configuration
type BackoffSpec struct {
	Initial    *string `json:"initial,omitempty"`
	Max        *string `json:"max,omitempty"`
	Multiplier *string `json:"multiplier,omitempty"`
}

// ResourceSpec defines resource requirements
type ResourceSpec struct {
	Requests *ResourceList `json:"requests,omitempty"`
	Limits   *ResourceList `json:"limits,omitempty"`
}

// ResourceList defines CPU and memory resources
type ResourceList struct {
	CPU    *string `json:"cpu,omitempty"`
	Memory *string `json:"memory,omitempty"`
}
