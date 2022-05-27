//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	aliyun_log_go_sdk "github.com/aliyun/aliyun-log-go-sdk"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IndexKey) DeepCopyInto(out *IndexKey) {
	*out = *in
	if in.Token != nil {
		in, out := &in.Token, &out.Token
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
	if in.CaseSensitive != nil {
		in, out := &in.CaseSensitive, &out.CaseSensitive
		*out = new(bool)
		**out = **in
	}
	if in.Type != nil {
		in, out := &in.Type, &out.Type
		*out = new(string)
		**out = **in
	}
	if in.DocValue != nil {
		in, out := &in.DocValue, &out.DocValue
		*out = new(bool)
		**out = **in
	}
	if in.Alias != nil {
		in, out := &in.Alias, &out.Alias
		*out = new(string)
		**out = **in
	}
	if in.Chn != nil {
		in, out := &in.Chn, &out.Chn
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IndexKey.
func (in *IndexKey) DeepCopy() *IndexKey {
	if in == nil {
		return nil
	}
	out := new(IndexKey)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InputDetail) DeepCopyInto(out *InputDetail) {
	*out = *in
	if in.LogType != nil {
		in, out := &in.LogType, &out.LogType
		*out = new(string)
		**out = **in
	}
	if in.LogPath != nil {
		in, out := &in.LogPath, &out.LogPath
		*out = new(string)
		**out = **in
	}
	if in.FilePattern != nil {
		in, out := &in.FilePattern, &out.FilePattern
		*out = new(string)
		**out = **in
	}
	if in.TopicFormat != nil {
		in, out := &in.TopicFormat, &out.TopicFormat
		*out = new(string)
		**out = **in
	}
	if in.TimeFormat != nil {
		in, out := &in.TimeFormat, &out.TimeFormat
		*out = new(string)
		**out = **in
	}
	if in.Preserve != nil {
		in, out := &in.Preserve, &out.Preserve
		*out = new(bool)
		**out = **in
	}
	if in.PreserveDepth != nil {
		in, out := &in.PreserveDepth, &out.PreserveDepth
		*out = new(int)
		**out = **in
	}
	if in.FileEncoding != nil {
		in, out := &in.FileEncoding, &out.FileEncoding
		*out = new(string)
		**out = **in
	}
	if in.DiscardUnmatch != nil {
		in, out := &in.DiscardUnmatch, &out.DiscardUnmatch
		*out = new(bool)
		**out = **in
	}
	if in.MaxDepth != nil {
		in, out := &in.MaxDepth, &out.MaxDepth
		*out = new(int)
		**out = **in
	}
	if in.TailExisted != nil {
		in, out := &in.TailExisted, &out.TailExisted
		*out = new(bool)
		**out = **in
	}
	if in.DiscardNonUtf8 != nil {
		in, out := &in.DiscardNonUtf8, &out.DiscardNonUtf8
		*out = new(bool)
		**out = **in
	}
	if in.DelaySkipBytes != nil {
		in, out := &in.DelaySkipBytes, &out.DelaySkipBytes
		*out = new(int)
		**out = **in
	}
	if in.IsDockerFile != nil {
		in, out := &in.IsDockerFile, &out.IsDockerFile
		*out = new(bool)
		**out = **in
	}
	if in.DockerIncludeLabel != nil {
		in, out := &in.DockerIncludeLabel, &out.DockerIncludeLabel
		*out = new(map[string]string)
		if **in != nil {
			in, out := *in, *out
			*out = make(map[string]string, len(*in))
			for key, val := range *in {
				(*out)[key] = val
			}
		}
	}
	if in.DockerExcludeLabel != nil {
		in, out := &in.DockerExcludeLabel, &out.DockerExcludeLabel
		*out = new(map[string]string)
		if **in != nil {
			in, out := *in, *out
			*out = make(map[string]string, len(*in))
			for key, val := range *in {
				(*out)[key] = val
			}
		}
	}
	if in.DockerIncludeEnv != nil {
		in, out := &in.DockerIncludeEnv, &out.DockerIncludeEnv
		*out = new(map[string]string)
		if **in != nil {
			in, out := *in, *out
			*out = make(map[string]string, len(*in))
			for key, val := range *in {
				(*out)[key] = val
			}
		}
	}
	if in.DockerExcludeEnv != nil {
		in, out := &in.DockerExcludeEnv, &out.DockerExcludeEnv
		*out = new(map[string]string)
		if **in != nil {
			in, out := *in, *out
			*out = make(map[string]string, len(*in))
			for key, val := range *in {
				(*out)[key] = val
			}
		}
	}
	if in.LogBeginRegex != nil {
		in, out := &in.LogBeginRegex, &out.LogBeginRegex
		*out = new(string)
		**out = **in
	}
	if in.Regex != nil {
		in, out := &in.Regex, &out.Regex
		*out = new(string)
		**out = **in
	}
	if in.Keys != nil {
		in, out := &in.Keys, &out.Keys
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InputDetail.
func (in *InputDetail) DeepCopy() *InputDetail {
	if in == nil {
		return nil
	}
	out := new(InputDetail)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogStore) DeepCopyInto(out *LogStore) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogStore.
func (in *LogStore) DeepCopy() *LogStore {
	if in == nil {
		return nil
	}
	out := new(LogStore)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LogStore) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogStoreList) DeepCopyInto(out *LogStoreList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]LogStore, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogStoreList.
func (in *LogStoreList) DeepCopy() *LogStoreList {
	if in == nil {
		return nil
	}
	out := new(LogStoreList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LogStoreList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogStoreSpec) DeepCopyInto(out *LogStoreSpec) {
	*out = *in
	in.ResourceSpec.DeepCopyInto(&out.ResourceSpec)
	in.ForProvider.DeepCopyInto(&out.ForProvider)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogStoreSpec.
func (in *LogStoreSpec) DeepCopy() *LogStoreSpec {
	if in == nil {
		return nil
	}
	out := new(LogStoreSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogStoreStatus) DeepCopyInto(out *LogStoreStatus) {
	*out = *in
	in.ResourceStatus.DeepCopyInto(&out.ResourceStatus)
	out.AtProvider = in.AtProvider
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogStoreStatus.
func (in *LogStoreStatus) DeepCopy() *LogStoreStatus {
	if in == nil {
		return nil
	}
	out := new(LogStoreStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogstoreIndex) DeepCopyInto(out *LogstoreIndex) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogstoreIndex.
func (in *LogstoreIndex) DeepCopy() *LogstoreIndex {
	if in == nil {
		return nil
	}
	out := new(LogstoreIndex)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LogstoreIndex) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogstoreIndexList) DeepCopyInto(out *LogstoreIndexList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]LogstoreIndex, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogstoreIndexList.
func (in *LogstoreIndexList) DeepCopy() *LogstoreIndexList {
	if in == nil {
		return nil
	}
	out := new(LogstoreIndexList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LogstoreIndexList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogstoreIndexObservation) DeepCopyInto(out *LogstoreIndexObservation) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogstoreIndexObservation.
func (in *LogstoreIndexObservation) DeepCopy() *LogstoreIndexObservation {
	if in == nil {
		return nil
	}
	out := new(LogstoreIndexObservation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogstoreIndexParameters) DeepCopyInto(out *LogstoreIndexParameters) {
	*out = *in
	if in.ProjectName != nil {
		in, out := &in.ProjectName, &out.ProjectName
		*out = new(string)
		**out = **in
	}
	if in.LogstoreName != nil {
		in, out := &in.LogstoreName, &out.LogstoreName
		*out = new(string)
		**out = **in
	}
	if in.Keys != nil {
		in, out := &in.Keys, &out.Keys
		*out = make(map[string]IndexKey, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogstoreIndexParameters.
func (in *LogstoreIndexParameters) DeepCopy() *LogstoreIndexParameters {
	if in == nil {
		return nil
	}
	out := new(LogstoreIndexParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogstoreIndexSpec) DeepCopyInto(out *LogstoreIndexSpec) {
	*out = *in
	in.ResourceSpec.DeepCopyInto(&out.ResourceSpec)
	in.ForProvider.DeepCopyInto(&out.ForProvider)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogstoreIndexSpec.
func (in *LogstoreIndexSpec) DeepCopy() *LogstoreIndexSpec {
	if in == nil {
		return nil
	}
	out := new(LogstoreIndexSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogstoreIndexStatus) DeepCopyInto(out *LogstoreIndexStatus) {
	*out = *in
	in.ResourceStatus.DeepCopyInto(&out.ResourceStatus)
	out.AtProvider = in.AtProvider
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogstoreIndexStatus.
func (in *LogstoreIndexStatus) DeepCopy() *LogstoreIndexStatus {
	if in == nil {
		return nil
	}
	out := new(LogstoreIndexStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Logtail) DeepCopyInto(out *Logtail) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Logtail.
func (in *Logtail) DeepCopy() *Logtail {
	if in == nil {
		return nil
	}
	out := new(Logtail)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Logtail) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogtailList) DeepCopyInto(out *LogtailList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Logtail, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogtailList.
func (in *LogtailList) DeepCopy() *LogtailList {
	if in == nil {
		return nil
	}
	out := new(LogtailList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LogtailList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogtailObservation) DeepCopyInto(out *LogtailObservation) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogtailObservation.
func (in *LogtailObservation) DeepCopy() *LogtailObservation {
	if in == nil {
		return nil
	}
	out := new(LogtailObservation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogtailParameters) DeepCopyInto(out *LogtailParameters) {
	*out = *in
	if in.InputType != nil {
		in, out := &in.InputType, &out.InputType
		*out = new(string)
		**out = **in
	}
	in.InputDetail.DeepCopyInto(&out.InputDetail)
	if in.OutputType != nil {
		in, out := &in.OutputType, &out.OutputType
		*out = new(string)
		**out = **in
	}
	out.OutputDetail = in.OutputDetail
	if in.LogSample != nil {
		in, out := &in.LogSample, &out.LogSample
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogtailParameters.
func (in *LogtailParameters) DeepCopy() *LogtailParameters {
	if in == nil {
		return nil
	}
	out := new(LogtailParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogtailSpec) DeepCopyInto(out *LogtailSpec) {
	*out = *in
	in.ResourceSpec.DeepCopyInto(&out.ResourceSpec)
	in.ForProvider.DeepCopyInto(&out.ForProvider)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogtailSpec.
func (in *LogtailSpec) DeepCopy() *LogtailSpec {
	if in == nil {
		return nil
	}
	out := new(LogtailSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LogtailStatus) DeepCopyInto(out *LogtailStatus) {
	*out = *in
	in.ResourceStatus.DeepCopyInto(&out.ResourceStatus)
	out.AtProvider = in.AtProvider
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LogtailStatus.
func (in *LogtailStatus) DeepCopy() *LogtailStatus {
	if in == nil {
		return nil
	}
	out := new(LogtailStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineGroup) DeepCopyInto(out *MachineGroup) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineGroup.
func (in *MachineGroup) DeepCopy() *MachineGroup {
	if in == nil {
		return nil
	}
	out := new(MachineGroup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MachineGroup) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineGroupBinding) DeepCopyInto(out *MachineGroupBinding) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineGroupBinding.
func (in *MachineGroupBinding) DeepCopy() *MachineGroupBinding {
	if in == nil {
		return nil
	}
	out := new(MachineGroupBinding)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MachineGroupBinding) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineGroupBindingList) DeepCopyInto(out *MachineGroupBindingList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MachineGroupBinding, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineGroupBindingList.
func (in *MachineGroupBindingList) DeepCopy() *MachineGroupBindingList {
	if in == nil {
		return nil
	}
	out := new(MachineGroupBindingList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MachineGroupBindingList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineGroupBindingObservation) DeepCopyInto(out *MachineGroupBindingObservation) {
	*out = *in
	if in.Configs != nil {
		in, out := &in.Configs, &out.Configs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineGroupBindingObservation.
func (in *MachineGroupBindingObservation) DeepCopy() *MachineGroupBindingObservation {
	if in == nil {
		return nil
	}
	out := new(MachineGroupBindingObservation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineGroupBindingParameters) DeepCopyInto(out *MachineGroupBindingParameters) {
	*out = *in
	if in.ProjectName != nil {
		in, out := &in.ProjectName, &out.ProjectName
		*out = new(string)
		**out = **in
	}
	if in.GroupName != nil {
		in, out := &in.GroupName, &out.GroupName
		*out = new(string)
		**out = **in
	}
	if in.ConfigName != nil {
		in, out := &in.ConfigName, &out.ConfigName
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineGroupBindingParameters.
func (in *MachineGroupBindingParameters) DeepCopy() *MachineGroupBindingParameters {
	if in == nil {
		return nil
	}
	out := new(MachineGroupBindingParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineGroupBindingSpec) DeepCopyInto(out *MachineGroupBindingSpec) {
	*out = *in
	in.ResourceSpec.DeepCopyInto(&out.ResourceSpec)
	in.ForProvider.DeepCopyInto(&out.ForProvider)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineGroupBindingSpec.
func (in *MachineGroupBindingSpec) DeepCopy() *MachineGroupBindingSpec {
	if in == nil {
		return nil
	}
	out := new(MachineGroupBindingSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineGroupBindingStatus) DeepCopyInto(out *MachineGroupBindingStatus) {
	*out = *in
	in.ResourceStatus.DeepCopyInto(&out.ResourceStatus)
	in.AtProvider.DeepCopyInto(&out.AtProvider)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineGroupBindingStatus.
func (in *MachineGroupBindingStatus) DeepCopy() *MachineGroupBindingStatus {
	if in == nil {
		return nil
	}
	out := new(MachineGroupBindingStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineGroupList) DeepCopyInto(out *MachineGroupList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MachineGroup, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineGroupList.
func (in *MachineGroupList) DeepCopy() *MachineGroupList {
	if in == nil {
		return nil
	}
	out := new(MachineGroupList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MachineGroupList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineGroupObservation) DeepCopyInto(out *MachineGroupObservation) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineGroupObservation.
func (in *MachineGroupObservation) DeepCopy() *MachineGroupObservation {
	if in == nil {
		return nil
	}
	out := new(MachineGroupObservation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineGroupParameters) DeepCopyInto(out *MachineGroupParameters) {
	*out = *in
	if in.Project != nil {
		in, out := &in.Project, &out.Project
		*out = new(string)
		**out = **in
	}
	if in.Logstore != nil {
		in, out := &in.Logstore, &out.Logstore
		*out = new(string)
		**out = **in
	}
	if in.Type != nil {
		in, out := &in.Type, &out.Type
		*out = new(string)
		**out = **in
	}
	if in.MachineIDType != nil {
		in, out := &in.MachineIDType, &out.MachineIDType
		*out = new(string)
		**out = **in
	}
	if in.MachineIDList != nil {
		in, out := &in.MachineIDList, &out.MachineIDList
		*out = new([]string)
		if **in != nil {
			in, out := *in, *out
			*out = make([]string, len(*in))
			copy(*out, *in)
		}
	}
	if in.Attribute != nil {
		in, out := &in.Attribute, &out.Attribute
		*out = new(aliyun_log_go_sdk.MachinGroupAttribute)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineGroupParameters.
func (in *MachineGroupParameters) DeepCopy() *MachineGroupParameters {
	if in == nil {
		return nil
	}
	out := new(MachineGroupParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineGroupSpec) DeepCopyInto(out *MachineGroupSpec) {
	*out = *in
	in.ResourceSpec.DeepCopyInto(&out.ResourceSpec)
	in.ForProvider.DeepCopyInto(&out.ForProvider)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineGroupSpec.
func (in *MachineGroupSpec) DeepCopy() *MachineGroupSpec {
	if in == nil {
		return nil
	}
	out := new(MachineGroupSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineGroupStatus) DeepCopyInto(out *MachineGroupStatus) {
	*out = *in
	in.ResourceStatus.DeepCopyInto(&out.ResourceStatus)
	out.AtProvider = in.AtProvider
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineGroupStatus.
func (in *MachineGroupStatus) DeepCopy() *MachineGroupStatus {
	if in == nil {
		return nil
	}
	out := new(MachineGroupStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OutputDetail) DeepCopyInto(out *OutputDetail) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OutputDetail.
func (in *OutputDetail) DeepCopy() *OutputDetail {
	if in == nil {
		return nil
	}
	out := new(OutputDetail)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Project) DeepCopyInto(out *Project) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Project.
func (in *Project) DeepCopy() *Project {
	if in == nil {
		return nil
	}
	out := new(Project)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Project) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProjectList) DeepCopyInto(out *ProjectList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Project, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProjectList.
func (in *ProjectList) DeepCopy() *ProjectList {
	if in == nil {
		return nil
	}
	out := new(ProjectList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ProjectList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProjectObservation) DeepCopyInto(out *ProjectObservation) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProjectObservation.
func (in *ProjectObservation) DeepCopy() *ProjectObservation {
	if in == nil {
		return nil
	}
	out := new(ProjectObservation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProjectParameters) DeepCopyInto(out *ProjectParameters) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProjectParameters.
func (in *ProjectParameters) DeepCopy() *ProjectParameters {
	if in == nil {
		return nil
	}
	out := new(ProjectParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProjectSpec) DeepCopyInto(out *ProjectSpec) {
	*out = *in
	in.ResourceSpec.DeepCopyInto(&out.ResourceSpec)
	out.ForProvider = in.ForProvider
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProjectSpec.
func (in *ProjectSpec) DeepCopy() *ProjectSpec {
	if in == nil {
		return nil
	}
	out := new(ProjectSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProjectStatus) DeepCopyInto(out *ProjectStatus) {
	*out = *in
	in.ResourceStatus.DeepCopyInto(&out.ResourceStatus)
	out.AtProvider = in.AtProvider
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProjectStatus.
func (in *ProjectStatus) DeepCopy() *ProjectStatus {
	if in == nil {
		return nil
	}
	out := new(ProjectStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StoreObservation) DeepCopyInto(out *StoreObservation) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StoreObservation.
func (in *StoreObservation) DeepCopy() *StoreObservation {
	if in == nil {
		return nil
	}
	out := new(StoreObservation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StoreParameters) DeepCopyInto(out *StoreParameters) {
	*out = *in
	if in.AutoSplit != nil {
		in, out := &in.AutoSplit, &out.AutoSplit
		*out = new(bool)
		**out = **in
	}
	if in.MaxSplitShard != nil {
		in, out := &in.MaxSplitShard, &out.MaxSplitShard
		*out = new(int)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StoreParameters.
func (in *StoreParameters) DeepCopy() *StoreParameters {
	if in == nil {
		return nil
	}
	out := new(StoreParameters)
	in.DeepCopyInto(out)
	return out
}
