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
	"strconv"
	"time"

	"github.com/pkg/errors"

	sdkerrors "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	aliredis "github.com/aliyun/alibaba-cloud-sdk-go/services/r-kvstore"

	"github.com/crossplane-contrib/provider-alibaba/apis/redis/v1alpha1"
)

var (
	// ErrDBInstanceNotFound indicates DBInstance not found
	ErrDBInstanceNotFound = errors.New("DBInstanceNotFound")
)

const (
	// DefaultReadTime indicates default connect timeout number
	DefaultReadTime = 60 * time.Second
	// PubilConnectionDomain indicates instances connect domain
	PubilConnectionDomain = "-pb.redis.rds.aliyuncs.com"
	// HTTPSScheme indicates request scheme
	HTTPSScheme = "https"
	// VPCNetworkType indicates network type by vpc
	VPCNetworkType = "VPC"
)

// Client defines Redis client operations
type Client interface {
	DescribeDBInstance(id string) (*DBInstance, error)
	CreateAccount(id, username, password string) error
	CreateDBInstance(*CreateRedisInstanceRequest) (*DBInstance, error)
	DeleteDBInstance(id string) error
	AllocateInstancePublicConnection(id string, port int) (string, error)
	ModifyDBInstanceConnectionString(id string, port int) (string, error)
	Update(id string, req *ModifyRedisInstanceRequest) error
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

// CreateRedisInstanceRequest defines the request info to create DB Instance
type CreateRedisInstanceRequest struct {
	Name           string
	InstanceType   string
	EngineVersion  string
	SecurityIPList string
	InstanceClass  string
	Password       string
	ChargeType     string
	Port           int
	NetworkType    string
	VpcID          string
	VSwitchID      string
}

// ModifyRedisInstanceRequest defines the request info to modify DB Instance
type ModifyRedisInstanceRequest struct {
	InstanceClass string
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
	request.Scheme = HTTPSScheme

	request.InstanceIds = id

	response, err := c.redisCli.DescribeInstances(request)
	if err != nil {
		return nil, errors.Wrap(err, "cannot describe redis instance")
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
	request := aliredis.CreateCreateInstanceRequest()
	request.Scheme = HTTPSScheme

	request.InstanceName = req.Name
	request.EngineVersion = req.EngineVersion
	request.InstanceClass = req.InstanceClass
	request.InstanceType = req.InstanceType
	request.ReadTimeout = DefaultReadTime
	request.ChargeType = req.ChargeType
	request.NetworkType = req.NetworkType

	if req.NetworkType == VPCNetworkType {
		request.VpcId = req.VpcID
		request.VSwitchId = req.VSwitchID
	}
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
	request.Scheme = HTTPSScheme
	request.InstanceId = id
	request.AccountName = user
	request.AccountPassword = pw
	request.ReadTimeout = DefaultReadTime

	_, err := c.redisCli.CreateAccount(request)
	return err
}

func (c *client) DeleteDBInstance(id string) error {
	request := aliredis.CreateDeleteInstanceRequest()
	request.Scheme = HTTPSScheme

	request.InstanceId = id

	_, err := c.redisCli.DeleteInstance(request)
	return err
}

// GenerateObservation is used to produce v1alpha1.RedisInstanceObservation from
// redis.DBInstance.
func GenerateObservation(db *DBInstance) v1alpha1.RedisInstanceObservation {
	return v1alpha1.RedisInstanceObservation{
		DBInstanceStatus: db.Status,
		DBInstanceID:     db.ID,
	}
}

// MakeCreateDBInstanceRequest generates CreateDBInstanceRequest
func MakeCreateDBInstanceRequest(name string, p *v1alpha1.RedisInstanceParameters) *CreateRedisInstanceRequest {
	return &CreateRedisInstanceRequest{
		Name:          name,
		InstanceType:  p.InstanceType,
		EngineVersion: p.EngineVersion,
		InstanceClass: p.InstanceClass,
		Port:          p.InstancePort,
		NetworkType:   p.NetworkType,
		VpcID:         p.VpcID,
		VSwitchID:     p.VSwitchID,
		ChargeType:    p.ChargeType,
	}
}

// IsErrorNotFound helper function to test for ErrCodeDBInstanceNotFoundFault error
func IsErrorNotFound(err error) bool {
	if err == nil {
		return false
	}
	// If the instance is already removed, errors should be ignored when deleting it.
	var srverr *sdkerrors.ServerError
	if !errors.As(err, &srverr) {
		return false || errors.Is(err, ErrDBInstanceNotFound)
	}

	return srverr.ErrorCode() == "InvalidInstanceId.NotFound"
}

func (c *client) AllocateInstancePublicConnection(id string, port int) (string, error) {
	request := aliredis.CreateAllocateInstancePublicConnectionRequest()
	request.Scheme = HTTPSScheme
	request.InstanceId = id
	request.ConnectionStringPrefix = id + PubilConnectionDomain
	request.Port = strconv.Itoa(port)
	request.ReadTimeout = DefaultReadTime
	_, err := c.redisCli.AllocateInstancePublicConnection(request)
	if err != nil {
		return "", err
	}
	return request.ConnectionStringPrefix, err
}

func (c *client) ModifyDBInstanceConnectionString(id string, port int) (string, error) {
	request := aliredis.CreateModifyDBInstanceConnectionStringRequest()
	request.Scheme = HTTPSScheme
	request.DBInstanceId = id
	request.CurrentConnectionString = id + PubilConnectionDomain
	request.Port = strconv.Itoa(port)
	request.ReadTimeout = DefaultReadTime
	_, err := c.redisCli.ModifyDBInstanceConnectionString(request)
	if err != nil {
		return "", err
	}
	return request.CurrentConnectionString, err
}

func (c *client) Update(id string, req *ModifyRedisInstanceRequest) error {
	if req.InstanceClass == "" {
		return errors.New("modify instances spec is require")
	}
	if req.InstanceClass != "" {
		return c.modifyInstanceSpec(id, req)
	}
	return nil
}

func (c *client) modifyInstanceSpec(id string, req *ModifyRedisInstanceRequest) error {
	request := aliredis.CreateModifyInstanceSpecRequest()
	request.Scheme = HTTPSScheme
	request.InstanceId = id
	request.InstanceClass = req.InstanceClass
	request.ReadTimeout = DefaultReadTime
	_, err := c.redisCli.ModifyInstanceSpec(request)
	return err
}
