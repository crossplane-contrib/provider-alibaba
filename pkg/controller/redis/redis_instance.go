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

package redis

import (
	"context"
	"time"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-alibaba/apis/redis/v1alpha1"
	aliv1beta1 "github.com/crossplane-contrib/provider-alibaba/apis/v1beta1"
	"github.com/crossplane-contrib/provider-alibaba/pkg/clients/redis"
)

const (
	controllerName = "AliCloud Controller"
	// Fall to connection instance error description
	errCreateInstanceConnectionFailed = "cannot instance connection"

	errNotInstance               = "managed resource is not an instance custom resource"
	errNoProvider                = "no provider config or provider specified"
	errCreateClient              = "cannot create redis client"
	errGetProviderConfig         = "cannot get provider config"
	errTrackUsage                = "cannot track provider config usage"
	errNoConnectionSecret        = "no connection secret specified"
	errGetConnectionSecret       = "cannot get connection secret"
	errInstanceIdEmpty           = "instance id is empty, maybe it's not created"
	errInstanceAlreadyCreated    = "instance id is not empty, maybe it's already been created"
	errStatusUpdate              = "failed to update CR status"
	errInstanceUpdateIPFailed    = "failed to update instance security IP list"
	errInstanceUpdateSpecFailed  = "failed to update instance spec"
	errInstanceUpdateParamFailed = "failed to update instance parameters"

	errCreateFailed        = "cannot create redis instance"
	errCreateAccountFailed = "cannot create redis account"
	errDeleteFailed        = "cannot delete redis instance"
	errDescribeFailed      = "cannot describe redis instance"
	errAccountNameInvalid  = "instance name is invalid"

	errFmtUnsupportedCredSource = "credentials source %q is not currently supported"
	errDuplicateConnectionPort  = "InvalidConnectionStringOrPort.Duplicate"
	errAccountNameDuplicate     = "InvalidAccountName.Duplicate"
)

