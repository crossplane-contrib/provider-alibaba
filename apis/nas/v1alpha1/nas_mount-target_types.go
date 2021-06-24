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
	runtimev1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// NASMountTargetList contains a list of NASMountTarget
type NASMountTargetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NASMountTarget `json:"items"`
}

// +kubebuilder:object:root=true

// NASMountTarget is a managed resource that represents an NASMountTarget instance
// +kubebuilder:printcolumn:name="FILE-SYSTEM-ID",type="string",JSONPath=".spec.atProvider.fileSystemID"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,alibaba},shortName=nasmt
type NASMountTarget struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NASMountTargetSpec   `json:"spec,omitempty"`
	Status NASMountTargetStatus `json:"status,omitempty"`
}

// NASMountTargetSpec defines the desired state of NASMountTarget
type NASMountTargetSpec struct {
	runtimev1.ResourceSpec `json:",inline"`
	ForProvider            NASMountTargetParameter `json:"forProvider"`
}

// NASMountTargetStatus defines the observed state of NASMountTarget
type NASMountTargetStatus struct {
	runtimev1.ResourceStatus `json:",inline"`
	AtProvider               NASMountTargetObservation `json:"atProvider,omitempty"`
}

// NASMountTargetParameter is the isolated place to store files
type NASMountTargetParameter struct {
	FileSystemID    *string `json:"fileSystemID"`
	AccessGroupName *string `json:"accessGroupName,omitempty"`
	NetworkType     *string `json:"networkType"`
	VpcID           *string `json:"vpcId,omitempty"`
	VSwitchID       *string `json:"vSwitchId,omitempty"`
	SecurityGroupID *string `json:"securityGroupId,omitempty"`
}

// NASMountTargetObservation is the representation of the current state that is observed.
type NASMountTargetObservation struct {
	MountTargetDomain *string `json:"mountTargetDomain,omitempty"`
}
