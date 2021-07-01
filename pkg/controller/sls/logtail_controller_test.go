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

	sdk "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	slsclient "github.com/crossplane/provider-alibaba/pkg/clients/sls"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	slsv1alpha1 "github.com/crossplane/provider-alibaba/apis/sls/v1alpha1"
)

const (
	badConfigName = "abc"
	configName    = "def"
)

var validLogtailCR = &slsv1alpha1.Logtail{
	ObjectMeta: metav1.ObjectMeta{
		Annotations: map[string]string{meta.AnnotationKeyExternalName: configName},
	},
	Spec: slsv1alpha1.LogtailSpec{
		ForProvider: slsv1alpha1.LogtailParameters{
		},
	},
}

func (c *fakeSDKClient) DescribeConfig(Logtail string, config string) (*sdk.LogConfig, error) {
	switch config {
	case "":
		return nil, sdk.Error{Code: slsclient.ErrCodeLogtailNotExist, HTTPCode: int32(0)}
	case badConfigName:
		return nil, errors.New("unknown error")
	default:
		return &sdk.LogConfig{
			Name: config,
		}, nil

	}
	return nil, nil
}

func (c *fakeSDKClient) CreateConfig(name string, config slsv1alpha1.LogtailParameters) error {
	return nil
}

func (c *fakeSDKClient) UpdateConfig(Logtail string, config *sdk.LogConfig) error {
	return nil
}

func (c *fakeSDKClient) DeleteConfig(Logtail string, config string) error {
	return nil
}

func TestLogtailObserve(t *testing.T) {
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
		"NotLogtail": {
			reason: "We should return an error if the supplied managed resource is not anLogtail",
			mg:     nil,
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.New(errNotLogtail),
			},
		},
		"LogtailNotFound": {
			reason: "Logtail name could not be found",
			mg:     &slsv1alpha1.Logtail{},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
				err: nil,
			},
		},
		"LogtailOtherError": {
			reason: "We should report an unknown error",
			mg: &slsv1alpha1.Logtail{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{meta.AnnotationKeyExternalName: badConfigName},
				},
				Spec: slsv1alpha1.LogtailSpec{ForProvider: slsv1alpha1.LogtailParameters{
				}}},
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.Wrap(errors.New("unknown error"), errDescribeLogtail),
			},
		},
		"LogtailSuccessfullyFound": {
			reason: "Observing a Logtail successfully should return an ExternalObservation and nil error",
			mg:     validLogtailCR,
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  false,
					ConnectionDetails: managed.ConnectionDetails{},
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			logtailExternal := &logtailExternal{client: &fakeSDKClient{}}
			got, err := logtailExternal.Observe(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestLogtailCreate(t *testing.T) {
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
		"NotLogtail": {
			reason: "Not an Logtail object",
			mg:     nil,
			want: want{
				o:   managed.ExternalCreation{},
				err: errors.New(errNotLogtail),
			},
		},
		"Success": {
			reason: "Creating anLogtail successfully",
			mg:     validLogtailCR,
			want: want{
				o: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{},
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			logtailExternal := &logtailExternal{client: &fakeSDKClient{}}
			got, err := logtailExternal.Create(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Create(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Create(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestLogtailUpdate(t *testing.T) {
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
		"NotLogtail": {
			reason: "Not an Logtail object",
			mg:     nil,
			want: want{
				o:   managed.ExternalUpdate{},
				err: nil,
			},
		},
		"Success": {
			reason: "Creating anLogtail successfully",
			mg:     validLogtailCR,
			want: want{
				o:   managed.ExternalUpdate{},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			logtailExternal := &logtailExternal{client: &fakeSDKClient{}}
			got, err := logtailExternal.Update(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Update(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Update(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestLogtailDelete(t *testing.T) {
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
		"NotLogtail": {
			reason: "Not an Logtail object",
			mg:     nil,
			want: want{
				err: errors.New(errNotLogtail),
			},
		},
		"Success": {
			reason: "Creating anLogtail successfully",
			mg:     validLogtailCR,
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			logtailExternal := &logtailExternal{client: &fakeSDKClient{}}
			err := logtailExternal.Delete(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Delete(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}
