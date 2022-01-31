/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ManagedMCGSpec defines the desired state of ManagedMCG
type ManagedMCGSpec struct{}

type ComponentState string

const (
	ComponentReady    ComponentState = "Ready"
	ComponentPending  ComponentState = "Pending"
	ComponentNotFound ComponentState = "NotFound"
	ComponentUnknown  ComponentState = "Unknown"
)

type ComponentStatus struct {
	State ComponentState `json:"state"`
}

type ComponentStatusMap struct {
	Noobaa ComponentStatus `json:"noobaa"`
	// Prometheus     ComponentStatus `json:"prometheus"`
	// Alertmanager   ComponentStatus `json:"alertmanager"`
}

// ManagedMCGStatus defines the observed state of ManagedMCG
type ManagedMCGStatus struct {
	Components ComponentStatusMap `json:"components"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=mmcg

// ManagedMCG is the Schema for the managedmcgs API
type ManagedMCG struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedMCGSpec   `json:"spec,omitempty"`
	Status ManagedMCGStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ManagedMCGList contains a list of ManagedMCG
type ManagedMCGList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedMCG `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManagedMCG{}, &ManagedMCGList{})
}
