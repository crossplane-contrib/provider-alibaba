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
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.dbInstanceStatus"
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

// RedisInstanceParameters define the desired state of an Redis instance.
type RedisInstanceParameters struct {
	// Engine is the name of the database engine to be used for this instance.
	// Engine is a required field.
	// +immutable
	// +kubebuilder:validation:Enum=Redis
	InstanceType string `json:"instanceType"`
	// EngineVersion indicates the database engine version.
	// Redis：4.0/5.0
	// +kubebuilder:validation:Enum="4.0";"5.0"
	EngineVersion string `json:"engineVersion"`

	// InstanceClass is the machine class of the instance, e.g. "redis.logic.sharding.2g.8db.0rodb.8proxy.default"
	InstanceClass string `json:"instanceClass"`

	// InstancePort is indicates the database service port
	// +optional
	InstancePort int `json:"port"`

	// PubliclyAccessible is Public network of service exposure
	PubliclyAccessible bool `json:"publiclyAccessible"`

	// ChargeType is indicates payment type
	// ChargeType：PrePaid/PostPaid
	// +optional
	// +kubebuilder:default="PostPaid"
	ChargeType string `json:"chargeType"`

	// MasterUsername is the name for the master user.
	// Constraints:
	//    * Required for Redis.
	//    * Must be 1 to 16 letters or numbers.
	//    * First character must be a letter.
	//    * Cannot be a reserved word for the chosen database engine.
	// +immutable
	// +optional
	MasterUsername string `json:"masterUsername"`

	// NetworkType is indicates service network type
	// NetworkType：CLASSIC/VPC
	// +optional
	// +kubebuilder:default="CLASSIC"
	NetworkType string `json:"networkType"`

	// VpcId is indicates VPC ID
	// +optional
	VpcID string `json:"vpcId"`

	// VSwitchId is indicates VSwitch ID
	// +optional
	VSwitchID string `json:"vSwitchId"`
}

// RedisInstanceObservation is the representation of the current state that is observed.
type RedisInstanceObservation struct {
	// DBInstanceStatus specifies the current state of this database.
	DBInstanceStatus string `json:"dbInstanceStatus,omitempty"`

	// DBInstanceID specifies the Redis instance ID.
	DBInstanceID string `json:"dbInstanceID"`

	// AccountReady specifies whether the initial user account (username + password) is ready
	AccountReady bool `json:"accountReady"`

	// ConnectionReady specifies whether the network connect is ready
	ConnectionReady bool `json:"connectionReady"`
}

// Endpoint is the redis endpoint
type Endpoint struct {
	// Address specifies the DNS address of the Redis instance.
	Address string `json:"address,omitempty"`

	// Port specifies the port that the database engine is listening on.
	Port string `json:"port,omitempty"`
}
