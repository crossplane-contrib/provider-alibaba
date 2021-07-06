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

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	slsv1alpha1 "github.com/crossplane/provider-alibaba/apis/sls/v1alpha1"
)

var (
	mgbProject                 = "abc"
	mgbGroup                   = "test-group"
	mgbConfig                  = "test-config"
	mgbBadProject              = "def"
	notExistedProject          = "not-found-abc"
	mgbOtherError              = "Some other error"
	validMachineGroupBindingCR = &slsv1alpha1.MachineGroupBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:        store,
			Annotations: map[string]string{meta.AnnotationKeyExternalName: mgbProject},
		},
		Spec: slsv1alpha1.MachineGroupBindingSpec{
			ForProvider: slsv1alpha1.MachineGroupBindingParameters{
				ProjectName: &mgbProject,
				GroupName:   &mgbGroup,
				ConfigName:  &mgbConfig,
			},
		},
		Status: slsv1alpha1.MachineGroupBindingStatus{
			AtProvider: slsv1alpha1.MachineGroupBindingObservation{
				Configs: []string{mgbConfig},
			},
		},
	}
)

func (c *fakeSDKClient) GetAppliedConfigs(projectName *string, groupName *string) ([]string, error) {
	switch *projectName {
	case mgbProject:
		return []string{mgbConfig}, nil
	case notExistedProject:
		return nil, nil
	case mgbBadProject:
		return nil, errors.New(mgbOtherError)
	default:
		return nil, nil
	}
}

func (c *fakeSDKClient) ApplyConfigToMachineGroup(projectName, groupName, confName *string) error {
	return nil
}

func (c *fakeSDKClient) RemoveConfigFromMachineGroup(projectName, groupName, confName *string) error {
	return nil
}

func TestMachineGroupBindingObserve(t *testing.T) {
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
		"NotMachineGroupBinding": {
			reason: "We should return an error if the supplied managed resource is not MachineGroupBinding",
			mg:     nil,
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.New(errNotMachineGroupBinding),
			},
		},
		"ExternalNameIsNotSet": {
			reason: "MachineGroupBinding's external name is not set",
			mg:     &slsv1alpha1.MachineGroupBinding{},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
				err: nil,
			},
		},
		"FailedToGetConfigs": {
			reason: "failed to get configs from a machine group",
			mg: &slsv1alpha1.MachineGroupBinding{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{meta.AnnotationKeyExternalName: mgbBadProject},
				},
				Spec: slsv1alpha1.MachineGroupBindingSpec{
					ForProvider: slsv1alpha1.MachineGroupBindingParameters{
						ProjectName: &mgbBadProject,
					},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
				err: errors.Wrap(errors.New(mgbOtherError), errDescribeMachineGroupBinding),
			},
		},
		"MachineGroupBindingNotFound": {
			reason: "MachineGroupBinding name could not be found",
			mg: &slsv1alpha1.MachineGroupBinding{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{meta.AnnotationKeyExternalName: notExistedProject},
				},
				Spec: slsv1alpha1.MachineGroupBindingSpec{
					ForProvider: slsv1alpha1.MachineGroupBindingParameters{
						ProjectName: &notExistedProject,
					},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: true,
				},
				err: nil,
			},
		},
		"MachineGroupBindingSuccessfullyFound": {
			reason: "Observing MachineGroupBinding successfully should return an ExternalObservation and nil error",
			mg:     validMachineGroupBindingCR,
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: GetMachineGroupBindingConnectionDetails([]string{mgbConfig})},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			external := &machineGroupBindingExternal{client: &fakeSDKClient{}}
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

func TestApplyConfigToMachineGroup(t *testing.T) {
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
		"NotMachineGroupBinding": {
			reason: "Not MachineGroupBinding object",
			mg:     nil,
			want: want{
				o:   managed.ExternalCreation{},
				err: errors.New(errNotMachineGroupBinding),
			},
		},
		"Success": {
			reason: "Creating MachineGroupBinding successfully",
			mg:     validMachineGroupBindingCR,
			want: want{
				o:   managed.ExternalCreation{},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			external := &machineGroupBindingExternal{client: &fakeSDKClient{}}
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

func TestRemoveConfigFromMachineGroup(t *testing.T) {
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
		"NotMachineGroupBinding": {
			reason: "Not MachineGroupBinding object",
			mg:     nil,
			want: want{
				err: errors.New(errNotMachineGroupBinding),
			},
		},
		"Success": {
			reason: "Creating MachineGroupBinding successfully",
			mg:     validMachineGroupBindingCR,
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			external := &machineGroupBindingExternal{client: &fakeSDKClient{}}
			err := external.Delete(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Delete(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}
