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
	errCreateIndex   = "failed to create Index"
	errDeleteIndex   = "failed to delete Index"
	errDescribeIndex = "failed to describe Index"
	errNotIndex      = "managed resource is not an Index custom resource"
)

// SetupIndex adds a controller that reconciles Index.
func SetupIndex(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(aliv1alpha1.IndexGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&aliv1alpha1.LogstoreIndex{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(aliv1alpha1.IndexGroupVersionKind),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithExternalConnecter(&indexConnector{
				client:      mgr.GetClient(),
				usage:       resource.NewProviderConfigUsageTracker(mgr.GetClient(), &v1alpha1.ProviderConfigUsage{}),
				NewClientFn: slsclient.NewClient,
			})))
}

// indexConnector stores Kubernetes client and SLS client
type indexConnector struct {
	client      client.Client
	usage       resource.Tracker
	NewClientFn func(accessKeyID, accessKeySecret, securityToken, region string) *slsclient.LogClient
}

// Connect initials cloud resource client
func (c *indexConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*aliv1alpha1.LogstoreIndex)
	if !ok {
		return nil, errors.New(errNotIndex)
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
	return &indexExternal{client: slsClient}, nil
}

// indexExternal includes external SLS client
type indexExternal struct {
	client slsclient.LogClientInterface
}

// Observe managed resource LogstoreIndex
func (e *indexExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*aliv1alpha1.LogstoreIndex)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotIndex)
	}

	if meta.GetExternalName(mg) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	index, err := e.client.DescribeIndex(cr.Spec.ForProvider.ProjectName, cr.Spec.ForProvider.LogstoreName)
	if err != nil {
		if slsclient.IsIndexNotFoundError(err) {
			return managed.ExternalObservation{ResourceExists: false, ResourceUpToDate: true}, nil
		}
		return managed.ExternalObservation{ResourceExists: false}, errors.Wrap(err, errDescribeIndex)
	}
	cr.Status.AtProvider = slsclient.GenerateIndexObservation(index)

	var upToDate = slsclient.IsIndexUpdateToDate(cr, index)
	if upToDate {
		cr.SetConditions(xpv1.Available())
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: GetIndexConnectionDetails(cr),
	}, nil
}

// Create managed resource LogstoreIndex
func (e *indexExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*aliv1alpha1.LogstoreIndex)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotIndex)
	}
	cr.SetConditions(xpv1.Creating())

	err := e.client.CreateIndex(cr.Spec.ForProvider)
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateIndex)
}

// Update managed resource LogstoreIndex
func (e *indexExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// TODO(zzxwll) need to add Update logic here
	return managed.ExternalUpdate{}, nil
}

// Delete managed resource LogstoreIndex
func (e *indexExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*aliv1alpha1.LogstoreIndex)
	if !ok {
		return errors.New(errNotIndex)
	}
	cr.SetConditions(xpv1.Deleting())
	if err := e.client.DeleteIndex(cr.Spec.ForProvider.ProjectName, cr.Spec.ForProvider.LogstoreName); err != nil {
		return errors.Wrap(err, errDeleteIndex)
	}
	return nil
}

// GetIndexConnectionDetails generates connection details
func GetIndexConnectionDetails(cr *aliv1alpha1.LogstoreIndex) managed.ConnectionDetails {
	cd := managed.ConnectionDetails{}
	return cd
}
