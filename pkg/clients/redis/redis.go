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
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/pkg/errors"

	sdkerrors "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	aliredis "github.com/aliyun/alibaba-cloud-sdk-go/services/r-kvstore"

	"github.com/crossplane-contrib/provider-alibaba/apis/redis/v1alpha1"
)

const (
	// HTTPSScheme indicates request scheme
	HTTPSScheme         = "https"
	ConnectTimeOut      = 60 * time.Second
	ReadTimeOut         = 60 * time.Second
	DefaultIPModifyMode = "Cover"

	// Effective time for Spec update
	EffectiveTimeImmediate = "Immediately"
	EffectiveTimeMaintain  = "MaintainTime"
	DefaultEffectiveTime   = "Immediately"

	// Architecture type
	RedisArchTypeCluster  = "cluster"
	RedisArchTypeStandard = "standard"
	RedisArchTypeRWSplit  = "rwsplit"

	// Errors
	errInstanceNotFound              = "InstanceNotFound"
	errInstanceNotFoundCode          = "InvalidInstanceId.NotFound"
	errDescribeInstanceFailed        = "cannot describe instance attributes"
	errGeneratePasswordFailed        = "cannot generate a password"
	errFailedToUnmarshalWantedConfig = "cannot unmarshal the json config from users input"
	errFailedToUnmarshalRealConfig   = "cannot unmarshal the json config from the real instance"
	errFailedToMarshalConfig         = "cannot marshal the json config for the update params request"
)

// Same server error but without requestID
type CleanedServerError struct {
	HttpStatus int
	HostId     string
	Code       string
	Message    string
	Comment    string
}

// Client defines Redis client operations
type Client interface {
	DescribeInstance(id string) (*aliredis.DBInstanceAttribute, *RedisConnection, error)
	CreateInstance(name string, parameters *v1alpha1.RedisInstanceParameters) (*aliredis.CreateInstanceResponse, *RedisConnection, error)
	DeleteInstance(id string) error

	ModifyInstanceSpecs(id string, req *aliredis.ModifyInstanceSpecRequest) error
	ModifyInstanceParams(id string, req *aliredis.ModifyInstanceConfigRequest) error
	ModifySecurityIps(id string, req *aliredis.ModifySecurityIpsRequest) error
}

// ModifyRedisInstanceRequest defines the request info to modify DB Instance
type ModifyRedisInstanceRequest struct {
	InstanceClass string
}

type client struct {
	redisCli *aliredis.Client
}

type RedisConnection struct {
	Username         string
	Password         string
	ConnectionDomain string
	Port             string
}

// NewClient creates new Redis RedisClient
func NewClient(ctx context.Context, accessKeyID, accessKeySecret, region string) (Client, error) {
	redisCli, err := aliredis.NewClientWithAccessKey(region, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, CleanError(err)
	}
	c := &client{redisCli: redisCli}
	return c, nil
}

func (c *client) DescribeInstance(id string) (*aliredis.DBInstanceAttribute, *RedisConnection, error) {
	request := aliredis.CreateDescribeInstanceAttributeRequest()
	request.Scheme = HTTPSScheme

	request.InstanceId = id

	response, err := c.redisCli.DescribeInstanceAttribute(request)
	if err != nil {
		return nil, nil, errors.Wrap(CleanError(err), errDescribeInstanceFailed)
	}
	if len(response.Instances.DBInstanceAttribute) == 0 {
		return nil, nil, errors.New(errInstanceNotFound)
	}

	attr := response.Instances.DBInstanceAttribute[0]

	conn := &RedisConnection{
		ConnectionDomain: attr.ConnectionDomain,
		Port:             strconv.FormatInt(attr.Port, 10),
	}

	return &attr, conn, nil
}

