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

	slsv1alpha1 "github.com/crossplane/provider-alibaba/apis/sls/v1alpha1"
	"github.com/crossplane/provider-alibaba/apis/v1alpha1"
	slsclient "github.com/crossplane/provider-alibaba/pkg/clients/sls"
	"github.com/crossplane/provider-alibaba/pkg/util"
)

const (
	errNotStore               = "managed resource is not a SLS store custom resource"
	errMaxSplitShardMustBeSet = "maxSplitShard must be set if autoSplit is true"
)

// SetupStore adds a controller that reconciles SLSStores.
func SetupStore(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(slsv1alpha1.StoreGroupKind)
	options := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&logStoreConnector{
			client:      mgr.GetClient(),
			usage:       resource.NewProviderConfigUsageTracker(mgr.GetClient(), &v1alpha1.ProviderConfigUsage{}),
			NewClientFn: slsclient.NewClient,
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&slsv1alpha1.Store{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(slsv1alpha1.StoreGroupVersionKind), options...))
}

type logStoreConnector struct {
	client      client.Client
	usage       resource.Tracker
	NewClientFn func(accessKeyID, accessKeySecret, securityToken, region string) *slsclient.LogClient
}

func (c *logStoreConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) { //nolint:gocyclo
	cr, ok := mg.(*slsv1alpha1.Store)
	if !ok {
		return nil, errors.New(errNotStore)
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
	return &storeExternal{client: slsClient}, nil
}

type storeExternal struct {
	client slsclient.LogClientInterface
}

func (e *storeExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*slsv1alpha1.Store)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotStore)
	}

	storeName := meta.GetExternalName(cr)
	project := cr.Spec.ForProvider.ProjectName

	store, err := e.client.DescribeStore(project, storeName)
	if slsclient.IsStoreNotFoundError(err) {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.Status.AtProvider = slsclient.GenerateStoreObservation(store)
	upToDate := slsclient.IsStoreUpdateToDate(cr, store)
	if upToDate {
		cr.SetConditions(xpv1.Available())
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: getStoreConnectionDetails(project, storeName),
	}, nil
}

func (e *storeExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*slsv1alpha1.Store)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotStore)
	}
	name := meta.GetExternalName(cr)
	store := &sdk.LogStore{
		Name:       name,
		TTL:        cr.Spec.ForProvider.TTL,
		ShardCount: cr.Spec.ForProvider.ShardCount,
	}
	if cr.Spec.ForProvider.AutoSplit != nil {
		store.AutoSplit = *cr.Spec.ForProvider.AutoSplit
		if store.AutoSplit && cr.Spec.ForProvider.MaxSplitShard == nil {
			return managed.ExternalCreation{}, errors.New(errMaxSplitShardMustBeSet)
		}
	}
	if cr.Spec.ForProvider.MaxSplitShard != nil {
		store.MaxSplitShard = *cr.Spec.ForProvider.MaxSplitShard
	}
	cr.SetConditions(xpv1.Creating())
	err := e.client.CreateStore(cr.Spec.ForProvider.ProjectName, store)
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	return managed.ExternalCreation{ConnectionDetails: getStoreConnectionDetails(cr.Spec.ForProvider.ProjectName, name)}, nil
}

func (e *storeExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*slsv1alpha1.Store)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotStore)
	}
	cr.Status.SetConditions(xpv1.Creating())
	err := e.client.UpdateStore(cr.Spec.ForProvider.ProjectName, meta.GetExternalName(cr), cr.Spec.ForProvider.TTL)
	return managed.ExternalUpdate{}, err
}

func (e *storeExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*slsv1alpha1.Store)
	if !ok {
		return errors.New(errNotStore)
	}
	cr.SetConditions(xpv1.Deleting())
	return e.client.DeleteStore(cr.Spec.ForProvider.ProjectName, meta.GetExternalName(cr))
}

func getStoreConnectionDetails(project, store string) managed.ConnectionDetails {
	cd := managed.ConnectionDetails{
		"LogStore": []byte(store),
		"Project":  []byte(project),
	}
	return cd
}
