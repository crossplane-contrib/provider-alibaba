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
	"errors"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/provider-alibaba/apis/oss/v1alpha1"
	ossclient "github.com/crossplane/provider-alibaba/pkg/clients/oss"
)

// BaseObserve is the common logic for controller Observe reconciling
func BaseObserve(mg resource.Managed, c ossclient.ClientInterface) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotOSS)
	}
	bucketSpec := cr.Spec.ForProvider.Bucket
	klog.InfoS("observing Bucket resource", "Name", bucketSpec.Name)

	bucket, err := c.Describe(bucketSpec.Name)
	if ossclient.IsNotFoundError(err) {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	if err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.Status.AtProvider = ossclient.GenerateObservation(*bucket)
	if bucketSpec.StorageClass != "" && bucketSpec.StorageClass != bucket.BucketInfo.StorageClass {
		cr.Status.AtProvider.Message += "[Warning] StorageClass is not allowed to update after creation; "
	}
	if bucketSpec.DataRedundancyType != "" && bucketSpec.DataRedundancyType != bucket.BucketInfo.RedundancyType {
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
		return managed.ExternalCreation{}, errors.New(errNotOSS)
	}
	klog.InfoS("creating Bucket resource", "Name", cr.Spec.ForProvider.Bucket.Name)
	cr.SetConditions(xpv1.Creating())
	if err := c.Create(cr.Spec.ForProvider.Bucket); err != nil {
		return managed.ExternalCreation{}, err
	}
	return managed.ExternalCreation{ConnectionDetails: GetConnectionDetails(cr)}, nil
}

// BaseUpdate is the base logic for controller Update reconciling
func BaseUpdate(mg resource.Managed, client ossclient.ClientInterface) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotOSS)
	}
	klog.InfoS("updating Bucket resource", "Name", cr.Spec.ForProvider.Bucket.Name)
	cr.Status.SetConditions(xpv1.Creating())
	got, err := client.Describe(cr.Spec.ForProvider.Bucket.Name)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	target := cr.Spec.ForProvider.Bucket

	if target.ACL != "" && target.ACL != got.BucketInfo.ACL {
		if err := client.Update(target.Name, target.ACL); err != nil {
			return managed.ExternalUpdate{}, err
		}
	}

	return managed.ExternalUpdate{}, nil
}

// BaseDelete is the common logic for controller Delete reconciling
func BaseDelete(mg resource.Managed, client ossclient.ClientInterface) error {
	cr, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return errors.New(errNotOSS)
	}
	klog.InfoS("deleting Bucket resource", "Name", cr.Spec.ForProvider.Bucket.Name)
	cr.SetConditions(xpv1.Deleting())
	if err := client.Delete(cr.Spec.ForProvider.Bucket.Name); err != nil && !ossclient.IsNotFoundError(err) {
		return err
	}
	return nil
}

// BaseSetupOSS is the base logic for controller SetupBucket
func BaseSetupOSS(mgr ctrl.Manager, l logging.Logger, o ...managed.ReconcilerOption) error {
	name := managed.ControllerName(v1alpha1.OSSGroupKind)
	o = append(
		o,
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Bucket{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.OSSGroupVersionKind), o...))
}

// GetConnectionDetails generates connection details
func GetConnectionDetails(cr *v1alpha1.Bucket) managed.ConnectionDetails {
	cd := managed.ConnectionDetails{
		"Bucket": []byte(cr.Spec.ForProvider.Bucket.Name),
	}
	if cr.Status.AtProvider.ExtranetEndpoint != "" {
		cd["ExtranetEndpoint"] = []byte(cr.Status.AtProvider.ExtranetEndpoint)
	}
	if cr.Status.AtProvider.IntranetEndpoint != "" {
		cd["IntranetEndpoint"] = []byte(cr.Status.AtProvider.IntranetEndpoint)
	}
	return cd
}
