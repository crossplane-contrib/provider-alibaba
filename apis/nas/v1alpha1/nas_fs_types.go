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

// NASFileSystemList contains a list of NASFileSystem
type NASFileSystemList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NASFileSystem `json:"items"`
}

// +kubebuilder:object:root=true

// NASFileSystem is a managed resource that represents an NASFileSystem instance
// +kubebuilder:printcolumn:name="FILE-SYSTEM-ID",type="string",JSONPath=".status.atProvider.fileSystemID"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,alibaba},shortName=nasfs
type NASFileSystem struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NASFileSystemSpec   `json:"spec,omitempty"`
	Status NASFileSystemStatus `json:"status,omitempty"`
}

// NASFileSystemSpec defines the desired state of NASFileSystem
type NASFileSystemSpec struct {
	runtimev1.ResourceSpec `json:",inline"`
	NASFileSystemParameter `json:",inline"`
}

// NASFileSystemStatus defines the observed state of NASFileSystem
type NASFileSystemStatus struct {
	runtimev1.ResourceStatus `json:",inline"`
	AtProvider               NASFileSystemObservation `json:"atProvider,omitempty"`
}

// NASFileSystemParameter is the isolated place to store files
type NASFileSystemParameter struct {
	FileSystemType *string `json:"fileSystemType,omitempty"`
	ChargeType     *string `json:"chargeType,omitempty"`
	StorageType    *string `json:"storageType"`
	ProtocolType   *string `json:"protocolType"`
	VpcID          *string `json:"vpcId,omitempty"`
	VSwitchID      *string `json:"vSwitchId,omitempty"`
}

// NASFileSystemObservation is the representation of the current state that is observed.
type NASFileSystemObservation struct {
	FileSystemID      string `json:"fileSystemID,omitempty"`
	MountTargetDomain string `json:"mountTargetDomain,omitempty"`
}
