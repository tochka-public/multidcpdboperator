package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MultidcPodDisruptionBudgetSpec defines the desired state of MultidcPodDisruptionBudget
type MultidcPodDisruptionBudgetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// An eviction is allowed if at least "minAvailable" pods selected by
	// "selector" will still be available after the eviction, i.e. even in the
	// absence of the evicted pod.
	// +optional
	MinAvailable string `json:"minAvailable,omitempty"`
	// which is a description of the number of pods from that set that can be unavailable after the eviction. It can be either an absolute number or a percentage.
	MaxUnavailable string            `json:"maxUnavailable,omitempty"`
	Selector       map[string]string `json:"selector"`
}

// MultidcPodDisruptionBudgetStatus defines the observed state of MultidcPodDisruptionBudget
type MultidcPodDisruptionBudgetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector
// MultidcPodDisruptionBudget is the Schema for the multidcpoddisruptionbudgets API
type MultidcPodDisruptionBudget struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MultidcPodDisruptionBudgetSpec   `json:"spec,omitempty"`
	Status MultidcPodDisruptionBudgetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MultidcPodDisruptionBudgetList contains a list of MultidcPodDisruptionBudget
type MultidcPodDisruptionBudgetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MultidcPodDisruptionBudget `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MultidcPodDisruptionBudget{}, &MultidcPodDisruptionBudgetList{})
}
