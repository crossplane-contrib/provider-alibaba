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
	"testing"

	sdk "github.com/alibabacloud-go/nas-20170626/v2/client"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/crossplane-contrib/provider-alibaba/apis/nas/v1alpha1"
)

type fakeSDKClient struct {
}

func (c *fakeSDKClient) DescribeFileSystems(fileSystemID, fileSystemType, vpcID *string) (*sdk.DescribeFileSystemsResponse, error) {
	switch *fileSystemID {
	case "123":
		return nil, errors.New("unknown error")
	default:
		body := &sdk.DescribeFileSystemsResponseBody{
			FileSystems: &sdk.DescribeFileSystemsResponseBodyFileSystems{
				FileSystem: []*sdk.DescribeFileSystemsResponseBodyFileSystemsFileSystem{}}, TotalCount: pointer.Int32Ptr(0)}
		NASFileSystemInfoResult := sdk.DescribeFileSystemsResponse{
			Body: body,
		}
		return &NASFileSystemInfoResult, nil
	}
}

func (c *fakeSDKClient) CreateFileSystem(fs v1alpha1.NASFileSystemParameter) (*sdk.CreateFileSystemResponse, error) {
	res := &sdk.CreateFileSystemResponse{Body: &sdk.CreateFileSystemResponseBody{FileSystemId: pointer.StringPtr("123456")}}
	return res, nil
}

func (c *fakeSDKClient) DeleteFileSystem(fileSystemID string) error {
	return nil
}

func TestObserve(t *testing.T) {
	var ctx = context.Background()

	invalidCR := &v1alpha1.NASFileSystem{}
	invalidCR.ObjectMeta.Annotations = map[string]string{meta.AnnotationKeyExternalName: "abc"}
	invalidCR.Status.AtProvider.FileSystemID = "123"

	validCR := &v1alpha1.NASFileSystem{Spec: v1alpha1.NASFileSystemSpec{}}
	validCR.Spec.FileSystemType = pointer.StringPtr("standard")
	validCR.ObjectMeta.Annotations = map[string]string{meta.AnnotationKeyExternalName: "def"}
	validCR.Status.AtProvider.FileSystemID = "456"

	type want struct {
		o   managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		reason string
		mg     resource.Managed
		want   want
	}{
		"NotAnNASFileSystem": {
			reason: "We should return an error if the supplied managed resource is not NASFileSystem",
			mg:     nil,
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.New(errNotNASFileSystem),
			},
		},
		"NASFileSystemNotFound": {
			reason: "We should report a NotFound error",
			mg:     &v1alpha1.NASFileSystem{},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
				err: nil,
			},
		},
		"NASFileSystemOtherError": {
			reason: "We should report an unknown error",
			mg:     invalidCR,
			want: want{
				o:   managed.ExternalObservation{ResourceExists: false},
				err: errors.Wrap(errors.New("unknown error"), errFailedToDescribeNASFileSystem),
			},
		},
		"Success": {
			reason: "Observing NASFileSystem successfully should return an ExternalObservation and nil error",
			mg:     validCR,
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: GetConnectionDetails(pointer.StringPtr("456"), validCR)},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			external := &External{ExternalClient: &fakeSDKClient{}}
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
	var ctx = context.Background()

	validCR := &v1alpha1.NASFileSystem{Spec: v1alpha1.NASFileSystemSpec{}}
	validCR.Spec.StorageType = pointer.StringPtr("standard")
	validCR.Spec.ProtocolType = pointer.StringPtr("nfs")
	validCR.ObjectMeta.Annotations = map[string]string{meta.AnnotationKeyExternalName: "def"}

	type want struct {
		o   managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		reason string
		mg     resource.Managed
		want   want
	}{
		"NotAnNASFileSystem": {
			reason: "Not NASFileSystem object",
			mg:     nil,
			want: want{
				o:   managed.ExternalCreation{},
				err: errors.New(errNotNASFileSystem),
			},
		},
		"Success": {
			reason: "Creating NASFileSystem successfully",
			mg:     validCR,
			want: want{
				o: managed.ExternalCreation{
					ConnectionDetails: GetConnectionDetails(pointer.StringPtr("123456"), validCR)},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			external := &External{ExternalClient: &fakeSDKClient{}}
			got, err := external.Create(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Create(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Create(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	var ctx = context.Background()

	validCR := &v1alpha1.NASFileSystem{
		Spec:       v1alpha1.NASFileSystemSpec{},
		ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{meta.AnnotationKeyExternalName: "def"}},
	}

	type want struct {
		err error
	}

	cases := map[string]struct {
		reason string
		mg     resource.Managed
		want   want
	}{
		"NotNASFileSystem": {
			reason: "Not NASFileSystem object",
			mg:     nil,
			want: want{
				err: errors.New(errNotNASFileSystem),
			},
		},
		"Success": {
			reason: "Deleting NASFileSystem successfully",
			mg:     validCR,
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			external := &External{ExternalClient: &fakeSDKClient{}}
			err := external.Delete(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Delete(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}
