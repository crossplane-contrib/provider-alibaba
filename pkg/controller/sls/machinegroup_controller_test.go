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
	"k8s.io/utils/pointer"

	slsv1alpha1 "github.com/crossplane/provider-alibaba/apis/sls/v1alpha1"
	slsclient "github.com/crossplane/provider-alibaba/pkg/clients/sls"
)

var (
	badMachineGroupName = "abc"
	machineGroupName    = "def"
	machineIDList       = []string{"192.168.2.1", "192.168.2.2"}
)

var validMachineGroupCR = &slsv1alpha1.MachineGroup{
	ObjectMeta: metav1.ObjectMeta{
		Annotations: map[string]string{meta.AnnotationKeyExternalName: machineGroupName},
	},
	Spec: slsv1alpha1.MachineGroupSpec{ForProvider: slsv1alpha1.MachineGroupParameters{
		Project:       &machineGroupName,
		Logstore:      pointer.StringPtr("ls"),
		MachineIDType: pointer.StringPtr("xxx"),
		MachineIDList: &machineIDList,
		Attribute: &sdk.MachinGroupAttribute{
			ExternalName: "xxx",
			TopicName:    "xxx",
		},
	}},
}

func (c *fakeSDKClient) DescribeMachineGroup(project *string, name string) (*sdk.MachineGroup, error) {
	switch name {
	case "":
		return nil, sdk.Error{Code: slsclient.ErrCodeMachineGroupNotExist, HTTPCode: int32(0)}
	case badMachineGroupName:
		return nil, errors.New("unknown error")
	default:
		return &sdk.MachineGroup{
			Name:          *project,
			MachineIDType: "xxx",
			MachineIDList: machineIDList,
			Attribute: sdk.MachinGroupAttribute{
				ExternalName: "xxx",
				TopicName:    "xxx",
			},
		}, nil

	}
}

func (c *fakeSDKClient) CreateMachineGroup(name string, param slsv1alpha1.MachineGroupParameters) error {
	return nil
}

func (c *fakeSDKClient) UpdateMachineGroup(project, logstore *string, machineGroup *sdk.MachineGroup) error {
	return nil
}

func (c *fakeSDKClient) DeleteMachineGroup(project *string, logstore string) error {
	return nil
}

func TestMachineGroupObserve(t *testing.T) {
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
		"NotMachineGroup": {
			reason: "We should return an error if the supplied managed resource is not anMachineGroup",
			mg:     nil,
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.New(errNotMachineGroup),
			},
		},
		"MachineGroupNotFound": {
			reason: "MachineGroup name could not be found",
			mg: &slsv1alpha1.MachineGroup{
				Spec: slsv1alpha1.MachineGroupSpec{ForProvider: slsv1alpha1.MachineGroupParameters{Project: &badMachineGroupName}},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
				err: nil,
			},
		},
		"MachineGroupOtherError": {
			reason: "We should report an unknown error",
			mg: &slsv1alpha1.MachineGroup{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{meta.AnnotationKeyExternalName: badMachineGroupName},
				},
				Spec: slsv1alpha1.MachineGroupSpec{ForProvider: slsv1alpha1.MachineGroupParameters{Project: &machineGroupName}},
			},
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.Wrap(errors.New("unknown error"), errDescribeMachineGroup),
			},
		},
		"MachineGroupSuccessfullyFound": {
			reason: "Observing a MachineGroup successfully should return an ExternalObservation and nil error",
			mg:     validMachineGroupCR,
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
			machineGroupExternal := &machineGroupExternal{client: &fakeSDKClient{}}
			got, err := machineGroupExternal.Observe(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestMachineGroupCreate(t *testing.T) {
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
		"NotMachineGroup": {
			reason: "Not an MachineGroup object",
			mg:     nil,
			want: want{
				o:   managed.ExternalCreation{},
				err: errors.New(errNotMachineGroup),
			},
		},
		"Success": {
			reason: "Creating anMachineGroup successfully",
			mg:     validMachineGroupCR,
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
			machineGroupExternal := &machineGroupExternal{client: &fakeSDKClient{}}
			got, err := machineGroupExternal.Create(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Create(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Create(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestMachineGroupUpdate(t *testing.T) {
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
		"NotMachineGroup": {
			reason: "Not an MachineGroup object",
			mg:     nil,
			want: want{
				o:   managed.ExternalUpdate{},
				err: nil,
			},
		},
		"Success": {
			reason: "Creating anMachineGroup successfully",
			mg:     validMachineGroupCR,
			want: want{
				o:   managed.ExternalUpdate{},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			machineGroupExternal := &machineGroupExternal{client: &fakeSDKClient{}}
			got, err := machineGroupExternal.Update(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Update(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Update(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestMachineGroupDelete(t *testing.T) {
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
		"NotMachineGroup": {
			reason: "Not an MachineGroup object",
			mg:     nil,
			want: want{
				err: errors.New(errNotMachineGroup),
			},
		},
		"Success": {
			reason: "Creating anMachineGroup successfully",
			mg:     validMachineGroupCR,
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			machineGroupExternal := &machineGroupExternal{client: &fakeSDKClient{}}
			err := machineGroupExternal.Delete(ctx, tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Delete(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}
