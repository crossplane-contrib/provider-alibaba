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
	sdk "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/pkg/errors"

	"github.com/crossplane-contrib/provider-alibaba/apis/sls/v1alpha1"
)

var (
	// ErrCodeLogstoreIndexNotExist is the error code when Logstore index doesn't exist
	ErrCodeLogstoreIndexNotExist = "IndexConfigNotExist"
	// ErrCreateIndex is the error when failed to create the resource
	ErrCreateIndex = "failed to create a Logstore index"
	// ErrDeleteIndex is the error when failed to delete the resource
	ErrDeleteIndex = "failed to delete the Logstore index"
)

// DescribeIndex describes SLS Logstore index
func (c *LogClient) DescribeIndex(project, logstore *string) (*sdk.Index, error) {
	index, err := c.Client.GetIndex(*project, *logstore)
	return index, errors.Wrap(err, ErrCodeLogstoreIndexNotExist)
}

// CreateIndex creates SLS Logstore index
//nolint:gocyclo
func (c *LogClient) CreateIndex(param v1alpha1.LogstoreIndexParameters) error {
	keys := map[string]sdk.IndexKey{}
	for name, v := range param.Keys {
		key := sdk.IndexKey{
			Token:         *v.Token,
			CaseSensitive: *v.CaseSensitive,
			Type:          *v.Type,
		}
		if v.DocValue != nil {
			key.DocValue = *v.DocValue
		}
		if v.Alias != nil {
			key.Alias = *v.Alias
		}
		if v.Chn != nil {
			key.Chn = *v.Chn
		}
		keys[name] = key
	}
	index := sdk.Index{
		Keys: keys,
	}
	err := c.Client.CreateIndex(*param.ProjectName, *param.LogstoreName, index)
	return errors.Wrap(err, ErrCreateIndex)
}

// UpdateIndex updates SLS Logstore index
func (c *LogClient) UpdateIndex(project, logstore *string, index *sdk.Index) error {
	// TODO(zzxwill) Need to implement Update SLS Logstore index
	return nil
}

// DeleteIndex deletes SLS Logstore index
func (c *LogClient) DeleteIndex(project, logstore *string) error {
	err := c.Client.DeleteIndex(*project, *logstore)
	return errors.Wrap(err, ErrDeleteIndex)
}

// GenerateIndexObservation is used to produce v1alpha1.LogstoreObservation
func GenerateIndexObservation(index *sdk.Index) v1alpha1.LogstoreIndexObservation {
	// TODO(zzxwill) Currently nothing is needed to set for observation
	return v1alpha1.LogstoreIndexObservation{}
}

// IsIndexUpdateToDate checks whether cr is up to date
func IsIndexUpdateToDate(cr *v1alpha1.LogstoreIndex, index *sdk.Index) bool {
	if index == nil {
		return false
	}

	// TODO(zzxwill) More strict comparison should be made between two keys
	if len(cr.Spec.ForProvider.Keys) != len(index.Keys) {
		return true
	}
	return true
}

// IsIndexNotFoundError is helper function to test whether SLS Logstore index cloud not be found
func IsIndexNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := errors.Cause(err).(*sdk.Error); ok && (e.Code == ErrCodeLogstoreIndexNotExist) {
		return true
	}
	return false
}
