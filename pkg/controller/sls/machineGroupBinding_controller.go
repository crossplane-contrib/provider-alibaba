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
	"strings"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	aliv1alpha1 "github.com/crossplane/provider-alibaba/apis/sls/v1alpha1"
	"github.com/crossplane/provider-alibaba/apis/v1beta1"
	slsclient "github.com/crossplane/provider-alibaba/pkg/clients/sls"
	"github.com/crossplane/provider-alibaba/pkg/util"
)

const (
	errCreateMachineGroupBinding   = "failed to create MachineGroupBinding"
	errDeleteMachineGroupBinding   = "failed to delete MachineGroupBinding"
	errDescribeMachineGroupBinding = "failed to describe MachineGroupBinding"
	errNotMachineGroupBinding      = "managed resource is not a MachineGroupBinding custom resource"
)

// SetupMachineGroupBinding adds a controller that reconciles MachineGroupBinding
func SetupMachineGroupBinding(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(aliv1alpha1.MachineGroupBindingGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&aliv1alpha1.MachineGroupBinding{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(aliv1alpha1.MachineGroupBindingGroupVersionKind),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithExternalConnecter(&machineGroupBindingConnector{
				client:      mgr.GetClient(),
				usage:       resource.NewProviderConfigUsageTracker(mgr.GetClient(), &v1beta1.ProviderConfigUsage{}),
				NewClientFn: slsclient.NewClient,
			})))
}

// machineGroupBindingConnector stores Kubernetes client and SLS client
type machineGroupBindingConnector struct {
	client      client.Client
	usage       resource.Tracker
	NewClientFn func(accessKeyID, accessKeySecret, securityToken, region string) *slsclient.LogClient
}

// Connect initials cloud resource client
func (c *machineGroupBindingConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*aliv1alpha1.MachineGroupBinding)
	if !ok {
		return nil, errors.New(errNotMachineGroupBinding)
	}

	info, err := util.PrepareClient(ctx, mg, cr, c.client, c.usage, cr.Spec.ProviderConfigReference.Name)
	if err != nil {
		return nil, err
	}

	slsClient := c.NewClientFn(info.AccessKeyID, info.AccessKeySecret,
		info.SecurityToken, info.Region)
	return &machineGroupBindingExternal{client: slsClient}, nil
}

// machineGroupBindingExternal includes external SLS client
type machineGroupBindingExternal struct {
	client slsclient.LogClientInterface
}

// Observe managed resource MachineGroupBinding
func (e *machineGroupBindingExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*aliv1alpha1.MachineGroupBinding)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotMachineGroupBinding)
	}

	if meta.GetExternalName(mg) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	configs, err := e.client.GetAppliedConfigs(cr.Spec.ForProvider.ProjectName, cr.Spec.ForProvider.GroupName)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, errors.Wrap(err, errDescribeMachineGroupBinding)
	}
	if len(configs) == 0 {
		return managed.ExternalObservation{ResourceExists: false, ResourceUpToDate: true}, nil
	}
	cr.Status.AtProvider = slsclient.GenerateMachineGroupBindingObservation(configs)
	cr.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  true,
		ConnectionDetails: GetMachineGroupBindingConnectionDetails(configs),
	}, nil
}

// Create managed resource MachineGroupBinding
func (e *machineGroupBindingExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*aliv1alpha1.MachineGroupBinding)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotMachineGroupBinding)
	}
	cr.SetConditions(xpv1.Creating())

	err := e.client.ApplyConfigToMachineGroup(cr.Spec.ForProvider.ProjectName, cr.Spec.ForProvider.GroupName,
		cr.Spec.ForProvider.ConfigName)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateMachineGroupBinding)
	}

	return managed.ExternalCreation{}, nil
}

// Update managed resource MachineGroupBinding
func (e *machineGroupBindingExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// TODO(zzxwll) need to add Update logic
	return managed.ExternalUpdate{}, nil
}

// Delete managed resource MachineGroupBinding which will remove config from a machine group
func (e *machineGroupBindingExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*aliv1alpha1.MachineGroupBinding)
	if !ok {
		return errors.New(errNotMachineGroupBinding)
	}
	cr.SetConditions(xpv1.Deleting())
	if err := e.client.RemoveConfigFromMachineGroup(cr.Spec.ForProvider.ProjectName, cr.Spec.ForProvider.GroupName,
		cr.Spec.ForProvider.ConfigName); err != nil {
		return errors.Wrap(err, errDeleteMachineGroupBinding)
	}
	return nil
}

// GetMachineGroupBindingConnectionDetails generates connection details
func GetMachineGroupBindingConnectionDetails(configs []string) managed.ConnectionDetails {
	cd := managed.ConnectionDetails{
		"Configs": []byte(strings.Join(configs, ", ")),
	}
	return cd
}
