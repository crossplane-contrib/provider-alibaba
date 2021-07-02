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
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/provider-alibaba/apis/database/v1alpha1"
	aliv1beta1 "github.com/crossplane/provider-alibaba/apis/v1beta1"
	"github.com/crossplane/provider-alibaba/pkg/clients/rds"
	"github.com/crossplane/provider-alibaba/pkg/util"
)

const (
	errNotRDSInstance           = "managed resource is not an RDS instance custom resource"
	errCreateRDSClient          = "cannot create RDS client"
	errTrackUsage               = "cannot track provider config usage"
	errCreateFailed             = "cannot create RDS instance"
	errCreateAccountFailed      = "cannot create RDS database account"
	errDeleteFailed             = "cannot delete RDS instance"
	errDescribeFailed           = "cannot describe RDS instance"
	errFmtUnsupportedCredSource = "no extraction handler registered for source: %s"
	errGetCredentials           = "cannot get credentials"
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
				client:       mgr.GetClient(),
				usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &aliv1beta1.ProviderConfigUsage{}),
				newRDSClient: rds.NewClient,
			}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	client       client.Client
	usage        resource.Tracker
	newRDSClient func(ctx context.Context, accessKeyID, accessKeySecret, securityToken, region string) (rds.Client, error)
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) { //nolint:gocyclo
	// TODO(negz): This connect method will be simpler once we no longer have to
	// account for the deprecated Provider type.
	cr, ok := mg.(*v1alpha1.RDSInstance)
	if !ok {
		return nil, errors.New(errNotRDSInstance)
	}

	clientEstablishmentInfo, err := util.PrepareClient(ctx, mg, cr, c.client, c.usage, cr.Spec.ProviderConfigReference.Name)
	if err != nil {
		return nil, err
	}

	rdsClient, err := c.newRDSClient(ctx, clientEstablishmentInfo.AccessKeyID, clientEstablishmentInfo.AccessKeySecret,
		clientEstablishmentInfo.SecurityToken, clientEstablishmentInfo.Region)
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
		cr.Status.SetConditions(xpv1.Available())
		pw, err = e.createAccountIfNeeded(cr)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errCreateAccountFailed)
		}
	default:
		cr.Status.SetConditions(xpv1.Unavailable())
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  true,
		ConnectionDetails: getConnectionDetails(pw, cr, instance),
	}, nil
}

func (e *external) createAccountIfNeeded(cr *v1alpha1.RDSInstance) (string, error) {
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

	// Any connection details emitted in ExternalClient are cumulative.
	return managed.ExternalCreation{ConnectionDetails: getConnectionDetails("", cr, instance)}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.RDSInstance)
	if !ok {
		return errors.New(errNotRDSInstance)
	}
	if cr.Status.AtProvider.DBInstanceStatus == v1alpha1.RDSInstanceStateDeleting {
		return nil
	}

	err := e.client.DeleteDBInstance(cr.Status.AtProvider.DBInstanceID)
	return errors.Wrap(resource.Ignore(rds.IsErrorNotFound, err), errDeleteFailed)
}

func getConnectionDetails(password string, cr *v1alpha1.RDSInstance, instance *rds.DBInstance) managed.ConnectionDetails {
	cd := managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretUserKey: []byte(cr.Spec.ForProvider.MasterUsername),
	}

	if password != "" {
		cd[xpv1.ResourceCredentialsSecretPasswordKey] = []byte(password)
	}

	if instance.Endpoint != nil {
		cd[xpv1.ResourceCredentialsSecretEndpointKey] = []byte(instance.Endpoint.Address)
		cd[xpv1.ResourceCredentialsSecretPortKey] = []byte(instance.Endpoint.Port)
	}

	return cd
}
