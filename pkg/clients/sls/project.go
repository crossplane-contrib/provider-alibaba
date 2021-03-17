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

	"github.com/crossplane/provider-alibaba/apis/sls/v1alpha1"
)

var (
	// ErrCodeProjectNotExist error code of ServerError when Project not found
	ErrCodeProjectNotExist = "ProjectNotExist"
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
	return c.Client.GetProject(name)
}

// Create creates SLS project
func (c *LogClient) Create(name, description string) (*sdk.LogProject, error) {
	return c.Client.CreateProject(name, description)
}

// Update updates SLS project's description
func (c *LogClient) Update(name, description string) (*sdk.LogProject, error) {
	return c.Client.UpdateProject(name, description)
}

// Delete deletes SLS project
func (c *LogClient) Delete(name string) error {
	return c.Client.DeleteProject(name)
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
	if e, ok := err.(sdk.Error); ok && (e.Code == ErrCodeProjectNotExist) {
		return true
	}
	if e, ok := err.(*sdk.Error); ok && (e.Code == ErrCodeProjectNotExist) {
		return true
	}
	return false
}
