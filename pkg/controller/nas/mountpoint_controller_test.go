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

func (c *fakeSDKClient) DescribeMountTargets(fileSystemID, mountTargetDomain *string) (*sdk.DescribeMountTargetsResponse, error) {
	switch *fileSystemID {
	case "123":
		return nil, errors.New("unknown error")
	default:
		body := &sdk.DescribeMountTargetsResponseBody{
			MountTargets: &sdk.DescribeMountTargetsResponseBodyMountTargets{
				MountTarget: []*sdk.DescribeMountTargetsResponseBodyMountTargetsMountTarget{}}, TotalCount: pointer.Int32Ptr(0)}
		NASMountTargetInfoResult := sdk.DescribeMountTargetsResponse{
			Body: body,
		}
		return &NASMountTargetInfoResult, nil
	}
}

func (c *fakeSDKClient) CreateMountTarget(fs v1alpha1.NASMountTargetParameter) (*sdk.CreateMountTargetResponse, error) {
	res := &sdk.CreateMountTargetResponse{Body: &sdk.CreateMountTargetResponseBody{MountTargetDomain: pointer.StringPtr("abc.com")}}
	return res, nil
}

func (c *fakeSDKClient) DeleteMountTarget(fileSystemID, mountTargetDomain *string) error {
	return nil
}

func TestMountTargetObserve(t *testing.T) {
	var ctx = context.Background()

	invalidCR := &v1alpha1.NASMountTarget{Spec: v1alpha1.NASMountTargetSpec{ForProvider: v1alpha1.NASMountTargetParameter{
		FileSystemID: pointer.StringPtr("123")}},
		Status: v1alpha1.NASMountTargetStatus{AtProvider: v1alpha1.NASMountTargetObservation{
			MountTargetDomain: pointer.StringPtr("abc.com"),
		}}}
	invalidCR.ObjectMeta.Annotations = map[string]string{meta.AnnotationKeyExternalName: "abc"}

	validCR := &v1alpha1.NASMountTarget{Spec: v1alpha1.NASMountTargetSpec{ForProvider: v1alpha1.NASMountTargetParameter{
		FileSystemID: pointer.StringPtr("456")}},
		Status: v1alpha1.NASMountTargetStatus{AtProvider: v1alpha1.NASMountTargetObservation{
			MountTargetDomain: pointer.StringPtr("abc.com"),
		}}}
	validCR.ObjectMeta.Annotations = map[string]string{meta.AnnotationKeyExternalName: "def"}

	type want struct {
		o   managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		reason string
		mg     resource.Managed
		want   want
	}{
		"NotNASMountTarget": {
			reason: "We should return an error if the supplied managed resource is not NASMountTarget",
			mg:     nil,
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.New(errNotNASMountTarget),
			},
		},
		"NASMountTargetNotFound": {
			reason: "We should report a NotFound error",
			mg:     &v1alpha1.NASMountTarget{},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
				err: nil,
			},
		},
		"NASMountTargetOtherError": {
			reason: "We should report an unknown error",
			mg:     invalidCR,
			want: want{
				o:   managed.ExternalObservation{ResourceExists: false},
				err: errors.Wrap(errors.New("unknown error"), errFailedToDescribeNASMountTarget),
			},
		},
		"Success": {
			reason: "Observing NASMountTarget successfully should return an ExternalObservation and nil error",
			mg:     validCR,
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: GetMountTargetConnectionDetails(validCR)},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			external := &mountTargetExternal{ExternalClient: &fakeSDKClient{}}
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

func TestMountTargetCreate(t *testing.T) {
	var ctx = context.Background()

	validCR := &v1alpha1.NASMountTarget{Spec: v1alpha1.NASMountTargetSpec{ForProvider: v1alpha1.NASMountTargetParameter{}}}
	validCR.Status.AtProvider.MountTargetDomain = pointer.StringPtr("abc.com")

	type want struct {
		o   managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		reason string
		mg     resource.Managed
		want   want
	}{
		"NotNASMountTarget": {
			reason: "Not NASMountTarget object",
			mg:     nil,
			want: want{
				o:   managed.ExternalCreation{},
				err: errors.New(errNotNASMountTarget),
			},
		},
		"Success": {
			reason: "Creating NASMountTarget successfully",
			mg:     validCR,
			want: want{
				o: managed.ExternalCreation{
					ConnectionDetails: GetMountTargetConnectionDetails(validCR)},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			external := &mountTargetExternal{ExternalClient: &fakeSDKClient{}}
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

func TestMountTargetDelete(t *testing.T) {
	var ctx = context.Background()

	validCR := &v1alpha1.NASMountTarget{
		Spec:       v1alpha1.NASMountTargetSpec{},
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
		"NotNASMountTarget": {
			reason: "Not NASMountTarget object",
			mg:     nil,
			want: want{
				err: errors.New(errNotNASMountTarget),
			},
		},
		"Success": {
			reason: "Deleting NASMountTarget successfully",
			mg:     validCR,
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			external := &mountTargetExternal{ExternalClient: &fakeSDKClient{}}
			err := external.Delete(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Delete(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}
