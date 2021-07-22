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
	"fmt"

	sdk "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-alibaba/apis/sls/v1alpha1"
)

var (
	// ErrCodeProjectNotExist error code of ServerError when Project not found
	ErrCodeProjectNotExist = "ProjectNotExist"
	// ErrFailedToGetSLSProject is the error of failing to get an SLS project
	ErrFailedToGetSLSProject = "FailedToGetSLSProject"
	// ErrFailedToCreateSLSProject is the error of failing to create an SLS project
	ErrFailedToCreateSLSProject = "FailedToCreateSLSProject"
	// ErrFailedToUpdateSLSProject is the error of failing to update an SLS project
	ErrFailedToUpdateSLSProject = "FailedToUpdateSLSProject"
	// ErrFailedToDeleteSLSProject is the error of failing to delete an SLS project
	ErrFailedToDeleteSLSProject = "FailedToDeleteSLSProject"

	// ErrCodeStoreNotExist error code of ServerError when LogStore not found
	ErrCodeStoreNotExist = "LogStoreNotExist"
	// ErrFailedToGetSLSStore is the error of failing to get an SLS store
	ErrFailedToGetSLSStore = "FailedToGetSLSStore"
	// ErrFailedToCreateSLSStore is the error of failing to create an SLS store
	ErrFailedToCreateSLSStore = "FailedToCreateSLSStore"
	// ErrFailedToUpdateSLSStore is the error of failing to update an SLS store
	ErrFailedToUpdateSLSStore = "FailedToUpdateSLSStore"
	// ErrFailedToDeleteSLSStore is the error of failing to delete an SLS store
	ErrFailedToDeleteSLSStore = "FailedToDeleteSLSStore"

	// ErrCodeLogtailNotExist is the error code when Logtail doesn't exist
	ErrCodeLogtailNotExist = "ConfigNotExist"
)

// LogClientInterface is the Log client interface
type LogClientInterface interface {
	Describe(name string) (*sdk.LogProject, error)
	Create(name, description string) (*sdk.LogProject, error)
	Update(name, description string) (*sdk.LogProject, error)
	Delete(name string) error

	DescribeStore(project string, logstore string) (*sdk.LogStore, error)
	CreateStore(project string, store *sdk.LogStore) error
	UpdateStore(project string, logstore string, ttl int) error
	DeleteStore(project string, logstore string) error

	DescribeConfig(project string, config string) (*sdk.LogConfig, error)
	CreateConfig(name string, config v1alpha1.LogtailParameters) error
	UpdateConfig(project string, config *sdk.LogConfig) error
	DeleteConfig(project string, config string) error

	DescribeIndex(project, logstore *string) (*sdk.Index, error)
	CreateIndex(param v1alpha1.LogstoreIndexParameters) error
	UpdateIndex(project, logstore *string, index *sdk.Index) error
	DeleteIndex(project, logstore *string) error

	DescribeMachineGroup(project *string, name string) (*sdk.MachineGroup, error)
	CreateMachineGroup(name string, param v1alpha1.MachineGroupParameters) error
	UpdateMachineGroup(project, logstore *string, machineGroup *sdk.MachineGroup) error
	DeleteMachineGroup(project *string, logstore string) error

	GetAppliedConfigs(projectName *string, groupName *string) ([]string, error)
	ApplyConfigToMachineGroup(projectName, groupName, confName *string) error
	RemoveConfigFromMachineGroup(projectName, groupName, confName *string) error
}

// LogClient is the SDK client of SLS
type LogClient struct {
	Client sdk.ClientInterface
}

// NewClient creates new SLS client
func NewClient(accessKeyID, accessKeySecret, securityToken, endpoint string) *LogClient {
	logClient := sdk.CreateNormalInterface(endpoint, accessKeyID, accessKeySecret, securityToken)
	return &LogClient{Client: logClient}
}

// ----------------------SLS Project------------------------------ //

// Describe describes SLS project
func (c *LogClient) Describe(name string) (*sdk.LogProject, error) {
	logProject, err := c.Client.GetProject(name)
	return logProject, errors.Wrap(err, ErrFailedToGetSLSProject)
}

// Create creates SLS project
func (c *LogClient) Create(name, description string) (*sdk.LogProject, error) {
	logProject, err := c.Client.CreateProject(name, description)
	return logProject, errors.Wrap(err, ErrFailedToCreateSLSProject)
}

