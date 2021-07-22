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
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	aliv1beta1 "github.com/crossplane/provider-alibaba/apis/v1beta1"
)

const (
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
