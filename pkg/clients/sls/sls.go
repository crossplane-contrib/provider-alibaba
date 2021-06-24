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
}

// LogClient is the SDK client of SLS
type LogClient struct {
	Client sdk.ClientInterface
}

// NewClient creates new SLS client
func NewClient(accessKeyID, accessKeySecret, securityToken, region string) *LogClient {
	endpoint := fmt.Sprintf("%s.log.aliyuncs.com", region)
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
