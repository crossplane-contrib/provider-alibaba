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

package alibabacloud

import (
	"context"
	"fmt"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/provider-alibaba/apis/oss/v1alpha1"
	"github.com/crossplane/provider-alibaba/apis/v1beta1"
)

func TestGetEndpoint(t *testing.T) {
	type want struct {
		endpoint string
		err      error
	}
	region := "cn-beijing"
	cr := v1alpha1.Bucket{TypeMeta: metav1.TypeMeta{Kind: "Bucket"}}

	cases := map[string]struct {
		res    runtime.Object
		region string
		want   want
	}{
		"NotExistedCloudResource": {
			res:    nil,
			region: region,
			want: want{
				endpoint: "",
				err:      errors.New(errCloudResourceNotSupported),
			},
		},
		"EmptyRegion": {
			res:    cr.DeepCopyObject(),
			region: "",
			want: want{
				endpoint: "",
				err:      errors.New(errRegionNotValid),
			},
		},
		"CloudResourceAndRegionAreValid": {
			res:    cr.DeepCopyObject(),
			region: region,
			want: want{
				endpoint: fmt.Sprintf("http://oss-%s.%s", region, Domain),
				err:      nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			endpoint, err := GetEndpoint(tc.res, tc.region)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\nc.GetEndpoint(...) -want error, +got error:\n%s\n", diff)
			}
			if diff := cmp.Diff(tc.want.endpoint, endpoint, test.EquateConditions()); diff != "" {
				t.Errorf("\nc.GetEndpoint(...) %s\n", diff)
			}
		})
	}
}

func TestGetCredentials(t *testing.T) {
	ctx := context.TODO()
	type args struct {
		client client.Client
		name   string
	}
	var pc v1beta1.ProviderConfig
	pc.Spec.Credentials.Source = "Secret"
	pc.Spec.Credentials.SecretRef = &xpv1.SecretKeySelector{
		Key: "credentials",
	}
	pc.Spec.Credentials.SecretRef.Name = "default"

	type want struct {
		cred *AlibabaCredentials
		err  error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"FailedToGetProviderConfig": {
			args: args{
				client: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						return errors.New("E1")
					}),
				},
				name: "abc",
			},
			want: want{
				cred: nil,
				err:  errors.Wrap(errors.New("E1"), ErrGetProviderConfig),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cred, err := GetCredentials(ctx, tc.args.client, tc.args.name)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\nGetCredentials(...) -want error, +got error:\n%s\n", diff)
			}
			if diff := cmp.Diff(tc.want.cred, cred, test.EquateConditions()); diff != "" {
				t.Errorf("\nGetEndpoint(...) %s\n", diff)
			}
		})
	}
}