func (c *client) CreateInstance(externalName string, p *v1alpha1.RedisInstanceParameters) (*aliredis.CreateInstanceResponse, *RedisConnection, error) {
	request := aliredis.CreateCreateInstanceRequest()
	request.ConnectTimeout = ConnectTimeOut
	request.ReadTimeout = ReadTimeOut

	// Seems regionID will be by default from the first part ZoneID
	// request.RegionID = p.RegionID
	request.Token = p.Token
	request.InstanceName = externalName
	request.InstanceClass = p.InstanceClass
	request.ZoneId = p.ZoneID
	request.ChargeType = p.ChargeType
	request.NodeType = p.NodeType
	request.NetworkType = p.NetworkType
	request.VpcId = p.VpcID
	request.VSwitchId = p.VSwitchID
	request.Period = p.Period
	request.BusinessInfo = p.BusinessInfo
	request.CouponNo = p.CouponNo
	request.SrcDBInstanceId = p.SrcDBInstanceId
	request.BackupId = p.BackupId
	request.InstanceType = p.InstanceType
	request.EngineVersion = p.EngineVersion
	request.PrivateIpAddress = p.PrivateIpAddress
	request.AutoUseCoupon = p.AutoUseCoupon
	request.AutoRenew = p.AutoRenew
	request.AutoRenewPeriod = p.AutoRenewPeriod
	request.ResourceGroupId = p.ResourceGroupId
	request.RestoreTime = p.RestoreTime
	request.DedicatedHostGroupId = p.DedicatedHostGroupId
	request.GlobalInstanceId = p.GlobalInstanceId
	request.GlobalInstance = requests.NewBoolean(p.GlobalInstance)
	request.SecondaryZoneId = p.SecondaryZoneID
	request.GlobalSecurityGroupIds = p.GlobalSecurityGroupIds
	request.Appendonly = p.Appendonly
	request.ConnectionStringPrefix = p.ConnectionStringPrefix
	request.ParamGroupId = p.ParamGroupId
	request.ClusterBackupId = p.ClusterBackupId

	if p.Port != nil {
		request.Port = strconv.Itoa(*p.Port)
	}

	if p.Capacity != nil {
		request.Capacity = requests.NewInteger(*p.Capacity)
	}

	if p.ShardCount != nil {
		request.ShardCount = requests.NewInteger(*p.ShardCount)
	}

	if p.ReadOnlyCount != nil {
		request.ReadOnlyCount = requests.NewInteger(*p.ReadOnlyCount)
	}

	// Password might be generated or provided by user
	var pw string
	var err error
	if p.Password == "" {
		pw, err = password.Generate()
		if err != nil {
			return nil, nil, errors.Wrap(err, errGeneratePasswordFailed)
		}
	}
	request.Password = pw

	requestTags := make([]aliredis.CreateInstanceTag, len(p.Tag))
	for _, tag := range p.Tag {
		requestTags = append(requestTags, aliredis.CreateInstanceTag{Key: tag.Key, Value: tag.Value})
	}
	request.Tag = &requestTags

	resp, err := c.redisCli.CreateInstance(request)
	if err != nil {
		return nil, nil, CleanError(err)
	}

	conn := &RedisConnection{
		Username:         resp.InstanceId, // By default user name will be the instance Id
		Password:         pw,
		ConnectionDomain: resp.ConnectionDomain,
		Port:             strconv.Itoa(resp.Port),
	}

	return resp, conn, nil
}

func (c *client) DeleteInstance(id string) error {
	request := aliredis.CreateDeleteInstanceRequest()
	request.Scheme = HTTPSScheme

	request.InstanceId = id

	_, err := c.redisCli.DeleteInstance(request)
	return CleanError(err)
}

// GenerateObservation is used to produce v1alpha1.RedisInstanceObservation from
// redis.DBInstance.
func GenerateObservation(attr *aliredis.DBInstanceAttribute) v1alpha1.RedisInstanceObservation {
	return v1alpha1.RedisInstanceObservation{
		InstanceStatus:    attr.InstanceStatus,
		ConnectionDomain:  attr.ConnectionDomain,
		Port:              strconv.FormatInt(attr.Port, 10),
		VpcAuthMode:       attr.VpcAuthMode,
		ArchitectureType:  attr.ArchitectureType,
		Config:            attr.Config,
		SecurityIPList:    attr.SecurityIPList,
		InstanceClass:     attr.InstanceClass,
		RealInstanceClass: attr.RealInstanceClass,
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
		return false || errors.Is(err, errors.New(errInstanceNotFound))
	}

	return srverr.ErrorCode() == errInstanceNotFoundCode
}

func (c *client) UpdateInstance(id string, req *ModifyRedisInstanceRequest) error {
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
	_, err := c.redisCli.ModifyInstanceSpec(request)
	return CleanError(err)
}

// Check if the whitelist IPs (IPv4) in Redis parameters are different than what are actually configured
// Return true if there are differences
func SecurityIpsNeedUpdate(attr *aliredis.DBInstanceAttribute, p *v1alpha1.RedisInstanceParameters) *aliredis.ModifySecurityIpsRequest {
	if p.SecurityIps == "" {
		return nil
	}

	ips := strings.Split(attr.SecurityIPList, ",")
	needUpdate := false
	for _, ip := range ips {
		if !strings.Contains(p.SecurityIps, ip) {
			needUpdate = true
			break
		}
	}

	if needUpdate {
		req := aliredis.CreateModifySecurityIpsRequest()
		req.SecurityIps = p.SecurityIps
		req.ModifyMode = DefaultIPModifyMode
		return req
	}

	return nil
}

// Send Modify IP request
func (c *client) ModifySecurityIps(id string, req *aliredis.ModifySecurityIpsRequest) error {
	req.InstanceId = id
	_, err := c.redisCli.ModifySecurityIps(req)
	return CleanError(err)
}

