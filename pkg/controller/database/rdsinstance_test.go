package database

import (
	"context"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	crossplanemeta "github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/provider-alibaba/apis/database/v1alpha1"
	aliv1beta1 "github.com/crossplane/provider-alibaba/apis/v1beta1"
	"github.com/crossplane/provider-alibaba/pkg/clients/rds"
	"github.com/crossplane/provider-alibaba/pkg/util"
)

const (
	testName                = "test"
	errExtractSecretKey     = "cannot extract from secret key when none specified"
	errGetCredentialsSecret = "cannot get credentials secret"
)

func TestConnector(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		client       client.Client
		usage        resource.Tracker
		newRDSClient func(ctx context.Context, accessKeyID, accessKeySecret, securityToken, region string) (rds.Client, error)
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	var configSpec = aliv1beta1.ProviderCredentials{Source: xpv1.CredentialsSourceSecret}
	configSpec.SecretRef = &xpv1.SecretKeySelector{
		SecretReference: xpv1.SecretReference{
			Name: "coolsecret",
		},
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   error
	}{
		"NotRDSInstance": {
			reason: "Should return an error if the supplied managed resource is not an RDSInstance",
			args: args{
				mg: nil,
			},
			want: errors.New(errNotRDSInstance),
		},
		"TrackProviderConfigUsageError": {
			reason: "Errors tracking a ProviderConfigUsage should be returned",
			fields: fields{
				usage: resource.TrackerFn(func(ctx context.Context, mg resource.Managed) error { return errBoom }),
			},
			args: args{
				mg: &v1alpha1.RDSInstance{
					Spec: v1alpha1.RDSInstanceSpec{
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
				mg: &v1alpha1.RDSInstance{
					Spec: v1alpha1.RDSInstanceSpec{
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{},
						},
					},
				},
			},
			want: errors.Wrap(errors.Wrap(errBoom, util.ErrGetProviderConfig), util.ErrPrepareClientEstablishmentInfo),
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
									Source: "wat",
								},
							},
						}
						return nil
					}),
				},
				usage: resource.TrackerFn(func(ctx context.Context, mg resource.Managed) error { return nil }),
			},
			args: args{
				ctx: context.TODO(),
				mg: &v1alpha1.RDSInstance{
					Spec: v1alpha1.RDSInstanceSpec{
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{},
						},
					},
				},
			},
			want: errors.Wrap(errors.Wrap(errors.Errorf(errFmtUnsupportedCredSource, "wat"), errGetCredentials), util.ErrPrepareClientEstablishmentInfo),
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
				mg: &v1alpha1.RDSInstance{
					Spec: v1alpha1.RDSInstanceSpec{
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{},
						},
					},
				},
			},
			want: errors.Wrap(errors.Wrap(errBoom, util.ErrGetProviderConfig), util.ErrPrepareClientEstablishmentInfo),
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
				mg: &v1alpha1.RDSInstance{
					Spec: v1alpha1.RDSInstanceSpec{
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{},
						},
					},
				},
			},
			want: errors.Wrap(errors.Wrap(errors.New(errExtractSecretKey), errGetCredentials), util.ErrPrepareClientEstablishmentInfo),
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
									Credentials: configSpec,
								},
							}
						}
						return nil
					}),
				},
				usage: resource.TrackerFn(func(ctx context.Context, mg resource.Managed) error { return nil }),
			},
			args: args{
				mg: &v1alpha1.RDSInstance{
					Spec: v1alpha1.RDSInstanceSpec{
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{},
						},
					},
				},
			},
			want: errors.Wrap(errors.Wrap(errors.Wrap(errBoom, errGetCredentialsSecret), errGetCredentials), util.ErrPrepareClientEstablishmentInfo),
		},
		"NewRDSClientError": {
			reason: "Errors getting a secret should be returned",
			fields: fields{
				client: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						if t, ok := obj.(*aliv1beta1.ProviderConfig); ok {
							*t = aliv1beta1.ProviderConfig{
								Spec: aliv1beta1.ProviderConfigSpec{
									Credentials: configSpec,
								},
							}
						}
						return nil
					}),
				},
				usage: resource.TrackerFn(func(ctx context.Context, mg resource.Managed) error { return nil }),
				newRDSClient: func(ctx context.Context, accessKeyID, accessKeySecret, securityToken, region string) (rds.Client, error) {
					return nil, errBoom
				},
			},
			args: args{
				mg: &v1alpha1.RDSInstance{
					Spec: v1alpha1.RDSInstanceSpec{
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{},
						},
					},
				},
			},
			want: errors.Wrap(errors.New(util.ErrAccessKeyNotComplete), util.ErrPrepareClientEstablishmentInfo),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{client: tc.fields.client, usage: tc.fields.usage, newRDSClient: tc.fields.newRDSClient}
			_, err := c.Connect(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nc.Connect(...) -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestExternalClientObserve(t *testing.T) {
	e := &external{client: &fakeRDSClient{}}
	obj := &v1alpha1.RDSInstance{
		Spec: v1alpha1.RDSInstanceSpec{
			ForProvider: v1alpha1.RDSInstanceParameters{
				MasterUsername: testName,
			},
		},
		Status: v1alpha1.RDSInstanceStatus{
			AtProvider: v1alpha1.RDSInstanceObservation{
				DBInstanceID: testName,
			},
		},
	}
	ob, err := e.Observe(context.Background(), obj)
	if err != nil {
		t.Fatal(err)
	}
	if obj.Status.AtProvider.DBInstanceStatus != v1alpha1.RDSInstanceStateRunning {
		t.Errorf("DBInstanceStatus (%v) should be %v", obj.Status.AtProvider.DBInstanceStatus, v1alpha1.RDSInstanceStateRunning)
	}
	if obj.Status.AtProvider.AccountReady != true {
		t.Error("AccountReady should be true")
	}
	if string(ob.ConnectionDetails[xpv1.ResourceCredentialsSecretUserKey]) != testName {
		t.Error("ConnectionDetails should include username=test")
	}
}

func TestExternalClientCreate(t *testing.T) {
	e := &external{client: &fakeRDSClient{}}
	obj := &v1alpha1.RDSInstance{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				crossplanemeta.AnnotationKeyExternalName: testName,
			},
		},
		Spec: v1alpha1.RDSInstanceSpec{
			ForProvider: v1alpha1.RDSInstanceParameters{
				MasterUsername:        testName,
				Engine:                "PostgreSQL",
				EngineVersion:         "10.0",
				SecurityIPList:        "0.0.0.0/0",
				DBInstanceClass:       "rds.pg.s1.small",
				DBInstanceStorageInGB: 20,
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
	e := &external{client: &fakeRDSClient{}}
	obj := &v1alpha1.RDSInstance{
		Status: v1alpha1.RDSInstanceStatus{
			AtProvider: v1alpha1.RDSInstanceObservation{
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
		cr *v1alpha1.RDSInstance
		i  *rds.DBInstance
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
				cr: &v1alpha1.RDSInstance{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							crossplanemeta.AnnotationKeyExternalName: testName,
						},
					},
					Spec: v1alpha1.RDSInstanceSpec{
						ForProvider: v1alpha1.RDSInstanceParameters{
							MasterUsername: testName,
						},
					},
				},
				i: &rds.DBInstance{
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
				cr: &v1alpha1.RDSInstance{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							crossplanemeta.AnnotationKeyExternalName: testName,
						},
					},
					Spec: v1alpha1.RDSInstanceSpec{
						ForProvider: v1alpha1.RDSInstanceParameters{
							MasterUsername: testName,
						},
					},
				},
				i: &rds.DBInstance{},
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
				cr: &v1alpha1.RDSInstance{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							crossplanemeta.AnnotationKeyExternalName: testName,
						},
					},
					Spec: v1alpha1.RDSInstanceSpec{
						ForProvider: v1alpha1.RDSInstanceParameters{
							MasterUsername: testName,
						},
					},
				},
				i: &rds.DBInstance{
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

type fakeRDSClient struct {
}

func (c *fakeRDSClient) DescribeDBInstance(id string) (*rds.DBInstance, error) {
	if id != testName {
		return nil, errors.New("DescribeDBInstance: client doesn't work")
	}
	return &rds.DBInstance{
		ID:     id,
		Status: v1alpha1.RDSInstanceStateRunning,
	}, nil
}

func (c *fakeRDSClient) CreateDBInstance(req *rds.CreateDBInstanceRequest) (*rds.DBInstance, error) {
	if req.Name != testName || req.Engine != "PostgreSQL" {
		return nil, errors.New("CreateDBInstance: client doesn't work")
	}
	return &rds.DBInstance{
		ID: testName,
		Endpoint: &v1alpha1.Endpoint{
			Address: "172.0.0.1",
			Port:    "8888",
		},
	}, nil
}

func (c *fakeRDSClient) CreateAccount(id, user, pw string) error {
	if id != testName {
		return errors.New("CreateAccount: client doesn't work")
	}
	return nil
}

func (c *fakeRDSClient) DeleteDBInstance(id string) error {
	if id != testName {
		return errors.New("DeleteDBInstance: client doesn't work")
	}
	return nil
}
