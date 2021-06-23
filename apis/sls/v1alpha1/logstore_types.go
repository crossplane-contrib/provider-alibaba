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

// LogStoreSpec defines the desired state of SLS LogStore
type LogStoreSpec struct {
	xpv1.ResourceSpec `json:",inline"`

	// ForProvider field is where use set parameters for SLS LogStore
	ForProvider StoreParameters `json:"forProvider"`
}

// StoreObservation is the representation of the current state that is observed.
type StoreObservation struct {
	// CreateTime is the time when the store was created
	CreateTime uint32 `json:"createTime"`

	// LastModifyTime is the time when the store was last modified
	LastModifyTime uint32 `json:"lastModifyTime"`
}

// LogStoreStatus defines the observed state of SLS LogStore
type LogStoreStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          StoreObservation `json:"atProvider,omitempty"`
}

// StoreParameters define the desired state of an SLS store.
type StoreParameters struct {
	// SLS project name
	// +kubebuilder:validation:MinLength:=3
	// +kubebuilder:validation:MaxLength:=63
	ProjectName string `json:"projectName"`

	// The data retention period. Unit: days. If you set the value to 3650, the data is permanently stored
	// +kubebuilder:validation:Minimum:=1
	// +kubebuilder:validation:Maximum:=3650
	TTL int `json:"ttl"`

	// The number of shards
	// +kubebuilder:validation:Minimum:=1
	// +kubebuilder:validation:Maximum:=10
	ShardCount int `json:"shardCount"`

	// Specifies whether to enable automatic sharding. Default value: false.
	// +optional
	// +kubebuilder:default:=false
	AutoSplit *bool `json:"autoSplit,omitempty"`

	// The maximum number of shards for automatic sharding.
	// +optional
	MaxSplitShard *int `json:"maxSplitShard,omitempty"`
}

// +kubebuilder:object:root=true

// LogStore is the Schema for the SLS Stores API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,alibaba}
type LogStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              LogStoreSpec   `json:"spec"`
	Status            LogStoreStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StoreList contains a list of LogStore
type StoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LogStore `json:"items"`
}
