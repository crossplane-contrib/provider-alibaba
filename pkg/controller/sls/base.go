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
	sdk "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/provider-alibaba/apis/sls/v1alpha1"
	slsclient "github.com/crossplane/provider-alibaba/pkg/clients/sls"
)

// BaseObserve is the common logic for controller Observe reconciling
func BaseObserve(mg resource.Managed, c slsclient.LogClientInterface) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotProject)
	}
	projectName := cr.Spec.ForProvider.ProjectName
	project, err := c.Describe(projectName)
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
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: getConnectionDetails(project),
	}, nil
}

// BaseCreate is the logic for Create reconciling
func BaseCreate(mg resource.Managed, c slsclient.LogClientInterface) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotProject)
	}
	name := cr.Spec.ForProvider.ProjectName
	description := cr.Spec.ForProvider.Description
	project, err := c.Create(name, description)
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	return managed.ExternalCreation{ConnectionDetails: getConnectionDetails(project)}, nil
}

// BaseUpdate is the base logic for controller Update reconciling
func BaseUpdate(mg resource.Managed, client slsclient.LogClientInterface) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotProject)
	}
	name := cr.Spec.ForProvider.ProjectName
	description := cr.Spec.ForProvider.Description
	got, err := client.Describe(name)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	if got.Description != description {
		if err != nil {
			return managed.ExternalUpdate{}, err
		}
	}
	return managed.ExternalUpdate{}, nil
}

// BaseDelete is the common logic for controller Delete reconciling
func BaseDelete(mg resource.Managed, client slsclient.LogClientInterface) error {
	cr, ok := mg.(*v1alpha1.Project)
	if !ok {
		return errors.New(errNotProject)
	}
	name := cr.Spec.ForProvider.ProjectName
	if err := client.Delete(name); err != nil && !slsclient.IsNotFoundError(err) {
		return err
	}
	return nil
}

// BaseSetupSLSProject is the base logic for controller SetupSLSProject
func BaseSetupSLSProject(mgr ctrl.Manager, l logging.Logger, o ...managed.ReconcilerOption) error {
	name := managed.ControllerName(v1alpha1.SLSProjectGroupKind)
	o = append(
		o,
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Project{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.SLSProjectGroupVersionKind), o...))
}

func getConnectionDetails(project *sdk.LogProject) managed.ConnectionDetails {
	cd := managed.ConnectionDetails{
		"Name":     []byte(project.Name),
		"Endpoint": []byte(project.Endpoint),
	}
	return cd
}
