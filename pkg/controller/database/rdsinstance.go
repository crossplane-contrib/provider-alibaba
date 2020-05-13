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

	sdkerror "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
)

const (
	errNotRDSInstance = "managed resource is not an RDS instance custom resource"

	errCreateRDSClient   = "cannot create RDS client"
	errGetProvider       = "cannot get provider"
	errGetProviderSecret = "cannot get provider secret"

	errCreateFailed        = "cannot create RDS instance"
	errCreateAccountFailed = "cannot create RDS database account"
	errDeleteFailed        = "cannot delete RDS instance"
	errDescribeFailed      = "cannot describe RDS instance"
)

// SetupRDSInstance adds a controller that reconciles RDSInstances.
func SetupRDSInstance(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.RDSInstanceGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.RDSInstance{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.RDSInstanceGroupVersionKind),
			managed.WithExternalConnecter(&connector{
				reader:       &kube2ConnectorReader{kube: mgr.GetClient()},
				newRDSClient: rds.NewClient,
			}),
			managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	reader       connectorReader
	newRDSClient func(ctx context.Context, accessKeyID, accessKeySecret, region string) (rds.Client, error)
}

type kube2ConnectorReader struct {
	kube client.Client
}

func (kr *kube2ConnectorReader) GetProvider(ctx context.Context, key client.ObjectKey) (*aliv1alpha1.Provider, error) {
	obj := &aliv1alpha1.Provider{}
	err := kr.kube.Get(ctx, key, obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (kr *kube2ConnectorReader) GetSecret(ctx context.Context, key client.ObjectKey) (*corev1.Secret, error) {
	obj := &corev1.Secret{}
	err := kr.kube.Get(ctx, key, obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

type connectorReader interface {
	GetProvider(ctx context.Context, key client.ObjectKey) (*aliv1alpha1.Provider, error)
	GetSecret(ctx context.Context, key client.ObjectKey) (*corev1.Secret, error)
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.RDSInstance)
	if !ok {
		return nil, errors.New(errNotRDSInstance)
	}

	p, err := c.reader.GetProvider(ctx, meta.NamespacedNameOf(cr.Spec.ProviderReference))
	if err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errGetProviderSecret)
	}

	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	s, err := c.reader.GetSecret(ctx, n)
	if err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}

	rdsClient, err := c.newRDSClient(ctx, string(s.Data["accessKeyId"]), string(s.Data["accessKeySecret"]), p.Spec.Region)
	return &external{client: rdsClient}, errors.Wrap(err, errCreateRDSClient)
}

type external struct {
	client rds.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.RDSInstance)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRDSInstance)
	}

	if cr.Status.AtProvider.DBInstanceID == "" {
		return managed.ExternalObservation{}, nil
	}

	instance, err := e.client.DescribeDBInstance(cr.Status.AtProvider.DBInstanceID)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(rds.IsErrorNotFound, err), errDescribeFailed)
	}

	cr.Status.AtProvider = rds.GenerateObservation(instance)

	var pw string
	switch cr.Status.AtProvider.DBInstanceStatus {
	case v1alpha1.RDSInstanceStateRunning:
		cr.Status.SetConditions(runtimev1alpha1.Available())
		resource.SetBindable(cr)
		pw, err = e.createAccountIfneeded(cr)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errCreateAccountFailed)
		}
	case v1alpha1.RDSInstanceStateCreating:
		cr.Status.SetConditions(runtimev1alpha1.Creating())
	case v1alpha1.RDSInstanceStateDeleting:
		cr.Status.SetConditions(runtimev1alpha1.Deleting())
	default:
		cr.Status.SetConditions(runtimev1alpha1.Unavailable())
	}

	ob := managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}

	if pw != "" {
		ob.ConnectionDetails = managed.ConnectionDetails{
			runtimev1alpha1.ResourceCredentialsSecretUserKey:     []byte(cr.Spec.ForProvider.MasterUsername),
			runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(pw),
		}
	}
	return ob, nil
}

func (e *external) createAccountIfneeded(cr *v1alpha1.RDSInstance) (string, error) {
	if cr.Status.AtProvider.AccountReady {
		return "", nil
	}
	pw, err := password.Generate()
	if err != nil {
		return "", err
	}
	err = e.client.CreateAccount(cr.Status.AtProvider.DBInstanceID, cr.Spec.ForProvider.MasterUsername, pw)
	if err != nil {
		// The previous request might fail due to timeout. That's fine we will eventually reconcile it.
		if sdkErr, ok := err.(sdkerror.Error); ok {
			if sdkErr.ErrorCode() == "InvalidAccountName.Duplicate" {
				cr.Status.AtProvider.AccountReady = true
				return "", nil
			}
		}
		return "", err
	}
	cr.Status.AtProvider.AccountReady = true
	return pw, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.RDSInstance)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRDSInstance)
	}

	cr.SetConditions(runtimev1alpha1.Creating())
	if cr.Status.AtProvider.DBInstanceStatus == v1alpha1.RDSInstanceStateCreating {
		return managed.ExternalCreation{}, nil
	}

	req := rds.MakeCreateDBInstanceRequest(meta.GetExternalName(cr), &cr.Spec.ForProvider)
	instance, err := e.client.CreateDBInstance(req)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	// The crossplane runtime will send status update back to apiserver.
	cr.Status.AtProvider.DBInstanceID = instance.ID

	// Need to handle DB Account (username and password) in another resource.
	conn := managed.ConnectionDetails{
		runtimev1alpha1.ResourceCredentialsSecretEndpointKey: []byte(instance.Endpoint.Address),
		runtimev1alpha1.ResourceCredentialsSecretPortKey:     []byte(instance.Endpoint.Port),
	}

	// Any connection details emitted in ExternalClient are cumulative.
	return managed.ExternalCreation{ConnectionDetails: conn}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.RDSInstance)
	if !ok {
		return errors.New(errNotRDSInstance)
	}
	cr.SetConditions(runtimev1alpha1.Deleting())
	if cr.Status.AtProvider.DBInstanceStatus == v1alpha1.RDSInstanceStateDeleting {
		return nil
	}

	err := e.client.DeleteDBInstance(cr.Status.AtProvider.DBInstanceID)
	return errors.Wrap(resource.Ignore(rds.IsErrorNotFound, err), errDeleteFailed)
}
