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

package rds

import (
	"context"
	"errors"
	"time"

	sdkerrors "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	alirds "github.com/aliyun/alibaba-cloud-sdk-go/services/rds"

	"github.com/crossplane/provider-alibaba/apis/database/v1alpha1"
)

var (
	// ErrDBInstanceNotFound indicates DBInstance not found
	ErrDBInstanceNotFound = errors.New("DBInstanceNotFound")
	// ErrCodeInstanceNotFound error code of ServerError when DBInstance not found
	ErrCodeInstanceNotFound = "InvalidDBInstanceId.NotFound"
)

const (
	httpsScheme = "https"
)

// Client defines RDS client operations
type Client interface {
	DescribeDBInstance(id string) (*DBInstance, error)
	CreateAccount(id, username, password string) error
	CreateDBInstance(*CreateDBInstanceRequest) (*DBInstance, error)
	DeleteDBInstance(id string) error
}

// DBInstance defines the DB instance information
type DBInstance struct {
	// Instance ID
	ID string

	// Database engine
	Engine string

	// Instance status
	Status string

	// Endpoint specifies the connection endpoint.
	Endpoint *v1alpha1.Endpoint
}

// CreateDBInstanceRequest defines the request info to create DB Instance
type CreateDBInstanceRequest struct {
	Name                  string
	Engine                string
	EngineVersion         string
	SecurityIPList        string
	DBInstanceClass       string
	DBInstanceStorageInGB int
}

type client struct {
	rdsCli *alirds.Client
}

// NewClient creates new RDS RDSClient
func NewClient(ctx context.Context, accessKeyID, accessKeySecret, securityToken, region string) (Client, error) {
	var (
		rdsCli *alirds.Client
		err    error
	)
	if securityToken != "" {
		rdsCli, err = alirds.NewClientWithStsToken(region, accessKeyID, accessKeySecret, securityToken)
	} else {
		rdsCli, err = alirds.NewClientWithAccessKey(region, accessKeyID, accessKeySecret)
	}

	if err != nil {
		return nil, err
	}
	c := &client{rdsCli: rdsCli}
	return c, nil
}

func (c *client) DescribeDBInstance(id string) (*DBInstance, error) {
	request := alirds.CreateDescribeDBInstancesRequest()
	request.Scheme = httpsScheme

	request.DBInstanceId = id

	response, err := c.rdsCli.DescribeDBInstances(request)
	if err != nil {
		return nil, err
	}
	if len(response.Items.DBInstance) == 0 {
		return nil, ErrDBInstanceNotFound
	}
	rsp := response.Items.DBInstance[0]
	in := &DBInstance{
		ID:     rsp.DBInstanceId,
		Engine: rsp.Engine,
		Status: rsp.DBInstanceStatus,
	}

	return in, nil
}

func (c *client) CreateDBInstance(req *CreateDBInstanceRequest) (*DBInstance, error) {
	request := alirds.CreateCreateDBInstanceRequest()
	request.Scheme = httpsScheme

	request.DBInstanceDescription = req.Name
	request.Engine = req.Engine
	request.EngineVersion = req.EngineVersion
	request.DBInstanceClass = req.DBInstanceClass
	request.DBInstanceStorage = requests.NewInteger(req.DBInstanceStorageInGB)
	request.SecurityIPList = req.SecurityIPList
	request.DBInstanceNetType = "Internet"
	request.PayType = "Postpaid"
	request.ReadTimeout = 60 * time.Second
	request.ClientToken = req.Name

	resp, err := c.rdsCli.CreateDBInstance(request)
	if err != nil {
		return nil, err
	}

	return &DBInstance{
		ID: resp.DBInstanceId,
		Endpoint: &v1alpha1.Endpoint{
			Address: resp.ConnectionString,
			Port:    resp.Port,
		},
	}, nil
}

func (c *client) CreateAccount(id, user, pw string) error {
	request := alirds.CreateCreateAccountRequest()
	request.Scheme = httpsScheme
	request.DBInstanceId = id
	request.AccountName = user
	request.AccountPassword = pw
	request.ReadTimeout = 60 * time.Second

	_, err := c.rdsCli.CreateAccount(request)
	return err
}

func (c *client) DeleteDBInstance(id string) error {
	request := alirds.CreateDeleteDBInstanceRequest()
	request.Scheme = httpsScheme

	request.DBInstanceId = id

	_, err := c.rdsCli.DeleteDBInstance(request)
	return err
}

// LateInitialize fills the empty fields in *v1alpha1.RDSInstanceParameters with
// the values seen in rds.DBInstance.
func LateInitialize(in *v1alpha1.RDSInstanceParameters, db *DBInstance) {
	in.Engine = db.Engine
}

// GenerateObservation is used to produce v1alpha1.RDSInstanceObservation from
// rds.DBInstance.
func GenerateObservation(db *DBInstance) v1alpha1.RDSInstanceObservation {
	return v1alpha1.RDSInstanceObservation{
		DBInstanceStatus: db.Status,
		DBInstanceID:     db.ID,
	}
}

// MakeCreateDBInstanceRequest generates CreateDBInstanceRequest
func MakeCreateDBInstanceRequest(name string, p *v1alpha1.RDSInstanceParameters) *CreateDBInstanceRequest {
	return &CreateDBInstanceRequest{
		Name:                  name,
		Engine:                p.Engine,
		EngineVersion:         p.EngineVersion,
		SecurityIPList:        p.SecurityIPList,
		DBInstanceClass:       p.DBInstanceClass,
		DBInstanceStorageInGB: p.DBInstanceStorageInGB,
	}
}

// IsErrorNotFound helper function to test for ErrCodeDBInstanceNotFoundFault error
func IsErrorNotFound(err error) bool {
	if err == nil {
		return false
	}
	// If instance already remove from console.  should ignore when delete instance
	if e, ok := err.(*sdkerrors.ServerError); ok && e.ErrorCode() == ErrCodeInstanceNotFound {
		return true
	}
	return errors.Is(err, ErrDBInstanceNotFound)
}
