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
	"errors"
	"testing"

	sdk "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"k8s.io/klog/v2"

	ossv1alpha1 "github.com/crossplane/provider-alibaba/apis/oss/v1alpha1"
	ossclient "github.com/crossplane/provider-alibaba/pkg/clients/oss"
)

type fakeSDKClient struct {
}

// Describe describes Bucket bucket
func (c *fakeSDKClient) Describe(name string) (*sdk.GetBucketInfoResult, error) {
	switch name {
	case "":
		return nil, sdk.ServiceError{Code: "NoSuchBucket"}
	case "abc":
		return nil, errors.New("unknown error")
	default:
		bucketInfoResult := sdk.GetBucketInfoResult{
			BucketInfo: sdk.BucketInfo{
				Name:         name,
				ACL:          "private",
				StorageClass: "Standard",
			},
		}
		return &bucketInfoResult, nil

	}
}

// Create creates Bucket bucket
func (c *fakeSDKClient) Create(bucket ossv1alpha1.BucketParameter) error {
	return nil
}

// Update sets bucket acl
func (c *fakeSDKClient) Update(name string, aclStr string) error {
	_, err := ossclient.ValidateOSSAcl(aclStr)
	if err != nil {
		klog.ErrorS(err, "Name", name, "ACL", aclStr)
		return err
	}
	return nil
}

// Delete deletes SLS project
func (c *fakeSDKClient) Delete(name string) error {
	return nil
}

func TestObserve(t *testing.T) {
	var (
		ctx = context.Background()
	)
	validCR := &ossv1alpha1.Bucket{}
	validCR.Spec.Name = "def"

	invalidCR := &ossv1alpha1.Bucket{}
	invalidCR.Spec.Name = "abc"

	type want struct {
		o   managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		reason string
		mg     resource.Managed
		want   want
	}{
		"NotAnOSS": {
			reason: "We should return an error if the supplied managed resource is not an Bucket bucket",
			mg:     nil,
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.New(errNotOSS),
			},
		},
		"OSSNotFound": {
			reason: "We should report a NotFound error",
			mg:     &ossv1alpha1.Bucket{},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
				err: nil,
			},
		},
		"OSSOtherError": {
			reason: "We should report an unknown error",
			mg:     invalidCR,
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.New("unknown error"),
			},
		},
		"Success": {
			reason: "Observing an Bucket bucket successfully should return an ExternalObservation and nil error",
			mg:     validCR,
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: GetConnectionDetails(validCR)},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			external := &external{client: &fakeSDKClient{}}
			got, err := external.Observe(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	var (
		ctx = context.Background()
	)

	spec := ossv1alpha1.BucketSpec{}
	spec.Name = "def"
	validCR := &ossv1alpha1.Bucket{Spec: spec}

	type want struct {
		o   managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		reason string
		mg     resource.Managed
		want   want
	}{
		"NotAnOSS": {
			reason: "Not an Bucket object",
			mg:     nil,
			want: want{
				o:   managed.ExternalCreation{},
				err: errors.New(errNotOSS),
			},
		},
		"Success": {
			reason: "Creating an Bucket bucket successfully",
			mg:     validCR,
			want: want{
				o: managed.ExternalCreation{
					ConnectionDetails: GetConnectionDetails(validCR)},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			external := &external{client: &fakeSDKClient{}}
			got, err := external.Create(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	var (
		ctx = context.Background()
	)

	spec := ossv1alpha1.BucketSpec{}
	spec.Name = "def"
	validCR := &ossv1alpha1.Bucket{Spec: spec}

	type want struct {
		o   managed.ExternalUpdate
		err error
	}

	cases := map[string]struct {
		reason string
		mg     resource.Managed
		want   want
	}{
		"NotAnOSS": {
			reason: "Not an Bucket object",
			mg:     nil,
			want: want{
				o:   managed.ExternalUpdate{},
				err: errors.New(errNotOSS),
			},
		},
		"Success": {
			reason: "Creating an Bucket bucket successfully",
			mg:     validCR,
			want: want{
				o:   managed.ExternalUpdate{},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			external := &external{client: &fakeSDKClient{}}
			got, err := external.Update(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	var (
		ctx = context.Background()
	)

	spec := ossv1alpha1.BucketSpec{}
	spec.Name = "def"
	validCR := &ossv1alpha1.Bucket{Spec: spec}

	type want struct {
		err error
	}

	cases := map[string]struct {
		reason string
		mg     resource.Managed
		want   want
	}{
		"NotAnOSS": {
			reason: "Not an Bucket object",
			mg:     nil,
			want: want{
				err: errors.New(errNotOSS),
			},
		},
		"Success": {
			reason: "Creating an Bucket bucket successfully",
			mg:     validCR,
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			external := &external{client: &fakeSDKClient{}}
			err := external.Delete(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}
