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

// LogstoreIndexSpec defines the desired state of SLS LogstoreIndex
type LogstoreIndexSpec struct {
	xpv1.ResourceSpec `json:",inline"`

	// ForProvider field is SLS LogstoreIndex parameters
	ForProvider LogstoreIndexParameters `json:"forProvider"`
}

// LogstoreIndexObservation is the representation of the current state that is observed.
type LogstoreIndexObservation struct {
}

// LogstoreIndexStatus defines the observed state of SLS LogstoreIndex
type LogstoreIndexStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          LogstoreIndexObservation `json:"atProvider,omitempty"`
}

// LogstoreIndexParameters define the desired state of an SLS LogstoreIndex.
type LogstoreIndexParameters struct {
	ProjectName  *string             `json:"projectName"`
	LogstoreName *string             `json:"logstoreName"`
	Keys         map[string]IndexKey `json:"keys"`
	// Confirmed with Alibaba Cloud SLS developer, using `line` index is not encouraged. So we don't support it.
}

// IndexKey is the index by key.
// Copied most of these fields from sdk.IndexKey and leave out the field `JsonKeys` which is not supported per SLS developer
type IndexKey struct {
	Token         *[]string `json:"token"` // tokens that split the log line.
	CaseSensitive *bool     `json:"caseSensitive"`
	Type          *string   `json:"type"` // text, long, double
	DocValue      *bool     `json:"docValue,omitempty"`
	Alias         *string   `json:"alias,omitempty"`
	Chn           *bool     `json:"chn,omitempty"` // parse chinese or not
}

// +kubebuilder:object:root=true

// LogstoreIndex is the Schema for the SLS LogstoreIndex API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,alibaba},shortName=index
type LogstoreIndex struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              LogstoreIndexSpec   `json:"spec"`
	Status            LogstoreIndexStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LogstoreIndexList contains a list of LogstoreIndex
type LogstoreIndexList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LogstoreIndex `json:"items"`
}
