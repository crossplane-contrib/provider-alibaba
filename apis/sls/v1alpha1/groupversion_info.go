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

// Package v1alpha1 contains API Schema definitions for the sls v1alpha1 API group
// +kubebuilder:object:generate=true
// +groupName=sls.alibaba.crossplane.io
package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "sls.alibaba.crossplane.io", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

var (
	// ProjectKind is the kind of Project
	ProjectKind = reflect.TypeOf(Project{}).Name()

	// ProjectGroupKind is the group and kind of Project
	ProjectGroupKind = schema.GroupKind{Group: GroupVersion.Group, Kind: ProjectKind}.String()

	// ProjectGroupVersionKind is the group, version and kind of Project
	ProjectGroupVersionKind = GroupVersion.WithKind(ProjectKind)
)

var (
	// StoreKind is the kind of Log LogStore
	StoreKind = reflect.TypeOf(LogStore{}).Name()

	// StoreGroupKind is the group and kind of LogStore
	StoreGroupKind = schema.GroupKind{Group: GroupVersion.Group, Kind: StoreKind}.String()

	// StoreGroupVersionKind is the group, version and kind of LogStore
	StoreGroupVersionKind = GroupVersion.WithKind(StoreKind)
)

func init() {
	SchemeBuilder.Register(&Project{}, &ProjectList{})
	SchemeBuilder.Register(&LogStore{}, &LogStoreList{})
}
