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

package sls

import (
	"reflect"

	sdk "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-alibaba/apis/sls/v1alpha1"
)

var (
	// ErrCodeMachineGroupNotExist is the error code when Logtail MachineGroup doesn't exist
	ErrCodeMachineGroupNotExist = "MachineGroupNotExist"
	// ErrCreateMachineGroup is the error when failed to create the resource
	ErrCreateMachineGroup = "failed to create a Logtail MachineGroup"
	// ErrDeleteMachineGroup is the error when failed to delete the resource
	ErrDeleteMachineGroup = "failed to delete the Logtail MachineGroup"
)

// DescribeMachineGroup describes SLS Logtail MachineGroup
func (c *LogClient) DescribeMachineGroup(project *string, name string) (*sdk.MachineGroup, error) {
	machineGroup, err := c.Client.GetMachineGroup(*project, name)
	return machineGroup, errors.Wrap(err, ErrCodeMachineGroupNotExist)
}

// CreateMachineGroup creates SLS Logtail MachineGroup
//nolint:gocyclo
func (c *LogClient) CreateMachineGroup(name string, param v1alpha1.MachineGroupParameters) error {
	machineGroup := &sdk.MachineGroup{
		Name:          name,
		MachineIDType: *param.MachineIDType,
		MachineIDList: *param.MachineIDList,
		Attribute:     *param.Attribute,
	}
	if param.Type != nil {
		machineGroup.Type = *param.Type
	}

	err := c.Client.CreateMachineGroup(*param.Project, machineGroup)
	return errors.Wrap(err, ErrCreateMachineGroup)
}

// UpdateMachineGroup updates SLS Logtail MachineGroup
func (c *LogClient) UpdateMachineGroup(project, logstore *string, machineGroup *sdk.MachineGroup) error {
	// TODO(zzxwill) Need to implement Update SLS Logtail MachineGroup
	return nil
}

// DeleteMachineGroup deletes SLS Logtail MachineGroup
func (c *LogClient) DeleteMachineGroup(project *string, machineGroup string) error {
	err := c.Client.DeleteMachineGroup(*project, machineGroup)
	return errors.Wrap(err, ErrDeleteMachineGroup)
}

// GenerateMachineGroupObservation is used to produce v1alpha1.LogstoreObservation
func GenerateMachineGroupObservation(machineGroup *sdk.MachineGroup) v1alpha1.MachineGroupObservation {
	// TODO(zzxwill) Currently nothing is needed to set for observation
	return v1alpha1.MachineGroupObservation{}
}

// IsMachineGroupUpdateToDate checks whether cr is up to date
func IsMachineGroupUpdateToDate(cr *v1alpha1.MachineGroup, machineGroup *sdk.MachineGroup) bool {
	if machineGroup == nil {
		return false
	}
	if machineGroup.MachineIDType != *cr.Spec.ForProvider.MachineIDType {
		return false
	}
	if !reflect.DeepEqual(machineGroup.MachineIDList, *cr.Spec.ForProvider.MachineIDList) {
		return false
	}
	if cr.Spec.ForProvider.Type != nil && machineGroup.Type != *cr.Spec.ForProvider.Type {
		return false
	}
	if machineGroup.Attribute != *cr.Spec.ForProvider.Attribute {
		return false
	}
	return true
}

// IsMachineGroupNotFoundError is helper function to test whether SLS Logtail MachineGroup cloud not be found
func IsMachineGroupNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := errors.Cause(err).(*sdk.Error); ok && (e.Code == ErrCodeMachineGroupNotExist) {
		return true
	}
	return false
}
