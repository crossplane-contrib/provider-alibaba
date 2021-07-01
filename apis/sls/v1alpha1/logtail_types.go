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

// LogtailSpec defines the desired state of SLS Logtail
type LogtailSpec struct {
	xpv1.ResourceSpec `json:",inline"`

	// ForProvider field is SLS Logtail parameters
	ForProvider LogtailParameters `json:"forProvider"`
}

// LogtailObservation is the representation of the current state that is observed.
type LogtailObservation struct {
	// CreateTime is the time the resource was created
	CreateTime uint32 `json:"createTime"`

	// LastModifyTime is the time when the resource was last modified
	LastModifyTime uint32 `json:"lastModifyTime"`
}

// LogtailStatus defines the observed state of SLS Logtail
type LogtailStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          LogtailObservation `json:"atProvider,omitempty"`
}

// LogtailParameters define the desired state of an SLS Logtail.
type LogtailParameters struct {
	// +kubebuilder:validation:Enum:=plugin;file
	InputType   *string     `json:"inputType"`
	InputDetail InputDetail `json:"inputDetail"`
	// +kubebuilder:validation:Enum:=LogService
	OutputType   *string      `json:"outputType"`
	OutputDetail OutputDetail `json:"outputDetail"`
	LogSample    *string      `json:"logSample,omitempty"`
}

// InputDetail defines all file input detail's basic config
type InputDetail struct {
	LogType     *string `json:"logType"`
	LogPath     *string `json:"logPath"`
	FilePattern *string `json:"filePattern"`
	TopicFormat *string `json:"topicFormat"`
	TimeFormat  *string `json:"timeFormat,omitempty"`
	// +kubebuilder:default:=false
	Preserve           *bool              `json:"preserve,omitempty"`
	PreserveDepth      *int               `json:"preserveDepth,omitempty"`
	FileEncoding       *string            `json:"fileEncoding,omitempty"`
	DiscardUnmatch     *bool              `json:"discardUnmatch,omitempty"`
	MaxDepth           *int               `json:"maxDepth,omitempty"`
	TailExisted        *bool              `json:"tailExisted,omitempty"`
	DiscardNonUtf8     *bool              `json:"discardNonUtf8,omitempty"`
	DelaySkipBytes     *int               `json:"delaySkipBytes,omitempty"`
	IsDockerFile       *bool              `json:"dockerFile,omitempty"`
	DockerIncludeLabel *map[string]string `json:"dockerIncludeLabel,omitempty"`
	DockerExcludeLabel *map[string]string `json:"dockerExcludeLabel,omitempty"`
	DockerIncludeEnv   *map[string]string `json:"dockerIncludeEnv,omitempty"`
	DockerExcludeEnv   *map[string]string `json:"dockerExcludeEnv,omitempty"`

	LogBeginRegex *string  `json:"logBeginRegex,omitempty"`
	Regex         *string  `json:"regex,omitempty"`
	Keys          []string `json:"keys"`
}

// OutputDetail defines output
type OutputDetail struct {
	ProjectName  string `json:"projectName"`
	LogStoreName string `json:"logstoreName"`
}

// +kubebuilder:object:root=true

// Logtail is the Schema for the SLS Logtail API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,alibaba},shortName=config
type Logtail struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              LogtailSpec   `json:"spec"`
	Status            LogtailStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LogtailList contains a list of Logtail
type LogtailList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Logtail `json:"items"`
}