// Check if the Redis instance spec needs to be updated
// Return nil if there is no difference, else return the modify instance spec request
func SpecsNeedUpdate(attr *aliredis.DBInstanceAttribute, p *v1alpha1.RedisInstanceParameters) *aliredis.ModifyInstanceSpecRequest {
	return calculateSpecDiff(attr, p)
}

// Calculate the difference between the resource spec and what's actually configured
// Only support certain updates on specs, and we update the specs one by one to avoid conflicts
// Note that Major and Minor versions, adding or deleting shards should be set by separate calls
func calculateSpecDiff(attr *aliredis.DBInstanceAttribute, p *v1alpha1.RedisInstanceParameters) *aliredis.ModifyInstanceSpecRequest {
	diff := aliredis.CreateModifyInstanceSpecRequest()

	// By default, effective time is "Immediately"
	if p.EffectiveTime != "" {
		diff.EffectiveTime = p.EffectiveTime
	}

	// The real instance class should be in the RealInstanceClass parameter if it's a sharded instance
	if attr.ArchitectureType == RedisArchTypeCluster {
		if attr.RealInstanceClass != p.InstanceClass {
			diff.InstanceClass = p.InstanceClass
			return diff
		}
	} else {
		if attr.InstanceClass != p.InstanceClass {
			diff.InstanceClass = p.InstanceClass
			return diff
		}
	}

	if p.ReadOnlyCount != nil && attr.ReadOnlyCount != *p.ReadOnlyCount {
		diff.ReadOnlyCount = requests.NewInteger(*p.ReadOnlyCount)
		return diff
	}

	return nil
}

// Send Modify Spec request
func (c *client) ModifyInstanceSpecs(id string, req *aliredis.ModifyInstanceSpecRequest) error {
	req.InstanceId = id
	_, err := c.redisCli.ModifyInstanceSpec(req)
	return CleanError(err)
}

// Check if the Parameter configuration needs to be updated
// Return empty string if there is no difference, else return the modify instance param request
func ParamsNeedUpdate(attr *aliredis.DBInstanceAttribute, p *v1alpha1.RedisInstanceParameters) (*aliredis.ModifyInstanceConfigRequest, error) {
	return calculateParamDiff(attr, p)
}

// Calculate the difference between the resource parameter configuration and what's actually configured
func calculateParamDiff(attr *aliredis.DBInstanceAttribute, p *v1alpha1.RedisInstanceParameters) (*aliredis.ModifyInstanceConfigRequest, error) {
	if p.ParameterConfig == "" {
		return nil, nil
	}

	var actual, wanted, diff map[string]interface{}
	var err error
	err = json.Unmarshal([]byte(attr.Config), &actual)
	if err != nil {
		return nil, errors.Wrap(err, errFailedToUnmarshalRealConfig)
	}

	err = json.Unmarshal([]byte(p.ParameterConfig), &wanted)
	if err != nil {
		return nil, errors.Wrap(err, errFailedToUnmarshalWantedConfig)
	}

	needUpdate := false
	diff = make(map[string]interface{})
	for key, val := range wanted {
		actualVal, ok := actual[key]
		// Update if the param is different, Or Create if the param does not exist
		if (ok && actualVal != val) || !ok {
			diff[key] = val
			needUpdate = true
		}
	}

	configStr, err := json.Marshal(diff)
	if err != nil {
		return nil, errors.Wrap(err, errFailedToMarshalConfig)
	}

	if needUpdate {
		req := aliredis.CreateModifyInstanceConfigRequest()
		req.Config = string(configStr)
		return req, nil
	}

	return nil, nil
}

// Send Modify Param request
func (c *client) ModifyInstanceParams(id string, req *aliredis.ModifyInstanceConfigRequest) error {
	req.InstanceId = id
	_, err := c.redisCli.ModifyInstanceConfig(req)
	return CleanError(err)
}

// 2024-05-14: Henry
// Try to remove requestID from AliCloud SDK errors
// Returning error with requestID will cause Crossplane reconciler to treat the errors
// as a sequence of unique errors and insert all errors into the retry queue, which
// immediately boomed the AliCloud rate limit.
// See more details of a similar issue in AWS controller:
// https://github.com/crossplane-contrib/provider-aws/issues/69
func CleanError(err error) error {
	if err == nil {
		return err
	}

	if aliCloudErr, ok := err.(*sdkerrors.ServerError); ok {
		cleanedErr := CleanedServerError{
			HttpStatus: aliCloudErr.HttpStatus(),
			HostId:     aliCloudErr.HostId(),
			Code:       aliCloudErr.ErrorCode(),
			Message:    aliCloudErr.Message(),
			Comment:    aliCloudErr.Comment(),
		}
		strData, err := json.Marshal(cleanedErr)
		if err != nil {
			return errors.Wrap(err, "Failed to marshal cleaned error from AliCloud SDK Error.")
		}
		return sdkerrors.NewServerError(aliCloudErr.HttpStatus(), string(strData), aliCloudErr.Comment())
	}

	return err
}
