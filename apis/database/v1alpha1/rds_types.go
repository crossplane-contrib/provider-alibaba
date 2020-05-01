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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// SQL database engines.
const (
	MysqlEngine      = "MySQL"
	PostgresqlEngine = "PostgreSQL"
)

// +kubebuilder:object:root=true

// RDSInstanceClassList contains a list of cloud memorystore resource classes.
type RDSInstanceClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RDSInstanceClass `json:"items"`
}

// +kubebuilder:object:root=true

// An RDSInstanceClass is a resource class. It defines the desired spec of
// resource claims that use it to dynamically provision a managed resource.
// +kubebuilder:printcolumn:name="PROVIDER-REF",type="string",JSONPath=".specTemplate.providerRef.name"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".specTemplate.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type RDSInstanceClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// SpecTemplate is a template for the spec of a dynamically provisioned
	// RDSInstance.
	SpecTemplate RDSInstanceClassSpecTemplate `json:"specTemplate"`
}

// +kubebuilder:object:root=true

// RDSInstanceList contains a list of RDSInstance
type RDSInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RDSInstance `json:"items"`
}

// +kubebuilder:object:root=true

// An RDSInstance is a managed resource that represents an RDS instance.
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.dbInstanceStatus"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="ENGINE",type="string",JSONPath=".spec.forProvider.engine"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".spec.forProvider.engineVersion"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type RDSInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RDSInstanceSpec   `json:"spec"`
	Status RDSInstanceStatus `json:"status,omitempty"`
}

// An RDSInstanceClassSpecTemplate is a template for the spec of a dynamically
// provisioned RDSInstance.
type RDSInstanceClassSpecTemplate struct {
	runtimev1alpha1.ClassSpecTemplate `json:",inline"`
	ForProvider                       RDSInstanceParameters `json:"forProvider"`
}

// An RDSInstanceSpec defines the desired state of an RDSInstance.
type RDSInstanceSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  RDSInstanceParameters `json:"forProvider"`
}

// An RDSInstanceStatus represents the observed state of an RDSInstance.
type RDSInstanceStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     RDSInstanceObservation `json:"atProvider,omitempty"`
}

// RDSInstanceParameters define the desired state of an RDS instance.
type RDSInstanceParameters struct {
	// Engine is the name of the database engine to be used for this instance.
	// Engine is a required field.
	// +immutable
	Engine string `json:"engine"`

	// EngineVersion indicates the database engine version.
	// MySQL：5.5/5.6/5.7/8.0
	// PostgreSQL：9.4/10.0/11.0/12.0
	EngineVersion string `json:"engineVersion"`

	// DBInstanceClass is the machine class of the instance, e.g. "rds.pg.s1.small"
	DBInstanceClass string `json:"dbInstanceClass"`

	// DBInstanceStorageInGB indicates the size of the storage in GB.
	// Increments by 5GB.
	// For "rds.pg.s1.small", the range is 20-600 (GB).
	// See https://help.aliyun.com/document_detail/26312.html
	DBInstanceStorageInGB int `json:"dbInstanceStorageInGB"`

	// SecurityIPList is the IP whitelist for RDS instances
	SecurityIPList string `json:"securityIPList"`

	// MasterUsername is the name for the master user.
	// MySQL
	// Constraints:
	//    * Required for MySQL.
	//    * Must be 1 to 16 letters or numbers.
	//    * First character must be a letter.
	//    * Cannot be a reserved word for the chosen database engine.
	// PostgreSQL
	// Constraints:
	//    * Required for PostgreSQL.
	//    * Must be 1 to 63 letters or numbers.
	//    * First character must be a letter.
	//    * Cannot be a reserved word for the chosen database engine.
	// +immutable
	// +optional
	MasterUsername string `json:"masterUsername"`
}

// RDS instance states.
const (
	// The instance is healthy and available
	RDSInstanceStateRunning = "Running"
	// The instance is being created. The instance is inaccessible while it is being created.
	RDSInstanceStateCreating = "Creating"
	// The instance is being deleted.
	RDSInstanceStateDeleting = "Deleting"
)

// RDSInstanceObservation is the representation of the current state that is observed.
type RDSInstanceObservation struct {
	// DBInstanceStatus specifies the current state of this database.
	DBInstanceStatus string `json:"dbInstanceStatus,omitempty"`

	// DBInstanceID specifies the DB instance ID.
	DBInstanceID string `json:"dbInstanceID"`

	// AccountReady specifies whether the initial user account (username + password) is ready
	AccountReady bool `json:"accountReady"`
}

// Endpoint is the database endpoint
type Endpoint struct {
	// Address specifies the DNS address of the DB instance.
	Address string `json:"address,omitempty"`

	// Port specifies the port that the database engine is listening on.
	Port string `json:"port,omitempty"`
}
