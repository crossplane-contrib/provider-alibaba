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

// ProjectSpec defines the desired state of SLS Project
type ProjectSpec struct {
	xpv1.ResourceSpec `json:",inline"`

	// ForProvider field is where use set parameters for SLS project
	ForProvider ProjectParameters `json:"forProvider"`
}

// ProjectObservation is the representation of the current state that is observed.
type ProjectObservation struct {
	// CreateTime is the time when the project was created
	CreateTime string `json:"createTime"`

	// LastModifyTime is the time when the project was last modified
	LastModifyTime string `json:"lastModifyTime"`

	// Owner is the ID of the Alibaba Cloud account that was used to create the project
	Owner string `json:"owner"`

	// Status is the the status of the project
	Status string `json:"status"`

	// Region is the region to which the project belongs
	Region string `json:"region"`
}

// ProjectStatus defines the observed state of SLS Project
type ProjectStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ProjectObservation `json:"atProvider,omitempty"`
}

// ProjectParameters define the desired state of an SLS project.
type ProjectParameters struct {
	ProjectName string `json:"name"`
	Description string `json:"description"`
}

// +kubebuilder:object:root=true

// Project is the Schema for the SLS Projects API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,alibaba}
type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ProjectSpec   `json:"spec"`
	Status            ProjectStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProjectList contains a list of Project
type ProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Project `json:"items"`
}
