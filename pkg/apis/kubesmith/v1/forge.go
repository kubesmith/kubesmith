package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ForgeSpec defines the specification for a Kubesmith Forge.
type ForgeSpec struct {
	//
}

// ForgeStatus captures the current status of a Kubesmith Forge.
type ForgeStatus struct {
	//
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Forge struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   ForgeSpec   `json:"spec"`
	Status ForgeStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ForgeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Forge `json:"items"`
}
