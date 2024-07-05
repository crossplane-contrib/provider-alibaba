package redis

import (
	"context"
	"strconv"
	"testing"

	aliredis "github.com/aliyun/alibaba-cloud-sdk-go/services/r-kvstore"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	crossplanemeta "github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1" //nolint:typecheck
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-alibaba/apis/redis/v1alpha1"
	aliv1beta1 "github.com/crossplane-contrib/provider-alibaba/apis/v1beta1"
	"github.com/crossplane-contrib/provider-alibaba/pkg/clients/redis"
)

const testId = "testId"
const testName = "testName"
const testStatus = "testEndpoint"
const testPassword = "testPassword"

const testPort = "8080"
const testAddress = "172.0.0.1"

var testPortInt = 8080

func TestConnector(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		client         client.Client
		usage          resource.Tracker
		newRedisClient func(ctx context.Context, accessKeyID, accessKeySecret, region string) (redis.Client, error)
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   error
	}{
		"NotRedisInstance": {
			reason: "Should return an error if the supplied managed resource is not an RedisInstance",
			args: args{
				mg: nil,
			},
			want: errors.New(errNotInstance),
		},
		"TrackProviderConfigUsageError": {
			reason: "Errors tracking a ProviderConfigUsage should be returned",
			fields: fields{
				usage: resource.TrackerFn(func(ctx context.Context, mg resource.Managed) error { return errBoom }),
			},
			args: args{
				mg: &v1alpha1.RedisInstance{
					Spec: v1alpha1.RedisInstanceSpec{
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{},
						},
					},
				},
			},
			want: errors.Wrap(errBoom, errTrackUsage),
		},
		"GetProviderConfigError": {
			reason: "Errors getting a ProviderConfig should be returned",
			fields: fields{
				client: &test.MockClient{
					MockGet: test.NewMockGetFn(errBoom),
				},
				usage: resource.TrackerFn(func(ctx context.Context, mg resource.Managed) error { return nil }),
			},
			args: args{
				mg: &v1alpha1.RedisInstance{
					Spec: v1alpha1.RedisInstanceSpec{
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{},
						},
					},
				},
			},
			want: errors.Wrap(errBoom, errGetProviderConfig),
		},
		"UnsupportedCredentialsError": {
			reason: "An error should be returned if the selected credentials source is unsupported",
			fields: fields{
				client: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						t := obj.(*aliv1beta1.ProviderConfig)
						*t = aliv1beta1.ProviderConfig{
							Spec: aliv1beta1.ProviderConfigSpec{
								Credentials: aliv1beta1.ProviderCredentials{
									Source: xpv1.CredentialsSource("wat"),
								},
							},
						}
						return nil
					}),
				},
				usage: resource.TrackerFn(func(ctx context.Context, mg resource.Managed) error { return nil }),
			},
			args: args{
				mg: &v1alpha1.RedisInstance{
					Spec: v1alpha1.RedisInstanceSpec{
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{},
						},
					},
				},
			},
			want: errors.Errorf(errFmtUnsupportedCredSource, "wat"),
		},
		"GetProviderError": {
			reason: "Errors getting a Provider should be returned",
			fields: fields{
				client: &test.MockClient{
					MockGet: test.NewMockGetFn(errBoom),
				},
				usage: resource.TrackerFn(func(ctx context.Context, mg resource.Managed) error { return nil }),
			},
			args: args{
				mg: &v1alpha1.RedisInstance{
					Spec: v1alpha1.RedisInstanceSpec{
						ResourceSpec: xpv1.ResourceSpec{
							ProviderReference: &xpv1.Reference{},
						},
					},
				},
			},
			want: errors.New(errNoProvider),
		},
		"NoConnectionSecretError": {
			reason: "An error should be returned if no connection secret was specified",
			fields: fields{
				client: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						t := obj.(*aliv1beta1.ProviderConfig)
						*t = aliv1beta1.ProviderConfig{
							Spec: aliv1beta1.ProviderConfigSpec{
								Credentials: aliv1beta1.ProviderCredentials{
									Source: xpv1.CredentialsSourceSecret,
								},
							},
						}
						return nil
					}),
				},
				usage: resource.TrackerFn(func(ctx context.Context, mg resource.Managed) error { return nil }),
			},
			args: args{
				mg: &v1alpha1.RedisInstance{
					Spec: v1alpha1.RedisInstanceSpec{
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{},
						},
					},
				},
			},
			want: errors.New(errNoConnectionSecret),
		},
		"GetConnectionSecretError": {
			reason: "Errors getting a secret should be returned",
			fields: fields{
				client: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						switch t := obj.(type) {
						case *corev1.Secret:
							return errBoom
						case *aliv1beta1.ProviderConfig:
							*t = aliv1beta1.ProviderConfig{
								Spec: aliv1beta1.ProviderConfigSpec{
									Credentials: aliv1beta1.ProviderCredentials{
										Source: xpv1.CredentialsSourceSecret,
									},
								},
							}
							t.Spec.Credentials.SecretRef = &xpv1.SecretKeySelector{
								SecretReference: xpv1.SecretReference{
									Name: "coolsecret",
								},
							}
						}
						return nil
					}),
				},
				usage: resource.TrackerFn(func(ctx context.Context, mg resource.Managed) error { return nil }),
			},
			args: args{
				mg: &v1alpha1.RedisInstance{
					Spec: v1alpha1.RedisInstanceSpec{
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{},
						},
					},
				},
			},
			want: errors.Wrap(errBoom, errGetConnectionSecret),
		},
		"NewRedisClientError": {
			reason: "Errors getting a secret should be returned",
			fields: fields{
				client: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						if t, ok := obj.(*aliv1beta1.ProviderConfig); ok {
							*t = aliv1beta1.ProviderConfig{
								Spec: aliv1beta1.ProviderConfigSpec{
									Credentials: aliv1beta1.ProviderCredentials{
										Source: xpv1.CredentialsSourceSecret,
									},
								},
							}
							t.Spec.Credentials.SecretRef = &xpv1.SecretKeySelector{
								SecretReference: xpv1.SecretReference{
									Name: "coolsecret",
								},
							}
						}
						return nil
					}),
				},
				usage: resource.TrackerFn(func(ctx context.Context, mg resource.Managed) error { return nil }),
				newRedisClient: func(ctx context.Context, accessKeyID, accessKeySecret, region string) (redis.Client, error) {
					return nil, errBoom
				},
			},
			args: args{
				mg: &v1alpha1.RedisInstance{
					Spec: v1alpha1.RedisInstanceSpec{
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{},
						},
					},
				},
			},
			want: errors.Wrap(errBoom, errCreateClient),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &redisConnector{kubeClient: tc.fields.client, usage: tc.fields.usage, newRedisClient: tc.fields.newRedisClient}
			_, err := c.Connect(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nc.Connect(...) -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	e := &external{redisClient: &fakeRedisClient{}}
	type want struct {
		ResourceExists   bool
		ResourceUpToDate bool
		err              error
	}

	cases := map[string]struct {
		mg   resource.Managed
		want want
	}{
		"InstancePort is not set": {
			mg: &v1alpha1.RedisInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name: testName,
					Annotations: map[string]string{
						crossplanemeta.AnnotationKeyExternalName: testId,
					},
				},
				Spec: v1alpha1.RedisInstanceSpec{
					ForProvider: v1alpha1.RedisInstanceParameters{},
				},
				Status: v1alpha1.RedisInstanceStatus{
					AtProvider: v1alpha1.RedisInstanceObservation{
						InstanceStatus:   testStatus,
						ConnectionDomain: testAddress,
						Port:             testPort,
					},
				},
			},
			want: want{
				ResourceExists: true, ResourceUpToDate: true, err: nil,
			},
		},
		"InstancePort is set": {
			mg: &v1alpha1.RedisInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name: testName,
					Annotations: map[string]string{
						crossplanemeta.AnnotationKeyExternalName: testId,
					},
				},
				Spec: v1alpha1.RedisInstanceSpec{
					ForProvider: v1alpha1.RedisInstanceParameters{
						Port: &testPortInt,
					},
				},
				Status: v1alpha1.RedisInstanceStatus{
					AtProvider: v1alpha1.RedisInstanceObservation{
						InstanceStatus:   testStatus,
						ConnectionDomain: testAddress,
						Port:             testPort,
					},
				},
			},
			want: want{
				ResourceExists: true, ResourceUpToDate: true, err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := e.Observe(context.Background(), tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", name, diff)
			}
			if diff := cmp.Diff(tc.want.ResourceUpToDate, got.ResourceUpToDate); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", name, diff)
			}
			if diff := cmp.Diff(tc.want.ResourceExists, got.ResourceExists); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", name, diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	e := &external{redisClient: &fakeRedisClient{}, kubeClient: &test.MockClient{
		MockStatusUpdate: test.NewMockStatusUpdateFn(nil, func(obj client.Object) error {
			return nil
		}),
	}}

	type want struct {
		u   managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		mg   resource.Managed
		want want
	}{
		"No a valid managed resource": {
			mg: nil,
			want: want{
				u: managed.ExternalCreation{}, err: errors.New(errNotInstance),
			},
		},
		"Successfully create a managed resource": {
			mg: &v1alpha1.RedisInstance{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						crossplanemeta.AnnotationKeyExternalName: testId,
					},
					Name: testName,
				},
				Spec: v1alpha1.RedisInstanceSpec{
					ForProvider: v1alpha1.RedisInstanceParameters{
						EngineVersion: "5.0",
						InstanceClass: "redis.logic.sharding.2g.8db.0rodb.8proxy.default",
						Port:          &testPortInt,
						Password:      testPassword,
						// PubliclyAccessible: true,
					},
				},
			},
			want: want{
				u: managed.ExternalCreation{
					ConnectionDetails: map[string][]byte{
						"username": []byte(testId),
						"password": []byte(testPassword),
						"endpoint": []byte(testAddress),
						"port":     []byte(strconv.Itoa(testPortInt)),
					}},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := e.Create(context.Background(), tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Update(...): -want error, +got error:\n%s\n", name, diff)
			}
			if diff := cmp.Diff(tc.want.u, got); diff != "" {
				t.Errorf("\n%s\ne.Update(...): -want, +got:\n%s\n", name, diff)
			}
		})
	}

}

