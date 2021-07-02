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
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"

	nas "github.com/crossplane/provider-alibaba/apis/nas/v1alpha1"
	oss "github.com/crossplane/provider-alibaba/apis/oss/v1alpha1"
	slb "github.com/crossplane/provider-alibaba/apis/slb/v1alpha1"
	sls "github.com/crossplane/provider-alibaba/apis/sls/v1alpha1"
)

// Domain is Alibaba Cloud Domain
var Domain = "aliyuncs.com"

var (
	errRegionNotValid            = "region is not valid"
	errCloudResourceNotSupported = "cloud resource is not supported"
)

// GetEndpoint gets endpoints for all cloud resources
func GetEndpoint(res runtime.Object, region string) (string, error) {
	if res == nil || res.GetObjectKind() == nil {
		return "", errors.New(errCloudResourceNotSupported)
	}

	if region == "" && res.GetObjectKind().GroupVersionKind().Kind != slb.CLBKind {
		return "", errors.New(errRegionNotValid)
	}

	var endpoint string
	switch res.GetObjectKind().GroupVersionKind().Kind {
	case oss.BucketKind:
		endpoint = fmt.Sprintf("http://oss-%s.%s", region, Domain)
	case nas.NASFileSystemKind, nas.NASMountTargetKind:
		endpoint = fmt.Sprintf("nas.%s.%s", region, Domain)
	case slb.CLBKind:
		endpoint = fmt.Sprintf("slb.%s", Domain)
	case sls.ProjectGroupKind:
		endpoint = fmt.Sprintf("%s.log.%s", region, Domain)
	default:
		return "", errors.New(errCloudResourceNotSupported)
	}
	return endpoint, nil
}
