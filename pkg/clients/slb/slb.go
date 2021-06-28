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

package nas

import (
	"context"

	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	sdk "github.com/alibabacloud-go/slb-20140515/v2/client"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-alibaba/apis/slb/v1alpha1"
)

const (
	errFailedToCreateSLBClient = "failed to crate SLB client"
)

// ClientInterface create a client inferface
type ClientInterface interface {
	DescribeLoadBalancers(region, loadBalancerID, vpcID, vSwitchID *string) (*sdk.DescribeLoadBalancersResponse, error)
	CreateLoadBalancer(clb v1alpha1.CLBParameter) (*sdk.CreateLoadBalancerResponse, error)
	DeleteLoadBalancer(region, loadBalancerID *string) error
}

// SDKClient is the SDK client for SLBLoadBalancer
type SDKClient struct {
	Client *sdk.Client
}

// NewClient will create SLB client
func NewClient(ctx context.Context, endpoint string, accessKeyID string, accessKeySecret string, securityToken string) (*SDKClient, error) {
	config := &openapi.Config{
		AccessKeyId:     &accessKeyID,
		AccessKeySecret: &accessKeySecret,
		SecurityToken:   &securityToken,
		Endpoint:        &endpoint,
	}
	client, err := sdk.NewClient(config)
	if err != nil {
		return nil, errors.Wrap(err, errFailedToCreateSLBClient)
	}
	return &SDKClient{Client: client}, nil
}

// DescribeLoadBalancers describes SLB LoadBalancer
func (c *SDKClient) DescribeLoadBalancers(region, loadBalancerID, vpcID, vSwitchID *string) (*sdk.DescribeLoadBalancersResponse, error) {
	describeLoadBalancersRequest := &sdk.DescribeLoadBalancersRequest{
		RegionId: region,
	}
	if loadBalancerID != nil {
		describeLoadBalancersRequest.LoadBalancerId = loadBalancerID
	}
	if vpcID != nil {
		describeLoadBalancersRequest.VpcId = vpcID
	}
	if vSwitchID != nil {
		describeLoadBalancersRequest.VSwitchId = vSwitchID
	}
	fs, err := c.Client.DescribeLoadBalancers(describeLoadBalancersRequest)
	if err != nil {
		return nil, err
	}
	return fs, nil
}

// CreateLoadBalancer creates SLBLoadBalancer
func (c *SDKClient) CreateLoadBalancer(clb v1alpha1.CLBParameter) (*sdk.CreateLoadBalancerResponse, error) {
	createLoadBalancerRequest := &sdk.CreateLoadBalancerRequest{
		RegionId:           clb.Region,
		AddressType:        clb.AddressType,
		InternetChargeType: clb.InternetChargeType,
		Bandwidth:          clb.Bandwidth,
		LoadBalancerName:   clb.LoadBalancerName,
		VpcId:              clb.VpcID,
		VSwitchId:          clb.VSwitchID,
		LoadBalancerSpec:   clb.LoadBalancerSpec,
	}
	res, err := c.Client.CreateLoadBalancer(createLoadBalancerRequest)
	return res, err
}

// DeleteLoadBalancer deletes SLBLoadBalancer
func (c *SDKClient) DeleteLoadBalancer(region, loadBalancerID *string) error {
	deleteLoadBalancerRequest := &sdk.DeleteLoadBalancerRequest{
		RegionId:       region,
		LoadBalancerId: loadBalancerID,
	}
	_, err := c.Client.DeleteLoadBalancer(deleteLoadBalancerRequest)
	return err
}

// GenerateObservation generates SLBLoadBalancerObservation from LoadBalancer information
func GenerateObservation(res *sdk.DescribeLoadBalancersResponse) v1alpha1.CLBObservation {
	observation := v1alpha1.CLBObservation{}
	if *res.Body.TotalCount == 0 {
		return observation
	}
	lb := res.Body.LoadBalancers.LoadBalancer[0]
	observation = v1alpha1.CLBObservation{
		LoadBalancerID:               lb.LoadBalancerId,
		CreateTime:                   lb.CreateTime,
		NetworkType:                  lb.NetworkType,
		MasterZoneID:                 lb.MasterZoneId,
		ModificationProtectionReason: lb.ModificationProtectionReason,
		ModificationProtectionStatus: lb.ModificationProtectionStatus,
		LoadBalancerStatus:           lb.LoadBalancerStatus,
		ResourceGroupID:              lb.ResourceGroupId,
		DeleteProtection:             lb.DeleteProtection,
	}
	return observation
}

// IsUpdateToDate checks whether cr is up to date
//nolint:gocyclo
func IsUpdateToDate(cr *v1alpha1.CLB, res *sdk.DescribeLoadBalancersResponse) bool {
	spec := cr.Spec.ForProvider
	if *res.Body.TotalCount == 0 {
		return false
	}
	lb := res.Body.LoadBalancers.LoadBalancer[0]

	if *spec.Region == *lb.RegionId && *spec.LoadBalancerSpec == *lb.LoadBalancerSpec &&
		*spec.InternetChargeType == *lb.InternetChargeType && *spec.AddressType == *lb.AddressType &&
		*spec.Address == *lb.Address && *spec.Bandwidth == *lb.Bandwidth && *spec.VpcID == *lb.VpcId &&
		*spec.VSwitchID == *lb.VSwitchId && *spec.LoadBalancerName == *lb.LoadBalancerName {
		return true
	}
	return false
}
