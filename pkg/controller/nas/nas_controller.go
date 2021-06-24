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
	errCreateClient             = "cannot create NAS client"
	errTrackUsage               = "cannot track provider config usage"
	errNoConnectionSecret       = "no connection secret specified"
	errGetConnectionSecret      = "cannot get connection secret"
	errFmtUnsupportedCredSource = "credentials source %q is not currently supported"
)

const (
	errFailedToCreateNASFileSystem   = "failed to create NAS filesystem"
	errFailedToDeleteNASFileSystem   = "failed to delete NAS filesystem"
	errFailedToDescribeNASFileSystem = "failed to describe NAS filesystem"
	errNotNASFileSystem              = "managed resource is not a NASFileSystem custom resource"
)

// SetupNASFileSystem adds a controller that reconciles NASFileSystem.
func SetupNASFileSystem(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.NASFileSystemGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.NASFileSystem{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.NASFileSystemGroupVersionKind),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithExternalConnecter(&Connector{
				Client:      mgr.GetClient(),
				Usage:       resource.NewProviderConfigUsageTracker(mgr.GetClient(), &aliv1alpha1.ProviderConfigUsage{}),
				NewClientFn: nasclient.NewClient,
			})))
}

// Connector stores Kubernetes client and NAS client
type Connector struct {
	Client      client.Client
	Usage       resource.Tracker
	NewClientFn func(ctx context.Context, endpoint, accessKeyID, accessKeySecret, stsToken string) (*nasclient.SDKClient, error)
}

// Connect initials cloud resource client
func (c *Connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.NASFileSystem)
	if !ok {
		return nil, errors.New(errNotNASFileSystem)
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
	return &External{ExternalClient: client}, errors.Wrap(err, errCreateClient)
}

// External includes external NAS client
type External struct {
	ExternalClient nasclient.ClientInterface
}

// Observe managed resource NAS filesystem
func (e *External) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.NASFileSystem)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotNASFileSystem)
	}

	fsID := cr.Status.AtProvider.FileSystemID
	if fsID == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	filesystem, err := e.ExternalClient.DescribeFileSystems(&fsID, cr.Spec.FileSystemType, cr.Spec.VpcID)
	if err != nil {
		// Managed resource `NASFileSystem` is special, the identifier of if `name` is different to the cloud resource identifier `FileSystemID`
		if nasclient.IsNotFoundError(err) {
			return managed.ExternalObservation{ResourceExists: false, ResourceUpToDate: true}, nil
		}
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	cr.Status.AtProvider = nasclient.GenerateObservation(filesystem)
	var upToDate = nasclient.IsUpdateToDate(cr, filesystem)
	if upToDate {
		cr.SetConditions(xpv1.Available())
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: GetConnectionDetails(cr),
	}, nil
}

// Create managed resource NASFilesystem
func (e *External) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.NASFileSystem)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotNASFileSystem)
	}
	cr.SetConditions(xpv1.Creating())
	filesystemParameter := v1alpha1.NASFileSystemParameter{
		FileSystemType: cr.Spec.FileSystemType,
		ChargeType:     cr.Spec.ChargeType,
		StorageType:    cr.Spec.StorageType,
		ProtocolType:   cr.Spec.ProtocolType,
		VpcID:          cr.Spec.VpcID,
		VSwitchID:      cr.Spec.VSwitchID,
	}
	res, err := e.ExternalClient.CreateFileSystem(filesystemParameter)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errFailedToCreateNASFileSystem)
	}
	fsRes, err := e.ExternalClient.DescribeFileSystems(res.Body.FileSystemId, cr.Spec.FileSystemType, cr.Spec.VpcID)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errFailedToDescribeNASFileSystem)
	}
	cr.Status.AtProvider = nasclient.GenerateObservation(fsRes)
	return managed.ExternalCreation{ConnectionDetails: GetConnectionDetails(cr)}, nil
}

// Update managed resource NASFilesystem
func (e *External) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

// Delete managed resource NASFilesystem
func (e *External) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.NASFileSystem)
	if !ok {
		return errors.New(errNotNASFileSystem)
	}
	cr.SetConditions(xpv1.Deleting())
	if err := e.ExternalClient.DeleteFileSystem(cr.Status.AtProvider.FileSystemID); err != nil {
		return errors.Wrap(err, errFailedToDeleteNASFileSystem)
	}
	return nil
}

// GetConnectionDetails generates connection details
func GetConnectionDetails(cr *v1alpha1.NASFileSystem) managed.ConnectionDetails {
	cd := managed.ConnectionDetails{
		"MountTargetDomain": []byte(cr.Status.AtProvider.MountTargetDomain),
	}
	if cr.Spec.FileSystemType != nil {
		cd["FileSystemType"] = []byte(*cr.Spec.FileSystemType)
	}
	return cd
}