// Update updates SLS project's description
func (c *LogClient) Update(name, description string) (*sdk.LogProject, error) {
	logProject, err := c.Client.UpdateProject(name, description)
	return logProject, errors.Wrap(err, ErrFailedToUpdateSLSProject)

}

// Delete deletes SLS project
func (c *LogClient) Delete(name string) error {
	err := c.Client.DeleteProject(name)
	return errors.Wrap(err, ErrFailedToDeleteSLSProject)
}

// GenerateObservation is used to produce v1alpha1.ProjectObservation
func GenerateObservation(project *sdk.LogProject) v1alpha1.ProjectObservation {
	return v1alpha1.ProjectObservation{
		CreateTime:     project.CreateTime,
		LastModifyTime: project.LastModifyTime,
		Owner:          project.Owner,
		Status:         project.Status,
		Region:         project.Status,
	}
}

// IsNotFoundError helper function to test for SLS project not found error
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := errors.Cause(err).(sdk.Error); ok && (e.Code == ErrCodeProjectNotExist) {
		return true
	}
	if e, ok := errors.Cause(err).(*sdk.Error); ok && (e.Code == ErrCodeProjectNotExist) {
		return true
	}
	return false
}

// ----------------------SLS LogStore------------------------------ //

// DescribeStore describes SLS store
func (c *LogClient) DescribeStore(project string, logstore string) (*sdk.LogStore, error) {
	logStore, err := c.Client.GetLogStore(project, logstore)
	return logStore, errors.Wrap(err, ErrFailedToGetSLSStore)
}

// CreateStore creates SLS store
func (c *LogClient) CreateStore(project string, logstore *sdk.LogStore) error {
	err := c.Client.CreateLogStoreV2(project, logstore)
	return errors.Wrap(err, ErrFailedToCreateSLSStore)
}

// UpdateStore updates SLS store's description
func (c *LogClient) UpdateStore(project string, logstore string, ttl int) error {
	err := c.Client.UpdateLogStore(project, logstore, ttl, 2)
	return errors.Wrap(err, ErrFailedToUpdateSLSStore)

}

// DeleteStore deletes SLS store
func (c *LogClient) DeleteStore(project string, logstore string) error {
	err := c.Client.DeleteLogStore(project, logstore)
	return errors.Wrap(err, ErrFailedToDeleteSLSStore)
}

// GenerateStoreObservation is used to produce v1alpha1.StoreObservation
func GenerateStoreObservation(store *sdk.LogStore) v1alpha1.StoreObservation {
	return v1alpha1.StoreObservation{
		CreateTime:     store.CreateTime,
		LastModifyTime: store.LastModifyTime,
	}
}

// IsStoreUpdateToDate checks whether cr is up to date
func IsStoreUpdateToDate(cr *v1alpha1.LogStore, store *sdk.LogStore) bool {
	if (cr.Name == store.Name) && (cr.Spec.ForProvider.TTL == store.TTL) {
		return true
	}
	return false
}

// IsStoreNotFoundError helper function to test for SLS project not found error
func IsStoreNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := errors.Cause(err).(*sdk.Error); ok && (e.Code == ErrCodeStoreNotExist) {
		return true
	}
	return false
}

// ----------------------SLS Logtail------------------------------ //

// DescribeConfig describes SLS Logtail config
func (c *LogClient) DescribeConfig(project string, config string) (*sdk.LogConfig, error) {
	logStore, err := c.Client.GetConfig(project, config)
	return logStore, errors.Wrap(err, ErrFailedToGetSLSStore)
}

