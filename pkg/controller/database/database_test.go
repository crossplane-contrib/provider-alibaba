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
	"github.com/crossplane/provider-alibaba/apis/database/v1alpha1"
	aliv1alpha1 "github.com/crossplane/provider-alibaba/apis/v1alpha1"
	"github.com/crossplane/provider-alibaba/pkg/clients/rds"
)

func TestConnector(t *testing.T) {
	c := &connector{
		reader:       &testConnectorReader{},
		newRDSClient: testNewRDSClient,
	}
	obj := &v1alpha1.RDSInstance{
		Spec: v1alpha1.RDSInstanceSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: &corev1.ObjectReference{},
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
				MasterUsername: "test",
			},
		},
		Status: v1alpha1.RDSInstanceStatus{
			AtProvider: v1alpha1.RDSInstanceObservation{
				DBInstanceID: "test",
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
	if string(ob.ConnectionDetails[runtimev1alpha1.ResourceCredentialsSecretUserKey]) != "test" {
		t.Error("ConnectionDetails should include username=test")
	}
}

func TestExternalClientCreate(t *testing.T) {
	e := &external{client: &fakeRDSClient{}}
	obj := &v1alpha1.RDSInstance{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				crossplanemeta.AnnotationKeyExternalName: "test",
			},
		},
		Spec: v1alpha1.RDSInstanceSpec{
			ForProvider: v1alpha1.RDSInstanceParameters{
				MasterUsername:        "test",
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
	if obj.Status.AtProvider.DBInstanceID != "test" {
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
				DBInstanceID: "test",
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
						Name:      "test",
						Namespace: "test",
					},
				},
			},
			Region: "test",
		},
	}
	return obj, nil
}

func (r *testConnectorReader) GetSecret(ctx context.Context, key client.ObjectKey) (*corev1.Secret, error) {
	if key.Name != "test" || key.Namespace != "test" {
		return nil, errors.New("GetSecret: reader doesn't work")
	}
	obj := &corev1.Secret{
		Data: map[string][]byte{
			"accessKeyId":     []byte("test"),
			"accessKeySecret": []byte("test"),
		},
	}
	return obj, nil
}

func testNewRDSClient(ctx context.Context, accessKeyID, accessKeySecret, region string) (rds.Client, error) {
	if accessKeyID != "test" || accessKeySecret != "test" || region != "test" {
		return nil, errors.New("testNewRDSClient: reader doesn't work")
	}
	return &fakeRDSClient{}, nil
}

type fakeRDSClient struct {
}

func (c *fakeRDSClient) DescribeDBInstance(id string) (*rds.DBInstance, error) {
	if id != "test" {
		return nil, errors.New("DescribeDBInstance: client doesn't work")
	}
	return &rds.DBInstance{
		ID:     id,
		Status: v1alpha1.RDSInstanceStateRunning,
	}, nil
}

func (c *fakeRDSClient) CreateDBInstance(req *rds.CreateDBInstanceRequest) (*rds.DBInstance, error) {
	if req.Name != "test" || req.Engine != "PostgreSQL" {
		return nil, errors.New("CreateDBInstance: client doesn't work")
	}
	return &rds.DBInstance{
		ID: "test",
		Endpoint: &v1alpha1.Endpoint{
			Address: "172.0.0.1",
			Port:    "8888",
		},
	}, nil
}

func (c *fakeRDSClient) CreateAccount(id, user, pw string) error {
	if id != "test" {
		return errors.New("CreateAccount: client doesn't work")
	}
	return nil
}

func (c *fakeRDSClient) DeleteDBInstance(id string) error {
	if id != "test" {
		return errors.New("DeleteDBInstance: client doesn't work")
	}
	return nil
}
