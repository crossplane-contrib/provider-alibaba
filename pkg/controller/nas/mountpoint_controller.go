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

package nas

import (
	"context"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/provider-alibaba/apis/nas/v1alpha1"
	aliv1alpha1 "github.com/crossplane/provider-alibaba/apis/v1alpha1"
	nasclient "github.com/crossplane/provider-alibaba/pkg/clients/nas"
	"github.com/crossplane/provider-alibaba/pkg/util"
)

const (
	errFailedToCreateNASMountTarget   = "failed to create NAS filesystem"
	errFailedToDeleteNASMountTarget   = "failed to delete NAS filesystem"
	errFailedToDescribeNASMountTarget = "failed to describe NAS filesystem"
	errNotNASMountTarget              = "managed resource is not a NASMountTarget custom resource"
)

// SetupNASMountTarget adds a controller that reconciles NASMountTarget.
func SetupNASMountTarget(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.NASMountTargetGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.NASMountTarget{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.NASMountTargetGroupVersionKind),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithExternalConnecter(&mtConnector{
				Client:      mgr.GetClient(),
				Usage:       resource.NewProviderConfigUsageTracker(mgr.GetClient(), &aliv1alpha1.ProviderConfigUsage{}),
				NewClientFn: nasclient.NewClient,
			})))
}

// mtConnector stores Kubernetes client and NAS client
type mtConnector struct {
	Client      client.Client
	Usage       resource.Tracker
	NewClientFn func(ctx context.Context, endpoint, accessKeyID, accessKeySecret, stsToken string) (*nasclient.SDKClient, error)
}

// Connect initials cloud resource client
func (c *mtConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.NASMountTarget)
	if !ok {
		return nil, errors.New(errNotNASMountTarget)
	}

	var (
		secretKeySelector *xpv1.SecretKeySelector
		region            string
	)

	if err := c.Usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackUsage)
	}

	providerConfig, err := util.GetProviderConfig(ctx, c.Client, cr.Spec.ProviderConfigReference.Name)
	if err != nil {
		return nil, err
	}

	if s := providerConfig.Spec.Credentials.Source; s != xpv1.CredentialsSourceSecret {
		return nil, errors.Errorf(errFmtUnsupportedCredSource, s)
	}
	secretKeySelector = providerConfig.Spec.Credentials.SecretRef
	region = providerConfig.Spec.Region

	if secretKeySelector == nil {
		return nil, errors.New(errNoConnectionSecret)
	}
	s := &corev1.Secret{}
	nn := types.NamespacedName{Namespace: secretKeySelector.Namespace, Name: secretKeySelector.Name}
	if err := c.Client.Get(ctx, nn, s); err != nil {
		return nil, errors.Wrap(err, errGetConnectionSecret)
	}

	endpoint, err := util.GetEndpoint(cr.DeepCopyObject(), region)
	if err != nil {
		return nil, err
	}

	client, err := c.NewClientFn(ctx, endpoint, string(s.Data[util.AccessKeyID]), string(s.Data[util.AccessKeySecret]), string(s.Data[util.SecurityToken]))
	return &mountTargetExternal{ExternalClient: client}, errors.Wrap(err, errCreateClient)
}

// mountTargetExternal includes external NAS client
type mountTargetExternal struct {
	ExternalClient nasclient.ClientInterface
}

// Observe managed resource NAS filesystem
func (e *mountTargetExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.NASMountTarget)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotNASMountTarget)
	}

	domain := cr.Status.AtProvider.MountTargetDomain
	if domain == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	mountTarget, err := e.ExternalClient.DescribeMountTargets(cr.Spec.ForProvider.FileSystemID, domain)
	if err != nil {
		// Managed resource `NASMountTarget` is special, the identifier of if `name` is different to the cloud resource identifier `MountTargetDomain`
		if nasclient.IsMountTargetNotFoundError(err) {
			return managed.ExternalObservation{ResourceExists: false, ResourceUpToDate: true}, nil
		}
		return managed.ExternalObservation{ResourceExists: false}, errors.Wrap(err, errFailedToDescribeNASMountTarget)
	}

	var upToDate = nasclient.IsMountTargetUpdateToDate(cr, mountTarget)
	if upToDate {
		cr.SetConditions(xpv1.Available())
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: GetMountTargetConnectionDetails(cr),
	}, nil
}

// Create managed resource NASFilesystem
func (e *mountTargetExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.NASMountTarget)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotNASMountTarget)
	}
	cr.SetConditions(xpv1.Creating())
	res, err := e.ExternalClient.CreateMountTarget(cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errFailedToCreateNASMountTarget)
	}
	cr.Status.AtProvider = nasclient.GenerateObservation4MountTarget(res)
	return managed.ExternalCreation{ConnectionDetails: GetMountTargetConnectionDetails(cr)}, nil
}

// Update managed resource NASFilesystem
func (e *mountTargetExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

// Delete managed resource NASFilesystem
func (e *mountTargetExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.NASMountTarget)
	if !ok {
		return errors.New(errNotNASMountTarget)
	}
	cr.SetConditions(xpv1.Deleting())
	if err := e.ExternalClient.DeleteMountTarget(cr.Spec.ForProvider.FileSystemID, cr.Status.AtProvider.MountTargetDomain); err != nil {
		return errors.Wrap(err, errFailedToDeleteNASMountTarget)
	}
	return nil
}

// GetMountTargetConnectionDetails generates connection details
func GetMountTargetConnectionDetails(cr *v1alpha1.NASMountTarget) managed.ConnectionDetails {
	cd := managed.ConnectionDetails{}

	if cr.Status.AtProvider.MountTargetDomain != nil {
		cd["MountTargetDomain"] = []byte(*cr.Status.AtProvider.MountTargetDomain)
	}
	return cd
}
