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

package util

import (
	"errors"
	"fmt"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/crossplane/provider-alibaba/apis/oss/v1alpha1"
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
			res:    nil,
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
