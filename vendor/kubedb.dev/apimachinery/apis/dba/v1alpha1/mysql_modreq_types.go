/*
Copyright The KubeDB Authors.

Licensed under the Apache License, ModificationRequest 2.0 (the "License");
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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceCodeMySQLModificationRequest     = "mymodreq"
	ResourceKindMySQLModificationRequest     = "MySQLModificationRequest"
	ResourceSingularMySQLModificationRequest = "mysqlmodificationrequest"
	ResourcePluralMySQLModificationRequest   = "mysqlmodificationrequests"
)

// MySQLModificationRequest defines a MySQL Modification Request object.

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=mysqlmodificationrequests,singular=mysqlmodificationrequest,shortName=mymodreq,categories={datastore,kubedb,appscode}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type MySQLModificationRequest struct {
	metav1.TypeMeta   `json:",inline,omitempty"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec              MySQLModificationRequestSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            MySQLModificationRequestStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// MySQLModificationRequestSpec is the spec for MySQLModificationRequest version
type MySQLModificationRequestSpec struct {
	// Specifies the database reference
	DatabaseRef v1.LocalObjectReference `json:"databaseRef" protobuf:"bytes,1,opt,name=databaseRef"`
	// Specifies the modification request type; ScaleUp, ScaleDown, Upgrade etc.
	Type ModificationRequestType `json:"type" protobuf:"bytes,2,opt,name=type"`
	// Specifies the field information that needed to be updated
	Update *UpdateSpec `json:"update,omitempty" protobuf:"bytes,3,opt,name=update"`
	//Specifies the scaling info of database
	Scale *ScaleSpec `json:"scale,omitempty" protobuf:"bytes,4,opt,name=scale"`
	//Specifies the current suffix of the StatefulSet
	CurStsSuffix int32 `json:"curStsSuffix,omitempty" protobuf:"varint,5,opt,name=curStsSuffix"`
}

type UpdateSpec struct {
	// Specifies the ElasticsearchVersion object name
	TargetVersion string `json:"targetVersion,omitempty" protobuf:"bytes,1,opt,name=targetVersion"`
}

// ScaleSpec contains the scaling information of the MySQL
type ScaleSpec struct {
	// Horizontal specifies the horizontal scaling.
	Horizontal *HorizontalScale `json:"horizontal,omitempty" protobuf:"bytes,1,opt,name=horizontal"`
	// Vertical specifies the vertical scaling.
	Vertical *VerticalScale `json:"vertical,omitempty" protobuf:"bytes,2,opt,name=vertical"`
	// specifies the weight of the current member/Node
	MemberWeight int32 `json:"memberWeight,omitempty" protobuf:"varint,3,opt,name=memberWeight"`
}

type HorizontalScale struct {
	// Number of nodes/members of the group
	Member *int32 `json:"member,omitempty" protobuf:"varint,1,opt,name=member"`
}

type VerticalScale struct {
	// Containers represents the containers specification for scaling the requested resources.
	Containers []v1.Container `json:"containers,omitempty" protobuf:"bytes,1,opt,name=containers"`
}

// MySQLModificationRequestStatus is the status for MySQLModificationRequest object
type MySQLModificationRequestStatus struct {
	// Specifies the current phase of the modification request
	// +optional
	Phase ModificationRequestPhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=ModificationRequestPhase"`
	// Specifies the reason behind the current status, if any
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,2,opt,name=reason"`
	// Conditions applied to the request, such as approval or denial.
	// +optional
	Conditions []MySQLModificationRequestCondition `json:"conditions,omitempty" protobuf:"bytes,3,rep,name=conditions"`
	// observedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,4,opt,name=observedGeneration"`
}

type MySQLModificationRequestCondition struct {
	// request approval state, currently Approved or Denied.
	Type RequestConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=RequestConditionType"`

	// brief reason for the request state
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,2,opt,name=reason"`

	// human readable message with details about the request state
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,3,opt,name=message"`

	// timestamp for the last update to this condition
	// +optional
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty" protobuf:"bytes,4,opt,name=lastUpdateTime"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MySQLModificationRequestList is a list of MySQLModificationRequests
type MySQLModificationRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Items is a list of MySQLModificationRequest CRD objects
	Items []MySQLModificationRequest `json:"items,omitempty" protobuf:"bytes,2,rep,name=items"`
}
