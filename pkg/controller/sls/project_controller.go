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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	slsv1alpha1 "github.com/crossplane/provider-alibaba/apis/sls/v1alpha1"
	"github.com/crossplane/provider-alibaba/apis/v1beta1"
	slsclient "github.com/crossplane/provider-alibaba/pkg/clients/sls"
	"github.com/crossplane/provider-alibaba/pkg/util"
)

const errNotProject = "managed resource is not a SLS project custom resource"

// SetupProject adds a controller that reconciles SLSProjects.
func SetupProject(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(slsv1alpha1.ProjectGroupKind)
	options := []managed.ReconcilerOption{
		managed.WithExternalConnecter(&connector{
			client:      mgr.GetClient(),
			usage:       resource.NewProviderConfigUsageTracker(mgr.GetClient(), &v1beta1.ProviderConfigUsage{}),
			NewClientFn: slsclient.NewClient,
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&slsv1alpha1.Project{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(slsv1alpha1.ProjectGroupVersionKind), options...))
}

type connector struct {
	client      client.Client
	usage       resource.Tracker
	NewClientFn func(accessKeyID, accessKeySecret, securityToken, region string) *slsclient.LogClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) { //nolint:gocyclo
	cr, ok := mg.(*slsv1alpha1.Project)
	if !ok {
		return nil, errors.New(errNotProject)
	}

	clientEstablishmentInfo, err := util.PrepareClient(ctx, mg, cr, c.client, c.usage, cr.Spec.ProviderConfigReference.Name)
	if err != nil {
		return nil, err
	}

	slsClient := c.NewClientFn(clientEstablishmentInfo.AccessKeyID, clientEstablishmentInfo.AccessKeySecret,
		clientEstablishmentInfo.SecurityToken, clientEstablishmentInfo.Region)
	return &external{client: slsClient}, nil
}

type external struct {
	client slsclient.LogClientInterface
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*slsv1alpha1.Project)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotProject)
	}
	projectName := meta.GetExternalName(cr)
	project, err := e.client.Describe(projectName)
	if slsclient.IsNotFoundError(err) {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	if err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.Status.AtProvider = slsclient.GenerateObservation(project)
	var upToDate bool
	if (projectName == project.Name) && (cr.Spec.ForProvider.Description == project.Description) {
		upToDate = true
		cr.SetConditions(xpv1.Available())
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: getConnectionDetails(project),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*slsv1alpha1.Project)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotProject)
	}
	name := meta.GetExternalName(cr)
	description := cr.Spec.ForProvider.Description
	project, err := e.client.Create(name, description)
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	return managed.ExternalCreation{ConnectionDetails: getConnectionDetails(project)}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*slsv1alpha1.Project)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotProject)
	}
	name := meta.GetExternalName(cr)
	description := cr.Spec.ForProvider.Description
	got, err := e.client.Update(name, description)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	if got.Description != description {
		return managed.ExternalUpdate{}, err
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*slsv1alpha1.Project)
	if !ok {
		return errors.New(errNotProject)
	}
	name := meta.GetExternalName(cr)
	if err := e.client.Delete(name); err != nil && !slsclient.IsNotFoundError(err) {
		return err
	}
	return nil
}

func getConnectionDetails(project *sdk.LogProject) managed.ConnectionDetails {
	cd := managed.ConnectionDetails{
		"Name":     []byte(project.Name),
		"Endpoint": []byte(project.Endpoint),
	}
	return cd
}
