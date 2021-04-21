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

package redis

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/crossplane/crossplane-runtime/pkg/password"

	sdkerrors "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"

	aliredis "github.com/aliyun/alibaba-cloud-sdk-go/services/r-kvstore"

	"github.com/crossplane/provider-alibaba/apis/database/v1alpha1"
	redisv1alpha1 "github.com/crossplane/provider-alibaba/apis/database/v1alpha1/redis"
)

var (
	// ErrDBInstanceNotFound indicates DBInstance not found
	ErrDBInstanceNotFound = errors.New("DBInstanceNotFound")
	// ErrCodeInstanceNotFound error code of ServerError when DBInstance not found
	ErrCodeInstanceNotFound = "InvalidDBInstanceId.NotFound"

	httpsScheme = "https"
)

// Client defines Redis client operations
type Client interface {
	DescribeDBInstance(id string) (*DBInstance, error)
	CreateAccount(id, username, password string) error
	CreateDBInstance(*CreateRedisInstanceRequest) (*DBInstance, error)
	DeleteDBInstance(id string) error
	AllocateInstancePublicConnection(id string, port int) (string, error)
	ModifyDBInstanceConnectionString(id string, port int) (string, error)
}

// DBInstance defines the DB instance information
type DBInstance struct {
	// Instance ID
	ID string

	// Instance status
	Status string

	// Endpoint specifies the connection endpoint.
	Endpoint *v1alpha1.Endpoint
}

// CreateDBInstanceRequest defines the request info to create DB Instance
type CreateRedisInstanceRequest struct {
	Name           string
	Engine         string
	EngineVersion  string
	SecurityIPList string
	InstanceClass  string
	Password       string
	Port           int
	Config         string
}

type client struct {
	redisCli *aliredis.Client
}

// NewClient creates new Redis RedisClient
func NewClient(ctx context.Context, accessKeyID, accessKeySecret, region string) (Client, error) {
	redisCli, err := aliredis.NewClientWithAccessKey(region, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, err
	}
	c := &client{redisCli: redisCli}
	return c, nil
}

func (c *client) DescribeDBInstance(id string) (*DBInstance, error) {
	request := aliredis.CreateDescribeInstancesRequest()
	request.Scheme = httpsScheme

	request.InstanceIds = id

	response, err := c.redisCli.DescribeInstances(request)
	if err != nil {
		return nil, err
	}
	if len(response.Instances.KVStoreInstance) == 0 {
		return nil, ErrDBInstanceNotFound
	}
	rsp := response.Instances.KVStoreInstance[0]
	in := &DBInstance{
		ID:     rsp.InstanceId,
		Status: rsp.InstanceStatus,
	}

	return in, nil
}

func (c *client) CreateDBInstance(req *CreateRedisInstanceRequest) (*DBInstance, error) {
	pw, err := password.Generate()
	if err != nil {
		return nil, err
	}
	request := aliredis.CreateCreateInstanceRequest()
	request.Scheme = "https"

	request.InstanceName = req.Name
	request.EngineVersion = req.EngineVersion
	request.InstanceClass = req.InstanceClass
	request.InstanceType = req.Engine
	request.ReadTimeout = 60 * time.Second
	request.Password = pw
	request.ChargeType = "PostPaid"

	resp, err := c.redisCli.CreateInstance(request)
	if err != nil {
		return nil, err
	}

	return &DBInstance{
		ID: resp.InstanceId,
		Endpoint: &v1alpha1.Endpoint{
			Address: resp.ConnectionDomain,
			Port:    strconv.Itoa(resp.Port),
		},
	}, nil
}

func (c *client) CreateAccount(id, user, pw string) error {
	request := aliredis.CreateCreateAccountRequest()
	request.Scheme = "https"
	request.InstanceId = id
	request.AccountName = user
	request.AccountPassword = pw
	request.ReadTimeout = 60 * time.Second

	_, err := c.redisCli.CreateAccount(request)
	return err
}

func (c *client) DeleteDBInstance(id string) error {
	request := aliredis.CreateDeleteInstanceRequest()
	request.Scheme = "https"

	request.InstanceId = id

	_, err := c.redisCli.DeleteInstance(request)
	return err
}

// GenerateObservation is used to produce v1alpha1.RedisInstanceObservation from
// redis.DBInstance.
func GenerateObservation(db *DBInstance) redisv1alpha1.RedisInstanceObservation {
	return redisv1alpha1.RedisInstanceObservation{
		DBInstanceStatus: db.Status,
		DBInstanceID:     db.ID,
	}
}

// MakeCreateDBInstanceRequest generates CreateDBInstanceRequest
func MakeCreateDBInstanceRequest(name string, p *redisv1alpha1.RedisInstanceParameters) *CreateRedisInstanceRequest {
	if p.Engine == "" {
		p.Engine = "Redis"
	}

	return &CreateRedisInstanceRequest{
		Name:           name,
		Engine:         p.Engine,
		EngineVersion:  p.EngineVersion,
		SecurityIPList: p.SecurityIPList,
		InstanceClass:  p.DBInstanceClass,
		Port:           p.DBInstancePort,
	}
}

// IsErrorNotFound helper function to test for ErrCodeDBInstanceNotFoundFault error
func IsErrorNotFound(err error) bool {
	if err == nil {
		return false
	}
	// If the instance is already removed, errors should be ignored when deleting it.
	if e, ok := err.(*sdkerrors.ServerError); ok && e.ErrorCode() == ErrCodeInstanceNotFound {
		return true
	}
	return errors.Is(err, ErrDBInstanceNotFound)
}

func (c *client) AllocateInstancePublicConnection(id string, port int) (string, error) {
	request := aliredis.CreateAllocateInstancePublicConnectionRequest()
	request.Scheme = "https"
	request.InstanceId = id
	request.ConnectionStringPrefix = id + "-pb.redis.rds.aliyuncs.com"
	request.Port = strconv.Itoa(port)
	request.ReadTimeout = 60 * time.Second
	_, err := c.redisCli.AllocateInstancePublicConnection(request)
	if err != nil {
		return "", err
	}
	return request.ConnectionStringPrefix, err
}

func (c *client) ModifyDBInstanceConnectionString(id string, port int) (string, error) {
	request := aliredis.CreateModifyDBInstanceConnectionStringRequest()
	request.Scheme = "https"
	request.DBInstanceId = id
	request.CurrentConnectionString = id + "-pb.redis.rds.aliyuncs.com"
	request.Port = strconv.Itoa(port)
	request.ReadTimeout = 60 * time.Second
	_, err := c.redisCli.ModifyDBInstanceConnectionString(request)
	if err != nil {
		return "", err
	}
	return request.CurrentConnectionString, err
}
