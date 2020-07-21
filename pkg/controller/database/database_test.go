package database

import (
	"context"
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	crossplanemeta "github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-alibaba/apis/database/v1alpha1"
	aliv1alpha1 "github.com/crossplane/provider-alibaba/apis/v1alpha1"
	"github.com/crossplane/provider-alibaba/pkg/clients/rds"
)

const testName = "test"

func TestConnector(t *testing.T) {
	c := &connector{
		reader:       &testConnectorReader{},
		newRDSClient: testNewRDSClient,
	}
	obj := &v1alpha1.RDSInstance{
		Spec: v1alpha1.RDSInstanceSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: runtimev1alpha1.Reference{},
			},
		},
	}
	ext, err := c.Connect(context.Background(), obj)
	if err != nil {
		t.Fatal(err)
	}
	e := ext.(*external)
	_, ok := e.client.(*fakeRDSClient)
	if !ok {
		t.Error("newRDSClient doesn't work")
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
	if obj.GetBindingPhase() != runtimev1alpha1.BindingPhaseUnbound {
		t.Errorf("Binding phase (%v) should be %v", obj.GetBindingPhase(), runtimev1alpha1.BindingPhaseUnbound)
	}
	if obj.Status.AtProvider.AccountReady != true {
		t.Error("AccountReady should be true")
	}
	if string(ob.ConnectionDetails[runtimev1alpha1.ResourceCredentialsSecretUserKey]) != testName {
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
	if string(ob.ConnectionDetails[runtimev1alpha1.ResourceCredentialsSecretEndpointKey]) != "172.0.0.1" ||
		string(ob.ConnectionDetails[runtimev1alpha1.ResourceCredentialsSecretPortKey]) != "8888" {
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

type testConnectorReader struct {
}

func (r *testConnectorReader) GetProvider(ctx context.Context, key client.ObjectKey) (*aliv1alpha1.Provider, error) {
	obj := &aliv1alpha1.Provider{
		Spec: aliv1alpha1.ProviderSpec{
			ProviderSpec: runtimev1alpha1.ProviderSpec{
				CredentialsSecretRef: &runtimev1alpha1.SecretKeySelector{
					SecretReference: runtimev1alpha1.SecretReference{
						Name:      testName,
						Namespace: testName,
					},
				},
			},
			Region: testName,
		},
	}
	return obj, nil
}

func (r *testConnectorReader) GetSecret(ctx context.Context, key client.ObjectKey) (*corev1.Secret, error) {
	if key.Name != testName || key.Namespace != testName {
		return nil, errors.New("GetSecret: reader doesn't work")
	}
	obj := &corev1.Secret{
		Data: map[string][]byte{
			"accessKeyId":     []byte(testName),
			"accessKeySecret": []byte(testName),
		},
	}
	return obj, nil
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
					runtimev1alpha1.ResourceCredentialsSecretUserKey:     []byte(testName),
					runtimev1alpha1.ResourceCredentialsSecretEndpointKey: []byte(address),
					runtimev1alpha1.ResourceCredentialsSecretPortKey:     []byte(port),
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
					runtimev1alpha1.ResourceCredentialsSecretUserKey:     []byte(testName),
					runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(password),
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
					runtimev1alpha1.ResourceCredentialsSecretUserKey:     []byte(testName),
					runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(password),
					runtimev1alpha1.ResourceCredentialsSecretEndpointKey: []byte(address),
					runtimev1alpha1.ResourceCredentialsSecretPortKey:     []byte(port),
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

func testNewRDSClient(ctx context.Context, accessKeyID, accessKeySecret, region string) (rds.Client, error) {
	if accessKeyID != testName || accessKeySecret != testName || region != testName {
		return nil, errors.New("testNewRDSClient: reader doesn't work")
	}
	return &fakeRDSClient{}, nil
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
