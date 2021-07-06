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
	sdk "github.com/aliyun/aliyun-log-go-sdk"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MachineGroupSpec defines the desired state of SLS MachineGroup
type MachineGroupSpec struct {
	xpv1.ResourceSpec `json:",inline"`

	// ForProvider field is SLS MachineGroup parameters
	ForProvider MachineGroupParameters `json:"forProvider"`
}

// MachineGroupObservation is the representation of the current state that is observed.
type MachineGroupObservation struct {
	// CreateTime is the time the resource was created
	CreateTime uint32 `json:"createTime"`

	// LastModifyTime is the time when the resource was last modified
	LastModifyTime uint32 `json:"lastModifyTime"`
}

// MachineGroupStatus defines the observed state of SLS MachineGroup
type MachineGroupStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          MachineGroupObservation `json:"atProvider,omitempty"`
}

// MachineGroupParameters define the desired state of an SLS MachineGroup.
type MachineGroupParameters struct {
	Project       *string                   `json:"project"`
	Logstore      *string                   `json:"logstore"`
	Type          *string                   `json:"type,omitempty"`
	MachineIDType *string                   `json:"machineIDType"`
	MachineIDList *[]string                 `json:"machineIDList"`
	Attribute     *sdk.MachinGroupAttribute `json:"attribute"`
}

// +kubebuilder:object:root=true

// MachineGroup is the Schema for the SLS MachineGroup API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,alibaba},shortName=machinegroup
type MachineGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MachineGroupSpec   `json:"spec"`
	Status            MachineGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MachineGroupList contains a list of MachineGroup
type MachineGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachineGroup `json:"items"`
}
