package redis

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	crossplanemeta "github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-alibaba/apis/redis/v1alpha1"
	aliv1alpha1 "github.com/crossplane/provider-alibaba/apis/v1alpha1"
	"github.com/crossplane/provider-alibaba/pkg/clients/redis"
)

const testName = "test"

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
					MockGet: test.NewMockGetFn(nil, func(obj runtime.Object) error {
						t := obj.(*aliv1alpha1.ProviderConfig)
						*t = aliv1alpha1.ProviderConfig{
							Spec: aliv1alpha1.ProviderConfigSpec{
								ProviderConfigSpec: xpv1.ProviderConfigSpec{
									Credentials: xpv1.ProviderCredentials{
										Source: xpv1.CredentialsSource("wat"),
									},
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
			want: errors.Wrap(errBoom, errGetProvider),
		},
		"NoConnectionSecretError": {
			reason: "An error should be returned if no connection secret was specified",
			fields: fields{
				client: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj runtime.Object) error {
						t := obj.(*aliv1alpha1.ProviderConfig)
						*t = aliv1alpha1.ProviderConfig{
							Spec: aliv1alpha1.ProviderConfigSpec{
								ProviderConfigSpec: xpv1.ProviderConfigSpec{
									Credentials: xpv1.ProviderCredentials{
										Source: xpv1.CredentialsSourceSecret,
									},
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
					MockGet: test.NewMockGetFn(nil, func(obj runtime.Object) error {
						switch t := obj.(type) {
						case *corev1.Secret:
							return errBoom
						case *aliv1alpha1.ProviderConfig:
							*t = aliv1alpha1.ProviderConfig{
								Spec: aliv1alpha1.ProviderConfigSpec{
									ProviderConfigSpec: xpv1.ProviderConfigSpec{
										Credentials: xpv1.ProviderCredentials{
											Source: xpv1.CredentialsSourceSecret,
											SecretRef: &xpv1.SecretKeySelector{
												SecretReference: xpv1.SecretReference{
													Name: "coolsecret",
												},
											},
										},
									},
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
					MockGet: test.NewMockGetFn(nil, func(obj runtime.Object) error {
						if t, ok := obj.(*aliv1alpha1.ProviderConfig); ok {
							*t = aliv1alpha1.ProviderConfig{
								Spec: aliv1alpha1.ProviderConfigSpec{
									ProviderConfigSpec: xpv1.ProviderConfigSpec{
										Credentials: xpv1.ProviderCredentials{
											Source: xpv1.CredentialsSourceSecret,
											SecretRef: &xpv1.SecretKeySelector{
												SecretReference: xpv1.SecretReference{
													Name: "coolsecret",
												},
											},
										},
									},
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
			c := &redisConnector{client: tc.fields.client, usage: tc.fields.usage, newRedisClient: tc.fields.newRedisClient}
			_, err := c.Connect(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nc.Connect(...) -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestExternalClientObserve(t *testing.T) {
	e := &external{client: &fakeRedisClient{}}
	obj := &v1alpha1.RedisInstance{
		Spec: v1alpha1.RedisInstanceSpec{
			ForProvider: v1alpha1.RedisInstanceParameters{
				MasterUsername: testName,
			},
		},
		Status: v1alpha1.RedisInstanceStatus{
			AtProvider: v1alpha1.RedisInstanceObservation{
				DBInstanceID: testName,
			},
		},
	}
	ob, err := e.Observe(context.Background(), obj)
	if err != nil {
		t.Fatal(err)
	}
	if obj.Status.AtProvider.DBInstanceStatus != v1alpha1.RedisInstanceStateRunning {
		t.Errorf("DBInstanceStatus (%v) should be %v", obj.Status.AtProvider.DBInstanceStatus, v1alpha1.RedisInstanceStateRunning)
	}
	if obj.Status.AtProvider.AccountReady != true {
		t.Error("AccountReady should be true")
	}
	if string(ob.ConnectionDetails[xpv1.ResourceCredentialsSecretUserKey]) != testName {
		t.Error("ConnectionDetails should include username=test")
	}
}

func TestExternalClientCreate(t *testing.T) {
	e := &external{client: &fakeRedisClient{}}
	obj := &v1alpha1.RedisInstance{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				crossplanemeta.AnnotationKeyExternalName: testName,
			},
		},
		Spec: v1alpha1.RedisInstanceSpec{
			ForProvider: v1alpha1.RedisInstanceParameters{
				MasterUsername:     testName,
				EngineVersion:      "5.0",
				InstanceClass:      "redis.logic.sharding.2g.8db.0rodb.8proxy.default",
				InstancePort:       8080,
				PubliclyAccessible: true,
			},
		},
	}
	ob, err := e.Create(context.Background(), obj)
	if err != nil {
		t.Fatal(err)
	}
	if obj.Status.AtProvider.DBInstanceID != testName {
		t.Error("DBInstanceID should be set to 'test'")
	}
	if string(ob.ConnectionDetails[xpv1.ResourceCredentialsSecretEndpointKey]) != "172.0.0.1" ||
		string(ob.ConnectionDetails[xpv1.ResourceCredentialsSecretPortKey]) != "8888" {
		t.Error("ConnectionDetails should include endpoint=172.0.0.1 and port=8888")
	}
}

func TestExternalClientDelete(t *testing.T) {
	e := &external{client: &fakeRedisClient{}}
	obj := &v1alpha1.RedisInstance{
		Status: v1alpha1.RedisInstanceStatus{
			AtProvider: v1alpha1.RedisInstanceObservation{
				DBInstanceID: testName,
			},
		},
	}
	err := e.Delete(context.Background(), obj)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetConnectionDetails(t *testing.T) {
	address := "0.0.0.0"
	port := "3346"
	password := "super-secret"

	type args struct {
		pw string
		cr *v1alpha1.RedisInstance
		i  *redis.DBInstance
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
				pw: "",
				cr: &v1alpha1.RedisInstance{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							crossplanemeta.AnnotationKeyExternalName: testName,
						},
					},
					Spec: v1alpha1.RedisInstanceSpec{
						ForProvider: v1alpha1.RedisInstanceParameters{
							MasterUsername: testName,
						},
					},
				},
				i: &redis.DBInstance{
					Endpoint: &v1alpha1.Endpoint{
						Address: address,
						Port:    port,
					},
				},
			},
			want: want{
				conn: managed.ConnectionDetails{
					xpv1.ResourceCredentialsSecretUserKey:     []byte(testName),
					xpv1.ResourceCredentialsSecretEndpointKey: []byte(address),
					xpv1.ResourceCredentialsSecretPortKey:     []byte(port),
				},
			},
		},
		"SuccessfulNoEndpoint": {
			args: args{
				pw: password,
				cr: &v1alpha1.RedisInstance{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							crossplanemeta.AnnotationKeyExternalName: testName,
						},
					},
					Spec: v1alpha1.RedisInstanceSpec{
						ForProvider: v1alpha1.RedisInstanceParameters{
							MasterUsername: testName,
						},
					},
				},
				i: &redis.DBInstance{},
			},
			want: want{
				conn: managed.ConnectionDetails{
					xpv1.ResourceCredentialsSecretUserKey:     []byte(testName),
					xpv1.ResourceCredentialsSecretPasswordKey: []byte(password),
				},
			},
		},
		"Successful": {
			args: args{
				pw: password,
				cr: &v1alpha1.RedisInstance{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							crossplanemeta.AnnotationKeyExternalName: testName,
						},
					},
					Spec: v1alpha1.RedisInstanceSpec{
						ForProvider: v1alpha1.RedisInstanceParameters{
							MasterUsername: testName,
						},
					},
				},
				i: &redis.DBInstance{
					Endpoint: &v1alpha1.Endpoint{
						Address: address,
						Port:    port,
					},
				},
			},
			want: want{
				conn: managed.ConnectionDetails{
					xpv1.ResourceCredentialsSecretUserKey:     []byte(testName),
					xpv1.ResourceCredentialsSecretPasswordKey: []byte(password),
					xpv1.ResourceCredentialsSecretEndpointKey: []byte(address),
					xpv1.ResourceCredentialsSecretPortKey:     []byte(port),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			conn := getConnectionDetails(tc.args.pw, tc.args.cr, tc.args.i)
			if diff := cmp.Diff(tc.want.conn, conn); diff != "" {
				t.Errorf("getConnectionDetails(...): -want, +got:\n%s", diff)
			}
		})
	}
}

type fakeRedisClient struct{}

func (c *fakeRedisClient) DescribeDBInstance(id string) (*redis.DBInstance, error) {
	if id != testName {
		return nil, errors.New("DescribeRedisInstance: client doesn't work")
	}
	return &redis.DBInstance{
		ID:     id,
		Status: v1alpha1.RedisInstanceStateRunning,
	}, nil
}

func (c *fakeRedisClient) CreateDBInstance(req *redis.CreateRedisInstanceRequest) (*redis.DBInstance, error) {
	if req.Name != testName {
		return nil, errors.New("CreateRedisInstance: client doesn't work")
	}
	return &redis.DBInstance{
		ID: testName,
		Endpoint: &v1alpha1.Endpoint{
			Address: "172.0.0.1",
			Port:    "8888",
		},
	}, nil
}

func (c *fakeRedisClient) CreateAccount(id, user, pw string) error {
	if id != testName {
		return errors.New("CreateAccount: client doesn't work")
	}
	return nil
}

func (c *fakeRedisClient) DeleteDBInstance(id string) error {
	if id != testName {
		return errors.New("DeleteRedisInstance: client doesn't work")
	}
	return nil
}

func (c *fakeRedisClient) AllocateInstancePublicConnection(id string, port int) (string, error) {
	if id != testName {
		return "nil", errors.New("AllocateInstancePublicConnection: client doesn't work")
	}
	return "", nil
}

func (c *fakeRedisClient) ModifyDBInstanceConnectionString(id string, port int) (string, error) {
	if id != testName {
		return "nil", errors.New("ModifyDBInstanceConnectionString: client doesn't work")
	}
	return "", nil
}

func (c *fakeRedisClient) Update(id string, req *redis.ModifyRedisInstanceRequest) error {
	if id != testName {
		return errors.New("Update: client doesn't work")
	}
	return nil
}