func TestUpdate(t *testing.T) {
	e := &external{redisClient: &fakeRedisClient{}}
	type want struct {
		u   managed.ExternalUpdate
		err error
	}

	cases := map[string]struct {
		mg   resource.Managed
		want want
	}{
		"No a valid managed resource": {
			mg: nil,
			want: want{
				u: managed.ExternalUpdate{}, err: errors.New(errNotInstance),
			},
		},
		"Successfully update a managed resource": {
			mg: &v1alpha1.RedisInstance{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{crossplanemeta.AnnotationKeyExternalName: testId},
				},
				Spec: v1alpha1.RedisInstanceSpec{
					ForProvider: v1alpha1.RedisInstanceParameters{
						InstanceClass: "class-test",
					},
				},
			},
			want: want{
				u: managed.ExternalUpdate{}, err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := e.Update(context.Background(), tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Update(...): -want error, +got error:\n%s\n", name, diff)
			}
			if diff := cmp.Diff(tc.want.u, got); diff != "" {
				t.Errorf("\n%s\ne.Update(...): -want, +got:\n%s\n", name, diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	e := &external{redisClient: &fakeRedisClient{}}
	type want struct {
		err error
	}

	cases := map[string]struct {
		mg   resource.Managed
		want want
	}{
		"No a valid managed resource": {
			mg: nil,
			want: want{
				err: errors.New(errNotInstance),
			},
		},
		"Managed resource is already in a delete state": {
			mg: &v1alpha1.RedisInstance{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{crossplanemeta.AnnotationKeyExternalName: testId},
				},
				Status: v1alpha1.RedisInstanceStatus{
					AtProvider: v1alpha1.RedisInstanceObservation{
						InstanceStatus:   v1alpha1.RedisInstanceStateDeleting,
						ConnectionDomain: testAddress,
						Port:             testPort,
					},
				},
			},
			want: want{
				err: nil,
			},
		},
		"Successfully delete a managed resource": {
			mg: &v1alpha1.RedisInstance{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{crossplanemeta.AnnotationKeyExternalName: testId},
				},
				Status: v1alpha1.RedisInstanceStatus{
					AtProvider: v1alpha1.RedisInstanceObservation{
						InstanceStatus:   testStatus,
						ConnectionDomain: testAddress,
						Port:             testPort,
					},
				},
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := e.Delete(context.Background(), tc.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Update(...): -want error, +got error:\n%s\n", name, diff)
			}
		})
	}
}

func TestGetConnectionDetails(t *testing.T) {
	address := testAddress
	port := testPort
	password := testPassword

	type args struct {
		id     string
		pw     string
		domain string
		port   string
	}
	type want struct {
		conn managed.ConnectionDetails
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"SuccessfulNoPassword": {
			args: args{
				id:     testId,
				pw:     "",
				domain: address,
				port:   port,
			},
			want: want{
				conn: managed.ConnectionDetails{
					xpv1.ResourceCredentialsSecretUserKey:     []byte(testId),
					xpv1.ResourceCredentialsSecretEndpointKey: []byte(address),
					xpv1.ResourceCredentialsSecretPortKey:     []byte(port),
				},
			},
		},
		"SuccessfulNoEndpoint": {
			args: args{
				id: testId,
				pw: password,
			},
			want: want{
				conn: managed.ConnectionDetails{
					xpv1.ResourceCredentialsSecretUserKey:     []byte(testId),
					xpv1.ResourceCredentialsSecretPasswordKey: []byte(password),
				},
			},
		},
		"Successful": {
			args: args{
				id:     testId,
				pw:     password,
				domain: address,
				port:   port,
			},
			want: want{
				conn: managed.ConnectionDetails{
					xpv1.ResourceCredentialsSecretUserKey:     []byte(testId),
					xpv1.ResourceCredentialsSecretPasswordKey: []byte(password),
					xpv1.ResourceCredentialsSecretEndpointKey: []byte(address),
					xpv1.ResourceCredentialsSecretPortKey:     []byte(port),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			conn := getConnectionDetails(&redis.RedisConnection{
				Username:         tc.args.id,
				Password:         tc.args.pw,
				ConnectionDomain: tc.args.domain,
				Port:             tc.args.port,
			})
			if diff := cmp.Diff(tc.want.conn, conn); diff != "" {
				t.Errorf("getConnectionDetails(...): -want, +got:\n%s", diff)
			}
		})
	}
}

type fakeRedisClient struct{}

func (c *fakeRedisClient) DescribeInstance(id string) (*aliredis.DBInstanceAttribute, *redis.RedisConnection, error) {
	if id != testId {
		return nil, nil, errors.New("DescribeRedisInstance: client doesn't work")
	}
	return &aliredis.DBInstanceAttribute{
			InstanceId:       id,
			InstanceStatus:   v1alpha1.RedisInstanceStateRunning,
			ConnectionDomain: testAddress,
			Port:             int64(testPortInt),
		}, &redis.RedisConnection{
			ConnectionDomain: testAddress,
			Port:             testPort,
		}, nil
}

func (c *fakeRedisClient) CreateInstance(instanceName string, p *v1alpha1.RedisInstanceParameters) (*aliredis.CreateInstanceResponse, *redis.RedisConnection, error) {
	if instanceName != testName {
		return nil, nil, errors.New("CreateRedisInstance: client doesn't work")
	}
	return &aliredis.CreateInstanceResponse{
			InstanceId:       testId,
			InstanceName:     instanceName,
			ConnectionDomain: testAddress,
			Port:             *p.Port,
		}, &redis.RedisConnection{
			Username:         testId,
			Password:         p.Password,
			ConnectionDomain: testAddress,
			Port:             strconv.Itoa(*p.Port),
		}, nil
}

func (c *fakeRedisClient) DeleteInstance(id string) error {
	if id != testId {
		return errors.New("DeleteRedisInstance: client doesn't work")
	}
	return nil
}

func (c *fakeRedisClient) ModifySecurityIps(id string, req *aliredis.ModifySecurityIpsRequest) error {
	if id != testId {
		return errors.New("ModifySecurityIps: client doesn't work")
	}
	return nil
}

func (c *fakeRedisClient) ModifyInstanceSpecs(id string, req *aliredis.ModifyInstanceSpecRequest) error {
	if id != testId {
		return errors.New("ModifyInstanceSpecs: client doesn't work")
	}
	return nil
}

func (c *fakeRedisClient) ModifyInstanceParams(id string, req *aliredis.ModifyInstanceConfigRequest) error {
	if id != testId {
		return errors.New("ModifyInstanceParams: client doesn't work")
	}
	return nil
}
