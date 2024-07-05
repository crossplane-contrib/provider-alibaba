/*


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

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// RedisInstance is the Schema for the redisinstances API
// An RedisInstance is a managed resource that represents an Redis instance.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.InstanceStatus"
// +kubebuilder:printcolumn:name="INSTANCE_TYPE",type="string",JSONPath=".spec.forProvider.instanceType"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".spec.forProvider.engineVersion"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,alibaba}
type RedisInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisInstanceSpec   `json:"spec,omitempty"`
	Status RedisInstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RedisInstanceList contains a list of RedisInstance
type RedisInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisInstance `json:"items"`
}

// RedisInstanceSpec defines the desired state of RedisInstance
type RedisInstanceSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       RedisInstanceParameters `json:"forProvider"`
}

// Redis instance states.
const (
	// The instance is healthy and available
	RedisInstanceStateRunning = "Normal"
	// The instance is being created. The instance is inaccessible while it is being created.
	RedisInstanceStateCreating = "Creating"
	// The instance is being deleted.
	RedisInstanceStateDeleting = "Flushing"
)

// RedisInstanceStatus defines the observed state of RedisInstance
type RedisInstanceStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          RedisInstanceObservation `json:"atProvider,omitempty"`
}

// A Tag is used to tag the Redis resources in AliCloud.
type Tag struct {
	// Key for the tag.
	Key string `json:"key"`

	// Value of the tag.
	// +optional
	Value string `json:"value,omitempty"`
}

// RedisInstanceParameters define the desired state of an Redis instance.
// Detailed information can be found in https://www.alibabacloud.com/help/en/redis/developer-reference/api-r-kvstore-2015-01-01-createinstance-redis
type RedisInstanceParameters struct {
	// The ID of the region where you want to create the instance.
	RegionID string `json:"regionId,omitempty"`

	// The client token that is used to ensure the idempotence of the request.
	// You can use the client to generate the value, but you must make sure that the token is
	// unique among different requests. The token is case-sensitive.
	// The token can contain only ASCII characters and cannot exceed 64 characters in length.
	// +optional
	Token string `json:"token,omitempty"`

	// The name of the instance. The name must be 2 to 80 characters in length and must start with a letter.
	// It cannot contain spaces or specific special characters. These special characters include @ / : = " < > { [ ] }
	// +optional
	InstanceName string `json:"instanceName,omitempty"`

	// The password that is used to connect to the instance. The password must be 8 to 32 characters
	// in length and must contain at least three of the following character types: uppercase letters,
	// lowercase letters, digits, and specific special characters. These special characters include ! @ # $ % ^ & * ( ) _ + - =
	// +optional
	Password string `json:"password,omitempty"`

	// The storage capacity of the instance. Unit: MB.
	// Note: You must specify at least one of the Capacity and InstanceClass parameters when you call this operation.
	// +optional
	Capacity *int `json:"capacity,omitempty"`

	// The instance type.
	// For example, redis.master.small.default indicates a Community Edition
	// standard master-replica instance that has 1 GB of memory.
	// For more information, see https://www.alibabacloud.com/help/en/redis/product-overview/overview-4.
	// +optional
	InstanceClass string `json:"instanceClass,omitempty"`

	// The primary zone ID of the instance.
	ZoneID string `json:"zoneId,omitempty"`

	// ChargeType is indicates payment type
	// Default value: PrePaid.
	// Valid values:
	//		PrePaid: subscription
	//		PostPaid: pay-as-you-go
	// +optional
	ChargeType string `json:"chargeType,omitempty"`

	// The node type.
	// Valid values:
	//		MASTER_SLAVE: high availability (master-replica)
	//		STAND_ALONE: standalone
	//		double: master-replica
	//		single: standalone
	// Note: To create a cloud-native instance, set this parameter to MASTER_SLAVE or STAND_ALONE.
	// To create a classic instance, set this parameter to double or single.
	// +optional
	NodeType string `json:"nodeType,omitempty"`

	// The network type of the instance.
	// Default value:
	//		VPC
	// Valid values:
	//		VPC
	// +optional
	NetworkType string `json:"networkType,omitempty"`

	// The ID of the virtual private cloud (VPC).
	// +optional
	VpcID string `json:"vpcId,omitempty"`

	// The ID of the vSwitch to which you want the instance to connect.
	// +optional
	VSwitchID string `json:"vSwitchId,omitempty"`

	// The subscription duration.
	// Valid values:
	//		1, 2, 3, 4, 5, 6, 7, 8, 9, 12, 24,36, and 60.
	// Unit:
	//		months.
	// Note: This parameter is available and required only if the ChargeType parameter is set to PrePaid.
	// +optional
	Period string `json:"period,omitempty"`

	// The ID of the promotional event or business information.
	// +optional
	BusinessInfo string `json:"businessInfo,omitempty"`

	// The coupon code.
	// Default value:
	//		default.
	// +optional
	CouponNo string `json:"couponNo,omitempty"`

	// The ID of the original instance. If you want to create an instance based on a
	// backup file of a specified instance, you can specify this parameter and use
	// the BackupId or RestoreTime parameter to specify the backup file.
	// +optional
	SrcDBInstanceId string `json:"srcDBInstanceId,omitempty"`

	// The ID of the backup file of the original instance. If you want to create
	// an instance based on a backup file of a specified instance, you can specify
	// this parameter after you specify the SrcDBInstanceId parameter.
	// Then, the system creates an instance based on the backup file that is specified by this parameter.
	// Note:
	// 		After you specify the SrcDBInstanceId parameter,
	//		you must use the BackupId or RestoreTime parameter to specify the backup file.
	// +optional
	BackupId string `json:"backupId,omitempty"`

	// The category of the instance.
	//	Default value:
	//		Redis
	//	Valid values:
	//		Redis
	//		Memcache
	// +optional
	InstanceType string `json:"instanceType,omitempty"`

	// The database engine version of the instance.
	// Valid values:
	//		4.0
	//		5.0
	//		6.0
	//		7.0
	// +optional
	EngineVersion string `json:"engineVersion,omitempty"`

	// The private IP address of the instance.
	// Note:
	//		The private IP address must be available within the CIDR block
	//		of the vSwitch to which to connect the instance.
	// +optional
	PrivateIpAddress string `json:"privateIpAddress,omitempty"`

	// Specifies whether to use a coupon.
	// Default value:
	//		false
	// Valid values:
	//		true: uses a coupon
	//		false: does not use a coupon
	// +optional
	AutoUseCoupon string `json:"autoUseCoupon,omitempty"`

	// Specifies whether to enable auto-renewal for the instance.
	// Default value:
	//		false
	//	Valid values:
	//		true: enables auto-renewal
	//		false: disables auto-renewal
	// +optional
	AutoRenew string `json:"autoRenew,omitempty"`

	// The subscription duration that is supported by auto-renewal.
	// Unit: months.
	// Valid values:
	// 		1, 2, 3, 6, and 12
	// Note:
	//		This parameter is required only if the AutoRenew parameter is set to true.
	// +optional
	AutoRenewPeriod string `json:"autoRenewPeriod,omitempty"`

	// The ID of the resource group.
	// +optional
	ResourceGroupId string `json:"resourceGroupId,omitempty"`

	// The point in time at which the specified original instance is backed up.
	// The point in time must be within the retention period of backup files of the
	// original instance. If you want to create an instance based on a backup file of
	// a specified instance, you can set this parameter to specify a point in time after
	// you set the SrcDBInstanceId parameter. Then, the system creates an instance based
	// on the backup file that was created at the specified point in time for the
	// original instance. Specify the time in the ISO 8601 standard in the
	// yyyy-MM-ddTHH:mm:ssZ format. The time must be in UTC.
	// Note:
	// 		After you specify the SrcDBInstanceId parameter, you must use the BackupId or
	//		RestoreTime parameter to specify the backup file.
	// +optional
	RestoreTime string `json:"restoreTime,omitempty"`

	// The ID of the dedicated cluster. This parameter is required if you create an
	// instance in a dedicated cluster.
	// +optional
	DedicatedHostGroupId string `json:"dedicatedHostGroupId,omitempty"`

	// The number of data shards.
	// This parameter is available only if you create a cluster instance that uses cloud disks.
	// +optional
	ShardCount *int `json:"shardCount,omitempty"`

	// The number of read-only nodes in the instance. This parameter is available
	// only if you create a read/write splitting instance that uses cloud disks.
	// Valid values:
	//		1 to 5
	// +optional
	ReadOnlyCount *int `json:"readOnlyCount,omitempty"`

	// The ID of the distributed instance. This parameter is available only on the China site (aliyun.com).
	// +optional
	GlobalInstanceId string `json:"globalInstanceId,omitempty"`

	// Specifies whether to use the new instance as the first child instance of the distributed instance.
	// Default value:
	//		false
	// Valid values:
	//		true: uses the new instance as the first child instance
	//		false: does not use the new instance as the first child instance
	// If you want to create an ApsaraDB for Redis Enhanced Edition (Tair) DRAM-based instance that
	// runs Redis 5.0, you must set this parameter to true.
	// This parameter is available only on the China site (aliyun.com).
	// +optional
	GlobalInstance bool `json:"globalInstance,omitempty"`

	// The secondary zone ID of the instance.
	// The master node and replica node of the instance can be deployed in different zones and disaster
	// recovery is implemented across zones.
	// Note:
	//		If you specify this parameter, the master node and replica node of the instance can be
	//		deployed in different zones and disaster recovery is implemented across zones.
	//		The instance can withstand failures in data centers.
	// +optional
	SecondaryZoneID string `json:"secondaryZoneId,omitempty"`

	// Port is indicates the database service port
	// Valid values:
	//		1024 to 65535
	// Default value:
	//		6379
	// +optional
	Port *int `json:"port,omitempty"`

	// The global IP whitelist template for the instance. Multiple IP whitelist templates should
	// be separated by English commas (,) and cannot be duplicated.
	// +optional
	GlobalSecurityGroupIds string `json:"globalSecurityGroupIds,omitempty"`

	// Specifies whether to enable append-only file (AOF) persistence for the instance.
	// Valid values:
	//		yes (default): enables AOF persistence
	//		no: disables AOF persistence
	// Description:
	//		This parameter is applicable to classic instances, and is unavailable for cloud-native instances.
	// +optional
	Appendonly string `json:"appendonly,omitempty"`

	// The operation that you want to perform. Set the value to AllocateInstancePublicConnection.
	// +optional
	ConnectionStringPrefix string `json:"connectionStringPrefix,omitempty"`

	// The parameter template ID, which must be globally unique.
	// +optional
	ParamGroupId string `json:"paramGroupId,omitempty"`

	// The tags of the instance.
	// +optional
	Tag []Tag `json:"tag,omitempty"`

	// The backup set ID.
	// +optional
	ClusterBackupId string `json:"clusterBackupId,omitempty"`

	// A list of IP addresses should be whitelisted.
	// Separate multiple IP addresses with a comma (,).
	// +optional
	SecurityIps string `json:"securityIps,omitempty"`

	// Effective time for Instance Updates
	// Valid values:
	//		Immediately (default): The configurations are immediately changed
	//		MaintainTime: The configurations are changed within the maintenance window.
	// +optional
	EffectiveTime string `json:"effectiveTime,omitempty"`

	// The instance parameter settings that you want to modify.
	// Specify this parameter in the JSON format.
	// The new value of a parameter overwrites the original value.
	// Supported parameters can be found here: https://www.alibabacloud.com/help/en/redis/user-guide/supported-parameters
	// +optional
	ParameterConfig string `json:"parameterConfig,omitempty"`
}

// RedisInstanceObservation is the representation of the current state that is observed.
type RedisInstanceObservation struct {
	// InstanceStatus specifies the current state of this database.
	InstanceStatus string `json:"InstanceStatus,omitempty"`

	// ConnectionDomain contains domain name used to connect with the Redis instance
	ConnectionDomain string `json:"connectionDomain,omitempty"`

	// Port contains the port number used to connect with the Redis instance
	Port string `json:"port,omitempty"`

	// Indicates whether password authentication is enabled. Valid values:
	//		Open: Password authentication is enabled.
	//		Close: Password authentication is disabled and password-free access is enabled.
	VpcAuthMode string `json:"vpcAuthMode,omitempty"`

	// The architecture of the instance. Valid values:
	//		cluster: cluster architecture
	//		standard: standard architecture
	//		rwsplit: read/write splitting architecture
	ArchitectureType string `json:"architectureType,omitempty"`

	// The parameter configurations of the instance in the JSON format.
	// For more information, see Parameter descriptions: https://www.alibabacloud.com/help/en/redis/user-guide/supported-parameters
	Config string `json:"config,omitempty"`

	// The IP addresses in the whitelist.
	SecurityIPList string `json:"securityIPList,omitempty"`

	// The instance type
	InstanceClass string `json:"instanceClass,omitempty"`

	// If the instance is a cluster instance that uses cloud disks,
	// this parameter indicates the actual instance type of individual shards in the instance.
	// The InstanceClass parameter indicates the virtual instance type.
	RealInstanceClass string `json:"realInstanceClass,omitempty"`
}
