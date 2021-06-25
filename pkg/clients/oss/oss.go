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

package oss

import (
	"context"

	sdk "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-alibaba/apis/oss/v1alpha1"
)

// ErrCodeNoSuchBucket is the error code "NoSuchBucket" returned by SDK
const ErrCodeNoSuchBucket = "NoSuchBucket"

// ClientInterface will help fakeOSSClient in unit tests
type ClientInterface interface {
	Describe(name string) (*sdk.GetBucketInfoResult, error)
	Create(name string, bucket v1alpha1.BucketParameter) error
	Update(name string, aclStr string) error
	Delete(name string) error
}

// SDKClient is the SDK client for Bucket
type SDKClient struct {
	Client *sdk.Client
}

// NewClient will create OSS client
func NewClient(ctx context.Context, endpoint string, accessKeyID string, accessKeySecret string, stsToken string) (*SDKClient, error) {
	var (
		client *sdk.Client
		err    error
	)

	if stsToken == "" {
		client, err = sdk.New(endpoint, accessKeyID, accessKeySecret)
	} else {
		client, err = sdk.New(endpoint, accessKeyID, accessKeySecret, sdk.SecurityToken(stsToken))
	}
	if err != nil {
		return nil, errors.Errorf("failed to crate Bucket client: %v", err)
	}
	return &SDKClient{Client: client}, nil
}

// Describe describes OSS bucket
func (c *SDKClient) Describe(name string) (*sdk.GetBucketInfoResult, error) {
	bucketInfoResult, err := c.Client.GetBucketInfo(name)
	if err != nil {
		return nil, err
	}
	return &bucketInfoResult, nil
}

// Create creates Bucket bucket
func (c *SDKClient) Create(name string, bucket v1alpha1.BucketParameter) error {
	var options []sdk.Option
	var (
		acl                sdk.ACLType
		storageClass       sdk.StorageClassType
		dataRedundancyType sdk.DataRedundancyType
		err                error
	)

	// validate ACL
	acl, err = ValidateOSSAcl(bucket.ACL)
	if err != nil {
		return err
	}
	options = append(options, sdk.ACL(acl))

	// validate StorageClass
	storageClass, err = validateOSSStorageClass(bucket.StorageClass)
	if err != nil {
		return err
	}
	options = append(options, sdk.StorageClass(storageClass))

	// validate DataRedundancyType
	dataRedundancyType, err = validateOSSDataRedundancyType(bucket.DataRedundancyType)
	if err != nil {
		return err
	}
	options = append(options, sdk.RedundancyType(dataRedundancyType))

	if err := c.Client.CreateBucket(name, options...); err != nil {
		return err
	}
	return nil
}

// Update sets bucket acl
func (c *SDKClient) Update(name string, aclStr string) error {
	acl, err := ValidateOSSAcl(aclStr)
	if err != nil {
		return err
	}
	return c.Client.SetBucketACL(name, acl)
}

// Delete deletes OSS Bucket
func (c *SDKClient) Delete(name string) error {
	return c.Client.DeleteBucket(name)
}

// IsNotFoundError checks whether the error is an NotFound error
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	e, ok := errors.Cause(err).(sdk.ServiceError)
	return ok && e.Code == ErrCodeNoSuchBucket
}

// GenerateObservation generates BucketObservation from bucket information
func GenerateObservation(r sdk.GetBucketInfoResult) v1alpha1.BucketObservation {
	return v1alpha1.BucketObservation{
		ExtranetEndpoint: r.BucketInfo.ExtranetEndpoint,
		IntranetEndpoint: r.BucketInfo.IntranetEndpoint,
	}
}

// ValidateOSSAcl validates Bucket ACL and convert it to sdk.ACLType if possible
func ValidateOSSAcl(aclStr string) (sdk.ACLType, error) {
	var acl sdk.ACLType
	switch aclStr {
	case string(sdk.ACLPublicRead):
		acl = sdk.ACLPublicRead
	case string(sdk.ACLPublicReadWrite):
		acl = sdk.ACLPublicReadWrite
	case string(sdk.ACLPrivate), "":
		acl = sdk.ACLPrivate
	default:
		err := errors.Errorf("bucket ACL %s is invalid. The ACL could only be public-read-write, public-read, and private", aclStr)
		return "", err
	}
	return acl, nil
}

func validateOSSStorageClass(storageClassStr string) (sdk.StorageClassType, error) {
	var storageClass sdk.StorageClassType
	switch storageClassStr {
	case string(sdk.StorageStandard), "":
		storageClass = sdk.StorageStandard
	case string(sdk.StorageIA):
		storageClass = sdk.StorageIA
	case string(sdk.StorageArchive):
		storageClass = sdk.StorageArchive
	case string(sdk.StorageColdArchive):
		storageClass = sdk.StorageColdArchive
	default:
		err := errors.Errorf("bucket StorageClass %s is invalid. It only supports could be Standard, IA, Archive, and ColdArchive", storageClassStr)
		return "", err
	}
	return storageClass, nil
}

func validateOSSDataRedundancyType(dataRedundancyTypeStr string) (sdk.DataRedundancyType, error) {
	var dataRedundancyType sdk.DataRedundancyType
	switch dataRedundancyTypeStr {
	case string(sdk.RedundancyLRS), "":
		dataRedundancyType = sdk.RedundancyLRS
	case string(sdk.RedundancyZRS):
		dataRedundancyType = sdk.RedundancyZRS
	default:
		return "", errors.Errorf("bucket DataRedundancyType %s is invalid. It only supports could be LRS and ZRS", dataRedundancyType)
	}
	return dataRedundancyType, nil
}

// IsUpdateToDate checks whether cr is up to date
func IsUpdateToDate(cr *v1alpha1.Bucket, bucket *sdk.GetBucketInfoResult) bool {
	if (cr.Spec.ACL == bucket.BucketInfo.ACL) || (cr.Spec.ACL == "" && bucket.BucketInfo.ACL == "private") {
		return true
	}
	return false
}
