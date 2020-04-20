/*
Copyright 2019 The Crossplane Authors.

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

package database

import (
	"context"

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/source"

	databasev1alpha1 "github.com/crossplane/crossplane/apis/database/v1alpha1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/claimbinding"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/claimdefaulting"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/claimscheduling"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-alibaba/apis/database/v1alpha1"
)

// SetupPostgreSQLInstanceClaimScheduling adds a controller that reconciles
// PostgreSQLInstance claims that include a class selector but omit their class
// and resource references by picking a random matching RDSInstanceClass, if
// any.
func SetupPostgreSQLInstanceClaimScheduling(mgr ctrl.Manager, l logging.Logger) error {
	name := claimscheduling.ControllerName(databasev1alpha1.PostgreSQLInstanceGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&databasev1alpha1.PostgreSQLInstance{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(claimscheduling.NewReconciler(mgr,
			resource.ClaimKind(databasev1alpha1.PostgreSQLInstanceGroupVersionKind),
			resource.ClassKind(v1alpha1.RDSInstanceClassGroupVersionKind),
			claimscheduling.WithLogger(l.WithValues("controller", name)),
			claimscheduling.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

// SetupPostgreSQLInstanceClaimDefaulting adds a controller that reconciles
// PostgreSQLInstance claims that omit their resource ref, class ref, and class
// selector by choosing a default RDSInstanceClass if one exists.
func SetupPostgreSQLInstanceClaimDefaulting(mgr ctrl.Manager, l logging.Logger) error {
	name := claimdefaulting.ControllerName(databasev1alpha1.PostgreSQLInstanceGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&databasev1alpha1.PostgreSQLInstance{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasNoClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(claimdefaulting.NewReconciler(mgr,
			resource.ClaimKind(databasev1alpha1.PostgreSQLInstanceGroupVersionKind),
			resource.ClassKind(v1alpha1.RDSInstanceClassGroupVersionKind),
			claimdefaulting.WithLogger(l.WithValues("controller", name)),
			claimdefaulting.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

// SetupPostgreSQLInstanceClaimBinding adds a controller that reconciles
// PostgreSQLInstance claims with RDSInstances, dynamically provisioning them if
// needed.
func SetupPostgreSQLInstanceClaimBinding(mgr ctrl.Manager, l logging.Logger) error {
	name := claimbinding.ControllerName(databasev1alpha1.PostgreSQLInstanceGroupKind)

	r := claimbinding.NewReconciler(mgr,
		resource.ClaimKind(databasev1alpha1.PostgreSQLInstanceGroupVersionKind),
		resource.ClassKind(v1alpha1.RDSInstanceClassGroupVersionKind),
		resource.ManagedKind(v1alpha1.RDSInstanceGroupVersionKind),
		claimbinding.WithManagedConfigurators(
			claimbinding.ManagedConfiguratorFn(ConfigurePostgreRDSInstance),
			claimbinding.ManagedConfiguratorFn(claimbinding.ConfigureReclaimPolicy),
			claimbinding.ManagedConfiguratorFn(claimbinding.ConfigureNames)),
		claimbinding.WithLogger(l.WithValues("controller", name)),
		claimbinding.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	)

	p := resource.NewPredicates(resource.AnyOf(
		resource.HasClassReferenceKind(resource.ClassKind(v1alpha1.RDSInstanceClassGroupVersionKind)),
		resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1alpha1.RDSInstanceGroupVersionKind)),
		resource.IsManagedKind(resource.ManagedKind(v1alpha1.RDSInstanceGroupVersionKind), mgr.GetScheme()),
	))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &v1alpha1.RDSInstance{}}, &resource.EnqueueRequestForClaim{}).
		For(&databasev1alpha1.PostgreSQLInstance{}).
		WithEventFilter(p).
		Complete(r)
}

// ConfigurePostgreRDSInstance configures the supplied resource (presumed
// to be a RDSInstance) using the supplied resource claim (presumed to be a
// PostgreSQLInstance) and resource class.
func ConfigurePostgreRDSInstance(_ context.Context, cm resource.Claim, cs resource.Class, mg resource.Managed) error {
	_, cmok := cm.(*databasev1alpha1.PostgreSQLInstance)
	if !cmok {
		return errors.Errorf("expected resource claim %s to be %s", cm.GetName(), databasev1alpha1.PostgreSQLInstanceGroupVersionKind)
	}

	rs, csok := cs.(*v1alpha1.RDSInstanceClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1alpha1.RDSInstanceClassGroupVersionKind)
	}

	i, mgok := mg.(*v1alpha1.RDSInstance)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1alpha1.RDSInstanceGroupVersionKind)
	}

	spec := &v1alpha1.RDSInstanceSpec{
		ResourceSpec: runtimev1alpha1.ResourceSpec{
			ReclaimPolicy: runtimev1alpha1.ReclaimRetain,
		},
		ForProvider: rs.SpecTemplate.ForProvider,
	}
	spec.ForProvider.Engine = v1alpha1.PostgresqlEngine

	spec.WriteConnectionSecretToReference = &runtimev1alpha1.SecretReference{
		Namespace: rs.SpecTemplate.WriteConnectionSecretsToNamespace,
		Name:      string(cm.GetUID()),
	}
	spec.ProviderReference = rs.SpecTemplate.ProviderReference
	spec.ReclaimPolicy = rs.SpecTemplate.ReclaimPolicy

	i.Spec = *spec

	return nil
}
