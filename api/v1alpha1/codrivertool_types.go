/*
Copyright 2025.

*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CoDriverToolSpec defines the desired state of CoDriverTool
type CoDriverToolSpec struct {
	// Name is the unique identifier for this power tool
	// +required
	Name string `json:"name"`

	// Image is the container image for this power tool
	// +required
	Image string `json:"image"`

	// SecurityContext defines the security context for this power tool
	// +required
	SecurityContext SecuritySpec `json:"securityContext"`

	// AllowedNamespaces restricts which namespaces can use this tool
	// If empty, tool can be used in any namespace
	// +optional
	AllowedNamespaces []string `json:"allowedNamespaces,omitempty"`

	// Description provides information about what this tool does
	// +optional
	Description *string `json:"description,omitempty"`

	// Version specifies the tool version
	// +optional
	Version *string `json:"version,omitempty"`

	// DefaultArgs provides default arguments for the tool
	// +optional
	DefaultArgs []string `json:"defaultArgs,omitempty"`

	// Resources defines the resource requirements for the ephemeral container
	// +optional
	Resources *ResourceSpec `json:"resources,omitempty"`
}

// CoDriverToolStatus defines the observed state of CoDriverTool
type CoDriverToolStatus struct {
	// Phase represents the current phase of the CoDriverTool
	// +optional
	Phase *string `json:"phase,omitempty"`

	// LastValidated indicates when this configuration was last validated
	// +optional
	LastValidated *metav1.Time `json:"lastValidated,omitempty"`

	// Conditions represent the latest available observations of the CoDriverTool's state
	// +optional
	Conditions []CoDriverToolCondition `json:"conditions,omitempty"`
}

// CoDriverToolCondition represents a condition of a CoDriverTool
type CoDriverToolCondition struct {
	Type               string      `json:"type"`
	Status             string      `json:"status"`
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`
	Reason             string      `json:"reason,omitempty"`
	Message            string      `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CoDriverTool is the Schema for the codrivertools API
type CoDriverTool struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of CoDriverTool
	// +required
	Spec CoDriverToolSpec `json:"spec"`

	// status defines the observed state of CoDriverTool
	// +optional
	Status CoDriverToolStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// CoDriverToolList contains a list of CoDriverTool
type CoDriverToolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CoDriverTool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CoDriverTool{}, &CoDriverToolList{})
}
