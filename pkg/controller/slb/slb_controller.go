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

package slb

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

	"github.com/crossplane/provider-alibaba/apis/slb/v1alpha1"
	aliv1alpha1 "github.com/crossplane/provider-alibaba/apis/v1alpha1"
	slbclient "github.com/crossplane/provider-alibaba/pkg/clients/slb"
	"github.com/crossplane/provider-alibaba/pkg/util"
)

const (
	errCreateClient             = "cannot create SLB client"
	errTrackUsage               = "cannot track provider config usage"
	errNoConnectionSecret       = "no connection secret specified"
	errGetConnectionSecret      = "cannot get connection secret"
	errFmtUnsupportedCredSource = "credentials source %q is not currently supported"
)

const (
	errFailedToCreateSLB   = "failed to create SLB"
	errFailedToDeleteSLB   = "failed to delete SLB"
	errFailedToDescribeSLB = "failed to describe SLB"
	errNotCLB              = "managed resource is not a CLB custom resource"
)

// SetupCLB adds a controller that reconciles CLB
func SetupCLB(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha1.CLBGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.CLB{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.CLBGroupVersionKind),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithExternalConnecter(&Connector{
				Client:      mgr.GetClient(),
				Usage:       resource.NewProviderConfigUsageTracker(mgr.GetClient(), &aliv1alpha1.ProviderConfigUsage{}),
				NewClientFn: slbclient.NewClient,
			})))
}

// Connector stores Kubernetes client and SLB client
type Connector struct {
	Client      client.Client
	Usage       resource.Tracker
	NewClientFn func(ctx context.Context, endpoint, accessKeyID, accessKeySecret, stsToken string) (*slbclient.SDKClient, error)
}

// Connect initials cloud resource client
func (c *Connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.CLB)
	if !ok {
		return nil, errors.New(errNotCLB)
	}

	var secretKeySelector *xpv1.SecretKeySelector
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

	if secretKeySelector == nil {
		return nil, errors.New(errNoConnectionSecret)
	}
	s := &corev1.Secret{}
	nn := types.NamespacedName{Namespace: secretKeySelector.Namespace, Name: secretKeySelector.Name}
	if err := c.Client.Get(ctx, nn, s); err != nil {
		return nil, errors.Wrap(err, errGetConnectionSecret)
	}

	endpoint, err := util.GetEndpoint(cr.DeepCopyObject(), "")
	if err != nil {
		return nil, err
	}

	client, err := c.NewClientFn(ctx, endpoint, string(s.Data[util.AccessKeyID]), string(s.Data[util.AccessKeySecret]), string(s.Data[util.SecurityToken]))
	return &External{ExternalClient: client}, errors.Wrap(err, errCreateClient)
}

// External includes external SLB client
type External struct {
	ExternalClient slbclient.ClientInterface
}

// Observe managed resource CLB
func (e *External) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.CLB)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCLB)
	}

	lbID := cr.Status.AtProvider.LoadBalancerID
	if lbID == nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	slb, err := e.ExternalClient.DescribeLoadBalancers(cr.Spec.ForProvider.Region, lbID, cr.Spec.ForProvider.VpcID,
		cr.Spec.ForProvider.VSwitchID)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, errors.Wrap(err, errFailedToDescribeSLB)
	}
	if *slb.Body.TotalCount == 0 {
		return managed.ExternalObservation{ResourceExists: false, ResourceUpToDate: true}, nil
	}

	cr.Status.AtProvider = slbclient.GenerateObservation(slb)
	var upToDate = slbclient.IsUpdateToDate(cr, slb)
	if upToDate {
		cr.SetConditions(xpv1.Available())
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: GetConnectionDetails(cr),
	}, nil
}

// Create managed resource CLB
func (e *External) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.CLB)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCLB)
	}
	cr.SetConditions(xpv1.Creating())
	res, err := e.ExternalClient.CreateLoadBalancer(cr.Name, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errFailedToCreateSLB)
	}
	lb, err := e.ExternalClient.DescribeLoadBalancers(cr.Spec.ForProvider.Region, res.Body.LoadBalancerId,
		cr.Spec.ForProvider.VpcID, cr.Spec.ForProvider.VSwitchID)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errFailedToDescribeSLB)
	}
	cr.Status.AtProvider = slbclient.GenerateObservation(lb)
	return managed.ExternalCreation{ConnectionDetails: GetConnectionDetails(cr)}, nil
}

// Update managed resource CLB
func (e *External) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

// Delete managed resource CLB
func (e *External) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.CLB)
	if !ok {
		return errors.New(errNotCLB)
	}
	cr.SetConditions(xpv1.Deleting())
	if err := e.ExternalClient.DeleteLoadBalancer(cr.Spec.ForProvider.Region, cr.Status.AtProvider.LoadBalancerID); err != nil {
		return errors.Wrap(err, errFailedToDeleteSLB)
	}
	return nil
}

// GetConnectionDetails generates connection details
func GetConnectionDetails(cr *v1alpha1.CLB) managed.ConnectionDetails {
	cd := managed.ConnectionDetails{
		"Address":        []byte(*cr.Status.AtProvider.Address),
		"LoadBalancerId": []byte(*cr.Status.AtProvider.LoadBalancerID),
	}
	return cd
}
