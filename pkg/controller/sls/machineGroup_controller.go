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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	aliv1alpha1 "github.com/crossplane/provider-alibaba/apis/sls/v1alpha1"
	"github.com/crossplane/provider-alibaba/apis/v1alpha1"
	slsclient "github.com/crossplane/provider-alibaba/pkg/clients/sls"
	"github.com/crossplane/provider-alibaba/pkg/util"
)

const (
	errCreateMachineGroup   = "failed to create MachineGroup"
	errDeleteMachineGroup   = "failed to delete MachineGroup"
	errDescribeMachineGroup = "failed to describe MachineGroup"
	errNotMachineGroup      = "managed resource is not an MachineGroup custom resource"
)

// SetupMachineGroup adds a controller that reconciles MachineGroup.
func SetupMachineGroup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(aliv1alpha1.MachineGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&aliv1alpha1.MachineGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(aliv1alpha1.MachineGroupVersionKind),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithExternalConnecter(&machineGroupConnector{
				client:      mgr.GetClient(),
				usage:       resource.NewProviderConfigUsageTracker(mgr.GetClient(), &v1alpha1.ProviderConfigUsage{}),
				NewClientFn: slsclient.NewClient,
			})))
}

// machineGroupConnector stores Kubernetes client and SLS client
type machineGroupConnector struct {
	client      client.Client
	usage       resource.Tracker
	NewClientFn func(accessKeyID, accessKeySecret, securityToken, region string) *slsclient.LogClient
}

// Connect initials cloud resource client
func (c *machineGroupConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*aliv1alpha1.MachineGroup)
	if !ok {
		return nil, errors.New(errNotMachineGroup)
	}

	var (
		sel    *xpv1.SecretKeySelector
		region string
	)

	switch {
	case cr.GetProviderConfigReference() != nil:
		if err := c.usage.Track(ctx, mg); err != nil {
			return nil, errors.Wrap(err, errTrackUsage)
		}

		pc := &v1alpha1.ProviderConfig{}
		if err := c.client.Get(ctx, types.NamespacedName{Name: cr.Spec.ProviderConfigReference.Name}, pc); err != nil {
			return nil, errors.Wrap(err, errGetProviderConfig)
		}
		if s := pc.Spec.Credentials.Source; s != xpv1.CredentialsSourceSecret {
			return nil, errors.Errorf(errFmtUnsupportedCredSource, s)
		}
		sel = pc.Spec.Credentials.SecretRef
		region = pc.Spec.Region
	default:
		return nil, errors.New(errNoProvider)
	}

	if sel == nil {
		return nil, errors.New(errNoConnectionSecret)
	}

	s := &corev1.Secret{}
	nn := types.NamespacedName{Namespace: sel.Namespace, Name: sel.Name}
	if err := c.client.Get(ctx, nn, s); err != nil {
		return nil, errors.Wrap(err, errGetConnectionSecret)
	}

	slsClient := c.NewClientFn(string(s.Data[util.AccessKeyID]), string(s.Data[util.AccessKeySecret]), string(s.Data[util.SecurityToken]), region)
	return &machineGroupExternal{client: slsClient}, nil
}

// machineGroupExternal includes external SLS client
type machineGroupExternal struct {
	client slsclient.LogClientInterface
}

// Observe managed resource LogstoreMachineGroup
func (e *machineGroupExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*aliv1alpha1.MachineGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotMachineGroup)
	}

	if meta.GetExternalName(mg) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	machineGroup, err := e.client.DescribeMachineGroup(cr.Spec.ForProvider.Project, meta.GetExternalName(mg))
	if err != nil {
		if slsclient.IsMachineGroupNotFoundError(err) {
			return managed.ExternalObservation{ResourceExists: false, ResourceUpToDate: true}, nil
		}
		return managed.ExternalObservation{ResourceExists: false}, errors.Wrap(err, errDescribeMachineGroup)
	}
	cr.Status.AtProvider = slsclient.GenerateMachineGroupObservation(machineGroup)

	var upToDate = slsclient.IsMachineGroupUpdateToDate(cr, machineGroup)
	if upToDate {
		cr.SetConditions(xpv1.Available())
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: GetMachineGroupConnectionDetails(cr),
	}, nil
}

// Create managed resource LogstoreMachineGroup
func (e *machineGroupExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*aliv1alpha1.MachineGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotMachineGroup)
	}
	cr.SetConditions(xpv1.Creating())

	err := e.client.CreateMachineGroup(meta.GetExternalName(mg), cr.Spec.ForProvider)
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateMachineGroup)
}

// Update managed resource LogstoreMachineGroup
func (e *machineGroupExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// TODO(zzxwill) need to add Update logic here
	return managed.ExternalUpdate{}, nil
}

// Delete managed resource LogstoreMachineGroup
func (e *machineGroupExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*aliv1alpha1.MachineGroup)
	if !ok {
		return errors.New(errNotMachineGroup)
	}
	cr.SetConditions(xpv1.Deleting())
	if err := e.client.DeleteMachineGroup(cr.Spec.ForProvider.Project, meta.GetExternalName(mg)); err != nil {
		return errors.Wrap(err, errDeleteMachineGroup)
	}
	return nil
}

// GetMachineGroupConnectionDetails generates connection details
func GetMachineGroupConnectionDetails(cr *aliv1alpha1.MachineGroup) managed.ConnectionDetails {
	cd := managed.ConnectionDetails{}
	return cd
}
