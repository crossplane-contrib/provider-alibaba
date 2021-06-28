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

package util

import (
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// ErrPrepareClientEstablishmentInfo is the error of failing to prepare all the information for establishing an SDK client
	ErrPrepareClientEstablishmentInfo string = "failed to prepare all the information for establishing an SDK client"
	errTrackUsage                     string = "cannot track provider config usage"
)

// ClientEstablishmentInfo represents all the information for establishing an SDK client
type ClientEstablishmentInfo struct {
	AlibabaCredentials `json:",inline"`
	Region             string `json:"region"`
	Endpoint           string `json:"endpoint"`
}

// PrepareClient will prepare all information to establish an Alibaba Cloud resource SDK client
func PrepareClient(ctx context.Context, mg resource.Managed, res runtime.Object, c client.Client, usage resource.Tracker, providerConfigName string) (*ClientEstablishmentInfo, error) {
	info := &ClientEstablishmentInfo{}

	if err := usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackUsage)
	}

	cred, err := GetCredentials(ctx, c, providerConfigName)
	if err != nil {
		return nil, errors.Wrap(err, ErrPrepareClientEstablishmentInfo)
	}
	info.AccessKeyID = cred.AccessKeyID
	info.AccessKeySecret = cred.AccessKeySecret
	info.SecurityToken = cred.SecurityToken

	region, err := GetRegion(ctx, c, providerConfigName)
	if err != nil {
		return nil, errors.Wrap(err, ErrPrepareClientEstablishmentInfo)
	}
	info.Region = region

	endpoint, err := GetEndpoint(res, region)
	if err != nil {
		return nil, err
	}
	info.Endpoint = endpoint

	return info, nil
}
