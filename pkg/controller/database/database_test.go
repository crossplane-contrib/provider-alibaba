package database

import (
	"context"
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
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

func TestExternal(t *testing.T) {

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
		return nil, errors.New("reader doesn't work")
	}
	return &fakeRDSClient{}, nil
}

type fakeRDSClient struct {
}

func (c *fakeRDSClient) DescribeDBInstance(id string) (*rds.DBInstance, error) {
	panic("")
}

func (c *fakeRDSClient) CreateDBInstance(req *rds.CreateDBInstanceRequest) (*rds.DBInstance, error) {
	panic("")
}

func (c *fakeRDSClient) CreateAccount(id, user, pw string) error {
	panic("")
}

func (c *fakeRDSClient) DeleteDBInstance(id string) error {
	panic("")
}
