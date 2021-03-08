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
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/provider-alibaba/apis/sls/v1alpha1"
	aliv1alpha1 "github.com/crossplane/provider-alibaba/apis/v1alpha1"
	"github.com/crossplane/provider-alibaba/pkg/clients/sls"
)

const (
	errNotSLSProject            = "managed resource is not an SLS project custom resource"
	errNoProvider               = "no provider config or provider specified"
	errGetProviderConfig        = "cannot get provider config"
	errTrackUsage               = "cannot track provider config usage"
	errNoConnectionSecret       = "no connection secret specified"
	errGetConnectionSecret      = "cannot get connection secret"
	errCreateFailed             = "cannot create SLS project"
	errDeleteFailed             = "cannot delete SLS project"
	errDescribeFailed           = "cannot describe SLS project"
	errFmtUnsupportedCredSource = "credentials source %q is not currently supported"
)

// SetupSLSProject adds a controller that reconciles SLSProjects.
func SetupSLSProject(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.SLSProjectGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.SLSProject{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.SLSProjectGroupVersionKind),
			managed.WithExternalConnecter(&connector{
				client: mgr.GetClient(),
				usage:  resource.NewProviderConfigUsageTracker(mgr.GetClient(), &aliv1alpha1.ProviderConfigUsage{}),
				Client: sls.NewClient,
			}),
			managed.WithLogger(l.WithValues("SLS Project Controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client client.Client
	usage  resource.Tracker
	Client func(accessKeyID, accessKeySecret, region string) sls.LogClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) { //nolint:gocyclo
	cr, ok := mg.(*v1alpha1.SLSProject)
	if !ok {
		return nil, errors.New(errNotSLSProject)
	}

	// provider has more than one kind of managed resource.
	var (
		sel    *xpv1.SecretKeySelector
		region string
	)

	switch {
	case cr.GetProviderConfigReference() != nil:
		if err := c.usage.Track(ctx, mg); err != nil {
			return nil, errors.Wrap(err, errTrackUsage)
		}

		pc := &aliv1alpha1.ProviderConfig{}
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

	slsClient := c.Client(string(s.Data["accessKeyId"]), string(s.Data["accessKeySecret"]), region)
	return &external{client: slsClient}, nil
}

type external struct {
	client sls.LogClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.SLSProject)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSLSProject)
	}

	if cr.Status.AtProvider.Name == "" {
		return managed.ExternalObservation{}, nil
	}

	project, err := e.client.Describe(cr.Status.AtProvider.Name)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(sls.IsErrorNotFound, err), errDescribeFailed)
	}

	cr.Status.AtProvider = sls.GenerateObservation(project)

	var pw string
	switch cr.Status.AtProvider.Status {
	case v1alpha1.SLSProjectStateDeleting:
		cr.Status.SetConditions(xpv1.Deleting())
	default:
		cr.Status.SetConditions(xpv1.Unavailable())
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  true,
		ConnectionDetails: getConnectionDetails(pw, cr),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.SLSProject)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSLSProject)
	}

	cr.SetConditions(xpv1.Creating())
	if cr.Status.AtProvider.Status == v1alpha1.SLSProjectStateCreating {
		return managed.ExternalCreation{}, nil
	}

	req := sls.CreateSLSProjectRequest{Name: cr.Spec.ForProvider.Name, Description: cr.Spec.ForProvider.Description}
	project, err := e.client.Create(req)

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	// The crossplane runtime will send status update back to apiserver.
	cr.Status.AtProvider.Name = project.Name
	cr.Status.AtProvider.Status = project.Status

	// Any connection details emitted in ExternalClient are cumulative.
	return managed.ExternalCreation{ConnectionDetails: getConnectionDetails("", cr)}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.SLSProject)
	if !ok {
		return errors.New(errNotSLSProject)
	}
	cr.SetConditions(xpv1.Deleting())
	if cr.Status.AtProvider.Status == v1alpha1.SLSProjectStateDeleting {
		return nil
	}

	err := e.client.Delete(cr.Status.AtProvider.Name)
	return errors.Wrap(resource.Ignore(sls.IsErrorNotFound, err), errDeleteFailed)
}

func getConnectionDetails(password string, cr *v1alpha1.SLSProject) managed.ConnectionDetails {
	cd := managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretUserKey: []byte(cr.Spec.ForProvider.Name),
	}

	if password != "" {
		cd[xpv1.ResourceCredentialsSecretPasswordKey] = []byte(password)
	}

	return cd
}
