/*
Copyright 2019 The Crossplane Authors.

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
	"errors"
	"fmt"

	sdkerrors "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	sdk "github.com/aliyun/aliyun-log-go-sdk"
	"k8s.io/klog/v2"

	"github.com/crossplane/provider-alibaba/apis/sls/v1alpha1"
)

var (
	// ErrSLSProjectNotFound indicates SLSProject not found
	ErrSLSProjectNotFound = errors.New("SLSProjectNotFound")
	// ErrCodeInstanceNotFound error code of ServerError when SLSProject not found
	ErrCodeInstanceNotFound = "InvalidSLSProjectId.NotFound"
)

// CreateSLSProjectRequest defines the request info to create DB Instance
type CreateSLSProjectRequest struct {
	Name        string
	Description string
}

// LogClient is the SDK client of SLS
type LogClient struct {
	Client sdk.ClientInterface
}

// NewClient creates new RDS RDSClient
func NewClient(accessKeyID, accessKeySecret, region string) LogClient {
	endpoint := fmt.Sprintf("%s.log.aliyuncs.com", region)
	logClient := sdk.CreateNormalInterface(endpoint, accessKeyID, accessKeySecret, "")
	c := LogClient{Client: logClient}
	return c
}

// Describe describes SLS project
func (c *LogClient) Describe(name string) (*sdk.LogProject, error) {
	project, err := c.Client.GetProject(name)
	if err != nil {
		klog.ErrorS(err, "create SLS project")
		return nil, err
	}
	return project, err
}

// Create creates SLS project
func (c *LogClient) Create(req CreateSLSProjectRequest) (*sdk.LogProject, error) {
	project, err := c.Client.CreateProject(req.Name, req.Description)
	if err != nil {
		klog.ErrorS(err, "create SLS project")
		return nil, err
	}
	return project, err
}

// Delete deletes SLS project
func (c *LogClient) Delete(name string) error {
	return c.Client.DeleteProject(name)
}

// GenerateObservation is used to produce v1alpha1.SLSProjectObservation from
// rds.SLSProject.
func GenerateObservation(project *sdk.LogProject) v1alpha1.SLSProjectObservation {
	return v1alpha1.SLSProjectObservation{
		Name:   project.Name,
		Status: project.Status,
	}
}

// IsErrorNotFound helper function to test for ErrCodeSLSProjectNotFoundFault error
func IsErrorNotFound(err error) bool {
	if err == nil {
		return false
	}
	// If instance already remove from console.  should ignore when delete instance
	if e, ok := err.(*sdkerrors.ServerError); ok && e.ErrorCode() == ErrCodeInstanceNotFound {
		return true
	}
	return errors.Is(err, ErrSLSProjectNotFound)
}
