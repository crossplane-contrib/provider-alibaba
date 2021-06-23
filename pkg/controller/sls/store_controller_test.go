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

	slsv1alpha1 "github.com/crossplane/provider-alibaba/apis/sls/v1alpha1"
	slsclient "github.com/crossplane/provider-alibaba/pkg/clients/sls"
)

var (
	project         = "abc"
	store           = "def"
	notExistedStore = "not-found-abc"
	someOtherError  = "Some other error"
	validStoreCR    = &slsv1alpha1.LogStore{
		ObjectMeta: metav1.ObjectMeta{
			Name:        store,
			Annotations: map[string]string{meta.AnnotationKeyExternalName: store},
		},
		Spec: slsv1alpha1.LogStoreSpec{
			ForProvider: slsv1alpha1.StoreParameters{
				ProjectName: project,
				TTL:         1,
				ShardCount:  2,
			},
		},
		Status: slsv1alpha1.LogStoreStatus{
			AtProvider: slsv1alpha1.StoreObservation{
				CreateTime:     123,
				LastModifyTime: 234,
			},
		},
	}
	validStore = &sdk.LogStore{Name: store, TTL: 1, ShardCount: 2}
)

func (c *fakeSDKClient) DescribeStore(project string, logstore string) (*sdk.LogStore, error) {
	switch logstore {
	case "":
		return nil, errors.Wrap(&sdk.Error{Code: slsclient.ErrCodeStoreNotExist}, "xxx")
	case notExistedStore:
		return nil, errors.New(someOtherError)
	default:
		return validStore, nil
	}
}

func (c *fakeSDKClient) CreateStore(project string, logstore *sdk.LogStore) error {
	return nil
}

func (c *fakeSDKClient) UpdateStore(project string, logstore string, ttl int) error {
	return nil
}

func (c *fakeSDKClient) DeleteStore(project string, logstore string) error {
	return nil
}

func TestStoreObserve(t *testing.T) {
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
		"NotSLSStore": {
			reason: "We should return an error if the supplied managed resource is not SLS store",
			mg:     nil,
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.New(errNotStore),
			},
		},
		"SLSStoreNotFound": {
			reason: "SLS LogStore name could not be found",
			mg:     &slsv1alpha1.LogStore{},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
				err: nil,
			},
		},
		"SLSStoreOtherError": {
			reason: "We should report an unknown error",
			mg: &slsv1alpha1.LogStore{
				ObjectMeta: metav1.ObjectMeta{
					Name:        notExistedStore,
					Annotations: map[string]string{meta.AnnotationKeyExternalName: notExistedStore},
				},
				Spec: slsv1alpha1.LogStoreSpec{ForProvider: slsv1alpha1.StoreParameters{
					ProjectName: "sls-project-test",
					TTL:         1,
					ShardCount:  2,
				}}},
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.New(someOtherError),
			},
		},
		"SLSStoreSuccessfullyFound": {
			reason: "Observing SLS store successfully should return an ExternalObservation and nil error",
			mg:     validStoreCR,
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: getStoreConnectionDetails(project, store)},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			external := &storeExternal{client: &fakeSDKClient{}}
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

func TestStoreCreate(t *testing.T) {
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
		"NotSLSStore": {
			reason: "Not LogStore object",
			mg:     nil,
			want: want{
				o:   managed.ExternalCreation{},
				err: errors.New(errNotStore),
			},
		},
		"Success": {
			reason: "Creating SLS store successfully",
			mg:     validStoreCR,
			want: want{
				o: managed.ExternalCreation{
					ConnectionDetails: getStoreConnectionDetails(project, store)},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			external := &storeExternal{client: &fakeSDKClient{}}
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

func TestStoreUpdate(t *testing.T) {
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
		"NotSLSStore": {
			reason: "Not LogStore object",
			mg:     nil,
			want: want{
				o:   managed.ExternalUpdate{},
				err: errors.New(errNotStore),
			},
		},
		"Success": {
			reason: "Creating SLS store successfully",
			mg:     validStoreCR,
			want: want{
				o:   managed.ExternalUpdate{},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			external := &storeExternal{client: &fakeSDKClient{}}
			got, err := external.Update(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Update(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Update(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestStoreDelete(t *testing.T) {
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
		"NotSLSStore": {
			reason: "Not LogStore object",
			mg:     nil,
			want: want{
				err: errors.New(errNotStore),
			},
		},
		"Success": {
			reason: "Creating SLS store successfully",
			mg:     validStoreCR,
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			external := &storeExternal{client: &fakeSDKClient{}}
			err := external.Delete(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Delete(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}
