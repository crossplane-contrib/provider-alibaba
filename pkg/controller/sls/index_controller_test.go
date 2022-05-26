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
	"testing"

	sdk "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	slsv1alpha1 "github.com/crossplane-contrib/provider-alibaba/apis/sls/v1alpha1"
	slsclient "github.com/crossplane-contrib/provider-alibaba/pkg/clients/sls"
)

var (
	badIndexName = "abc"
	indexName    = "def"
)

var validIndexCR = &slsv1alpha1.LogstoreIndex{
	ObjectMeta: metav1.ObjectMeta{
		Annotations: map[string]string{meta.AnnotationKeyExternalName: indexName},
	},
	Spec: slsv1alpha1.LogstoreIndexSpec{
		ForProvider: slsv1alpha1.LogstoreIndexParameters{
			ProjectName: &indexName,
		},
	},
}

func (c *fakeSDKClient) DescribeIndex(project, logstore *string) (*sdk.Index, error) {
	switch *project {
	case "":
		return nil, sdk.Error{Code: slsclient.ErrCodeLogstoreIndexNotExist, HTTPCode: int32(0)}
	case badIndexName:
		return nil, errors.New("unknown error")
	default:
		return &sdk.Index{}, nil
	}
}

func (c *fakeSDKClient) CreateIndex(param slsv1alpha1.LogstoreIndexParameters) error {
	return nil
}

func (c *fakeSDKClient) UpdateIndex(project, logstore *string, index *sdk.Index) error {
	return nil
}

func (c *fakeSDKClient) DeleteIndex(project, logstore *string) error {
	return nil
}

func TestIndexObserve(t *testing.T) {
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
		"NotIndex": {
			reason: "We should return an error if the supplied managed resource is not anIndex",
			mg:     nil,
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.New(errNotIndex),
			},
		},
		"IndexNotFound": {
			reason: "Index name could not be found",
			mg:     &slsv1alpha1.LogstoreIndex{},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
				err: nil,
			},
		},
		"IndexOtherError": {
			reason: "We should report an unknown error",
			mg: &slsv1alpha1.LogstoreIndex{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{meta.AnnotationKeyExternalName: badIndexName},
				},
				Spec: slsv1alpha1.LogstoreIndexSpec{ForProvider: slsv1alpha1.LogstoreIndexParameters{ProjectName: &badIndexName}}},
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.Wrap(errors.New("unknown error"), errDescribeIndex),
			},
		},
		"IndexSuccessfullyFound": {
			reason: "Observing a Index successfully should return an ExternalObservation and nil error",
			mg:     validIndexCR,
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: managed.ConnectionDetails{},
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			indexExternal := &indexExternal{client: &fakeSDKClient{}}
			got, err := indexExternal.Observe(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestIndexCreate(t *testing.T) {
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
		"NotIndex": {
			reason: "Not an Index object",
			mg:     nil,
			want: want{
				o:   managed.ExternalCreation{},
				err: errors.New(errNotIndex),
			},
		},
		"Success": {
			reason: "Creating anIndex successfully",
			mg:     validIndexCR,
			want: want{
				o: managed.ExternalCreation{
					ConnectionDetails: nil,
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			indexExternal := &indexExternal{client: &fakeSDKClient{}}
			got, err := indexExternal.Create(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Create(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Create(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestIndexUpdate(t *testing.T) {
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
		"NotIndex": {
			reason: "Not an Index object",
			mg:     nil,
			want: want{
				o:   managed.ExternalUpdate{},
				err: nil,
			},
		},
		"Success": {
			reason: "Creating anIndex successfully",
			mg:     validIndexCR,
			want: want{
				o:   managed.ExternalUpdate{},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			indexExternal := &indexExternal{client: &fakeSDKClient{}}
			got, err := indexExternal.Update(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Update(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Update(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestIndexDelete(t *testing.T) {
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
		"NotIndex": {
			reason: "Not an Index object",
			mg:     nil,
			want: want{
				err: errors.New(errNotIndex),
			},
		},
		"Success": {
			reason: "Creating anIndex successfully",
			mg:     validIndexCR,
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			indexExternal := &indexExternal{client: &fakeSDKClient{}}
			err := indexExternal.Delete(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Delete(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}