// SetupRedisInstance adds a controller that reconciles RedisInstances.
func SetupRedisInstance(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.RedisInstanceGroupKind)

	connector := &redisConnector{
		kubeClient:     mgr.GetClient(),
		usage:          resource.NewProviderConfigUsageTracker(mgr.GetClient(), &aliv1beta1.ProviderConfigUsage{}),
		newRedisClient: redis.NewClient,
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithExternalConnecter(connector),
		managed.WithLogger(l.WithValues(controllerName, name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		// Use empty initializer to make sure the externalName is not initialized
		// externalName is going to be updated after a successful creation
		managed.WithInitializers(),
		managed.WithCreationGracePeriod(3 * time.Minute),
	}

	r := managed.NewReconciler(
		mgr,
		resource.ManagedKind(v1alpha1.RedisInstanceGroupVersionKind),
		reconcilerOpts...,
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.RedisInstance{}).
		Complete(r)
}

type redisConnector struct {
	kubeClient     client.Client
	usage          resource.Tracker
	newRedisClient func(ctx context.Context, accessKeyID, accessKeySecret, region string) (redis.Client, error)
}

func (c *redisConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) { //nolint:gocyclo
	// account for the deprecated Provider type.
	cr, ok := mg.(*v1alpha1.RedisInstance)
	if !ok {
		return nil, errors.New(errNotInstance)
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

		pc := &aliv1beta1.ProviderConfig{}
		if err := c.kubeClient.Get(ctx, types.NamespacedName{Name: cr.Spec.ProviderConfigReference.Name}, pc); err != nil {
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
	if err := c.kubeClient.Get(ctx, nn, s); err != nil {
		return nil, errors.Wrap(err, errGetConnectionSecret)
	}

	redisClient, err := c.newRedisClient(ctx, string(s.Data["accessKeyId"]), string(s.Data["accessKeySecret"]), region)
	return &external{redisClient: redisClient, kubeClient: c.kubeClient}, errors.Wrap(err, errCreateClient)
}

type external struct {
	redisClient redis.Client
	kubeClient  client.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.RedisInstance)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotInstance)
	}

	instanceId := meta.GetExternalName(cr)
	// externalName will not be initialized until a successful creation
	if instanceId == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	resp, conn, err := e.redisClient.DescribeInstance(instanceId)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(redis.IsErrorNotFound, err), errDescribeFailed)
	}

	cr.Status.AtProvider = redis.GenerateObservation(resp)

	switch cr.Status.AtProvider.InstanceStatus {
	case v1alpha1.RedisInstanceStateRunning:
		cr.Status.SetConditions(xpv1.Available())
	case v1alpha1.RedisInstanceStateCreating:
		cr.Status.SetConditions(xpv1.Creating())
	case v1alpha1.RedisInstanceStateDeleting:
		cr.Status.SetConditions(xpv1.Deleting())
	default:
		cr.Status.SetConditions(xpv1.Unavailable())
	}

	var securityIpsNeedUpdate bool
	var specsNeedUpdate bool
	var paramsNeedUpdate bool
	if cr.Status.AtProvider.InstanceStatus == v1alpha1.RedisInstanceStateRunning {
		// Check for Spec update
		specsNeedUpdate = redis.SpecsNeedUpdate(resp, &cr.Spec.ForProvider) != nil

		// TODO: Check for Sharding update

		// TODO: Check for Version update

		// Check for parameter update
		paramsUpdateRequest, err := redis.ParamsNeedUpdate(resp, &cr.Spec.ForProvider)
		if err != nil {
			return managed.ExternalObservation{}, err
		}
		paramsNeedUpdate = paramsUpdateRequest != nil

		// Check for whitelist IPs update
		securityIpsNeedUpdate = redis.SecurityIpsNeedUpdate(resp, &cr.Spec.ForProvider) != nil
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  !securityIpsNeedUpdate && !specsNeedUpdate && !paramsNeedUpdate,
		ConnectionDetails: getConnectionDetails(conn),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.RedisInstance)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotInstance)
	}

	cr.Status.SetConditions(xpv1.Creating())
	if err := e.kubeClient.Status().Update(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errStatusUpdate)
	}

	resp, conn, err := e.redisClient.CreateInstance(cr.GetObjectMeta().GetName(), &cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	meta.SetExternalName(cr, resp.InstanceId)

	// Any connection details emitted in ExternalClient are cumulative.
	return managed.ExternalCreation{ConnectionDetails: getConnectionDetails(conn)}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.RedisInstance)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotInstance)
	}

	instanceId := meta.GetExternalName(cr)

	resp, _, err := e.redisClient.DescribeInstance(instanceId)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	// Only update the instance when it's in normal state
	if resp.InstanceStatus != v1alpha1.RedisInstanceStateRunning {
		return managed.ExternalUpdate{}, nil
	}

	// Check and update spec if needed
	specsUpdateRequest := redis.SpecsNeedUpdate(resp, &cr.Spec.ForProvider)
	if specsUpdateRequest != nil {
		err = e.redisClient.ModifyInstanceSpecs(instanceId, specsUpdateRequest)
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errInstanceUpdateSpecFailed)
		}
		return managed.ExternalUpdate{}, nil
	}

	// Check and update parameters if needed
	paramsUpdateRequest, err := redis.ParamsNeedUpdate(resp, &cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	if paramsUpdateRequest != nil {
		err = e.redisClient.ModifyInstanceParams(instanceId, paramsUpdateRequest)
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errInstanceUpdateParamFailed)
		}
		return managed.ExternalUpdate{}, nil
	}

	// Check and update security IPs if needed
	securityIpsUpdate := redis.SecurityIpsNeedUpdate(resp, &cr.Spec.ForProvider)
	if securityIpsUpdate != nil {
		err = e.redisClient.ModifySecurityIps(instanceId, securityIpsUpdate)
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errInstanceUpdateIPFailed)
		}
		return managed.ExternalUpdate{}, nil
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.RedisInstance)
	if !ok {
		return errors.New(errNotInstance)
	}

	cr.SetConditions(xpv1.Deleting())

	err := e.redisClient.DeleteInstance(meta.GetExternalName(cr))
	return errors.Wrap(resource.Ignore(redis.IsErrorNotFound, err), errDeleteFailed)
}

func getConnectionDetails(c *redis.RedisConnection) managed.ConnectionDetails {
	cd := managed.ConnectionDetails{}

	if c.Username != "" {
		cd[xpv1.ResourceCredentialsSecretUserKey] = []byte(c.Username)
	}

	if c.Password != "" {
		cd[xpv1.ResourceCredentialsSecretPasswordKey] = []byte(c.Password)
	}

	if c.ConnectionDomain != "" {
		cd[xpv1.ResourceCredentialsSecretEndpointKey] = []byte(c.ConnectionDomain)
	}

	if c.Port != "" {
		cd[xpv1.ResourceCredentialsSecretPortKey] = []byte(c.Port)
	}

	return cd
}
