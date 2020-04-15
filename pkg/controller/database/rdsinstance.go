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
	"reflect"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/provider-alibaba/apis/database/v1alpha1"
	aliv1alpha1 "github.com/crossplane/provider-alibaba/apis/v1alpha1"
	"github.com/crossplane/provider-alibaba/pkg/clients/rds"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errNotRDSInstance   = "managed resource is not an RDS instance custom resource"
	errKubeUpdateFailed = "cannot update RDS instance custom resource"

	errCreateRDSClient   = "cannot create RDS client"
	errGetProvider       = "cannot get provider"
	errGetProviderSecret = "cannot get provider secret"

	errCreateFailed   = "cannot create RDS instance"
	errDeleteFailed   = "cannot delete RDS instance"
	errDescribeFailed = "cannot describe RDS instance"
)

// SetupRDSInstance adds a controller that reconciles RDSInstances.
func SetupRDSInstance(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.RDSInstanceGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.RDSInstance{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.RDSInstanceGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), log: l.WithValues("connector", "rds")}),
			managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube client.Client
	log  logging.Logger
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	c.log.Info("Connect", "ns/name", mg.GetNamespace()+"/"+mg.GetName())
	cr, ok := mg.(*v1alpha1.RDSInstance)
	if !ok {
		return nil, errors.New(errNotRDSInstance)
	}

	p := &aliv1alpha1.Provider{}
	err := c.kube.Get(ctx, meta.NamespacedNameOf(cr.Spec.ProviderReference), p)
	if err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errGetProviderSecret)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.kube.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}

	rdsClient, err := rds.NewClient(ctx, string(s.Data["accessKeyId"]), string(s.Data["accessSecret"]), p.Spec.Region)
	return &external{client: rdsClient, kube: c.kube, log: c.log}, errors.Wrap(err, errCreateRDSClient)
}

type external struct {
	client rds.Client
	kube   client.Client
	log    logging.Logger
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.RDSInstance)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRDSInstance)
	}

	instance, err := e.client.DescribeDBInstance(meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(rds.IsErrorNotFound, err), errDescribeFailed)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	rds.LateInitialize(&cr.Spec.ForProvider, instance)
	if !reflect.DeepEqual(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}
	cr.Status.AtProvider = rds.GenerateObservation(instance)

	switch cr.Status.AtProvider.DBInstanceStatus {
	case v1alpha1.RDSInstanceStateRunning:
		cr.Status.SetConditions(runtimev1alpha1.Available())
		resource.SetBindable(cr)
	case v1alpha1.RDSInstanceStateCreating:
		cr.Status.SetConditions(runtimev1alpha1.Creating())
	case v1alpha1.RDSInstanceStateDeleting:
		cr.Status.SetConditions(runtimev1alpha1.Deleting())
	default:
		cr.Status.SetConditions(runtimev1alpha1.Unavailable())
	}

	upToDate := rds.IsUpToDate(cr.Spec.ForProvider, instance)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	e.log.Info("Create RDS", "name", mg.GetName())
	cr, ok := mg.(*v1alpha1.RDSInstance)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRDSInstance)
	}

	cr.SetConditions(runtimev1alpha1.Creating())
	if cr.Status.AtProvider.DBInstanceStatus == v1alpha1.RDSInstanceStateCreating {
		return managed.ExternalCreation{}, nil
	}
	pw, err := password.Generate()
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	dbusername := cr.Spec.ForProvider.MasterUsername

	req := rds.MakeCreateDBInstanceRequest(meta.GetExternalName(cr), dbusername, pw, &cr.Spec.ForProvider)
	instance, err := e.client.CreateDBInstance(req)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	// emit username and password
	conn := managed.ConnectionDetails{
		runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(pw),
		runtimev1alpha1.ResourceCredentialsSecretUserKey:     []byte(dbusername),
		runtimev1alpha1.ResourceCredentialsSecretEndpointKey: []byte(instance.Endpoint.Address),
		runtimev1alpha1.ResourceCredentialsSecretPortKey:     []byte(instance.Endpoint.Port),
	}

	// Any connection details emitted in ExternalClient are cumulative.
	return managed.ExternalCreation{ConnectionDetails: conn}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	e.log.Info("Update RDS not implemented", "name", mg.GetName())
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	e.log.Info("Delete RDS", "name", mg.GetName())
	cr, ok := mg.(*v1alpha1.RDSInstance)
	if !ok {
		return errors.New(errNotRDSInstance)
	}
	cr.SetConditions(runtimev1alpha1.Deleting())
	if cr.Status.AtProvider.DBInstanceStatus == v1alpha1.RDSInstanceStateDeleting {
		return nil
	}

	err := e.client.DeleteDBInstance(meta.GetExternalName(cr))
	return errors.Wrap(resource.Ignore(rds.IsErrorNotFound, err), errDeleteFailed)
}
