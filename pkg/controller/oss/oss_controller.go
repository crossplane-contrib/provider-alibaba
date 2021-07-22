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

package oss

import (
	"context"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/provider-alibaba/apis/oss/v1alpha1"
	aliv1beta1 "github.com/crossplane/provider-alibaba/apis/v1beta1"
	ossclient "github.com/crossplane/provider-alibaba/pkg/clients/oss"
	"github.com/crossplane/provider-alibaba/pkg/util"
)

const (
	errCreateClient           = "cannot create OSS client"
	errFailedToCreateBucket   = "failed to create OSS bucket"
	errFailedToUpdateBucket   = "failed to update OSS bucket"
	errFailedToDeleteBucket   = "failed to delete OSS bucket"
	errFailedToDescribeBucket = "failed to describe OSS bucket"
	errNotBucket              = "managed resource is not a Bucket custom resource"
)

// SetupBucket adds a controller that reconciles Bucket.
func SetupBucket(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.BucketGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Bucket{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.BucketGroupVersionKind),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithExternalConnecter(&Connector{
				Client:      mgr.GetClient(),
				Usage:       resource.NewProviderConfigUsageTracker(mgr.GetClient(), &aliv1beta1.ProviderConfigUsage{}),
				NewClientFn: ossclient.NewClient,
			})))
}

// Connector stores Kubernetes client and oss client
type Connector struct {
	Client      client.Client
	Usage       resource.Tracker
	NewClientFn func(ctx context.Context, endpoint, accessKeyID, accessKeySecret, stsToken string) (*ossclient.SDKClient, error)
}

// Connect initials cloud resource client
func (c *Connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return nil, errors.New(errNotBucket)
	}

	info, err := util.PrepareClient(ctx, mg, cr, c.Client, c.Usage, cr.Spec.ProviderConfigReference.Name)
	if err != nil {
		return nil, err
	}

	ossClient, err := c.NewClientFn(ctx, info.Endpoint, info.AccessKeyID, info.AccessKeySecret, info.SecurityToken)
	return &External{ExternalClient: ossClient}, errors.Wrap(err, errCreateClient)
}

// External includes external OSS client
type External struct {
	ExternalClient ossclient.ClientInterface
}

// Observe managed resource OSS bucket
func (e *External) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotBucket)
	}

	bucket, err := e.ExternalClient.Describe(meta.GetExternalName(cr))
	if ossclient.IsNotFoundError(err) {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errFailedToDescribeBucket)
	}

	cr.Status.AtProvider = ossclient.GenerateObservation(*bucket)
	if cr.Spec.StorageClass != "" && cr.Spec.StorageClass != bucket.BucketInfo.StorageClass {
		cr.Status.AtProvider.Message += "[Warning] StorageClass is not allowed to update after creation; "
	}
	if cr.Spec.DataRedundancyType != "" && cr.Spec.DataRedundancyType != bucket.BucketInfo.RedundancyType {
		cr.Status.AtProvider.Message += "[Warning] DataRedundancyType is not allowed to update after creation; "
	}
	var upToDate = ossclient.IsUpdateToDate(cr, bucket)
	if upToDate {
		cr.SetConditions(xpv1.Available())
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: GetConnectionDetails(cr),
	}, nil
}

// Create managed resource OSS bucket
func (e *External) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotBucket)
	}
	cr.SetConditions(xpv1.Creating())
	bucketParameter := v1alpha1.BucketParameter{
		ACL:                cr.Spec.ACL,
		StorageClass:       cr.Spec.StorageClass,
		DataRedundancyType: cr.Spec.DataRedundancyType,
	}
	name := meta.GetExternalName(cr)
	if err := e.ExternalClient.Create(name, bucketParameter); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errFailedToCreateBucket)
	}
	return managed.ExternalCreation{ConnectionDetails: GetConnectionDetails(cr)}, nil
}

// Update managed resource OSS bucket
func (e *External) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotBucket)
	}
	cr.Status.SetConditions(xpv1.Creating())
	got, err := e.ExternalClient.Describe(meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errFailedToDescribeBucket)
	}

	if cr.Spec.ACL != "" && cr.Spec.ACL != got.BucketInfo.ACL {
		if err := e.ExternalClient.Update(meta.GetExternalName(cr), cr.Spec.ACL); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errFailedToUpdateBucket)
		}
	}

	return managed.ExternalUpdate{}, nil
}

// Delete managed resource OSS bucket
func (e *External) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return errors.New(errNotBucket)
	}
	cr.SetConditions(xpv1.Deleting())
	if err := e.ExternalClient.Delete(meta.GetExternalName(cr)); err != nil && !ossclient.IsNotFoundError(err) {
		return errors.Wrap(err, errFailedToDeleteBucket)
	}
	return nil
}

// GetConnectionDetails generates connection details
func GetConnectionDetails(cr *v1alpha1.Bucket) managed.ConnectionDetails {
	cd := managed.ConnectionDetails{
		"Bucket": []byte(meta.GetExternalName(cr)),
	}
	if cr.Status.AtProvider.ExtranetEndpoint != "" {
		cd["ExtranetEndpoint"] = []byte(cr.Status.AtProvider.ExtranetEndpoint)
	}
	if cr.Status.AtProvider.IntranetEndpoint != "" {
		cd["IntranetEndpoint"] = []byte(cr.Status.AtProvider.IntranetEndpoint)
	}
	return cd
}
