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
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/provider-alibaba/apis/oss/v1alpha1"
	ossclient "github.com/crossplane/provider-alibaba/pkg/clients/oss"
)

const (
	errFailedToCreateBucket   = "failed to create OSS bucket"
	errFailedToUpdateBucket   = "failed to update OSS bucket"
	errFailedToDeleteBucket   = "failed to delete OSS bucket"
	errFailedToDescribeBucket = "failed to describe OSS bucket"
)

const errNotBucket = "managed resource is not a Bucket custom resource"

// BaseObserve is the common logic for controller Observe reconciling
func BaseObserve(mg resource.Managed, c ossclient.ClientInterface) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotBucket)
	}

	bucket, err := c.Describe(meta.GetExternalName(cr))
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

// BaseCreate is the logic for Create reconciling
func BaseCreate(mg resource.Managed, c ossclient.ClientInterface) (managed.ExternalCreation, error) {
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
	if err := c.Create(name, bucketParameter); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errFailedToCreateBucket)
	}
	return managed.ExternalCreation{ConnectionDetails: GetConnectionDetails(cr)}, nil
}

// BaseUpdate is the base logic for controller Update reconciling
func BaseUpdate(mg resource.Managed, client ossclient.ClientInterface) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotBucket)
	}
	cr.Status.SetConditions(xpv1.Creating())
	got, err := client.Describe(meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errFailedToDescribeBucket)
	}

	if cr.Spec.ACL != "" && cr.Spec.ACL != got.BucketInfo.ACL {
		if err := client.Update(meta.GetExternalName(cr), cr.Spec.ACL); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errFailedToUpdateBucket)
		}
	}

	return managed.ExternalUpdate{}, nil
}

// BaseDelete is the common logic for controller Delete reconciling
func BaseDelete(mg resource.Managed, client ossclient.ClientInterface) error {
	cr, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return errors.New(errNotBucket)
	}
	cr.SetConditions(xpv1.Deleting())
	if err := client.Delete(meta.GetExternalName(cr)); err != nil && !ossclient.IsNotFoundError(err) {
		return errors.Wrap(err, errFailedToDeleteBucket)
	}
	return nil
}

// BaseSetupOSS is the base logic for controller SetupBucket
func BaseSetupOSS(mgr ctrl.Manager, l logging.Logger, o ...managed.ReconcilerOption) error {
	name := managed.ControllerName(v1alpha1.BucketGroupKind)
	o = append(
		o,
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Bucket{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.BucketGroupVersionKind), o...))
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
