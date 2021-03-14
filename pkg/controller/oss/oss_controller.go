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

package oss

import (
	"context"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/provider-alibaba/apis/oss/v1alpha1"
	aliv1alpha1 "github.com/crossplane/provider-alibaba/apis/v1alpha1"
	ossclient "github.com/crossplane/provider-alibaba/pkg/clients/oss"
	"github.com/crossplane/provider-alibaba/pkg/util"
)

const (
	errNotOSS                   = "managed resource is not an Bucket custom resource"
	errCreateBucket             = "cannot create Bucket bucket"
	errNoProvider               = "no provider config or provider specified"
	errTrackUsage               = "cannot track provider config usage"
	errNoConnectionSecret       = "no connection secret specified"
	errGetConnectionSecret      = "cannot get connection secret"
	errFmtUnsupportedCredSource = "credentials source %q is not currently supported"
)

// SetupBucket adds a controller that reconciles Bucket.
func SetupBucket(mgr ctrl.Manager, l logging.Logger) error {
	options := []managed.ReconcilerOption{managed.WithExternalConnecter(&Connector{
		Client:      mgr.GetClient(),
		Usage:       resource.NewProviderConfigUsageTracker(mgr.GetClient(), &aliv1alpha1.ProviderConfigUsage{}),
		NewClientFn: ossclient.NewClient,
	})}

	return BaseSetupOSS(mgr, l, options...)
}

// Connector stores Kubernetes client and oss client
type Connector struct {
	Client      client.Client
	Usage       resource.Tracker
	NewClientFn func(ctx context.Context, endpoint, accessKeyID, accessKeySecret, stsToken string) (*ossclient.SDKClient, error)
}

// Connect initials cloud resource client
func (c *Connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return nil, errors.New(errNotOSS)
	}

	var (
		secretKeySelector *xpv1.SecretKeySelector
		region            string
	)

	switch {
	case cr.GetProviderConfigReference() != nil:
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
	default:
		return nil, errors.New(errNoProvider)
	}

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

	ossClient, err := c.NewClientFn(ctx, endpoint, string(s.Data["accessKeyId"]), string(s.Data["accessKeySecret"]), "")
	return &external{client: ossClient}, errors.Wrap(err, errCreateBucket)
}

type external struct {
	client ossclient.ClientInterface
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	return BaseObserve(mg, e.client)
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	return BaseCreate(mg, e.client)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return BaseUpdate(mg, e.client)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	return BaseDelete(mg, e.client)
}
