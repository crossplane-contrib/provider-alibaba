/*

 Copyright 2021 The Crossplane Authors.

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
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MachineGroupBindingSpec defines the desired state of SLS MachineGroupBinding
type MachineGroupBindingSpec struct {
	xpv1.ResourceSpec `json:",inline"`

	// ForProvider field is where use set parameters for SLS MachineGroupBinding
	ForProvider MachineGroupBindingParameters `json:"forProvider"`
}

// MachineGroupBindingObservation is the representation of the current state that is observed.
type MachineGroupBindingObservation struct {
	Configs []string `json:"configs"`
}

// MachineGroupBindingStatus defines the observed state of SLS MachineGroupBinding
type MachineGroupBindingStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          MachineGroupBindingObservation `json:"atProvider,omitempty"`
}

// MachineGroupBindingParameters define the desired state of an SLS store.
type MachineGroupBindingParameters struct {
	// SLS project name
	// +kubebuilder:validation:MinLength:=3
	// +kubebuilder:validation:MaxLength:=63
	ProjectName *string `json:"projectName"`

	GroupName *string `json:"groupName"`

	ConfigName *string `json:"configName"`
}

// +kubebuilder:object:root=true

// MachineGroupBinding is the Schema for the SLS MachineGroupBindings API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,alibaba}
type MachineGroupBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MachineGroupBindingSpec   `json:"spec"`
	Status            MachineGroupBindingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MachineGroupBindingList contains a list of MachineGroupBinding
type MachineGroupBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachineGroupBinding `json:"items"`
}
