/*
Copyright 2021 Pavan.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ApplicationSpec defines the desired state of Application
type ApplicationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Application. Edit application_types.go to remove/update
	ApplicationName string        `json:"applicationName,omitempty"`
	Cloud           CloudSelector `json:"cloud,omitempty"`
}

type CloudSelector struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Product. Edit product_types.go to remove/update
	Aws AwsSpec `json:"aws,omitempty"`
}

type AwsSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Product. Edit product_types.go to remove/update
	S3 string `json:"s3,omitempty"`
}

// ApplicationStatus defines the observed state of Application
type ApplicationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	DeploymentCreationTime metav1.Time `json:"deploymentCreationTime,omitempty"`
	ServiceCreationTime    metav1.Time `json:"serviceCreationTime,omitempty"`
	HpaCreationTime        metav1.Time `json:"hpaCreationTime,omitempty"`
	IngressCreationTime    metav1.Time `json:"ingressCreationTime,omitempty"`
	S3BucketName           string      `json:"s3BucketName,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:printcolumn:name="S3BucketName",type="string",JSONPath=".status.s3BucketName",description="Name of the S3 bucket"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//Priority set to 1 so that it should be show when -o wide
//+kubebuilder:printcolumn:name="CloudProvider",type="string",priority=1,JSONPath=".spec.cloud"
//+kubebuilder:subresource:status

// Application is the Schema for the applications API
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}
