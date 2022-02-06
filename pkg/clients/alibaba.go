/*
Copyright 2020 The Crossplane Authors.

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

package alibabacloud

import (
	"fmt"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nas "github.com/crossplane/provider-alibaba/apis/nas/v1alpha1"
	oss "github.com/crossplane/provider-alibaba/apis/oss/v1alpha1"
	slb "github.com/crossplane/provider-alibaba/apis/slb/v1alpha1"
	sls "github.com/crossplane/provider-alibaba/apis/sls/v1alpha1"
	aliv1beta1 "github.com/crossplane/provider-alibaba/apis/v1beta1"
)

// Domain is Alibaba Cloud Domain
const Domain = "aliyuncs.com"

const (
	// ErrPrepareClientEstablishmentInfo is the error of failing to prepare all the information for establishing an SDK client
	ErrPrepareClientEstablishmentInfo = "failed to prepare all the information for establishing an SDK client"
	errTrackUsage                     = "cannot track provider config usage"

	errRegionNotValid            = "region is not valid"
	errCloudResourceNotSupported = "cloud resource is not supported"

	// ErrGetProviderConfig is the error of getting provider config
	ErrGetProviderConfig = "failed to get ProviderConfig"
	// ErrGetCredentials is the error of getting credentials
	ErrGetCredentials             = "cannot get credentials"
	errFailedToExtractCredentials = "failed to extract Alibaba credentials"
	// ErrAccessKeyNotComplete is the error of not existing of AccessKeyID or AccessKeySecret
	ErrAccessKeyNotComplete = "AccessKeyID or AccessKeySecret not existed"
)

// AlibabaCredentials represents ak/sk, stsToken(maybe) information
type AlibabaCredentials struct {
	AccessKeyID     string `yaml:"accessKeyId"`
	AccessKeySecret string `yaml:"accessKeySecret"`
	SecurityToken   string `yaml:"securityToken"`
}

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

// GetProviderConfig gets ProviderConfig
func GetProviderConfig(ctx context.Context, k8sClient client.Client, providerConfigName string) (*aliv1beta1.ProviderConfig, error) {
	providerConfig := &aliv1beta1.ProviderConfig{}
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: providerConfigName}, providerConfig); err != nil {
		return nil, errors.Wrap(err, ErrGetProviderConfig)
	}
	return providerConfig, nil
}

// GetCredentials gets Alibaba credentials from ProviderConfig
func GetCredentials(ctx context.Context, client client.Client, providerConfigName string) (*AlibabaCredentials, error) {
	pc, err := GetProviderConfig(ctx, client, providerConfigName)
	if err != nil {
		return nil, err
	}

	cd := pc.Spec.Credentials
	data, err := resource.CommonCredentialExtractor(ctx, cd.Source, client, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, ErrGetCredentials)
	}

	var cred AlibabaCredentials
	if err := yaml.Unmarshal(data, &cred); err != nil {
		return nil, errors.Wrap(err, errFailedToExtractCredentials)
	}
	if cred.AccessKeyID == "" || cred.AccessKeySecret == "" {
		return nil, errors.New(ErrAccessKeyNotComplete)
	}

	return &cred, nil
}

// GetRegion gets regions from ProviderConfig
func GetRegion(ctx context.Context, client client.Client, providerConfigName string) (string, error) {
	pc, err := GetProviderConfig(ctx, client, providerConfigName)
	if err != nil {
		return "", err
	}
	return pc.Spec.Region, nil
}
