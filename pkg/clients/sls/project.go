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
)

// LogClientInterface will help fakeOSSClient in unit tests
type LogClientInterface interface {
	Describe(name string) (*sdk.LogProject, error)
	Create(name, description string) (*sdk.LogProject, error)
	Update(name, description string) (*sdk.LogProject, error)
	Delete(name string) error
}

// LogClient is the SDK client of SLS
type LogClient struct {
	Client sdk.ClientInterface
}

// NewClient creates new SLS client
func NewClient(accessKeyID, accessKeySecret, region string) *LogClient {
	endpoint := fmt.Sprintf("%s.log.aliyuncs.com", region)
	logClient := sdk.CreateNormalInterface(endpoint, accessKeyID, accessKeySecret, "")
	return &LogClient{Client: logClient}
}

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
