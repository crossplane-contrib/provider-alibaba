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

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	alirds "github.com/aliyun/alibaba-cloud-sdk-go/services/rds"

	"github.com/crossplane/provider-alibaba/apis/database/v1alpha1"
)

var (
	// ErrDBInstanceNotFound indicates DBInstance not found
	ErrDBInstanceNotFound = errors.New("DBInstanceNotFound")
)

// Client defines RDS client operations
type Client interface {
	DescribeDBInstance(id string) (*DBInstance, error)
	CreateDBInstance(*CreateDBInstanceRequest) (*DBInstance, error)
	DeleteDBInstance(id string) error
}

// DBInstance defines the DB instance information
type DBInstance struct {
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
	Username              string
	Password              string
}

type client struct {
	rdsCli *alirds.Client
}

// NewClient creates new RDS RDSClient
func NewClient(ctx context.Context, accessKeyID, accessSecret, region string) (Client, error) {
	rdsCli, err := alirds.NewClientWithAccessKey(region, accessKeyID, accessSecret)
	if err != nil {
		return nil, err
	}
	c := &client{rdsCli: rdsCli}
	return c, nil
}

func (c *client) DescribeDBInstance(id string) (*DBInstance, error) {
	request := alirds.CreateDescribeDBInstancesRequest()
	request.Scheme = "https"

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
		Engine: rsp.Engine,
		Status: rsp.DBInstanceStatus,
	}

	return in, nil
}

func (c *client) CreateDBInstance(req *CreateDBInstanceRequest) (*DBInstance, error) {
	request := alirds.CreateCreateDBInstanceRequest()
	request.Scheme = "https"

	request.DBInstanceDescription = req.Name
	request.Engine = req.Engine
	request.EngineVersion = req.EngineVersion
	request.DBInstanceClass = req.DBInstanceClass
	request.DBInstanceStorage = requests.NewInteger(req.DBInstanceStorageInGB)
	request.SecurityIPList = req.SecurityIPList
	request.DBInstanceNetType = "Internet"
	request.PayType = "Postpaid"

	resp, err := c.rdsCli.CreateDBInstance(request)
	if err != nil {
		return nil, err
	}

	accReq := alirds.CreateCreateAccountRequest()
	accReq.Scheme = "https"
	accReq.DBInstanceId = resp.DBInstanceId
	accReq.AccountName = req.Username
	accReq.AccountPassword = req.Password

	_, err = c.rdsCli.CreateAccount(accReq)
	if err != nil {
		return nil, err
	}

	return &DBInstance{
		Endpoint: &v1alpha1.Endpoint{
			Address: resp.ConnectionString,
			Port:    resp.Port,
		},
	}, nil
}

func (c *client) DeleteDBInstance(id string) error {
	request := alirds.CreateDeleteDBInstanceRequest()
	request.Scheme = "https"

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
func GenerateObservation(db *DBInstance) *v1alpha1.RDSInstanceObservation {
	return &v1alpha1.RDSInstanceObservation{
		DBInstanceStatus: db.Status,
	}
}

// MakeCreateDBInstanceRequest generates CreateDBInstanceRequest
func MakeCreateDBInstanceRequest(name, username, password string, p *v1alpha1.RDSInstanceParameters) *CreateDBInstanceRequest {
	return &CreateDBInstanceRequest{
		Name:                  name,
		Engine:                p.Engine,
		EngineVersion:         p.EngineVersion,
		SecurityIPList:        p.SecurityIPList,
		DBInstanceClass:       p.DBInstanceClass,
		DBInstanceStorageInGB: p.DBInstanceStorageInGB,
		Username:              username,
		Password:              password,
	}
}

// IsErrorNotFound helper function to test for ErrCodeDBInstanceNotFoundFault error
func IsErrorNotFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrDBInstanceNotFound)
}

// IsUpToDate checks whether there is a change in any of the modifiable fields.
func IsUpToDate(p v1alpha1.RDSInstanceParameters, db *DBInstance) bool {
	return db != nil
}
