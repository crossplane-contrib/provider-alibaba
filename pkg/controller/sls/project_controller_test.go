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
	"context"
	"errors"
	"testing"

	sdk "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	slsv1alpha1 "github.com/crossplane/provider-alibaba/apis/sls/v1alpha1"
	slsclient "github.com/crossplane/provider-alibaba/pkg/clients/sls"
)

var (
	slsProjectDescription = "test project"
	slsProjectEndpoint    = "xxx.com"
	validCR               = &slsv1alpha1.SLSProject{Spec: slsv1alpha1.SLSProjectSpec{ForProvider: slsv1alpha1.SLSProjectParameters{
		Name:        "def",
		Description: slsProjectDescription,
	}}}

	validProject = &sdk.LogProject{Name: "def", Endpoint: slsProjectEndpoint}
)

type fakeSDKClient struct {
}

// Describe describes SLSProject bucket
func (c *fakeSDKClient) Describe(name string) (*sdk.LogProject, error) {
	switch name {
	case "":
		return nil, sdk.Error{Code: slsclient.ErrCodeProjectNotExist, HTTPCode: int32(0)}
	case "abc":
		return nil, errors.New("unknown error")
	default:
		return &sdk.LogProject{
			Name:        name,
			Description: slsProjectDescription,
			Endpoint:    slsProjectEndpoint,
		}, nil

	}
}

// Create creates SLSProject bucket
func (c *fakeSDKClient) Create(name, description string) (*sdk.LogProject, error) {
	return validProject, nil
}

// Update sets bucket acl
func (c *fakeSDKClient) Update(name, description string) (*sdk.LogProject, error) {
	return validProject, nil
}

// Delete deletes SLS project
func (c *fakeSDKClient) Delete(name string) error {
	return nil
}

func TestObserve(t *testing.T) {
	var (
		ctx = context.Background()
	)

	type want struct {
		o   managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		reason string
		mg     resource.Managed
		want   want
	}{
		"NotAnSLSProject": {
			reason: "We should return an error if the supplied managed resource is not an SLSProject bucket",
			mg:     nil,
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.New(errNotSLSProject),
			},
		},
		"SLSProjectNotFound": {
			reason: "SLS Project name could not be found",
			mg:     &slsv1alpha1.SLSProject{},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
				err: nil,
			},
		},
		"SLSProjectOtherError": {
			reason: "We should report an unknown error",
			mg: &slsv1alpha1.SLSProject{Spec: slsv1alpha1.SLSProjectSpec{ForProvider: slsv1alpha1.SLSProjectParameters{
				Name:        "abc",
				Description: "test project",
			}}},
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.New("unknown error"),
			},
		},
		"Success": {
			reason: "Observing an SLSProject bucket successfully should return an ExternalObservation and nil error",
			mg:     validCR,
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: getConnectionDetails(validProject)},
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

	type want struct {
		o   managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		reason string
		mg     resource.Managed
		want   want
	}{
		"NotAnSLSProject": {
			reason: "Not an SLSProject object",
			mg:     nil,
			want: want{
				o:   managed.ExternalCreation{},
				err: errors.New(errNotSLSProject),
			},
		},
		"Success": {
			reason: "Creating an SLSProject bucket successfully",
			mg:     validCR,
			want: want{
				o: managed.ExternalCreation{
					ConnectionDetails: getConnectionDetails(validProject)},
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

	type want struct {
		o   managed.ExternalUpdate
		err error
	}

	cases := map[string]struct {
		reason string
		mg     resource.Managed
		want   want
	}{
		"NotAnSLSProject": {
			reason: "Not an SLSProject object",
			mg:     nil,
			want: want{
				o:   managed.ExternalUpdate{},
				err: errors.New(errNotSLSProject),
			},
		},
		"Success": {
			reason: "Creating an SLSProject bucket successfully",
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

	type want struct {
		err error
	}

	cases := map[string]struct {
		reason string
		mg     resource.Managed
		want   want
	}{
		"NotAnSLSProject": {
			reason: "Not an SLSProject object",
			mg:     nil,
			want: want{
				err: errors.New(errNotSLSProject),
			},
		},
		"Success": {
			reason: "Creating an SLSProject bucket successfully",
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