// CreateConfig creates SLS Logtail config
//nolint:gocyclo
func (c *LogClient) CreateConfig(name string, t v1alpha1.LogtailParameters) error {
	in := t.InputDetail
	inputDetail := sdk.RegexConfigInputDetail{}
	switch {
	case *t.InputType == "file" && *in.LogType == "common_reg_log":
		inputDetail.LogPath = *in.LogPath
		inputDetail.FilePattern = *in.FilePattern
		inputDetail.LogType = *in.LogType
		inputDetail.TopicFormat = *in.TopicFormat

		if in.Preserve != nil {
			inputDetail.Preserve = *in.Preserve
		}
		if in.PreserveDepth != nil {
			inputDetail.PreserveDepth = *in.PreserveDepth
		}
		if in.FileEncoding != nil {
			inputDetail.FileEncoding = *in.FileEncoding
		}
		if in.DiscardUnmatch != nil {
			inputDetail.DiscardNonUtf8 = *in.DiscardUnmatch
		}
		if in.MaxDepth != nil {
			inputDetail.MaxDepth = *in.MaxDepth
		}
		if in.TailExisted != nil {
			inputDetail.TailExisted = *in.TailExisted
		}
		if in.DiscardNonUtf8 != nil {
			inputDetail.DiscardNonUtf8 = *in.DiscardNonUtf8
		}
		if in.DelaySkipBytes != nil {
			inputDetail.DelaySkipBytes = *in.DelaySkipBytes
		}
		if in.IsDockerFile != nil {
			inputDetail.IsDockerFile = *in.IsDockerFile
		}
		if in.DockerIncludeEnv != nil {
			inputDetail.DockerIncludeEnv = *in.DockerIncludeEnv
		}
		if in.DockerIncludeLabel != nil {
			inputDetail.DockerIncludeLabel = *in.DockerIncludeLabel
		}
		if in.DockerExcludeLabel != nil {
			inputDetail.DockerExcludeLabel = *in.DockerExcludeLabel
		}
		if in.DockerExcludeEnv != nil {
			inputDetail.DockerExcludeEnv = *in.DockerExcludeEnv
		}

		inputDetail.Key = in.Keys
		if in.LogBeginRegex != nil {
			inputDetail.LogBeginRegex = *in.LogBeginRegex
		}
		if in.Regex != nil {
			inputDetail.Regex = *in.Regex
		}
	case *t.InputType != "file":
		return fmt.Errorf("InputType %s is not supported", *t.InputType)
	case *in.LogType == "common_reg_log":
		return fmt.Errorf("LogType %s is not supported", *in.LogType)
	}

	outputDetail := sdk.OutputDetail{
		ProjectName:  t.OutputDetail.ProjectName,
		LogStoreName: t.OutputDetail.LogStoreName,
	}
	config := &sdk.LogConfig{
		Name:         name,
		InputType:    *t.InputType,
		InputDetail:  inputDetail,
		OutputType:   *t.OutputType,
		OutputDetail: outputDetail,
	}
	if t.LogSample != nil {
		config.LogSample = *t.LogSample
	}
	err := c.Client.CreateConfig(t.OutputDetail.ProjectName, config)
	return errors.Wrap(err, ErrFailedToCreateSLSStore)
}

// UpdateConfig updates SLS Logtail config's description
func (c *LogClient) UpdateConfig(project string, config *sdk.LogConfig) error {
	err := c.Client.UpdateConfig(project, config)
	return errors.Wrap(err, ErrFailedToUpdateSLSStore)

}

// DeleteConfig deletes SLS Logtail config
func (c *LogClient) DeleteConfig(project string, config string) error {
	err := c.Client.DeleteConfig(project, config)
	return errors.Wrap(err, ErrFailedToDeleteSLSStore)
}

// GenerateLogtailObservation is used to produce v1alpha1.LogtailObservation
func GenerateLogtailObservation(config *sdk.LogConfig) v1alpha1.LogtailObservation {
	// TODO(zzxwill) Currently nothing is needed to set for observation
	return v1alpha1.LogtailObservation{}
}

// IsLogtailUpdateToDate checks whether cr is up to date
func IsLogtailUpdateToDate(cr *v1alpha1.Logtail, config *sdk.LogConfig) bool {
	if config == nil {
		return false
	}
	if cr.Name != config.Name {
		return false
	}
	if *cr.Spec.ForProvider.InputType != config.InputType {
		return false
	}

	if *cr.Spec.ForProvider.OutputType != config.OutputType {
		return false
	}
	if cr.Spec.ForProvider.OutputDetail.ProjectName != config.OutputDetail.ProjectName ||
		cr.Spec.ForProvider.OutputDetail.LogStoreName != config.OutputDetail.LogStoreName {
		return false
	}
	return true
}

// IsLogtailNotFoundError helper function to test for SLS project not found error
func IsLogtailNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := errors.Cause(err).(*sdk.Error); ok && (e.Code == ErrCodeLogtailNotExist) {
		return true
	}
	return false
}
